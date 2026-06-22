# Data Model: Mermaid Rendering Support

This feature does not introduce persistent application data entities. It reuses the existing `PreviewPayload` shape and adds only frontend runtime state.

## Existing Entity: PreviewPayload (unchanged)

- **Source**: `app.go` `PreviewPayload` struct
- **Meaning**: Wire payload from Go backend to React frontend carrying rendered HTML.
- **Field of interest**: `HTML` — contains `<pre><code class="language-mermaid">...</code></pre>` for Mermaid blocks, exactly as goldmark emits them. No schema change required.

## Runtime Entity: Mermaid Diagram Container (frontend-only, transient)

- **Location**: `frontend/src/App.tsx` render tree, inside `previewRef.current`
- **Lifecycle**:
  1. On `contentHtml` change, scan `pre > code.language-mermaid` nodes.
  2. For each node, capture `textContent` (preserving `%%{...}%%` directives), replace the `<pre>` with a `<div class="md-mermaid" data-theme="<theme>">` placeholder.
  3. Call `mermaid.run({ nodes })` against the placeholders.
  4. On theme change, clear rendered SVG and re-run.
  5. On next `contentHtml` change, the previous placeholders are discarded with the innerHTML replacement, so no explicit cleanup is required.
- **Validation**: After render, the placeholder contains an `<svg>` child. On error, it contains a `<div class="md-mermaid-error">` with the failure message.

## Runtime Entity: Mermaid Theme Mapping (frontend-only, transient)

- **Mapping**:
  - `github-light` → Mermaid `default`
  - `github-dark` → Mermaid `dark`
  - `github-sepia` → Mermaid `default` with a sepia-tinted container background via CSS
- **Validation**: The active `theme` state in `App.tsx` is the single source of truth; the mapping is a pure function of that state.

## Asset: Exported HTML Mermaid Runtime

- **Location**: `app.go` `exportHTMLTemplate`
- **Change**: Add a CDN `<script>` tag for Mermaid and an inline initializer that scans `pre > code.language-mermaid` after DOM ready.
- **Validation**: Opening the exported file in a browser renders Mermaid blocks as SVG when online.
