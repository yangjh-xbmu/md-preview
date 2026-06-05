# md-preview

Quickly preview Markdown files as rendered HTML with GitHub-style styling.

`md-preview` starts a desktop app by default to show rendered Markdown in a local window.
Optionally, you can use `--browser` to run the original local HTML preview server.

## Prerequisites

Go 1.25.7 or newer.

## Install

```bash
go install github.com/yangjh-xbmu/md-preview@latest
```

From a local checkout:

```bash
go install .
```

## Run

```bash
md-preview README.md
```

During development:

```bash
go run . README.md
```

By default, the tool opens a local desktop preview window and refreshes content when the Markdown file changes.

## Options

```text
Usage: md-preview [--browser] [--host 127.0.0.1] [--port 17776] [--no-open] [--watch=false] <file.md>
```

| Option | Default | Description |
| --- | --- | --- |
| `--browser` | `false` | Start the browser-based preview mode. |
| `--host` | `127.0.0.1` | HTTP bind host for browser mode. Use another value only when you intentionally want a wider bind address. |
| `--port` | `17776` | HTTP bind port. Use `0` to let the OS choose an available port. |
| `--no-open` | `false` | Print the preview URL without opening the browser. |
| `--watch` | `true` | Poll the file modification status from the page and reload after changes. |

## Rendering

Desktop mode uses `fyne.io/fyne/v2` for a native window and live refreshes rendered markdown.
Browser mode renders HTML with `github.com/yuin/goldmark` and sanitizes generated HTML with `github.com/microcosm-cc/bluemonday`.

Desktop output is plain Markdown rendering, while browser output shows full GitHub-style HTML.

## Common Problems

`file does not exist`: check the path passed to `md-preview`.

`expected a Markdown file, got directory`: pass a file path instead of a folder.

`unsupported file extension`: use `.md` or `.markdown`.

`port 17776 on host 127.0.0.1 is not available`: choose another port with `--port`, or stop the process already using that port.
`Could not open browser automatically`: usually means default browser launcher is unavailable.

## License

MIT
