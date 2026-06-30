# Feature Specification: Auto Update

**Feature Branch**: `004-auto-update`

**Created**: 2026-06-30

**Status**: Draft

**Input**: User description: "Increase a feature that checks version information on every startup and automatically updates itself, enabled by default, but can be disabled from the menu. Strictly use the Spec Kit flow."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Stay Current Automatically (Priority: P1)

As a desktop user, I want the app to check for a newer official release when it starts and update itself without extra manual searching, so I can keep using the latest stable version with minimal interruption.

**Why this priority**: This is the core value of the feature. Without startup checking and update execution, the menu setting has no practical effect.

**Independent Test**: Run the app in an environment where a newer official release is reported. The user receives a clear update status and the app prepares the newer version without blocking Markdown preview startup.

**Acceptance Scenarios**:

1. **Given** automatic updates are enabled and a newer stable release exists, **When** the app starts, **Then** it checks version information and begins the update flow without preventing the current document from loading.
2. **Given** automatic updates are enabled and the installed version is current, **When** the app starts, **Then** it reports that the app is up to date and does not download anything.
3. **Given** automatic updates are enabled and the version check cannot complete, **When** the app starts, **Then** Markdown preview remains usable and the user sees a non-blocking failure status.

---

### User Story 2 - Disable Automatic Updates From Menu (Priority: P2)

As a user on a controlled network or a machine where updates are managed manually, I want to turn off automatic update checks from the app menu, so startup does not contact the release service until I re-enable it.

**Why this priority**: The user explicitly requires the default-on behavior to be reversible from the menu.

**Independent Test**: Toggle the menu setting off, restart the app, and confirm no startup update check occurs. Toggle it back on and confirm the next startup checks again.

**Acceptance Scenarios**:

1. **Given** automatic updates are enabled, **When** the user disables them from the menu, **Then** the setting persists and startup checks stop.
2. **Given** automatic updates are disabled, **When** the user re-enables them from the menu, **Then** the app resumes startup checks.
3. **Given** the menu is opened, **When** the user views the update setting, **Then** its current enabled or disabled state is visible.

---

### User Story 3 - Understand Update Status (Priority: P3)

As a user, I want clear update status messages, so I know whether the app is current, downloading an update, ready to restart, or unable to update.

**Why this priority**: Status visibility reduces confusion, but it depends on the update check and setting behavior existing first.

**Independent Test**: Simulate current, available, download failure, and disabled states, then confirm the displayed status is accurate and non-blocking.

**Acceptance Scenarios**:

1. **Given** a newer release is available, **When** the update flow starts, **Then** the user sees progress or state text that distinguishes checking, downloading, ready, and failed states.
2. **Given** automatic updates are disabled, **When** the app starts, **Then** the user can still see that startup update checks are disabled.

### Edge Cases

- The release service is unreachable, slow, rate-limited, or returns malformed version data.
- The newest release has no compatible asset for the current operating system and CPU architecture.
- The downloaded update is incomplete, corrupt, or does not match the expected release asset.
- The app is run from a location where replacing the application is not permitted.
- The installed build has no embedded version string or uses a development version.
- The user disables automatic updates while an update check is already in progress.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The app MUST check official release version information during startup when automatic updates are enabled.
- **FR-002**: Automatic update checks MUST be enabled by default for first-time users.
- **FR-003**: Users MUST be able to disable and re-enable automatic update checks from the app menu.
- **FR-004**: The automatic update preference MUST persist across app restarts.
- **FR-005**: A startup update check MUST NOT block Markdown preview loading or normal document interaction.
- **FR-006**: The app MUST only offer or install a newer stable release than the currently running version.
- **FR-007**: The app MUST ignore draft releases, prereleases, and releases without a compatible asset.
- **FR-008**: The app MUST display a clear non-blocking status for checking, up to date, update available, downloading, ready to restart, disabled, and failure states.
- **FR-009**: The app MUST fail safely when checking, downloading, or replacing the app fails, leaving the currently running app usable.
- **FR-010**: The app MUST avoid repeated automatic downloads for the same failed version during a single app session.
- **FR-011**: The app MUST provide a way for users to manually trigger a version check from the menu even when startup checks are disabled.

### Key Entities

- **Update Preference**: The user's persistent choice for whether startup update checks are enabled.
- **Installed Version**: The version identifier for the currently running application.
- **Release Version**: The latest official stable version reported by the release source.
- **Update Status**: The current user-visible state of update checking, downloading, readiness, or failure.
- **Release Asset**: The downloadable package that matches the current operating system and architecture.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: On a normal connection, startup preview loading remains available within 3 seconds even when version checking is in progress.
- **SC-002**: 100% of disabled-startup launches skip automatic release checks while still allowing manual checks.
- **SC-003**: 100% of failed update checks leave the current app session usable and show a non-blocking failure message.
- **SC-004**: In controlled tests, the app correctly distinguishes current, newer, incompatible, disabled, and failed release states.
- **SC-005**: The menu setting persists correctly across at least three consecutive app restarts.

## Assumptions

- The official release source is the project's existing public release channel.
- The first implementation targets stable release assets produced by the current release workflow.
- Restarting the app may be required to complete an installed update.
- The app may need to defer replacement if the running executable cannot be overwritten immediately.
- Network access may be unavailable. That condition is treated as a normal non-blocking failure, not as an application error.
