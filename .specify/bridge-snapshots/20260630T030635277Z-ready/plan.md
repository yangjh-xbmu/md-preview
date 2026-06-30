# Implementation Plan: Code Block Rendering Regression

**Branch**: `003-codeblock-rendering-regression` | **Date**: 2026-06-30 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/003-codeblock-rendering-regression/spec.md`

## Summary

Restore the bad release to the previous stable version, then reapply the code block no-shadow fix as a scoped, test-backed rendering change. The implementation must cover both the React preview stylesheet and the exported HTML stylesheet, and verification must prove that footnotes and Mermaid behavior from `v0.1.0` remain intact. Local visual validation must use a freshly rebuilt Wails binary, not the stale `build/bin/md-preview.exe` from 2026-06-12.

## Technical Context

**Language/Version**: Go 1.23, TypeScript 5.4, Wails v2.11.0, React 18, Vite 3

**Primary Dependencies**: goldmark with GFM, Footnote, wikilink; bluemonday; Prism.js; Mermaid; github-markdown-css

**Storage**: Filesystem Markdown input and generated build artifacts only

**Testing**: `go test ./...`, `npm --prefix frontend run build`, `wails build`, scriptable CSS inspection, fresh README preview smoke test

**Target Platform**: Windows desktop build via Wails WebView2; exported HTML in modern browsers

**Project Type**: Desktop app with Go backend and React frontend

**Performance Goals**: No measurable render-time regression; CSS-only code block style changes should not affect Markdown conversion time

**Constraints**:
- Keep changes narrow and dependency-free.
- Do not execute arbitrary user-provided script as part of Markdown rendering.
- Preserve current footnote, Mermaid, frontmatter, wiki link, export, and print behavior.
- Validate with current build artifacts before release.

**Scale/Scope**: One CSS behavior fix in the frontend, matching exported HTML style in `app.go`, focused regression tests in `main_test.go`, README update only if user-visible behavior changes.

## Constitution Check

The repository constitution file still contains placeholders, so project constraints from `AGENTS.md` and `CLAUDE.md` are authoritative:

- **Small and dependency-light**: PASS. No new dependencies are needed.
- **Prefer established libraries**: PASS. Existing Prism, goldmark, bluemonday, and Mermaid behavior remains unchanged.
- **No arbitrary scripts from Markdown**: PASS. The fix is CSS and template styling only.
- **Docs and verification discipline**: PASS. README and Spec Kit docs are updated; validation includes Go, frontend, Wails build, and smoke preview.

No complexity violations identified.

## Project Structure

### Documentation (this feature)

```text
specs/003-codeblock-rendering-regression/
├── checklists/
│   └── requirements.md
├── spec.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
└── tasks.md
```

### Source Code (repository root)

```text
app.go                    # exportHTMLTemplate inline CSS gains code no-shadow rules
frontend/src/App.css      # React preview CSS gains scoped Prism no-shadow rules
main_test.go              # regression tests for export CSS and preserved footnote behavior
README.md                 # feature wording if needed
AGENTS.md                 # Spec Kit plan pointer
```

**Structure Decision**: Keep the fix inside existing rendering style surfaces. Do not add a new frontend module because the problem is a Prism theme override, not a rendering lifecycle feature.

## Complexity Tracking

No complexity violations. The work is intentionally limited to CSS overrides, exported HTML parity, and tests.
