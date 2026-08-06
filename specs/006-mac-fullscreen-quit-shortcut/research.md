# Research: macOS Full Screen and Quit Shortcuts

## Decision 1: Reuse the existing Wails native window runtime

- **Decision**: Keep `WindowFullscreen`, `WindowUnfullscreen`, and `WindowIsFullscreen` as the full-screen implementation, and use the existing `Quit` runtime wrapper for `Ctrl+Q`.
- **Rationale**: The checked-in Wails runtime declarations and wrappers already expose all four operations. The pinned Wails v2.12.0 macOS backend delegates full-screen transitions to the native `NSWindow` and reports the native full-screen style mask. Its unfullscreen path reapplies the saved window constraints, so it already provides the requested restore behavior.
- **Alternatives considered**:
  - Browser Fullscreen API: rejected because it would fullscreen the webview content instead of the native desktop window and would add browser permission and exit-state concerns.
  - New Go binding or Objective-C bridge: rejected because the existing Wails API already reaches the native window and a second bridge would add unnecessary code.
  - Startup-only full screen: rejected because the user described a keyboard-triggered `Option+F11` action and the existing menu already provides a toggle.

## Decision 2: Treat macOS Option as the existing keyboard event `altKey`

- **Decision**: Recognize `event.altKey && event.key === "F11"` as the macOS shortcut while preserving the existing unmodified `F11` branch.
- **Rationale**: The webview delivers macOS Option as the standard `KeyboardEvent.altKey` modifier. Both paths can call the same `toggleFullscreen` function, keeping state updates and error handling in one place.
- **Alternatives considered**:
  - Detect the platform with a new backend environment binding: rejected because the key event already contains the needed modifier and the behavior is harmless on desktop platforms that expose the same event.
  - Duplicate full-screen logic for Option+F11: rejected because it would create two state and error-handling paths.

## Decision 3: Handle Ctrl+Q before the general modifier dispatcher

- **Decision**: Match `Ctrl+Q` before the existing `Ctrl` or `Cmd` shortcut branch, call `Quit`, prevent the default event, and return.
- **Rationale**: This makes the quit action independent of focused controls and prevents the event from falling through to unrelated keyboard handling. The exact modifier check avoids treating `Cmd+Q` or `Ctrl+Alt+Q` as the requested shortcut.
- **Alternatives considered**:
  - Add a visible Quit menu item only: rejected because it does not satisfy the direct shortcut request.
  - Use `Cmd+Q` instead: rejected because the requested shortcut is explicitly `Ctrl+Q`.

## Evidence inspected

- `frontend/src/App.tsx`: existing `F11` handler, `toggleFullscreen`, fullscreen state synchronization, and menu hint.
- `frontend/wailsjs/runtime/runtime.d.ts` and `runtime.js`: existing declarations and wrapper for `Quit` and full-screen operations.
- Wails v2.12.0 module source under the local Go module cache: macOS `NSWindow` full-screen and quit implementation.
- `README.md` and `CLAUDE.md`: current user-facing shortcut and documentation conventions.
