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

Press <kbd>Enter</kbd> on the splash, <kbd>Tab</kbd>/<kbd>Shift+Tab</kbd> or the
digit keys to switch tabs, <kbd>↑</kbd>/<kbd>↓</kbd> to navigate lists,
<kbd>q</kbd> to quit.

## Configuration

Everything user-facing is in `content.toml`. The shipped
[`content.example.toml`](./content.example.toml) is heavily commented and covers
every supported block:

- `name`, `tagline` — required identity.
- `[splash]` — entry-screen title and CTA.
- `[about]` — bio paragraphs on the Start tab.
- `[theme]` — six optional hex colors.
- `[[tabs]]` — which tabs appear and in what order.
- `[[cv]]` — one block per role.
- `[[projects]]` — one block per project.
- `[[contact]]` — one block per link; add as many as you like.

Empty sections automatically hide their tab. So a minimal `content.toml` can be
as short as:

```toml
name    = "Jane Doe"
tagline = "Software engineer."

[about]
lines = ["Available for consulting."]
```

…and you'll get a working site with just the Start tab.

### Environment variables (all optional)

| Variable          | Default                 | Purpose                                  |
|-------------------|-------------------------|------------------------------------------|
| `SSHCV_HOST`      | `0.0.0.0`               | bind address                             |
| `SSHCV_PORT`      | `2222`                  | bind port                                |
| `SSHCV_HOST_KEY`  | `/data/host_key`        | ed25519 host key (generated on first run)|
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

```sh
make ci      # fmt + vet + test + build
make up      # docker compose up -d --build
make ssh     # ssh -p 2222 localhost
```

Requires Go 1.23+ for local builds, or just Docker for `make up`.

## License

MIT. See [LICENSE](./LICENSE).
