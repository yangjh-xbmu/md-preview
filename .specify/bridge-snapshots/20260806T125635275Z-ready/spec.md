# Feature Specification: macOS Full Screen and Quit Shortcuts

**Feature Branch**: `006-mac-fullscreen-quit-shortcut`

**Created**: 2026-08-06

**Status**: Draft

**Input**: User description: "F11 可以全屏，在 mac 中是 Option+F11，另外增加 Ctrl+Q 退出快捷键。"

## User Scenarios & Testing

### User Story 1 - Toggle native full screen on macOS (Priority: P1)

As a macOS user previewing a Markdown document, I want both `F11` and `Option+F11` to toggle the app's native full-screen window state, so that I can use the shortcut available on my keyboard without relying on the menu.

**Why this priority**: Full-screen preview is the primary requested macOS interaction improvement. It must work from the existing normal window state and remain reversible.

**Independent Test**: Open the desktop app with a Markdown file, press `Option+F11`, verify that the app enters native full screen, press the same shortcut again, and verify that the previous window state is restored. Repeat with `F11`.

**Acceptance Scenarios**:

1. **Given** the app is in a normal window on macOS, **When** the user presses `Option+F11`, **Then** the app enters native full screen and the interface reports the updated full-screen state.
2. **Given** the app is in native full screen, **When** the user presses `Option+F11`, **Then** the app exits full screen and restores its prior window dimensions and position.
3. **Given** the app is in either normal or full-screen state, **When** the user presses `F11`, **Then** the app toggles the same full-screen state as the menu action.
4. **Given** the full-screen shortcut is pressed while the Markdown preview, menu, or another app control has focus, **When** the key event reaches the app, **Then** the shortcut is handled consistently and does not trigger an unrelated browser or control action.

### User Story 2 - Quit the preview with a keyboard shortcut (Priority: P1)

As a user who has finished previewing a document, I want `Ctrl+Q` to quit md-preview immediately, so that I can close the app without opening the menu or clicking the window close control.

**Why this priority**: A direct quit shortcut is the second explicit request and is useful regardless of which part of the preview currently has focus.

**Independent Test**: Open the desktop app with a Markdown file, focus the preview and then a menu control, press `Ctrl+Q` in each state, and verify that the md-preview window and process close.

**Acceptance Scenarios**:

1. **Given** md-preview is open on macOS, **When** the user presses `Ctrl+Q`, **Then** the application quits and its window closes.
2. **Given** a menu is open or a text/control element has focus, **When** the user presses `Ctrl+Q`, **Then** the application still quits rather than performing another shortcut action.
3. **Given** the user presses another existing shortcut such as `Ctrl+O`, **When** the key event is handled, **Then** the existing action continues to work and the application does not quit.

## Edge Cases

- Pressing `Option+F11` while already in native full screen must exit full screen rather than creating another window state.
- The user may need to hold `Fn` on hardware keyboards that expose the top-row key as a media key; the application receives the resulting `F11` key event in the same way.
- A `Ctrl+Q` event must not be confused with `Cmd+Q`, `Ctrl+Alt+Q`, or another modifier combination.
- If a full-screen runtime operation fails, the preview remains usable and reports a clear failure state instead of becoming unresponsive.
- The Markdown content, file watcher, theme, navigation history, export, print, Mermaid rendering, and update behavior remain unchanged.

## Requirements

### Functional Requirements

- **FR-001**: The application MUST keep `F11` as a full-screen toggle for desktop preview windows.
- **FR-002**: On macOS, the application MUST treat `Option+F11` as an equivalent full-screen toggle.
- **FR-003**: Entering full screen MUST cover the native app window, and exiting full screen MUST restore the window state that existed before entering full screen.
- **FR-004**: The full-screen menu action and keyboard shortcuts MUST keep the visible full-screen state in sync with the actual window state.
- **FR-005**: The application MUST quit when the user presses `Ctrl+Q` while md-preview is focused.
- **FR-006**: `Ctrl+Q` MUST work while the preview, menu, or another app control has focus and MUST prevent the event from triggering an unrelated in-app action.
- **FR-007**: The application MUST continue to distinguish `Ctrl+Q` from other modifier combinations and MUST preserve all existing keyboard shortcuts.
- **FR-008**: The user documentation MUST list the macOS full-screen and quit shortcuts.
- **FR-009**: The implementation MUST NOT change Markdown rendering, sanitization, file watching, theme switching, navigation, export, print, Mermaid rendering, or update behavior.

### Key Entities

- **Window State**: The current normal or native full-screen state of the md-preview window, including the state restored after exiting full screen.
- **Keyboard Shortcut Event**: A key event identified by its key and modifier combination, including `F11`, `Option+F11`, and `Ctrl+Q`.

## Success Criteria

### Measurable Outcomes

- **SC-001**: On macOS, 100% of manual trials using `Option+F11` enter or exit native full screen within one second when the app window is responsive.
- **SC-002**: `F11` continues to toggle full screen in 100% of the existing full-screen smoke trials.
- **SC-003**: `Ctrl+Q` closes the md-preview window and process within one second in 100% of manual trials from the preview and menu focus states.
- **SC-004**: Existing automated Go tests and the frontend production build pass after the shortcut changes.
- **SC-005**: No regression is observed in the Markdown preview, file watching, navigation, theme, export, print, Mermaid, or update flows covered by the existing smoke checks.

## Assumptions

- macOS is the target platform for the new `Option+F11` behavior; existing `F11` behavior remains available on other desktop platforms.
- The macOS webview exposes the Option key as the standard browser `altKey` modifier and exposes the top-row key as `F11` when the user presses the hardware function key combination required by the keyboard.
- The existing Wails window runtime is the source of truth for full-screen and quit operations; no new dependency or persistent preference is needed.
- md-preview does not edit the source Markdown file, so quitting does not require an unsaved-changes confirmation flow.
- The current menu action remains available as a non-keyboard fallback.
