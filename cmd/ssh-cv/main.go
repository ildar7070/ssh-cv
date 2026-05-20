// Command ssh-cv serves a terminal-style personal site over SSH.
//
//	ssh user@host -p 2222
//
// On connect, visitors land on a splash screen, then navigate a tabbed TUI
// whose tabs, theme, and content are all driven by content.toml.
//
// Configuration env vars (all optional):
//
//	SSHCV_HOST       bind address      (default 0.0.0.0)
//	SSHCV_PORT       bind port         (default 2222)
//	SSHCV_HOST_KEY   ed25519 host key  (default .ssh/id_ed25519, generated on first run)
//	SSHCV_CONTENT    content TOML path (default content.toml)
//	SSHCV_LOG_LEVEL  debug|info|warn|error (default info)
package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
	"github.com/muesli/termenv"

	"github.com/ildar7070/ssh-cv/internal/content"
	"github.com/ildar7070/ssh-cv/internal/tui"
)

// version is set at build time via -ldflags "-X main.version=...".
var version = "dev"

const (
	idleTimeout     = 5 * time.Minute
	shutdownTimeout = 10 * time.Second
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-v", "--version", "version":
			fmt.Println("ssh-cv", version)
			return
		}
	}

	configureLogging()

	host := envDefault("SSHCV_HOST", "0.0.0.0")
	port := envDefault("SSHCV_PORT", "2222")
	hostKey := envDefault("SSHCV_HOST_KEY", ".ssh/id_ed25519")
	contentPath := envDefault("SSHCV_CONTENT", "content.toml")

	profile, err := content.Load(contentPath)
	if err != nil {
		log.Fatal("load content", "err", err)
	}

	// Anonymous access: by registering no auth handlers, the underlying
	// gliderlabs/ssh server skips authentication entirely. See
	// charmbracelet/ssh server.go: "When both PasswordHandler and
	// PublicKeyHandler are nil, no client authentication is performed."
	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
		wish.WithHostKeyPath(hostKey),
		wish.WithIdleTimeout(idleTimeout),
		wish.WithMiddleware(
			// MiddlewareWithColorProfile sets the MINIMUM color profile the
			// renderer will use. The default `Middleware` pins it to Ascii
			// (no colors), which silently strips all Foreground/Background
			// styling — passing TrueColor lets the client's terminal use
			// its native color depth.
			bubbletea.MiddlewareWithColorProfile(handler(profile), termenv.TrueColor),
			activeterm.Middleware(),
			logging.Middleware(),
		),
	)
	if err != nil {
		log.Fatal("create server", "err", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	log.Info("starting SSH server", "addr", s.Addr, "content", contentPath)
	go func() {
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Error("server error", "err", err)
			cancel()
		}
	}()

	<-ctx.Done()
	log.Info("shutting down")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()
	if err := s.Shutdown(shutdownCtx); err != nil {
		log.Warn("shutdown", "err", err)
	}
}

func handler(p *content.Profile) bubbletea.Handler {
	return func(sess ssh.Session) (tea.Model, []tea.ProgramOption) {
		_, _, active := sess.Pty()
		if !active {
			return nil, nil
		}
		// MakeRenderer reads the client's TERM and builds a Lipgloss
		// renderer with the correct color profile for THIS session. The
		// global default renderer would inherit the server process's env
		// (no TERM in distroless → Ascii → all colors stripped).
		r := bubbletea.MakeRenderer(sess)
		return tui.New(p, r), []tea.ProgramOption{tea.WithAltScreen()}
	}
}

func configureLogging() {
	switch os.Getenv("SSHCV_LOG_LEVEL") {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}
}

func envDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
