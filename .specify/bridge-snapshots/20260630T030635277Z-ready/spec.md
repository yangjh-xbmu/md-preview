# Feature Specification: Code Block Rendering Regression

**Feature Branch**: `003-codeblock-rendering-regression`

**Created**: 2026-06-30

**Status**: Draft

**Input**: User description: "Rollback the previous code block shadow change, strictly use the Speckit workflow, and complete development so code blocks have no shadow while existing footnote rendering continues to work."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Read code blocks without shadow artifacts (Priority: P1)

As a Markdown reader, I want fenced code blocks to render without text-shadow or glow artifacts so YAML and similar configuration snippets remain crisp in the preview.

**Why this priority**: This is the reported visual defect and the direct user-visible value.

**Independent Test**: Open a Markdown file containing a YAML fenced code block in the desktop preview. The code text and line numbers render without CSS text-shadow in Light, Dark, and Sepia themes.

**Acceptance Scenarios**:

1. **Given** a Markdown file with a `yaml` fenced code block, **When** the preview renders in Light theme, **Then** no code token or line number has a visible shadow effect.
2. **Given** the same file, **When** the theme changes to Dark or Sepia, **Then** the code block remains readable and still has no shadow effect.

---

### User Story 2 - Preserve existing Markdown features (Priority: P1)

As a user of md-preview, I want existing footnote and Mermaid support to keep working after the code style fix so a narrow visual change does not regress document rendering.

**Why this priority**: The previous attempt was suspected to have broken footnotes. Regression protection is part of this fix.

**Independent Test**: Render a Markdown file containing a footnote, a Mermaid block, and a normal code block. Footnotes render with links and endnotes, Mermaid renders as a diagram, and the code block has no shadow.

**Acceptance Scenarios**:

1. **Given** a Markdown document with `[^note]` and `[^note]: content`, **When** the backend loads the file, **Then** the generated HTML includes the footnote reference, backlink, and footnotes container.
2. **Given** a Markdown document with a `mermaid` code block, **When** the frontend preview renders, **Then** the Mermaid block is still excluded from Prism line-number and copy-button code block processing.
3. **Given** a Markdown document with both footnotes and code blocks, **When** export HTML is generated, **Then** exported footnote styles and code block no-shadow styles are both present.

---

### User Story 3 - Verify with current build artifacts (Priority: P2)

As a maintainer, I want validation to use freshly built application artifacts so local preview checks reflect the code under test, not an older binary.

**Why this priority**: A stale `build/bin/md-preview.exe` produced misleading evidence during the failed release attempt.

**Independent Test**: Run the documented verification commands, rebuild the desktop app, then open the README with the freshly built binary and confirm it uses the current code.

**Acceptance Scenarios**:

1. **Given** source changes are complete, **When** `go test ./...`, `npm --prefix frontend run build`, and `wails build` run, **Then** all required checks pass before completion.
2. **Given** the desktop binary has been rebuilt, **When** README preview is launched from `build/bin/md-preview.exe`, **Then** the binary modification time is newer than the implementation commit time.

---

### User Story 4 - Keep copy controls separate from code text (Priority: P1)

As a Markdown reader, I want the code block copy control to sit in its own control area so the button does not touch or cover the first line of code.

**Why this priority**: The no-shadow fix still leaves the code block UI visually cramped. The copy button overlapping text makes code blocks harder to read and interact with.

**Independent Test**: Render a Markdown code block with a long first line. The copy button appears in a separated top control area with reserved vertical space, and the first code line starts below that area.

**Acceptance Scenarios**:

1. **Given** a Markdown code block whose first line extends near the right side, **When** the preview renders, **Then** the copy button does not overlap or touch the code text.
2. **Given** Light, Dark, or Sepia theme, **When** the copy button is shown, **Then** its control area remains visually separated from the code content.

### Edge Cases

- Prism's default theme applies `text-shadow` on `code[class*="language-"]` and nested token spans.
- Prism line-number rows generate separate pseudo-content for numbers.
- Mermaid blocks use `language-mermaid` but are diagrams, not normal code blocks.
- Exported HTML has its own inline style template and must not rely only on bundled React CSS.
- Copy controls are absolutely positioned inside code blocks and need reserved space.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The preview MUST override Prism code block text-shadow for code elements, nested token spans, and line-number text.
- **FR-002**: The preview MUST preserve readable code block backgrounds and borders in Light, Dark, and Sepia themes.
- **FR-003**: The implementation MUST NOT remove or weaken footnote rendering, sanitization, classes, ids, roles, or styles.
- **FR-004**: Mermaid code blocks MUST remain excluded from normal Prism line-number and copy-button decoration.
- **FR-005**: Exported standalone HTML MUST include the same no-shadow code block behavior and existing footnote styles.
- **FR-006**: Verification MUST include backend tests, frontend production build, Wails desktop build, and a fresh README preview launch.
- **FR-007**: Release MUST NOT happen from stale local binaries or without all checks passing.
- **FR-008**: The preview MUST reserve enough top space for code block copy controls so the controls do not overlap code text.
- **FR-009**: The copy control area MUST be visibly separated from code content across Light, Dark, and Sepia themes.

### Key Entities *(include if feature involves data)*

- **Rendered Markdown Document**: The HTML payload generated from a Markdown file, including code blocks, footnotes, and Mermaid blocks.
- **Preview Theme**: Light, Dark, or Sepia visual mode applied by the frontend preview shell.
- **Exported HTML Document**: Standalone HTML generated from the preview pipeline with inline styles and runtime scripts.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Automated tests confirm footnote HTML still includes footnote references, footnotes container, target ids, backlinks, and roles.
- **SC-002**: Automated or scriptable inspection confirms no `text-shadow` remains on rendered Prism code block tokens or line-number elements.
- **SC-003**: `go test ./...`, `npm --prefix frontend run build`, and `wails build` all complete successfully.
- **SC-004**: A freshly rebuilt `build/bin/md-preview.exe` opens `README.md` after the build step.
- **SC-005**: Automated CSS inspection confirms code blocks reserve vertical space for copy controls and define a separate control strip.

## Assumptions

- The user wants the bad `v0.1.1` release rolled back to `v0.1.0` before a corrected release is attempted.
- The fix should be CSS-scoped and minimal unless tests prove a deeper rendering pipeline bug exists.
- Existing Mermaid and footnote behavior from `v0.1.0` is the baseline that must be preserved.
