# Feature Specification: Mermaid Rendering Support

**Feature Branch**: `002-mermaid-support`

**Created**: 2026-06-22

**Status**: Draft

**Input**: User request: "增加对 mermaid 语法的渲染支持" with a sample Mermaid diagram illustration.

## User Scenarios & Testing

### User Story 1 - Render Mermaid diagrams inline (Priority: P1)

As a user writing Markdown that contains ` ```mermaid ` fenced code blocks, I want the preview to render those blocks as SVG diagrams (flowcharts, sequence diagrams, class diagrams, state diagrams, Gantt charts, pie charts, etc.) so that I can read the visual diagram instead of the raw Mermaid source.

**Why this priority**: The request is solely about diagram rendering. Replacing raw code blocks with rendered SVG is the smallest independently valuable change.

**Independent Test**: Open a Markdown file containing a ` ```mermaid ` block with a simple flowchart. The preview should show the diagram as SVG, not the Mermaid source code.

**Acceptance Scenarios**:

1. **Given** a Markdown file with a ` ```mermaid ` fenced code block containing valid flowchart syntax, **When** the preview loads, **Then** the block is rendered as an inline SVG diagram inside the preview panel.
2. **Given** a Markdown file with multiple ` ```mermaid ` blocks, **When** the preview loads, **Then** each block renders its own SVG diagram independently.
3. **Given** a Markdown file with a ` ```mermaid ` block followed by a regular ` ```go ` block, **When** the preview loads, **Then** the Mermaid block renders as SVG and the Go block retains Prism syntax highlighting.
4. **Given** the active theme is `github-dark`, **When** a Mermaid block renders, **Then** the diagram uses a dark-friendly palette consistent with the surrounding preview.
5. **Given** the user switches theme from light to dark or sepia, **When** the theme change event fires, **Then** existing Mermaid diagrams re-render with the new theme without requiring a file reload.
6. **Given** a Markdown file with a ` ```mermaid ` block containing invalid syntax, **When** the preview loads, **Then** the block shows a readable error message in place of the diagram, and other blocks on the page still render.
7. **Given** the user exports the preview to a standalone HTML file, **When** the exported file is opened in a browser, **Then** Mermaid blocks render as SVG diagrams without requiring the desktop app.

### Edge Cases

- Mermaid block with empty content: render nothing or a placeholder, do not throw.
- Mermaid block whose first line is a directive (`%%{init: ...}%%`): preserve the directive so Mermaid honors it.
- File watch reload: when the file changes and the preview re-renders, Mermaid blocks must be re-rendered from scratch, never from stale SVG left in the DOM.
- Print / PDF: Mermaid SVG must be visible in print output, not hidden by print CSS.
- Drag-and-drop a new file: Mermaid blocks in the new file must render normally.

## Requirements

### Functional Requirements

- **FR-001**: The preview MUST render any ` ```mermaid ` fenced code block as an inline SVG diagram using the Mermaid library.
- **FR-002**: The Mermaid renderer MUST coexist with Prism.js syntax highlighting so non-Mermaid code blocks retain their existing highlighting and copy button.
- **FR-003**: The Mermaid renderer MUST respond to theme changes (`github-light`, `github-dark`, `github-sepia`) by re-rendering existing diagrams with a theme-appropriate palette.
- **FR-004**: The Mermaid renderer MUST show a readable in-page error message when a block contains invalid syntax, without breaking other blocks on the page.
- **FR-005**: The exported standalone HTML file MUST render Mermaid blocks as SVG when opened in a browser, without depending on the desktop app.
- **FR-006**: The implementation MUST NOT alter the Markdown rendering pipeline, CLI behavior, file watching, wiki link resolution, footnote rendering, or frontmatter display.
- **FR-007**: The implementation MUST NOT execute arbitrary user-provided JavaScript embedded in Mermaid source beyond what the Mermaid library itself interprets as diagram syntax.
- **FR-008**: The Markdown pipeline MUST still sanitize HTML with bluemonday. Mermaid rendering happens in the frontend after sanitization, on `pre > code.language-mermaid` nodes produced by goldmark.

## Success Criteria

### Measurable Outcomes

- **SC-001**: A Markdown file with a ` ```mermaid` flowchart renders an `<svg>` element inside the preview panel.
- **SC-002**: A Markdown file with mixed ` ```mermaid ` and ` ```go ` blocks renders both as SVG diagram and highlighted code respectively.
- **SC-003**: Switching theme re-renders all Mermaid diagrams with the new palette within one second.
- **SC-004**: Exported HTML opened in a standalone browser renders Mermaid blocks as SVG.
- **SC-005**: `go test ./...`, `npm --prefix frontend run build`, and `wails build` all complete successfully after the change.

## Assumptions

- Users author Mermaid using the standard ` ```mermaid ` fenced code block convention recognized by GitHub.
- The desktop app's webview can load the Mermaid library bundled by Vite; no CDN access is required at preview time.
- The exported HTML file may load Mermaid from a public CDN, since it is a standalone document opened in a browser, parallel to how GitHub renders Mermaid on export.
- Mermaid version 10+ is acceptable; the API `mermaid.run({ nodes })` is the supported render entry point.
- Sepia theme has no native Mermaid counterpart; it falls back to the default light palette with CSS-tinted background.
