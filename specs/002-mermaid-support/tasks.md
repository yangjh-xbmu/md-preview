# Tasks: Mermaid Rendering Support

**Input**: Design documents from `specs/002-mermaid-support/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, quickstart.md

**Tests**: `go test ./...`, `npm --prefix frontend run build`, `wails build`, plus manual smoke test per `quickstart.md`.

**Organization**: Tasks are grouped by phase. The single user story (Mermaid rendering) is implemented as a vertical slice: dependency → helper → React wiring → export → docs → validation.

## Phase 1: Setup

**Purpose**: Add the Mermaid dependency and confirm it builds into the frontend bundle.

- [x] T001 Run `npm --prefix frontend install mermaid` and verify `package.json` + `package-lock.json` reflect the addition
- [x] T002 Run `npm --prefix frontend run build` to confirm Vite can bundle Mermaid without TypeScript errors (required upgrading `typescript` to 5.4 and adding `frontend/src/mermaid-shim.d.ts` via `tsconfig.json` `paths`)

## Phase 2: Foundational

**Purpose**: Build the Mermaid helper module that the React layer will call.

- [x] T003 Create `frontend/src/mermaid.ts` exporting `mermaidThemeFor(theme)`, `reinitForTheme(theme)`, and `renderMermaidBlocks(root, theme)`
- [x] T004 Implement `mermaidThemeFor` mapping: `github-light` → `default`, `github-dark` → `dark`, `github-sepia` → `default`
- [x] T005 Implement `renderMermaidBlocks(root, theme)`: scan `pre > code.language-mermaid`, replace each `<pre>` with a `<div class="md-mermaid">` placeholder, call `mermaid.render`, catch per-block errors into `<div class="md-mermaid-error">`. Source preserved in `data-mermaid-source` for theme-switch re-render.

## Phase 3: User Story 1 — Render Mermaid diagrams inline (Priority: P1)

**Goal**: Wire the helper into `App.tsx` so Mermaid blocks render on preview load and on theme change.

**Independent Test**: Open `specs/002-mermaid-support/sample.md` (per `quickstart.md`). Mermaid blocks render as SVG; non-Mermaid code blocks retain Prism highlighting; theme switch re-renders diagrams.

### Implementation for User Story 1

- [x] T006 [US1] In `App.tsx`, add a new `useEffect` keyed on `[contentHtml, theme]` that calls `reinitForTheme(theme)` and `renderMermaidBlocks(previewRef.current, theme)`, declared before the Prism effect so Mermaid blocks are replaced before Prism scans
- [x] T007 [US1] The same effect handles theme re-render: when only `theme` changes, existing `.md-mermaid` placeholders are re-rendered from `data-mermaid-source`
- [x] T008 [US1] Add `.md-mermaid`, `.md-mermaid-error`, `.md-mermaid-empty`, and sepia-tinted container rules to `frontend/src/App.css`
- [x] T009 [US1] Prism effect updated to skip `language-mermaid` code blocks so the copy button and line-numbers class are not applied to diagrams

## Phase 4: Export Support

**Purpose**: Make exported standalone HTML render Mermaid when opened in a browser.

- [x] T010 [US1] Update `exportHTMLTemplate` in `app.go` to include a Mermaid CDN `<script>` tag with `defer`
- [x] T011 [US1] Add an inline initializer script in the template that scans `pre > code.language-mermaid` on `DOMContentLoaded` and replaces each with a rendered SVG, mirroring the frontend logic; theme initialized from a `%s` placeholder filled with `default` or `dark`
- [x] T012 [US1] Added Go tests: `TestLoadMarkdownPreservesMermaidCodeBlock`, `TestExportHTMLIncludesMermaidRuntime`, `TestExportHTMLMermaidThemeMatchesPreviewTheme`. Updated `TestExportHTMLWritesFileWithThemeAndSanitization` to scope the `<script>` check to user-provided content.

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Update docs and confirm scope discipline.

- [x] T013 Update `README.md` feature list to mention Mermaid rendering (added Mermaid example section)
- [x] T014 Update `CLAUDE.md` implementation notes with the Mermaid integration point
- [x] T015 Update `AGENTS.md` Domain Map with `frontend/src/mermaid.ts` and `frontend/src/mermaid-shim.d.ts`
- [x] T016 Confirm `git status` shows only Mermaid-related changes plus spec artifacts

## Phase 6: Validation

**Purpose**: Run the full verification suite end to end.

- [x] T017 Run `go test ./...`
- [x] T018 Run `npm --prefix frontend run build`
- [x] T019 Run `wails build`
- [ ] T020 Manual smoke test per `quickstart.md` (render, theme switch, export, error case) — left for the user to run interactively

## Dependencies & Execution Order

### Phase Dependencies

- **Setup**: No dependencies.
- **Foundational**: Depends on Setup (mermaid package installed).
- **User Story 1**: Depends on Foundational (helper module exists).
- **Export Support**: Depends on Foundational (mirrors frontend scan logic).
- **Polish**: Can run after User Story 1 implementation is functionally complete.
- **Validation**: Depends on all implementation phases.

### User Story Dependencies

- **User Story 1 (P1)**: No dependency on other user stories. Export Support extends US1 to standalone HTML and is tracked under the same user story.

## Implementation Strategy

Implement the vertical slice bottom-up: install dep → helper module → React wiring → export HTML → docs → validate. Stop if Vite cannot bundle Mermaid, if Mermaid fails to initialize in the webview, or if the export template change breaks existing export tests.
