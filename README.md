# md-preview

Small desktop app for rendering a local Markdown file in a GitHub-style preview window.

## Requirements

- Go 1.22+
- Node.js (for frontend build)
- [Wails](https://wails.io/) CLI

## Install and run

```bash
go mod download
wails build
./md-preview <file.md>
```

`<file.md>` can also be `.markdown`.

## CLI options

- `--watch=false` disables file watching
- `--browser` is kept for compatibility and currently maps to desktop mode

Examples:

```bash
md-preview README.md
md-preview --watch=false notes.markdown
```

## Notes

- The app uses `goldmark` + `github.com/yuin/goldmark/extension` for Markdown rendering.
- Output is sanitized with `github.com/microcosm-cc/bluemonday`.
- Rendering and updates are handled in Go. The frontend listens to `markdown-updated` events from Wails.

## Development

```bash
wails dev
```

You can also build the frontend manually:

```bash
cd frontend
npm install
npm run build
```

## Troubleshooting

- `file does not exist`: check the Markdown path and permissions.
- `unsupported file extension`: use `.md` or `.markdown`.
- `expected a Markdown file, got directory`: pass a file path instead of a folder.
