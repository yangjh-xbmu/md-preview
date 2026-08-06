# Tasks: macOS Full Screen and Quit Shortcuts

**Input**: Design documents from `/specs/006-mac-fullscreen-quit-shortcut/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, quickstart.md

**Organization**: Tasks are grouped by user story so each shortcut behavior can be implemented and validated independently.

## Phase 1: Setup and baseline

**Purpose**: Confirm the existing runtime surface and establish a clean validation baseline without changing the user's unrelated worktree modifications.

- [ ] T001 [P] Run `go test ./...` against `main_test.go` and `update_test.go` from the repository root and record the baseline result.
- [ ] T002 [P] Run `npm --prefix frontend run build` for `frontend/src/App.tsx` and record the baseline TypeScript and Vite result.
- [ ] T003 [P] Verify the existing `Quit`, `WindowFullscreen`, `WindowUnfullscreen`, and `WindowIsFullscreen` exports in `frontend/wailsjs/runtime/runtime.d.ts` and `frontend/wailsjs/runtime/runtime.js` before changing application code.

**Checkpoint**: Existing tests, frontend build, and required Wails runtime exports are available.

## Phase 2: User Story 1, macOS full-screen toggle (Priority: P1)

**Goal**: Keep `F11` working and make `Option+F11` call the same native full-screen toggle while preserving state synchronization and other shortcuts.

**Independent Test**: Build the frontend, launch the macOS desktop app with `README.md`, and verify `F11` and `Option+F11` both enter and exit native full screen.

### Implementation for User Story 1

- [ ] T004 [US1] Add a failing zero-dependency TypeScript test in `frontend/tests/keyboard.test.ts` covering unmodified `F11`, `Option+F11`, and rejected extra modifiers by importing the planned predicates from `frontend/src/keyboard.ts`.
- [ ] T005 [US1] Implement the tested full-screen shortcut predicate in `frontend/src/keyboard.ts` so `F11` and `altKey + F11` are accepted while unrelated modifier combinations are rejected.

**Checkpoint**: The app has one full-screen behavior path for the existing and macOS shortcuts.

## Phase 3: User Story 2, Ctrl+Q quit (Priority: P1)

**Goal**: Quit md-preview from the preview or menu focus state with the exact `Ctrl+Q` combination while preserving all other shortcuts.

**Independent Test**: Build and launch the macOS desktop app, focus the preview and a menu control in separate launches, press `Ctrl+Q`, and confirm the window and process exit.

### Implementation for User Story 2

- [ ] T006 [US2] Extend `frontend/tests/keyboard.test.ts` with a failing exact `Ctrl+Q` predicate test covering case normalization and rejected `Cmd+Q`, `Ctrl+Alt+Q`, and `Ctrl+Shift+Q` combinations.
- [ ] T007 [US2] Implement the tested quit shortcut predicate in `frontend/src/keyboard.ts` and wire both predicates into `frontend/src/App.tsx`, importing the existing `Quit` runtime function and returning before the general modifier dispatch.
- [ ] T008 [US2] Update the full-screen menu hint in `frontend/src/App.tsx` to show `F11 / ⌥F11` while keeping the existing menu action and `fullscreen` state label unchanged.
- [ ] T009 [US2] Update the shortcut table in `README.md` with `F11 / Option+F11` for full screen and `Ctrl+Q` for quitting, without changing the existing shortcut meanings.

**Checkpoint**: Both requested shortcuts are available, documented, and isolated from existing keyboard actions.

## Phase 4: Polish and cross-cutting validation

**Purpose**: Keep developer documentation current and validate the complete feature against the plan.

- [ ] T010 [P] Update the keyboard handling note in `CLAUDE.md` to record the macOS full-screen and quit shortcut behavior and the relevant verification commands.
- [ ] T011 [P] Run `go test ./...` against `main_test.go` and `update_test.go` after implementation.
- [ ] T012 [P] Run `npm --prefix frontend run build` for `frontend/src/App.tsx` after implementation.
- [ ] T013 [P] Run `wails build` using `wails.json` and confirm the desktop package completes.
- [ ] T014 Run the macOS manual smoke test in `specs/006-mac-fullscreen-quit-shortcut/quickstart.md`, including preview focus, menu focus, full-screen restoration, and process exit checks.

## Dependencies and execution order

### Phase dependencies

- **Setup and baseline**: no feature dependencies; complete before implementation.
- **User Story 1**: depends on T001, T002, and T003.
- **User Story 2**: depends on User Story 1 because both changes share the same keyboard handler in `frontend/src/App.tsx`.
- **Polish and validation**: depends on both user stories being implemented.

### Parallel opportunities

- T001, T002, and T003 can run in parallel.
- T010, T011, T012, and T013 can run in parallel after T009, although T014 depends on a successful desktop build from T013.

## Implementation strategy

1. Establish the Go and frontend baseline.
2. Make the smallest frontend-only full-screen change and preserve the existing menu path.
3. Add the existing runtime `Quit` import and exact `Ctrl+Q` branch.
4. Update user and developer documentation.
5. Run automated validation, desktop packaging, and the macOS smoke test.

The MVP is the two shortcut behaviors in `frontend/src/App.tsx`, with the README update required for user discoverability. No backend or dependency changes are in scope.
