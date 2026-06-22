# Testing Checklist: Mermaid Rendering Support

**Purpose**: Validate test coverage and verification approach
**Created**: 2026-06-22
**Feature**: [spec.md](../spec.md)

## Automated Coverage

- [x] Existing Go tests still pass (`go test ./...` → `ok github.com/yangjh-xbmu/md-preview`)
- [x] New Go test asserts exported HTML contains the Mermaid initializer script tag (`TestExportHTMLIncludesMermaidRuntime`)
- [x] New Go test asserts a `language-mermaid` code block survives sanitization and appears in the exported HTML (`TestExportHTMLIncludesMermaidRuntime`, `TestLoadMarkdownPreservesMermaidCodeBlock`)
- [x] New Go test asserts exported HTML still sanitizes user-provided `<script>` tags (updated `TestExportHTMLWritesFileWithThemeAndSanitization` to scope the check to user content)
- [x] New Go test asserts exported HTML picks the correct Mermaid theme per preview theme (`TestExportHTMLMermaidThemeMatchesPreviewTheme`)
- [x] `npm --prefix frontend run build` completes without TypeScript errors (TypeScript upgraded to 5.4; `mermaid-shim.d.ts` added)
- [x] `wails build` completes successfully (`build/bin/md-preview.exe` rebuilt)

## Manual Smoke Test (per quickstart.md)

- [ ] Mermaid flowchart renders as inline SVG in the preview
- [ ] Multiple Mermaid blocks on the same page each render independently
- [ ] Mixed Mermaid + Go blocks: Mermaid renders as SVG, Go block retains Prism highlight + copy button
- [ ] Theme switch (light → dark → sepia) re-renders Mermaid with appropriate palette
- [ ] Invalid Mermaid syntax shows an in-page error placeholder, other blocks still render
- [ ] File watch reload re-renders Mermaid from scratch
- [ ] Drag-and-drop a new file with Mermaid blocks renders correctly
- [ ] Exported HTML opened in a browser renders Mermaid blocks as SVG

## Regression Guards

- [x] Footnote rendering unchanged (existing footnote test still passes)
- [x] Wiki link resolution unchanged (existing wikilink tests still pass)
- [x] Frontmatter table still renders (no changes to `FrontmatterTable.tsx`)
- [x] Code block copy button still works on non-Mermaid blocks (Prism effect only skips `language-mermaid`)
- [x] Print / PDF still hides chrome and shows Mermaid SVG (no print rule added that would hide `.md-mermaid`)
- [x] CLI argument parsing unchanged (`parseArgs` untouched)

## Test Data

- [ ] A representative Mermaid sample (flowchart, sequence, class, state, gantt, pie) is used during smoke testing
- [ ] A Mermaid block with a `%%{init: ...}%%` directive is used to verify directives are honored
- [ ] A Mermaid block with invalid syntax is used to verify error handling

## Notes

Frontend rendering is not covered by automated tests (no Jest / Vitest setup in the project). Smoke testing covers the frontend behavior. Go-side tests cover the export template and sanitization guarantees.
