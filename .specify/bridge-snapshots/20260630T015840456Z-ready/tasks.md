# Tasks: Code Block Rendering Regression

**Input**: Design documents from `specs/003-codeblock-rendering-regression/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, quickstart.md

**Tests**: Required. This feature uses TDD because the previous release attempt produced a regression concern.

**Organization**: Tasks are grouped by user story and must remain independently verifiable.

## Phase 1: Setup

**Purpose**: Confirm the rollback baseline and isolate implementation work.

- [ ] T001 Verify `v0.1.1` draft release and tag are removed, and `v0.1.0` is the latest release
- [ ] T002 Verify the work branch is `003-codeblock-rendering-regression` and `main` contains the revert commit
- [ ] T003 Confirm unrelated generated-file diffs remain outside this feature branch commit scope

## Phase 2: Foundational

**Purpose**: Add regression coverage before reapplying the visual fix.

- [ ] T004 [P] Add or extend Go tests in `main_test.go` proving exported HTML contains code block no-shadow rules
- [ ] T005 [P] Confirm existing Go tests in `main_test.go` cover footnote rendering and export footnote styles
- [ ] T006 [P] Add a scriptable frontend CSS inspection step for `frontend/src/App.css` to assert Prism text-shadow overrides exist

## Phase 3: User Story 1 - Read code blocks without shadow artifacts (Priority: P1)

**Goal**: Normal Prism-highlighted code blocks render without text-shadow in all themes.

**Independent Test**: Inspect `frontend/src/App.css` and rendered/exported style rules for `text-shadow: none` on code, token descendants, and line-number pseudo-content.

### Tests for User Story 1

- [ ] T007 [US1] Run the CSS inspection before implementation and confirm it fails for missing no-shadow rules
- [ ] T008 [US1] Run the export style Go test before implementation and confirm it fails for missing no-shadow rules

### Implementation for User Story 1

- [ ] T009 [US1] Add scoped Prism no-shadow rules to `frontend/src/App.css`
- [ ] T010 [US1] Add matching no-shadow rules to `exportHTMLTemplate` in `app.go`
- [ ] T011 [US1] Keep Light, Dark, and Sepia code block borders/backgrounds readable in `frontend/src/App.css` and `app.go`

## Phase 4: User Story 2 - Preserve existing Markdown features (Priority: P1)

**Goal**: Footnotes and Mermaid behavior from `v0.1.0` still work after the visual fix.

**Independent Test**: Backend tests pass for footnote HTML and exported styles; Mermaid blocks remain excluded from Prism code-block decoration in `frontend/src/App.tsx`.

### Tests for User Story 2

- [ ] T012 [US2] Run `go test ./...` and confirm footnote tests pass
- [ ] T013 [US2] Inspect `frontend/src/App.tsx` to confirm `language-mermaid` is skipped by Prism decoration

### Implementation for User Story 2

- [ ] T014 [US2] Preserve existing footnote sanitization and styles in `app.go`, `frontend/src/App.css`, and tests
- [ ] T015 [US2] Preserve existing Mermaid rendering and Prism-skip behavior in `frontend/src/App.tsx`

## Phase 5: User Story 3 - Verify with current build artifacts (Priority: P2)

**Goal**: Final preview uses freshly built artifacts.

**Independent Test**: `wails build` updates `build/bin/md-preview.exe`, then README preview launches from that binary.

### Implementation for User Story 3

- [ ] T016 [US3] Run `npm --prefix frontend install` if dependencies are missing
- [ ] T017 [US3] Run `npm --prefix frontend run build`
- [ ] T018 [US3] Run `wails build`
- [ ] T019 [US3] Launch freshly built `build/bin/md-preview.exe README.md`
- [ ] T020 [US3] Confirm `build/bin/md-preview.exe` timestamp is newer than the implementation changes

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, release hygiene, and final verification.

- [ ] T021 Update `README.md` feature wording only if the user-visible behavior needs documentation
- [ ] T022 Ensure `AGENTS.md` Speckit block points to `specs/003-codeblock-rendering-regression/plan.md`
- [ ] T023 Run `go test ./...`
- [ ] T024 Run `npm --prefix frontend run build`
- [ ] T025 Run `wails build`
- [ ] T026 Review `git diff` to confirm only feature-related files are included

## Dependencies & Execution Order

### Phase Dependencies

- **Setup**: No dependencies.
- **Foundational**: Depends on Setup.
- **User Story 1**: Depends on failing regression checks from Foundational.
- **User Story 2**: Can run after User Story 1 implementation because it validates preserved behavior.
- **User Story 3**: Depends on implementation being complete.
- **Polish**: Depends on all user stories.

### User Story Dependencies

- **US1**: No dependency on other user stories. This is the MVP.
- **US2**: Depends on US1 changes to verify no regressions.
- **US3**: Depends on US1 and US2 to verify the packaged app.

## Parallel Opportunities

- T004, T005, and T006 can be prepared in parallel because they inspect different surfaces.
- T012 and T013 can run in parallel after implementation because they inspect backend tests and frontend logic.

## Implementation Strategy

Rebuild trust in this order: rollback verified, tests added, code changed narrowly, preserved features checked, Wails binary rebuilt, README preview launched from the fresh binary. Do not release until all tasks are checked off.
