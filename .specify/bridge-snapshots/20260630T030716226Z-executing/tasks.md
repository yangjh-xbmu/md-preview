# Tasks: Auto Update

**Input**: Design documents from `specs/004-auto-update/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/update.md, quickstart.md

**Tests**: Required. This feature changes startup behavior, release networking, persistence, and executable replacement handoff, so backend tests and frontend build validation are mandatory.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing.

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Confirm release contract, version injection path, and current generated bindings before feature code changes.

- [ ] T001 Verify current GitHub release asset naming and compatible Windows asset shape for `md-preview-v<VERSION>-windows-amd64.zip`
- [ ] T002 Add application version injection plan to `.github/workflows/release.yml` and local Wails build command expectations
- [ ] T003 [P] Confirm frontend Wails binding files that must be updated for new backend methods in `frontend/wailsjs/go/main/App.*`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Shared backend models, settings persistence, status state, and release parsing needed by every story.

- [ ] T004 Add update data types, default constants, and app version variable in `update.go` and `main.go`
- [ ] T005 Add persistent update settings load/save helpers in `update.go`
- [ ] T006 Add update status state storage and event emission helpers in `update.go`
- [ ] T007 [P] Add settings default and persistence tests in `update_test.go`
- [ ] T008 [P] Add semantic version comparison and platform asset selection tests in `update_test.go`

**Checkpoint**: Backend can remember the startup preference, compare versions, and identify compatible assets without touching the UI.

---

## Phase 3: User Story 1 - Stay Current Automatically (Priority: P1) MVP

**Goal**: Startup checks run asynchronously by default, find newer stable releases, download and stage compatible assets, and never block Markdown preview.

**Independent Test**: Use an HTTP test server that reports current, newer, malformed, and incompatible releases; run `go test ./...` and confirm startup check methods return non-blocking statuses.

### Tests for User Story 1

- [ ] T009 [P] [US1] Add latest release metadata tests in `update_test.go`
- [ ] T010 [P] [US1] Add download, digest verification, and extraction tests in `update_test.go`
- [ ] T011 [US1] Add startup check non-blocking behavior test in `update_test.go`

### Implementation for User Story 1

- [ ] T012 [US1] Implement release metadata fetch and stable release filtering in `update.go`
- [ ] T013 [US1] Implement compatible asset download, SHA-256 verification, and Windows zip extraction in `update.go`
- [ ] T014 [US1] Implement startup update check wiring in `app.go`
- [ ] T015 [US1] Implement staged Windows install handoff and safe unsupported fallback in `update.go`

**Checkpoint**: User Story 1 is functional and testable independently through backend methods and startup behavior.

---

## Phase 4: User Story 2 - Disable Automatic Updates From Menu (Priority: P2)

**Goal**: Users can disable and re-enable startup update checks from the existing app menu, and the preference persists across restarts.

**Independent Test**: Toggle the setting through frontend bindings, restart, and confirm startup checks are skipped while manual checks still work.

### Tests for User Story 2

- [ ] T016 [P] [US2] Add backend toggle and manual-check-while-disabled tests in `update_test.go`

### Implementation for User Story 2

- [ ] T017 [US2] Add Wails-bound update settings and manual check methods in `app.go`
- [ ] T018 [US2] Update generated frontend bindings in `frontend/wailsjs/go/main/App.js` and `frontend/wailsjs/go/main/App.d.ts`
- [ ] T019 [US2] Add update toggle and manual check controls to `frontend/src/App.tsx`
- [ ] T020 [US2] Add compact update menu styling in `frontend/src/App.css`

**Checkpoint**: The existing app menu can disable, re-enable, and manually trigger update checks.

---

## Phase 5: User Story 3 - Understand Update Status (Priority: P3)

**Goal**: Users can see clear update states for disabled, checking, up to date, available, downloading, ready, and failed outcomes.

**Independent Test**: Simulate backend status events and confirm the UI displays the status text without breaking existing menu or preview interactions.

### Tests for User Story 3

- [ ] T021 [P] [US3] Add backend status transition tests in `update_test.go`
- [ ] T022 [US3] Add frontend build validation coverage through `npm --prefix frontend run build`

### Implementation for User Story 3

- [ ] T023 [US3] Subscribe to `update-status-changed` and render update status in `frontend/src/App.tsx`
- [ ] T024 [US3] Add install/restart action for ready updates in `frontend/src/App.tsx` and backend binding in `app.go`

**Checkpoint**: Update status is visible and actionable without disrupting Markdown preview.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, release metadata, generated artifacts, and final verification.

- [ ] T025 [P] Update user-facing automatic update documentation in `README.md`
- [ ] T026 Update release workflow build command in `.github/workflows/release.yml` to inject `main.appVersion`
- [ ] T027 Run `go test ./...`
- [ ] T028 Run `npm --prefix frontend run build`
- [ ] T029 Run `wails build -ldflags "-X main.appVersion=0.1.1"` and restore unrelated generated side effects if any
- [ ] T030 Run quickstart scenarios from `specs/004-auto-update/quickstart.md`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies.
- **Foundational (Phase 2)**: Depends on setup and blocks all user stories.
- **User Story 1 (Phase 3)**: Depends on foundation and is the MVP.
- **User Story 2 (Phase 4)**: Depends on foundation and integrates with User Story 1 methods.
- **User Story 3 (Phase 5)**: Depends on status events from User Story 1 and menu surface from User Story 2.
- **Polish (Phase 6)**: Depends on selected stories being complete.

### User Story Dependencies

- **User Story 1 (P1)**: Can start after foundation.
- **User Story 2 (P2)**: Can start after foundation but is most useful after US1 has update methods.
- **User Story 3 (P3)**: Depends on US1 status transitions and US2 visible menu controls.

### Parallel Opportunities

- T003 can run alongside T001 and T002.
- T007 and T008 can run in parallel after T004-T006.
- T009 and T010 can run in parallel before US1 implementation.
- T016 can run while frontend UI work is pending.
- T021 can run alongside documentation updates.

## Implementation Strategy

Complete the MVP first: foundation, then US1 backend update checking and safe staging. Add menu preference controls next, then status visibility and install action. Do not run release or publish steps until all validation tasks are complete.
