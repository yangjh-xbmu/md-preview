# Research: Set md-preview App Icon

## Decision: Reuse the existing Wails Windows icon location

**Rationale**: The repository already contains `build/windows/icon.ico`, and Wails builds the Windows application using this packaging asset. Replacing that file is the narrowest path that satisfies the feature without changing app code or metadata.

**Alternatives considered**: Adding a new icon path or editing `wails.json` would add configuration churn without evidence that the current Wails convention is insufficient.

## Decision: Generate a multi-size `.ico` from the supplied SVG

**Rationale**: Windows icons need multiple display sizes. The current asset is an `.ico` with multiple embedded sizes, so the replacement should preserve that behavior.

**Alternatives considered**: Copying the SVG into the project alone would not update the packaged Windows icon. A single-size `.ico` could work in some contexts but would degrade display quality.

## Decision: Validate through file inspection and full build

**Rationale**: This feature has no runtime logic. The relevant checks are whether the icon file is valid and whether the Wails build packages successfully.

**Alternatives considered**: Adding automated unit tests would not directly validate packaged icon appearance and would create unnecessary code for an asset-only change.
