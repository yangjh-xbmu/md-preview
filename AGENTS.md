# md-preview

## Position

Small Go CLI for previewing a local Markdown file as rendered GitHub-style HTML.

## Logic

The CLI validates one `.md` or `.markdown` file, starts a local HTTP server, renders the file with goldmark GFM support, sanitizes generated HTML with bluemonday, and serves a built-in preview page. When watch mode is enabled, the browser polls `/status` and reloads after the file version changes.

## Constraints

- Bind to `127.0.0.1` by default.
- Keep the project small and dependency-light.
- Prefer established Markdown and HTML sanitization libraries.
- Do not execute arbitrary user-provided script as part of Markdown rendering.
- Keep user documentation in `README.md` and development notes in `CLAUDE.md`.

## Domain Map

| Area | File | Purpose |
| --- | --- | --- |
| CLI and server | `main.go` | Argument parsing, validation, rendering, HTTP handlers, page template, browser launch |
| Tests | `main_test.go` | CLI parsing, file validation, rendering safety, HTTP behavior |
| User docs | `README.md` | Install, run, options, troubleshooting |
| Dev docs | `CLAUDE.md` | Implementation notes and verification commands |
