# Implementation Plan: macOS Full Screen and Quit Shortcuts

**Branch**: `006-mac-fullscreen-quit-shortcut` | **Date**: 2026-08-06 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/006-mac-fullscreen-quit-shortcut/spec.md`

## Summary

Extend the existing frontend keyboard handler so `Option+F11` is recognized as the macOS equivalent of the current `F11` full-screen shortcut, and add an exact `Ctrl+Q` branch that calls the already generated Wails `Quit` runtime function. Keep the existing menu action and full-screen state synchronization unchanged, then update the shortcut documentation and validate with the current Go, frontend, desktop build, and macOS smoke checks.

## Technical Context

**Language/Version**: Go 1.23 module; TypeScript 5.4; React 18

**Primary Dependencies**: Wails v2.12.0 runtime bindings, Vite 3, existing React frontend; no new dependency

**Storage**: N/A. Shortcut handling and window state are transient.

**Testing**: zero-dependency TypeScript shortcut predicate test compiled with the existing TypeScript compiler; `go test ./...`; `npm --prefix frontend run build`; `wails build`; manual macOS window and keyboard smoke test

**Target Platform**: macOS desktop for `Option+F11` and `Ctrl+Q`; existing desktop platforms retain `F11` behavior

**Project Type**: Wails desktop application with React frontend and Go backend

**Performance Goals**: A recognized shortcut starts its window or quit action immediately and completes within one second during manual smoke testing.

**Constraints**: Reuse the existing Wails runtime APIs; keep the project dependency-light; preserve existing shortcut behavior and Markdown preview features; do not modify the rendering pipeline or add platform-specific native code unless verification proves the existing runtime path insufficient.

**Scale/Scope**: One small frontend shortcut helper, one event-handler integration, one generated runtime import use, shortcut/menu documentation updates, and focused validation artifacts. No new backend binding, persistent state, or public service contract.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

The repository constitution is still the unfilled template and defines no active project principles. The repository instructions therefore provide the effective gates:

- **Small and dependency-light**: Reuse existing Wails bindings and add no dependency.
- **Established runtime behavior**: Use the existing native Wails full-screen and quit operations rather than browser-only fullscreen APIs or custom native code.
- **Safe rendering**: Shortcut changes do not execute Markdown content or alter sanitization.
- **Documentation ownership**: Update `README.md` for user-facing shortcut changes and keep implementation notes in `CLAUDE.md` if needed.

Result: PASS. No violation or complexity exception is required.

## Phase 0: Research Summary

Research is recorded in [research.md](research.md). The key decision is to use the existing Wails runtime surface from the React event handler. The pinned Wails source confirms that macOS full screen maps to the native `NSWindow` full-screen transition and that `Quit` closes the application through the platform frontend.

## Phase 1: Design

### Data and state

No persistent data model is introduced. The design uses the existing `fullscreen` React state as the visible representation of the actual native window mode and treats the Wails runtime as the source of truth. See [data-model.md](data-model.md).

### Shortcut handling

1. Add pure shortcut predicates in `frontend/src/keyboard.ts` for `F11` or `Option+F11` and exact `Ctrl+Q`.
2. Prove each predicate with a zero-dependency TypeScript test before wiring the predicates into the UI.
3. Use the full-screen predicate in the existing `toggleFullscreen` branch and the quit predicate before the general `Ctrl` or `Cmd` shortcut dispatch. Prevent the browser default and return after each match.
4. Keep `Ctrl+O`, `Ctrl+S`, `Ctrl+P`, `Ctrl+T`, `Alt+ArrowLeft`, and `Alt+ArrowRight` behavior unchanged.
5. Show both macOS full-screen forms and the quit shortcut in the menu hint and README shortcut table.

### Error handling

The existing `toggleFullscreen` error handling remains the fallback for runtime failures. Quit has no user-facing error path because Wails `Quit` is a synchronous void runtime action. The handler must return after handling each recognized shortcut so one key event cannot trigger another action.

### Interfaces and contracts

No new public interface or project-owned external contract is introduced. The existing generated Wails runtime surface already contains `WindowFullscreen`, `WindowUnfullscreen`, `WindowIsFullscreen`, and `Quit`; the implementation only consumes the existing `Quit` export.

### Validation

The automated gate is the existing Go test suite plus the TypeScript and Vite production build. The desktop packaging gate is `wails build`. The macOS manual gate launches the built app, verifies both full-screen shortcuts from the preview and menu focus states, confirms restoration after exit, and relaunches the app to verify `Ctrl+Q` closes it.

## Project Structure

### Documentation (this feature)

```text
specs/006-mac-fullscreen-quit-shortcut/
├── spec.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── checklists/
│   └── requirements.md
└── tasks.md
```

### Source Code (repository root)

```text
frontend/src/App.tsx    # Keyboard event handling, fullscreen action, menu shortcut hints
frontend/src/keyboard.ts  # Pure keyboard shortcut predicates
frontend/tests/keyboard.test.ts  # Zero-dependency shortcut predicate regression test
README.md               # User-facing shortcut table and feature note
CLAUDE.md               # Developer verification notes if the behavior needs a durable note
main_test.go            # Existing Go regression suite, unchanged unless a backend regression test is useful
frontend/wailsjs/runtime/runtime.d.ts  # Existing Quit declaration, verify only
frontend/wailsjs/runtime/runtime.js    # Existing Quit wrapper, verify only
```

**Structure Decision**: Keep window actions and state in the existing React shell because keyboard events and full-screen state already live in `frontend/src/App.tsx`. Isolate only the modifier matching in a small pure helper so it can be tested without a browser or new test dependency. Reuse the generated Wails runtime wrapper instead of adding a Go binding or a new module. Update only the user documentation required by the shortcut change.

## Complexity Tracking

No constitution violations or additional complexity require justification.
