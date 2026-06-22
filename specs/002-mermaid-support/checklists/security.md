# Security Checklist: Mermaid Rendering Support

**Purpose**: Validate that Mermaid rendering does not weaken the existing security posture
**Created**: 2026-06-22
**Feature**: [spec.md](../spec.md)

## Sanitization Boundary

- [x] The Go backend still runs bluemonday on rendered HTML before sending it to the frontend (`renderMarkdown` unchanged)
- [x] The bluemonday policy is unchanged; no new elements or attributes are whitelisted globally (`markdownPolicy()` untouched)
- [x] Mermaid source text crosses the sanitization boundary inside a `pre > code.language-mermaid` node, exactly as goldmark emits it; no bypass path is introduced (verified by `TestLoadMarkdownPreservesMermaidCodeBlock`)
- [x] Mermaid SVG output is generated in the frontend DOM after sanitization and is never re-injected into the Go-side HTML pipeline

## Script Execution

- [x] User-provided `<script>` tags in Markdown continue to be stripped by bluemonday before reaching the frontend (verified by `TestExportHTMLWritesFileWithThemeAndSanitization`)
- [x] Mermaid interprets its own DSL; no arbitrary JavaScript from the Markdown body is executed as a side effect of Mermaid rendering
- [x] Mermaid directives (`%%{init: ...}%%`) are passed to the library as diagram source, not evaluated as JS by md-preview code (`renderMermaidBlocks` reads `textContent` and hands it to `mermaid.render`)
- [x] The exported HTML initializer script only reads `textContent` of `language-mermaid` code blocks and hands it to `mermaid.render`; it does not `eval` user content
- [x] Mermaid `securityLevel: 'strict'` is set in both the frontend helper and the exported HTML initializer, blocking HTML labels and event bindings in diagrams

## Export Surface

- [x] The Mermaid CDN script in exported HTML is loaded over HTTPS from `cdn.jsdelivr.net/npm/mermaid@11`
- [x] The exported HTML initializer uses `textContent` (not `innerHTML`) when extracting Mermaid source from code blocks
- [x] The exported HTML initializer does not introduce a new XSS vector: Mermaid's own output is scoped to the placeholder container

## Dependency Supply Chain

- [x] The `mermaid` npm package is pinned to a major version in `package.json` (`"mermaid": "^11.15.0"`)
- [x] `npm --prefix frontend install` completes without integrity errors
- [x] No new Go dependencies are added to `go.mod`

## Notes

Mermaid has had historical clickjacking / XSS advisories on user-controlled diagrams; this is acceptable for a local single-user preview tool. The sanitization boundary (bluemonday still strips `<script>` etc. before Mermaid sees anything) remains intact, and Mermaid's own `securityLevel: 'strict'` blocks HTML labels and event bindings inside diagrams.
