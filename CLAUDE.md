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
contents, and theme all come from `content.toml`. Each tab is a `[[sections]]`
block whose `type` selects a pluggable renderer (`text`, `list`, `links`, …).
List-style sections use a 30/70 list/detail layout.

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

Four packages:

- **`cmd/ssh-cv`** — entrypoint. Wires Wish middleware (`bubbletea`,
  `activeterm`, `logging`), reads `SSHCV_*` env vars, loads content, sets
  `IdleTimeout`, handles graceful shutdown on SIGINT/SIGTERM. Anonymous SSH
  access works because **no auth handlers are registered**: gliderlabs/ssh
  skips authentication when both `PasswordHandler` and `PublicKeyHandler` are
  nil. Adding any auth handler will silently break anonymous access.

- **`internal/content`** — loads `content.toml` into a `Profile`. The schema
  is intentionally generic:
  - `Splash` (title, CTA) — defaults fall back from `name`.
  - `Theme` (six optional hex colors).
  - `[]Section` — ordered tabs. Each section has `id`, `type`, `label` plus
    optional `lines` (text) or `items` (list/links). Renderers decide which
    fields they consume; unknown fields are ignored.
  - `Profile.Validate()` produces descriptive errors for missing name, empty
    section IDs, duplicate section IDs, unknown section types, and malformed
    hex colors. Type-specific validation is delegated to the
    `SectionValidator` wired in by `internal/tui/sections`.
  - `Profile.VisibleSections()` filters out sections whose backing data is
    empty (judged by the registered renderer).

  Content is loaded once at startup; restart the process after editing
  `content.toml`. `content` does **not** import the TUI — the renderer
  registry hooks itself in via `content.SetSectionValidator` from an
  `init()` in `internal/tui/sections`.

- **`internal/tui/sections`** — pluggable renderer registry. One file per
  section type (`text.go`, `list.go`, `links.go`); each `init()` calls
  `Register`. The `Renderer` interface exposes `Type`, `Validate`,
  `IsEmpty`, `Render`, `FooterHint`, and `HandleKey`. Shared layout
  primitives (`renderDetailBlock`, `place`, `clamp`) live in `detail.go`.
  `styles.go` holds the Lipgloss style set (bound to a per-session renderer
  + `content.Theme`).

- **`internal/tui`** — thin Bubble Tea shell:
  - `model.go` — root Model. `tabs` is `[]content.Section` resolved from
    `Profile.VisibleSections()`. Per-section selection is a
    `map[string]int` keyed by Section.ID. App-level keys (quit, tab
    switching, digit shortcuts) live here; section-local keys go through
    the active renderer's `HandleKey`.
  - `view.go` — splash + app shell (header / body / footer). `bodyView`
    looks up the renderer with `sections.Get(currentSection.Type)` and
    delegates. `footerView` pulls the hint from the renderer, falling back
    to a default.

Visitor flow: SSH connect → Wish accepts (no auth) → `activeterm` middleware
rejects clients without a PTY → `bubbletea` middleware starts a fresh Model
per session → splash → Enter → tabbed app → visitor presses q/esc to
disconnect.

## Adding a new section type

The framework's expressivity comes from `content.toml`, but new renderer
types are easy to plug in:

1. Add a new file in `internal/tui/sections/`, e.g. `table.go`.
2. Implement the `Renderer` interface (`Type`, `Validate`, `IsEmpty`,
   `Render`, `FooterHint`, `HandleKey`).
3. Call `Register(yourRenderer{})` from an `init()` function.
4. If the new type needs fields the existing `Section`/`Item` structs don't
   carry, add them as optional fields in `internal/content/content.go`.
   Keep the additions opt-in; existing renderers must keep working.
5. Add a documented example block in `content.example.toml`.

Nothing in the core TUI needs to change.

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
