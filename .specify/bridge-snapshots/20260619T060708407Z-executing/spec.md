# Feature Specification: Set md-preview App Icon

**Feature Branch**: `001-set-app-icon`

**Created**: 2026-06-19

**Status**: Draft

**Input**: User description: "Use `D:\Users\yangjh\Desktop\Inbox\md-preview-app-icon-final-left-layout.svg` as the md-preview app icon. All development must go through the speckit bridge workflow."

## User Scenarios & Testing

### User Story 1 - Recognizable desktop app icon (Priority: P1)

As a user launching or locating md-preview on Windows, I want the desktop application to show the supplied md-preview artwork as its icon, so that the packaged app is visually identifiable and no longer uses the previous icon.

**Why this priority**: The request is solely about application identity. Replacing the packaged icon is the smallest independently valuable change.

**Independent Test**: Build the desktop app and inspect the packaged Windows icon asset. The generated application should use the supplied artwork as its icon source.

**Acceptance Scenarios**:

1. **Given** the supplied SVG artwork exists, **When** the app icon asset is regenerated, **Then** the Windows icon asset represents that supplied artwork.
2. **Given** the Windows icon asset has been regenerated, **When** the app is built, **Then** the Wails build completes and packages the updated icon asset.

### Edge Cases

- If the supplied SVG is missing or unreadable, implementation must stop before replacing the current icon.
- If the icon conversion tool cannot produce a valid multi-size Windows icon, implementation must stop before marking the feature complete.

## Requirements

### Functional Requirements

- **FR-001**: The desktop app MUST use the supplied md-preview SVG artwork as the source for its packaged application icon.
- **FR-002**: The Windows icon asset MUST remain a valid `.ico` file with multiple sizes suitable for desktop display.
- **FR-003**: The app build MUST continue to complete successfully after the icon is replaced.
- **FR-004**: The change MUST NOT alter Markdown rendering behavior, command-line behavior, preview UI behavior, or application metadata unrelated to the icon.

## Success Criteria

### Measurable Outcomes

- **SC-001**: The Windows icon asset is recognized as a valid icon file after replacement.
- **SC-002**: The desktop application build completes successfully after the replacement.
- **SC-003**: The implementation touches only icon-related assets and required Spec Kit workflow artifacts, aside from pre-existing unrelated working tree changes.

## Assumptions

- The supplied SVG at `D:\Users\yangjh\Desktop\Inbox\md-preview-app-icon-final-left-layout.svg` is the final intended artwork.
- Windows packaging uses the existing Wails icon asset location already present in the repository.
- Linux and macOS icon packaging are out of scope unless already driven by the same existing source asset.
