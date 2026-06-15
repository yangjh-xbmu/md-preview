# md-preview

## Position

Small Go and Wails app for previewing a local Markdown file as rendered GitHub-style HTML.

## Logic

The CLI validates one `.md` or `.markdown` file, starts the Wails desktop app, renders the file with goldmark GFM and footnote support, sanitizes generated HTML with bluemonday, and serves the React preview surface. When watch mode is enabled, the desktop backend polls the file version and emits reload events after the file changes.

## Constraints

- Bind to `127.0.0.1` by default.
- Keep the project small and dependency-light.
- Prefer established Markdown and HTML sanitization libraries.
- Do not execute arbitrary user-provided script as part of Markdown rendering.
- Keep user documentation in `README.md` and development notes in `CLAUDE.md`.

## Domain Map

| Area | File | Purpose |
| --- | --- | --- |
| CLI entry | `main.go` | Argument parsing, file validation, compatibility flags, Wails app startup |
| Preview styles | `frontend/src/App.css` | Desktop shell, Markdown content, footnotes, themes, print and frontmatter styling |
| Desktop backend | `app.go` | Wails binding, Markdown rendering, footnote sanitization, file watching, export and print actions |
| Tests | `main_test.go` | CLI parsing, file validation, rendering safety, HTTP behavior |
| User docs | `README.md` | Install, run, options, troubleshooting |
| Dev docs | `CLAUDE.md` | Implementation notes and verification commands |
