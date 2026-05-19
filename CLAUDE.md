# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

`ssh-cv` is a reusable framework: a single-purpose SSH server that renders a
navigable TUI personal site to anyone who connects. The repo is meant to be
published as a Docker image (GHCR); end users edit a `content.toml` and run the
container — they do not fork the code.

Built with Charm Wish (SSH framework) + Bubble Tea (TUI) + Lipgloss (styling).
No real shell is exposed — the SSH session runs one Bubble Tea program and
disconnects when the visitor quits.

Visitor flow: connect → splash screen → press Enter → tabbed app whose tabs,
contents, and theme all come from `content.toml`. List-based tabs (CV,
Projects) use a 30/70 list/detail layout.

## Commands

```sh
# Local Go (Go 1.23+ required).
make run        # go run ./cmd/ssh-cv — listens on :2222, host key in .ssh/
make build      # builds bin/ssh-cv
make tidy       # go mod tidy
make vet        # go vet ./...
make test       # go test ./...
make ci         # fmt + vet + test + build
make ssh        # ssh -p 2222 localhost (with relaxed host-key checks)

# Docker / compose (preferred — Go does not need to be installed locally).
make docker     # docker build -t ssh-cv .
make up         # docker compose up -d --build
make down
make logs       # follow container logs
```

To smoke-test inside a fresh Go toolchain without installing Go locally:

```sh
docker run --rm -v "$PWD":/src -w /src golang:1.23-alpine \
  sh -c "go build -o /tmp/ssh-cv ./cmd/ssh-cv && echo BUILD_OK"
```

Run a single test by package or name:

```sh
go test ./internal/tui -run TestName
```

## Architecture

Three packages:

- **`cmd/ssh-cv`** — entrypoint. Wires Wish middleware (`bubbletea`,
  `activeterm`, `logging`), reads `SSHCV_*` env vars, loads content, sets
  `IdleTimeout`, handles graceful shutdown on SIGINT/SIGTERM. Anonymous SSH
  access works because **no auth handlers are registered**: gliderlabs/ssh
  skips authentication when both `PasswordHandler` and `PublicKeyHandler` are
  nil. Adding any auth handler will silently break anonymous access.

- **`internal/content`** — loads `content.toml` into a `Profile` struct. The
  schema is entirely data-driven:
  - `Splash` (title, CTA) — defaults fall back from `name`.
  - `Theme` (six optional hex colors).
  - `[]TabSpec` — which built-in tab renderers appear and in what order.
    Defaults to all four (start, cv, projects, contact).
  - `[]ContactLink` — contact rows. New platforms = one TOML block, no code.
  - `Profile.Validate()` produces descriptive errors for unknown tab IDs,
    duplicate tabs, missing name, and malformed hex colors.
  - `Profile.VisibleTabs()` filters out tabs whose backing section is empty.

  Content is loaded once at startup; restart the process after editing
  `content.toml`.

- **`internal/tui`** — Bubble Tea program. Split by concern:
  - `model.go` — root Model. `tabs` is resolved from `Profile.VisibleTabs()`
    at construction. Per-list selection is a `map[TabID]int`, not fixed
    struct fields. Digit shortcuts `1`–`9` dynamically map to the Nth visible
    tab. `bodyView` dispatches via a switch on `currentTab()`.
  - `view.go` — render functions. `View()` dispatches to splash or app;
    `appView` composes header + body + footer; `listDetailView` is the
    generic 30/70 layout reused by CV and Projects. `renderDetailBlock`
    is a shared helper for both CV and project detail panes. Magic
    numbers live as named constants at the top.
  - `styles.go` — Lipgloss styles bound to a per-session renderer **and** to
    `content.Theme`. Empty theme fields fall back to the built-in palette
    (constants at the top of the file).

Visitor flow: SSH connect → Wish accepts (no auth) → `activeterm` middleware
rejects clients without a PTY → `bubbletea` middleware starts a fresh Model
per session → splash → Enter → tabbed app → visitor presses q/esc to
disconnect.

## Adding a new built-in tab

This is rare — the framework's expressivity comes from `content.toml`, not
from new tabs. But if you must:

1. Add a new `TabID` constant in `internal/content/content.go` and append it
   to `BuiltinTabs`.
2. Add a default label in `defaultTabLabel`.
3. Teach `HasContent` how to tell if its section is non-empty.
4. Add a case in `internal/tui/view.go`'s `bodyView` and a renderer.
5. Update `content.example.toml` with documentation for the new section.

## Content

User content lives in `content.toml` at the repo root for local dev.
`content.example.toml` is the public template (baked into the image as
`/app/content.toml`); operators override it with a bind mount. **Do not
hard-code bio text in Go.**

## Host key

The server needs a stable SSH host key so visitors don't get MITM warnings on
every restart. Default path is `.ssh/id_ed25519` locally, `/data/host_key` in
the container (compose mounts a named volume at `/data`). Wish generates one
automatically on first start if missing.

Environment overrides: `SSHCV_HOST`, `SSHCV_PORT`, `SSHCV_HOST_KEY`,
`SSHCV_CONTENT`, `SSHCV_LOG_LEVEL`.

## Release

Image is published to `ghcr.io/<owner>/ssh-cv` by `.github/workflows/release.yml`
on every `v*` git tag. Multi-arch (linux/amd64, linux/arm64). Tags applied:
`vX.Y.Z`, `X.Y`, `X`, `latest`.

CI (`.github/workflows/ci.yml`) runs vet + test + build on every PR and push
to main.

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
- Apple Terminal does not support OSC 8 hyperlinks even on macOS 26. Contact
  links are rendered as plain underlined URLs so Apple Terminal's built-in
  URL detector can pick them up for Cmd+click.
