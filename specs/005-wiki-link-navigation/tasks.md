# Tasks: Wiki Link Navigation Fix

**Input**: Design documents from `/specs/005-wiki-link-navigation/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md

**Organization**: This is a single-user-story bug fix; tasks are grouped into Foundational and User Story 1 phases.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to
- Include exact file paths in descriptions

---

## Phase 1: Foundational (Blocking Prerequisites)

**Purpose**: Ensure the existing test/build pipeline passes before changing behavior.

**⚠️ CRITICAL**: No implementation work should begin until this phase confirms a clean baseline.

- [x] T001 Run `go test ./...` and confirm all existing tests pass before changes
- [x] T002 Run `npm --prefix frontend install` and `npm --prefix frontend run build` to confirm the frontend builds cleanly

**Checkpoint**: Baseline tests and build are green.

---

## Phase 2: User Story 1 - Click any wiki link and navigate reliably (Priority: P1) 🎯 MVP

**Goal**: Wiki links with ASCII, Chinese-character, space-encoded, `.html`, and extension-less targets resolve to the correct local Markdown file; missing targets show a friendly message.

**Independent Test**: Run the new and existing Go tests for `ResolveWikiLink`; verify the frontend build still succeeds.

### Tests for User Story 1

- [x] T003 [P] [US1] Add Go unit tests in `main_test.go` covering: ASCII target (`README`), Chinese target (`Git基本概念与常用命令`), space-encoded target (`My%20Note`), `.html` suffix target, `.md` suffix target, missing target, and invalid extension target

### Implementation for User Story 1

- [x] T004 [US1] Update `ResolveWikiLink` in `app.go` to decode percent-encoded `href` values with `net/url.PathUnescape` before extension normalization and file resolution
- [x] T005 [US1] Verify the frontend click handler in `frontend/src/App.tsx` still correctly skips external/anchor/mailto links and shows the existing "Wiki link target not found" message for unresolved targets

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently.

---

## Phase 3: Polish & Cross-Cutting Concerns

**Purpose**: Final validation and documentation.

- [x] T006 [P] Run `go test ./...` and confirm all tests pass after the change
- [x] T007 [P] Run `npm --prefix frontend run build` and confirm the frontend build succeeds after the change
- [x] T008 [P] Optionally update `README.md` or `CLAUDE.md` if user-facing behavior or developer notes need clarification
- [ ] T009 Run the manual smoke test steps in `quickstart.md`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Foundational (Phase 1)**: No dependencies - confirms a green baseline
- **User Story 1 (Phase 2)**: Depends on Foundational phase completion
- **Polish (Phase 3)**: Depends on User Story 1 completion

### Parallel Opportunities

- T001 and T002 can run in parallel.
- T003 and T004/T005 can run in parallel after T001/T002, but T004 must complete before T003 can pass (TDD-style: write tests, watch them fail, then implement).
- T006, T007, and T008 can run in parallel after implementation.

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: confirm baseline is green.
2. Complete Phase 2: add tests, then fix `ResolveWikiLink` decoding.
3. Complete Phase 3: run full validation and optional docs update.
4. **STOP and VALIDATE**: `go test ./...`, frontend build, and manual smoke test all pass.

---

## Notes

- The fix is intentionally small: one standard-library import, one resolver function change, and focused tests.
- No new Wails bindings, IPC events, or frontend structural changes are required.
