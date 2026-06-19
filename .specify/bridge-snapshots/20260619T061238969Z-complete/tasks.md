# Tasks: Set md-preview App Icon

**Input**: Design documents from `specs/001-set-app-icon/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, quickstart.md

**Tests**: This asset-only feature uses validation commands instead of new automated tests.

**Organization**: Tasks are grouped by the single user story so the feature remains independently verifiable.

## Phase 1: Setup

**Purpose**: Confirm inputs and current packaging path before replacing assets.

- [x] T001 Verify the supplied SVG exists and is readable at `D:/Users/yangjh/Desktop/Inbox/md-preview-app-icon-final-left-layout.svg`
- [x] T002 Verify the current Wails icon assets exist at `build/appicon.png` and `build/windows/icon.ico`

## Phase 2: Foundational

**Purpose**: Establish a valid conversion path before modifying the packaged icon.

- [x] T003 Identify available local tooling for SVG-to-ICO conversion without adding project dependencies

## Phase 3: User Story 1 - Recognizable desktop app icon (Priority: P1)

**Goal**: Package md-preview with the supplied final app icon artwork.

**Independent Test**: `file build/appicon.png build/windows/icon.ico` identifies a valid PNG source and multi-size Windows icon resource, and `wails build` completes successfully.

### Implementation for User Story 1

- [x] T004 [US1] Convert `D:/Users/yangjh/Desktop/Inbox/md-preview-app-icon-final-left-layout.svg` into `build/appicon.png` and a multi-size Windows icon at `build/windows/icon.ico`
- [x] T005 [US1] Inspect `build/appicon.png` and `build/windows/icon.ico` to confirm they are valid icon assets
- [x] T006 [US1] Run `go test ./...`
- [x] T007 [US1] Run `npm --prefix frontend run build`
- [x] T008 [US1] Run `wails build`

## Phase 4: Polish & Cross-Cutting Concerns

**Purpose**: Record completion state and avoid unrelated churn.

- [x] T009 Update `specs/001-set-app-icon/tasks.md` checkboxes for completed tasks
- [x] T010 Confirm git status separates this feature from unrelated pre-existing working tree changes

## Dependencies & Execution Order

### Phase Dependencies

- **Setup**: No dependencies.
- **Foundational**: Depends on setup confirmation.
- **User Story 1**: Depends on conversion tooling confirmation.
- **Polish**: Depends on successful validation.

### User Story Dependencies

- **User Story 1 (P1)**: No dependency on other user stories.

## Implementation Strategy

Complete the single user story as the MVP. Stop if the source SVG is unavailable, conversion cannot produce a valid `.ico`, or the build fails.
