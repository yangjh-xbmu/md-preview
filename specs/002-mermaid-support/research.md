# Research: Mermaid Rendering Support

## Decision: Render Mermaid in the React frontend, not in the Go backend

**Rationale**: The project already does client-side syntax highlighting with Prism.js in `frontend/src/App.tsx`. Mermaid follows the same pattern: goldmark emits a `pre > code.language-mermaid` node, the frontend scans for those nodes and replaces them with rendered SVG. This keeps the Go backend dependency-light (no `mermaid-cli` / Node.js toolchain added to the build), preserves the existing bluemonday sanitization boundary, and matches the "established Markdown and HTML sanitization libraries" constraint in `AGENTS.md`.

**Alternatives considered**:
- Server-side rendering with `mermaid-cli` (mmdc) would require a Node.js runtime at build time and per-render invocations, conflicting with the "keep the project small and dependency-light" constraint.
- A goldmark extension that pre-renders Mermaid to SVG in Go would still need to shell out to `mmdc` or embed a JS runtime, adding heavy deps for no gain.

## Decision: Bundle Mermaid via npm, not via CDN, for the desktop preview

**Rationale**: The desktop app runs in a Wails webview with no guaranteed internet access. Bundling `mermaid` through `frontend/package.json` and Vite keeps preview rendering fully offline and reproducible, matching how Prism.js is already bundled.

**Alternatives considered**:
- Loading Mermaid from a CDN at preview time would introduce a network dependency and a startup delay, and would break offline use.

## Decision: Use a public CDN for Mermaid in the exported standalone HTML

**Rationale**: The exported HTML is a standalone document opened in a browser, parallel to GitHub's rendered output. Embedding the full Mermaid library inline would balloon the export file size by hundreds of kilobytes for every export. A CDN `<script>` tag with a `defer` initializer keeps exports small and still renders Mermaid when the file is opened online.

**Alternatives considered**:
- Inlining the Mermaid library into every exported HTML file would roughly double or triple the export size and complicate the `exportHTMLTemplate` with a large embedded blob.
- Skipping Mermaid rendering in exports entirely would violate FR-005.

## Decision: Keep the bluemonday policy unchanged

**Rationale**: The existing `markdownPolicy()` already permits `pre`, `code`, and `class="language-*"` on `code`. Goldmark emits ` ```mermaid ` blocks as `<pre><code class="language-mermaid">...</code></pre>`, which already passes sanitization unchanged. Mermaid's SVG output is generated in the frontend DOM after sanitization, so it never crosses the bluemonday boundary.

**Alternatives considered**:
- Loosening bluemonday to allow SVG elements globally would weaken the sanitization guarantee for non-Mermaid content. Not needed.

## Decision: Re-render Mermaid on theme change

**Rationale**: Mermaid diagrams are SVG with theme-dependent colors (background, node fill, line color). The app already emits a `theme-changed` event from the backend and a `SetTheme` binding from the frontend. Hooking into the existing `theme` state in `App.tsx` to call `mermaid.initialize` with the new theme and re-run on existing blocks is the narrowest change.

**Alternatives considered**:
- Leaving diagrams with the initial theme would produce visible color clashes in dark mode (light SVG on dark background).
- Reloading the file on theme change would discard scroll position and TOC state.

## Decision: Map `github-sepia` to Mermaid's `default` theme

**Rationale**: Mermaid has no native sepia theme. The default palette renders acceptably on the sepia background. CSS will scope a tinted container background for sepia to integrate the diagram visually.

**Alternatives considered**:
- Forcing dark theme for sepia would clash with the warm background.
- Defining a custom Mermaid theme for sepia would add complexity without proportional value.

## Decision: Preserve Mermaid directives and error handling in-page

**Rationale**: Mermaid supports `%%{init: ...}%%` directives at the top of a block. Passing the raw text (including directives) to `mermaid.render` ensures those directives take effect. When syntax is invalid, Mermaid throws; catching the error and replacing the block with a styled error placeholder satisfies FR-004 without breaking sibling blocks.

**Alternatives considered**:
- Stripping directives would silently drop user configuration.
- Surfacing errors through `window.alert` or the status bar would lose the per-block context.
