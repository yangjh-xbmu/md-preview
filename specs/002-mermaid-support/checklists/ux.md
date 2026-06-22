# UX Checklist: Mermaid Rendering Support

**Purpose**: Validate user experience quality of the Mermaid rendering feature
**Created**: 2026-06-22
**Feature**: [spec.md](../spec.md)

## Visual Integration

- [x] Rendered Mermaid SVG is centered and consistent with the surrounding Markdown body width (`.md-mermaid { text-align: center; margin: 1rem 0; }`)
- [x] Mermaid SVG does not overflow the preview panel horizontally; large diagrams scroll (`overflow-x: auto`) and scale (`svg { max-width: 100% }`)
- [x] Sepia theme applies a tinted background to the Mermaid container (`.markdown-body.theme-github-sepia .md-mermaid { background: rgba(234, 213, 167, 0.35); }`)
- [x] Dark theme diagram palette matches the dark preview background (`mermaidThemeFor('github-dark') → 'dark'`)

## Interaction

- [x] Mermaid SVG is selectable and copyable like other preview content (auto-copy `mouseup` handler is on `previewRef`, not scoped to non-Mermaid only)
- [x] Mermaid blocks do not show the Prism "Copy" button (Prism effect skips `language-mermaid`)
- [x] Mermaid blocks do not show Prism line numbers (same skip; `pre` is replaced before Prism runs)
- [ ] Switching theme re-renders all visible Mermaid diagrams within 1 second (manual smoke test)
- [ ] File watch reload re-renders Mermaid blocks from scratch; no stale SVG remains (manual smoke test)

## Error States

- [x] Invalid Mermaid syntax shows a readable error message in place of the diagram (per-block `try/catch` in `renderMermaidBlocks`)
- [x] The error message stays inside the diagram's container, not in the global status bar (`.md-mermaid-error` is the container itself)
- [x] One bad Mermaid block does not prevent other valid blocks on the same page from rendering (`Promise.all` over per-block `try/catch`)
- [x] Empty Mermaid block does not throw or crash the preview (early return with `.md-mermaid-empty` placeholder)

## Print / Export

- [x] Mermaid SVG is visible in print output (no `@media print` rule hides `.md-mermaid`)
- [x] Exported standalone HTML renders Mermaid blocks when opened in a browser (CDN script + initializer in `exportHTMLTemplate`)
- [ ] Exported HTML degrades gracefully if offline (Mermaid CDN unreachable): raw code remains visible, not a blank hole (manual smoke test)

## Accessibility

- [x] Mermaid container has `role="img"` so screen readers announce it as a figure (`container.setAttribute("role", "img")`)
- [x] Error placeholder has `role="alert"` so screen readers announce the failure (set in the `catch` branch)

## Notes

Mermaid's own SVG also emits `role="img"` and `aria-roledescription`; our container-level role is a stable outer wrapper that survives re-render.
