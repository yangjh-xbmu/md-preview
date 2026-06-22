# Implementation Plan: Mermaid Rendering Support

**Branch**: `002-mermaid-support` | **Date**: 2026-06-22 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/002-mermaid-support/spec.md`

## Summary

Add Mermaid diagram rendering to md-preview by bundling the `mermaid` npm package in the frontend and scanning `pre > code.language-mermaid` nodes after each preview render. Theme switches re-initialize Mermaid and re-run against existing placeholders. Exported HTML loads Mermaid from a public CDN and runs the same scan on DOM ready. No Go backend changes to the rendering pipeline or sanitization policy are required; the only backend touch is appending Mermaid runtime tags to `exportHTMLTemplate`.

## Technical Context

**Language/Version**: Go 1.23, TypeScript 4.6, Wails v2.11.0, React 18, Vite 3

**Primary Dependencies** (new):
- `mermaid` (npm, v11+) — client-side diagram rendering

**Primary Dependencies** (existing, unchanged):
- `github.com/yuin/goldmark` + extensions (GFM, Footnote, wikilink)
- `github.com/microcosm-cc/bluemonday`
- `prismjs`, `github-markdown-css`

**Storage**: Filesystem assets only; no new persistent state.

**Testing**:
- `go test ./...` — backend regression (no new Go tests required; existing tests must still pass)
- `npm --prefix frontend run build` — TypeScript and Vite build must succeed
- `wails build` — desktop packaging must succeed
- Manual smoke test per `quickstart.md`

**Target Platform**: Windows desktop packaging (Wails webview); exported HTML is browser-portable.

**Project Type**: Desktop app with React frontend.

**Performance Goals**: Mermaid rendering is per-block and lazy; initial render should complete within 1 second for typical files (≤10 diagrams). Re-render on theme switch should complete within 1 second.

**Constraints** (from `AGENTS.md`):
- Keep the project small and dependency-light — Mermaid is the canonical library for this job; one npm dep is acceptable.
- Do not execute arbitrary user-provided script as part of Markdown rendering — Mermaid interprets its own DSL, not arbitrary JS; bluemonday still sanitizes the static HTML.
- Bind to 127.0.0.1 by default — unaffected.
- Update `README.md` feature list and shortcut table for new functionality (no new shortcut, but the feature list gains an entry).

**Scale/Scope**: One npm dependency, one new frontend module (`mermaid.ts` helper), one updated React effect in `App.tsx`, one template update in `app.go`, one CSS block in `App.css`, README update.

## Constitution Check

The project has no `.speckit/constitution.md`. Repository constraints from `AGENTS.md` apply:

- **"Keep the project small and dependency-light"** — Mermaid is the standard library for this feature; the alternative (server-side `mmdc`) would add a heavier Node.js toolchain. Bundling one npm package is the lighter option.
- **"Prefer established Markdown and HTML sanitization libraries"** — goldmark and bluemonday are unchanged. Mermaid is the established library for diagram rendering.
- **"Do not execute arbitrary user-provided script as part of Markdown rendering"** — Mermaid interprets a constrained DSL, not arbitrary JS. User-supplied `<script>` tags continue to be stripped by bluemonday before reaching the frontend.
- **"Keep user documentation in README.md and development notes in CLAUDE.md"** — README gets a feature entry; CLAUDE.md is updated if implementation notes warrant it.

No complexity violations identified.

## Project Structure

### Documentation (this feature)

```text
specs/002-mermaid-support/
├── spec.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── tasks.md
└── checklists/
    ├── requirements.md
    ├── ux.md
    ├── security.md
    └── testing.md
```

### Source Code (repository root)

```text
frontend/
├── package.json          (+ mermaid dependency)
├── src/
│   ├── App.tsx           (new useEffect for mermaid scan; theme re-render)
│   ├── App.css           (new .md-mermaid, .md-mermaid-error rules)
│   └── mermaid.ts        (new helper: initialize, renderBlocks, theme mapping)

app.go                    (exportHTMLTemplate gains mermaid CDN + initializer)
README.md                 (feature list entry)
CLAUDE.md                 (implementation note, if needed)
```

**Structure Decision**: Put Mermaid init/render logic in a dedicated `frontend/src/mermaid.ts` module so `App.tsx` stays focused on UI orchestration. The module exports a single `renderMermaidBlocks(root, theme)` function plus a `reinitForTheme(theme)` function. `App.tsx` calls these from the existing `contentHtml` effect and a new `theme` effect.

## Complexity Tracking

- One new runtime concern: Mermaid render lifecycle. Mitigated by scoping all Mermaid calls behind `mermaid.ts` and never letting Mermaid globals leak into `App.tsx`.
- One new external script in exported HTML: Mermaid CDN. Mitigated by `defer` + `DOMContentLoaded` guard + a `pre > code.language-mermaid` scan that matches the desktop path.
- No new Go API surface. No new Wails bindings. No new IPC events.
- No bluemonday policy changes. No goldmark extension changes.
