# Data Model: Wiki Link Navigation Fix

This feature does not introduce persistent application data entities. It reuses the existing `PreviewPayload` shape and the existing `ResolveWikiLink` Wails binding.

## Existing Entity: PreviewPayload (unchanged)

- **Source**: `app.go` `PreviewPayload` struct
- **Meaning**: Wire payload from Go backend to React frontend carrying rendered HTML and the current file path.
- **Field of interest**: `FilePath` — used by `ResolveWikiLink` to determine the base directory for relative wiki link targets. No schema change required.

## Existing Entity: ResolveWikiLink Binding (updated behavior)

- **Source**: `app.go` `ResolveWikiLink(href string) string`
- **Meaning**: Resolves a wiki link `href` emitted by goldmark to an absolute local Markdown file path.
- **Lifecycle**:
  1. Trim whitespace from `href`.
  2. Decode percent-encoded sequences with `url.PathUnescape`.
  3. Normalize `.html`/`.htm` suffix to `.md`, or append `.md` if no extension is present.
  4. Join the decoded target with the directory of the currently previewed file.
  5. Validate the candidate with `validateMarkdownFile`.
  6. Return the absolute path, or an empty string if resolution fails.
- **Validation**: New unit tests verify resolution for ASCII, Chinese, space-encoded, `.html`, `.md`, missing, and invalid-extension targets.

## Runtime Entity: Navigation History (unchanged)

- **Source**: `frontend/src/App.tsx`
- **Meaning**: Stack of previously previewed file paths used for Alt+← / Alt+→ navigation.
- **Interaction**: Successful wiki link navigation pushes the resolved path onto the history stack via the existing `applyPayload` flow.
