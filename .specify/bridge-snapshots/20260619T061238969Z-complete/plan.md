# Implementation Plan: Set md-preview App Icon

**Branch**: `001-set-app-icon` | **Date**: 2026-06-19 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/001-set-app-icon/spec.md`

## Summary

Replace md-preview's packaged application icon with the supplied final SVG artwork. The technical approach is to convert the SVG into the existing Wails common app icon source and Windows icon asset path, preserve multi-size `.ico` output, and verify that the project still builds.

## Technical Context

**Language/Version**: Go, TypeScript, Wails v2.11.0

**Primary Dependencies**: Wails desktop packaging, existing frontend build pipeline, local Python image tooling when needed

**Storage**: Filesystem assets only

**Testing**: Icon file inspection, `go test ./...`, `npm --prefix frontend run build`, `wails build`

**Target Platform**: Windows desktop packaging

**Project Type**: Desktop app

**Performance Goals**: No runtime performance impact

**Constraints**: Keep the change limited to icon packaging assets. Do not alter Markdown rendering, CLI behavior, or preview UI behavior.

**Scale/Scope**: One supplied SVG source artwork, one common app icon source, one Windows packaged icon asset, one desktop build verification.

## Constitution Check

The current constitution is still template text and defines no enforceable project-specific gates. Existing repository constraints apply: keep the project small, dependency-light, and do not alter unrelated behavior.

## Project Structure

### Documentation (this feature)

```text
specs/001-set-app-icon/
├── spec.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
└── tasks.md
```

### Source Code (repository root)

```text
build/
├── appicon.png
└── windows/
    └── icon.ico

D:/Users/yangjh/Desktop/Inbox/
└── md-preview-app-icon-final-left-layout.svg
```

**Structure Decision**: Reuse Wails' existing icon paths, `build/appicon.png` and `build/windows/icon.ico`, because Wails documents `appicon.png` as the source used to recreate a missing Windows icon and `icon.ico` as the Windows build icon.

## Complexity Tracking

No constitution violations or added complexity.
