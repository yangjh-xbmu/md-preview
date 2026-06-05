> This file is for people and AI coding assistants developing, debugging, and extending this repository.
> End-user documentation belongs in [README.md](./README.md).

# md-preview Development Notes

## Product Goal

Build a small Go CLI that previews Markdown files as rendered HTML, using GitHub-style Markdown rendering and GitHub-like CSS as the default theme.

The tool is for fast local reading. It should show rendered content, not Markdown source.

## Expected Shape

- CLI binary name: `md-preview`
- Primary command: `md-preview <file.md>`
- Starts a local HTTP server.
- Opens the rendered preview in the default browser when possible.
- Watches the Markdown file and refreshes the browser view after edits.
- Uses GitHub-style Markdown extensions where practical.
- Uses GitHub-like CSS as the built-in default theme.

## Current Implementation

- `main.go` contains the CLI entry point, argument parsing, file validation, Markdown rendering, HTTP handlers, browser launching, and the built-in page template.
- `main_test.go` covers argument parsing, file validation errors, GFM rendering, HTML sanitization, and HTTP handler behavior.
- Markdown rendering uses `github.com/yuin/goldmark` with `extension.GFM`.
- Rendered HTML is sanitized with `github.com/microcosm-cc/bluemonday`.
- Browser refresh uses a lightweight page poll to `/status`. When the Markdown file modification version changes, the page reloads.
- The default bind address is `127.0.0.1:17776`.

## CLI Options

```text
Usage: md-preview [--host 127.0.0.1] [--port 17776] [--no-open] [--watch=false] <file.md>
```

## Development Principles

- Keep the tool small and dependency-light.
- Prefer established Go libraries for Markdown parsing and HTML sanitization.
- Do not build an Electron app or a full editor.
- Do not show the source Markdown as the primary experience.
- Keep generated HTML local-only by default.

## Initial Acceptance Checklist

- `go test ./...` passes.
- `go run . README.md` starts a preview server.
- The rendered page looks close to GitHub Markdown: readable width, GitHub-ish typography, tables, code blocks, blockquotes, task lists.
- Missing file, directory input, non-Markdown extension, and occupied port errors are clear.
- Browser refresh works through lightweight polling while `--watch` is enabled.

## Verification

```bash
go test ./...
go run . --no-open --port 0 README.md
```

## Commit Message Style

Use short conventional commits, for example `feat: add markdown preview server`.
