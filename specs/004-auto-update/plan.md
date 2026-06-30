# Implementation Plan: Auto Update

**Branch**: `004-auto-update` | **Date**: 2026-06-30 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/004-auto-update/spec.md`

## Summary

Add startup update checking and self-update support to md-preview. The backend owns the persistent update preference, version lookup, asset selection, download, checksum verification, and replacement handoff. The frontend exposes the status and menu controls. Startup checks remain asynchronous so Markdown preview is usable even when the network is slow or unavailable.

## Technical Context

**Language/Version**: Go 1.23, TypeScript 5.4, Wails v2.11.0, React 18, Vite 3

**Primary Dependencies**: Go standard library networking, JSON, archive, checksum, filesystem, and process APIs; existing Wails backend bindings and React frontend

**Storage**: User config file under the OS user config directory for update preferences and per-session in-memory status for update attempts

**Testing**: `go test ./...`, focused HTTP test server coverage for update metadata and asset download behavior, TypeScript production build, Wails build, release workflow dry-run review

**Target Platform**: Windows desktop first, using the existing Windows release asset naming; non-Windows builds must fail safely when automatic installation is unsupported or no compatible asset is available

**Project Type**: Desktop app with Go backend and React frontend

**Performance Goals**: Startup preview remains available within 3 seconds while update checking runs in the background; update check timeout is bounded; no Markdown render-time regression

**Constraints**:
- Keep the project small and dependency-light.
- Do not execute arbitrary Markdown-provided script.
- Do not block Markdown preview startup on network access.
- Only consume official stable GitHub release metadata and compatible release assets.
- Preserve current footnote, Mermaid, frontmatter, wiki link, export, print, theme, and code block behavior.

**Scale/Scope**: Backend update service and config helpers, frontend menu/status controls, Wails bindings, release workflow version injection, focused tests, README documentation, and Spec Kit artifacts.

## Constitution Check

The repository constitution file still contains placeholders, so project constraints from `AGENTS.md` and `CLAUDE.md` are authoritative:

- **Small and dependency-light**: PASS. No new project dependency is planned. The update checker uses standard library APIs.
- **Prefer established libraries**: PASS. Existing rendering libraries remain untouched. Release metadata uses the project's current GitHub release channel.
- **No arbitrary user script execution**: PASS. The feature does not alter Markdown rendering or sanitization.
- **Docs and verification discipline**: PASS. README and quickstart are updated; validation includes backend tests, frontend build, Wails build, and workflow review.

No complexity violations identified.

## Project Structure

### Documentation (this feature)

```text
specs/004-auto-update/
├── checklists/
│   └── requirements.md
├── spec.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── update.md
└── tasks.md
```

### Source Code (repository root)

```text
app.go                         # Wails-bound update preference, status, check, download, install, and restart methods
main.go                        # application version variable and startup wiring
main_test.go                   # backend update service and config tests
frontend/src/App.tsx           # menu controls and status display for updates
frontend/src/App.css           # compact menu/status styling for update controls
frontend/wailsjs/go/main/App.* # Wails binding declarations used by frontend build
.github/workflows/release.yml  # inject release tag version into built binaries
README.md                      # user documentation for automatic updates
AGENTS.md                      # Spec Kit plan pointer
```

**Structure Decision**: Keep update behavior in the existing Go backend because it needs filesystem, executable, release asset, and persistent preference access. Keep the visible controls in the existing custom app menu to match current UI conventions. Do not introduce a native Wails menu for this feature because the current app already exposes a cross-platform in-app menu.

## Complexity Tracking

No complexity violations. The only platform-specific behavior is executable replacement. It is isolated behind a small installer step and must fail safely when unsupported.
