# Research: Wiki Link Navigation Fix

## Decision: Decode the wiki link `href` with `net/url.PathUnescape`

**Rationale**: The goldmark wikilink extension emits `href` values that are URL-encoded. The current `ResolveWikiLink` only manually decodes `%20`, `%5B`, and `%5D`, so links with Chinese characters or other encoded symbols fail to resolve. Using `url.PathUnescape` from the Go standard library handles all percent-encoded sequences consistently and is safer than maintaining a manual decoding table.

**Alternatives considered**:
- Expanding the manual replacement list to cover common UTF-8 byte sequences would be fragile and incomplete.
- Using `url.QueryUnescape` would incorrectly treat `+` as a space, which is inappropriate for path segments.

## Decision: Keep the resolver behavior for `.html` and extension-less targets

**Rationale**: The existing logic already maps `.html`/`.htm` to `.md` and appends `.md` when no extension is present. This matches the default output of the goldmark wikilink extension and common Obsidian conventions. Only the decoding step is missing.

**Alternatives considered**:
- Changing the wikilink extension to emit `.md` directly would require a custom resolver/renderer and touch more code than necessary.

## Decision: Add unit tests directly in `main_test.go`

**Rationale**: `main_test.go` already contains CLI, validation, rendering, and HTTP tests. Adding a small set of resolver tests there keeps the test suite organized and avoids creating a one-off test file for a single function.

**Alternatives considered**:
- Creating `app_test.go` would be reasonable for a larger refactor, but for one function it adds unnecessary file overhead.

## Decision: Make minimal or no frontend changes

**Rationale**: The frontend already intercepts clicks on non-external anchors and calls `ResolveWikiLink`. The bug is in how the backend decodes the `href`. The only possible frontend touch is to ensure the "not found" status message is user-friendly, which the current code already does.

**Alternatives considered**:
- Moving resolution logic to the frontend would require exposing the filesystem to the webview and would violate the existing security boundary.

## Decision: Do not loosen bluemonday

**Rationale**: The issue is not sanitization; the rendered anchor elements already reach the frontend. The fix happens at resolution time after the user clicks.

**Alternatives considered**:
- Allowing additional `href` schemes in bluemonday is unnecessary and would not fix the resolution bug.
