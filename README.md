# md-preview

Quickly preview Markdown files as rendered HTML with GitHub-style styling.

`md-preview` starts a local HTTP server, renders a Markdown file with GitHub Flavored Markdown support, and opens the preview in your default browser.

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

By default the server binds to `127.0.0.1:17776`, opens the preview page automatically, and refreshes the browser when the Markdown file changes.

## Options

```text
Usage: md-preview [--host 127.0.0.1] [--port 17776] [--no-open] [--watch=false] <file.md>
```

| Option | Default | Description |
| --- | --- | --- |
| `--host` | `127.0.0.1` | HTTP bind host. Use another value only when you intentionally want a wider bind address. |
| `--port` | `17776` | HTTP bind port. Use `0` to let the OS choose an available port. |
| `--no-open` | `false` | Print the preview URL without opening the browser. |
| `--watch` | `true` | Poll the file modification status from the page and reload after changes. |

## Rendering

The renderer uses `github.com/yuin/goldmark` with GitHub Flavored Markdown extensions. The generated HTML is sanitized with `github.com/microcosm-cc/bluemonday`, so arbitrary scripts are not enabled by default.

The built-in theme is local CSS that follows GitHub Markdown conventions for page width, headings, tables, code blocks, blockquotes, lists, and task lists.

## Common Problems

`file does not exist`: check the path passed to `md-preview`.

`expected a Markdown file, got directory`: pass a file path instead of a folder.

`unsupported file extension`: use `.md` or `.markdown`.

`port 17776 on host 127.0.0.1 is not available`: choose another port with `--port`, or stop the process already using that port.

## License

MIT
