# This file is for people and AI coding assistants developing, debugging, and extending this repository.
> End-user documentation belongs in [README.md](./README.md).

# md-preview Development Notes

## Product Goal

Provide a small Go CLI that previews Markdown as rendered content with GitHub-style style by default in a local desktop window.

## Expected Shape

- CLI binary name: `md-preview`
- Primary command: `md-preview <file.md>`
- Starts desktop preview window by default and refreshes when file changes.
- Supports `--browser` to use local HTTP preview mode on `127.0.0.1`.
- Supports `--host`, `--port`, `--no-open`, and `--watch=false`.
- Renders Markdown for desktop preview and HTML for browser mode with a sanitizer.

## Implementation Notes

- `main.go` holds CLI parsing, validation, rendering, and HTTP handlers.
- Desktop mode uses `fyne.io/fyne/v2` for native rendering and a polling-based watcher.
- Browser mode keeps the old `goldmark + bluemonday` path, serving a local preview page and `/status` endpoint.
- Validation rejects non-existing files, directories, and unsupported extensions.

## Development Principles

- Keep the project small and dependency-light.
- Prefer established libraries for Markdown parsing and HTML sanitization.
- Do not execute arbitrary user-provided scripts.
- Keep generated HTML local-only by default.

## Verification

```bash
go test ./...

# Desktop default

go run . README.md

# Browser mode

go run . --browser --no-open --port 0 README.md
```
