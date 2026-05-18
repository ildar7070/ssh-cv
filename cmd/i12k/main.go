// Command i12k serves a terminal-style personal intro over SSH.
//
//	ssh user@host -p 2222
//
// On connect, visitors see a typewriter animation playing through whoami,
// about, skills, projects, and contact. Content lives in content.toml.
package main

import (
	"context"
	"errors"
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

	"github.com/ildar7070/i12k/internal/content"
	"github.com/ildar7070/i12k/internal/tui"
)

func main() {
	host := envDefault("I12K_HOST", "0.0.0.0")
	port := envDefault("I12K_PORT", "2222")
	hostKey := envDefault("I12K_HOST_KEY", ".ssh/id_ed25519")
	contentPath := envDefault("I12K_CONTENT", "content.toml")

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

	log.Info("starting SSH server", "addr", s.Addr)
	go func() {
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Error("server error", "err", err)
			cancel()
		}
	}()

	<-ctx.Done()
	log.Info("shutting down")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	_ = s.Shutdown(shutdownCtx)
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

func envDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

