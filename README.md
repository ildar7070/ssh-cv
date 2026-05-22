# ssh-cv

> A terminal-first personal site over SSH. Visitors connect, navigate your CV,
> projects, and contact info with the keyboard. No browser, no JavaScript, no
> tracking.

Built with Go, [Charm Wish](https://github.com/charmbracelet/wish),
[Bubble Tea](https://github.com/charmbracelet/bubbletea), and
[Lipgloss](https://github.com/charmbracelet/lipgloss).

## Quickstart

1. Grab the example content file and edit it:

   ```sh
   curl -O https://raw.githubusercontent.com/ildar7070/ssh-cv/main/content.example.toml
   mv content.example.toml content.toml
   $EDITOR content.toml
   ```

2. Start the container:

   ```sh
   docker run -d \
     --name ssh-cv \
     -p 2222:2222 \
     -v "$PWD/content.toml":/app/content.toml:ro \
     -v ssh-cv-data:/data \
     ghcr.io/ildar7070/ssh-cv:latest
   ```

3. Connect:

   ```sh
   ssh -p 2222 localhost
   ```

The image ships no content of its own — mounting your `content.toml` at
`/app/content.toml` is required. Start it without one and the server exits
immediately with a message telling you what's missing.

Press <kbd>Enter</kbd> on the splash, <kbd>Tab</kbd>/<kbd>Shift+Tab</kbd> or the
digit keys to switch tabs, <kbd>↑</kbd>/<kbd>↓</kbd> to navigate lists,
<kbd>t</kbd> to toggle between dark and light, <kbd>q</kbd> to quit.

## Configuration

Everything user-facing is in `content.toml`. The shipped
[`content.example.toml`](./content.example.toml) is heavily commented and covers
every supported block:

- `name` — required identity (`tagline` is optional).
- `[splash]` — entry-screen title and CTA.
- `[theme]` — six optional hex colors. Either set them flat (the dark palette)
  or split them into `[theme.dark]` and `[theme.light]`. Visitors toggle between
  dark and light at runtime with <kbd>t</kbd>; sessions start in dark. If you
  omit `[theme.light]`, a built-in light palette is used.
- `[[sections]]` — the ordered list of tabs. Each block becomes one tab.

### Sections

Each `[[sections]]` block is a tab, rendered in the order listed. Every section
declares an `id` (stable, unique), a `type` (which renderer to use), and an
optional `label` (tab header text — defaults to the capitalised `id`).

Three renderer types ship today:

| `type`  | Layout                                                            | Good for            |
|---------|------------------------------------------------------------------|---------------------|
| `text`  | heading from `name`/`tagline` + free-form `lines`; optional `ascii`| intro / about     |
| `list`  | 30/70 list-detail split; `items` with title/subtitle/meta/bullets| CV, projects, edu   |
| `links` | label/value rows; values are underlined for click-to-open        | contact             |

A `text` section can carry an optional `ascii` field — a TOML multi-line string
of ASCII art. It renders to the right of the text when the terminal is wide
enough, and below the text otherwise:

```toml
[[sections]]
id   = "start"
type = "text"
ascii = """
  /\\_/\\
 ( o.o )
  > ^ <
"""
lines = ["Available for consulting."]
```

```toml
[[sections]]
id    = "experience"
type  = "list"
label = "Experience"

[[sections.items]]
title    = "Senior Software Engineer"
subtitle = "Acme Corp"
meta     = "2022 — Present"
bullets  = ["Led the platform migration.", "Mentored four engineers."]
```

Adding a new renderer is a code change (one file in
[`internal/tui/sections/`](./internal/tui/sections)) — the `content.toml` schema
itself doesn't change.

Empty sections automatically hide their tab. So a minimal `content.toml` can be
as short as:

```toml
name = "Jane Doe"

[[sections]]
id    = "start"
type  = "text"
lines = ["Available for consulting."]
```

…and you'll get a working site with just the Start tab.

### Environment variables (all optional)

| Variable          | Default                 | Purpose                                  |
|-------------------|-------------------------|------------------------------------------|
| `SSHCV_HOST`      | `0.0.0.0`               | bind address                             |
| `SSHCV_PORT`      | `2222`                  | bind port                                |
| `SSHCV_HOST_KEY`  | `/data/keys/host_key`   | ed25519 host key (generated on first run)|
| `SSHCV_CONTENT`   | `/app/content.toml`     | path to content TOML                     |
| `SSHCV_LOG_LEVEL` | `info`                  | `debug`, `info`, `warn`, `error`         |

## Using compose

```yaml
services:
  ssh-cv:
    image: ghcr.io/ildar7070/ssh-cv:latest
    restart: unless-stopped
    ports:
      - "2222:2222"
    volumes:
      - ./content.toml:/app/content.toml:ro
      - ssh-cv-data:/data

volumes:
  ssh-cv-data:
```

## Building from source

Seed your content file once (the example is a fully documented template):

```sh
cp content.example.toml content.toml
$EDITOR content.toml
```

Run it locally with Go (requires Go 1.25+):

```sh
go run ./cmd/ssh-cv
```

Or build and run via Docker Compose:

```sh
docker compose up -d --build       # build the image and start it
ssh -p 2222 localhost              # connect
docker compose down                # stop it
```

Checks and a versioned binary:

```sh
go vet ./... && go test ./...
go build -ldflags="-X main.version=$(git describe --tags --always --dirty)" \
  -o bin/ssh-cv ./cmd/ssh-cv
```

The `-ldflags` injects the version (from `git describe`), so `ssh-cv --version`
reports the running build. Released images pin their base layers by digest for
reproducible, multi-arch (`amd64`/`arm64`) builds.

## License

MIT. See [LICENSE](./LICENSE).
