# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

A single-purpose SSH server that renders a navigable TUI personal site to
anyone who connects. Built with Charm Wish (SSH framework) + Bubble Tea
(TUI) + Lipgloss (styling). No real shell is exposed — the SSH session runs
one Bubble Tea program and disconnects when the visitor quits.

Visitor flow: connect → splash screen ("i12k / Press Enter to start") →
press Enter → tabbed app (Start, CV, Projects, Contact). CV and Projects
use a 30/70 list/detail layout; ↑/↓ navigate, the detail pane updates
live.

## Commands

```sh
# Local Go (Go 1.23+ required).
make run        # go run ./cmd/i12k — listens on :2222, host key in .ssh/
make build      # builds bin/i12k
make tidy       # go mod tidy
make vet        # go vet ./...
make test       # go test ./...
make ssh        # ssh -p 2222 localhost (with relaxed host-key checks)

# Docker / compose (preferred — Go does not need to be installed locally).
make docker     # docker build -t i12k .
make up         # docker compose up -d --build
make down
make logs       # follow container logs
```

To smoke-test inside a fresh Go toolchain without installing Go locally:

```sh
docker run --rm -v "$PWD":/src -w /src golang:1.23-alpine \
  sh -c "go build -o /tmp/i12k ./cmd/i12k && echo BUILD_OK"
```

Run a single test by package or name:

```sh
go test ./internal/tui -run TestName
```

## Architecture

Three layers, each in its own package:

- **`cmd/i12k`** — entrypoint. Wires Wish middleware (`bubbletea`,
  `activeterm`, `logging`), loads content, handles graceful shutdown on
  SIGINT/SIGTERM. Anonymous SSH access works because **no auth handlers are
  registered**: gliderlabs/ssh skips authentication when both
  `PasswordHandler` and `PublicKeyHandler` are nil. Adding any auth handler
  will silently break anonymous access.

- **`internal/content`** — loads `content.toml` into a `Profile` struct
  (name, tagline, about, skills, projects, contact). Content is loaded once
  at startup; restart the process after editing `content.toml`.

- **`internal/tui`** — the Bubble Tea program. Split by concern:
  - `model.go` — root Model. Owns the mode (splash/app), the active tab,
    per-list selection indices (`cvIdx`, `projectsIdx`), and terminal size.
    All key handling lives here. Sub-views are pure render functions and
    do **not** hold their own state, which keeps Update small.
  - `view.go` — render functions. `View()` dispatches to `splashView` or
    `appView`; `appView` composes header + body + footer; `bodyView`
    dispatches to the per-tab renderer. `listDetailView` is the generic
    30/70 list+detail layout reused by CV and Projects.
  - `styles.go` — Lipgloss styles and the color palette in one place.
    Edit colors here; do not redefine styles inline in `view.go`.

Visitor flow: SSH connect → Wish accepts (no auth) → `activeterm` middleware
rejects clients without a PTY → `bubbletea` middleware starts a fresh Model
per session → splash → Enter → tabbed app → visitor presses q/esc to
disconnect.

### Adding a new tab

1. Add a `tab` constant in `model.go` and append it to `allTabs`.
2. Add a case to `tab.String()`.
3. Add a digit shortcut in `updateApp`.
4. Add a case in `bodyView` that returns the rendered view.
5. If it's a list-based tab, extend `listLen` and `selectionPtr`, and add
   an index field on `Model`.

## Content

All personal copy lives in `content.toml` at the repo root. The Docker image
bakes a copy in at `/app/content.toml`, and compose bind-mounts the local
file over it for hot edits. **Do not hard-code bio text in Go.**

## Host key

The server needs a stable SSH host key so visitors don't get MITM warnings on
every restart. Default path is `.ssh/id_ed25519` locally, `/data/host_key` in
the container (compose mounts a named volume at `/data`). Wish generates one
automatically on first start if missing.

Environment overrides: `I12K_HOST`, `I12K_PORT`, `I12K_HOST_KEY`,
`I12K_CONTENT`.

## Gotchas

- The Dockerfile has a dedicated `data` stage that creates `/data` with UID
  `65532` (distroless nonroot). Without it, the binary fails at startup with
  `mkdir /data: permission denied` because nonroot can't create the
  directory in a distroless rootfs.
- Testing SSH from macOS: `timeout` is not installed. Use background +
  `sleep` + `kill` to bound the connection, or install `coreutils` for
  `gtimeout`.
- The TUI uses `tea.WithAltScreen()`, so logs piped from `ssh` look like raw
  cursor-positioning escape sequences. That's expected; in a real terminal
  the UI renders cleanly.
- The TUI falls back to "resize your terminal" below 40×12. Pipes,
  non-interactive ssh, and PTY-less expect spawns trigger this fallback
  because the reported window size is 0×0. To smoke-test interactively,
  use `expect` with `set stty_init "rows 40 cols 120"` before `spawn ssh`,
  or just connect normally from a real terminal.
