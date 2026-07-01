# Implementation Plan: Wiki Link Navigation Fix

**Branch**: `005-wiki-link-navigation` | **Date**: 2026-07-01 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/005-wiki-link-navigation/spec.md`

## Summary

Fix wiki link navigation so that links produced by the goldmark wikilink extension resolve correctly regardless of URL-encoded characters, spaces, or non-ASCII file names. The Go backend `ResolveWikiLink` binding will decode the incoming `href` with `url.PathUnescape` before resolving it to a local `.md` file. The frontend click handler is already in place and only needs to keep treating external/anchor/mailto links as non-wiki links. New Go unit tests will cover ASCII, Chinese-character, space-encoded, `.html`, `.md`, and missing-target cases.

## Technical Context

**Language/Version**: Go 1.23, TypeScript 4.6, Wails v2.11.0, React 18, Vite 3

**Primary Dependencies** (existing, unchanged):
- `github.com/yuin/goldmark` with GFM, Footnote, and `go.abhg.dev/goldmark/wikilink`
- `github.com/microcosm-cc/bluemonday`
- `prismjs`, `github-markdown-css`

**Storage**: Filesystem assets only; no new persistent state.

**Testing**:
- `go test ./...` — backend regression plus new `ResolveWikiLink` cases
- `npm --prefix frontend run build` — TypeScript and Vite build must succeed
- `wails build` — desktop packaging must succeed
- Manual smoke test per `quickstart.md`

**Target Platform**: Windows desktop packaging (Wails webview).

**Project Type**: Desktop app with React frontend.

**Performance Goals**: Wiki link resolution is synchronous file-system lookup; must complete within 100 ms for typical local files.

**Constraints** (from `AGENTS.md`):
- Keep the project small and dependency-light — no new dependencies; reuse Go standard library `net/url`.
- Do not execute arbitrary user-provided script as part of Markdown rendering — unaffected.
- Bind to 127.0.0.1 by default — unaffected.
- Keep user documentation in `README.md` and development notes in `CLAUDE.md` — README gets a feature note if warranted.

**Scale/Scope**: One function change in `app.go`, one set of new test cases in `main_test.go`, possible minor frontend message cleanup, no new dependencies.

## Constitution Check

The project has no populated `.specify/memory/constitution.md`. Repository constraints from `AGENTS.md` apply:

- **"Keep the project small and dependency-light"** — the fix uses only the Go standard library `net/url` and existing resolver logic. No new packages.
- **"Prefer established Markdown and HTML sanitization libraries"** — goldmark and bluemonday are unchanged.
- **"Do not execute arbitrary user-provided script as part of Markdown rendering"** — no script execution; this is purely link resolution.
- **"Keep user documentation in README.md and development notes in CLAUDE.md"** — README and CLAUDE.md are updated only if the behavior materially changes user-facing docs.

No complexity violations identified.

## Project Structure

### Documentation (this feature)

```text
specs/005-wiki-link-navigation/
├── spec.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
└── checklists/
    └── requirements.md
```

### Source Code (repository root)

```text
app.go                    (update ResolveWikiLink to use url.PathUnescape)
main_test.go              (add ResolveWikiLink test cases)
frontend/src/App.tsx      (minor status-message consistency, if needed)
README.md                 (optional: feature note)
CLAUDE.md                 (optional: implementation note)
```

**Structure Decision**: Keep the resolver in `app.go` where it already lives; add focused unit tests in `main_test.go` rather than a new test file. Frontend link handling stays in `App.tsx` and requires no structural change.

## Complexity Tracking

- One new standard-library import (`net/url`).
- No new Wails bindings.
- No new IPC events.
- No bluemonday policy changes.
- No goldmark extension changes.
