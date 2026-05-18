# i12k

A terminal-first personal site. Visitors meet me by SSH-ing in:

```
ssh i12k.dev
```

On connect, visitors see a splash screen, press Enter, and land in a tabbed
TUI:

- **Start** — name, tagline, short about.
- **CV** — list of roles on the left, details (company, period, summary,
  bullets) on the right. Navigate with ↑/↓.
- **Projects** — same list/detail layout for selected projects.
- **Contact** — email, GitHub, LinkedIn.

Switch tabs with `Tab` / `Shift+Tab` or the number keys `1`–`4`. Quit with
`q` or `Ctrl+C`.

No real shell is exposed. The session runs a Bubble Tea program through
[Charm Wish](https://github.com/charmbracelet/wish); there are no commands the
visitor can run.

## Run it

```sh
docker compose up -d --build
ssh -p 2222 localhost
```

The compose file persists the SSH host key in a named volume so the server's
identity is stable across restarts, and mounts `content.toml` read-only so
edits don't require an image rebuild.

## Edit your bio

All personal content lives in [`content.toml`](./content.toml) — name, tagline,
about lines, skills, projects, contact. Edit and restart the container:

```sh
docker compose restart
```

## Local development (without Docker)

Requires Go 1.23+.

```sh
make run    # starts the server on :2222, writes a host key to .ssh/
make ssh    # connects from another terminal
```

## Deploy

The Dockerfile produces a distroless image (~15 MB). Any host that can run a
container and expose port 2222 works — Fly.io, Railway, a VPS. Mount a volume
at `/data` to keep the host key stable.
