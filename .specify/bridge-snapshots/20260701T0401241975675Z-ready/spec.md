# Feature Specification: Wiki Link Navigation Fix

**Feature Branch**: `005-wiki-link-navigation`

**Created**: 2026-07-01

**Status**: Draft

**Input**: User description: "笔记中 wiki 链接、相关文档的链接目前点击后，有的可以点击，有的是 html 出错。应该是 wiki 链接的形式，点击后能正常跳转。"

## User Scenarios & Testing

### User Story 1 - Click any wiki link and navigate reliably (Priority: P1)

As a user previewing a Markdown note that contains wiki links (for example in a "相关笔记" / related notes section), I want every wiki link to navigate to the target Markdown file when clicked, so that I can move between related notes without leaving the preview window.

**Why this priority**: The request is specifically about fixing broken wiki link navigation; making all wiki links reliably clickable is the smallest independently valuable change.

**Independent Test**: Open a Markdown file containing wiki links with ASCII targets (e.g. `[[README]]`) and non-ASCII targets (e.g. `[[Git基本概念与常用命令]]`). Clicking each link loads the corresponding `.md` file in the preview.

**Acceptance Scenarios**:

1. **Given** a Markdown file with a wiki link `[[README]]`, **When** the user clicks the rendered link, **Then** the preview loads `README.md` from the same directory.
2. **Given** a Markdown file with a wiki link containing Chinese characters such as `[[Git基本概念与常用命令]]`, **When** the user clicks the rendered link, **Then** the preview loads the matching `.md` file instead of showing an error or doing nothing.
3. **Given** a Markdown file with a wiki link containing spaces such as `[[My Note]]`, **When** the user clicks the rendered link, **Then** the preview loads the matching `.md` file.
4. **Given** a Markdown file with a wiki link whose rendered `href` ends with `.html`, **When** the user clicks the rendered link, **Then** the application resolves it to the corresponding `.md` file and navigates there.
5. **Given** a wiki link whose target file does not exist, **When** the user clicks the rendered link, **Then** a clear status message informs the user that the target was not found, and the current preview remains unchanged.
6. **Given** a rendered link that is an external URL (`http://` or `https://`), an email link (`mailto:`), or an in-page anchor (`#...`), **When** the user clicks it, **Then** the link behaves with its default browser-like behavior and is not treated as a wiki link.

### Edge Cases

- Wiki link target with mixed-case extension (`.MD`, `.Markdown`) must be accepted.
- Wiki link target with URL-encoded characters (e.g. `%20`, `%E4%B8%AD`) must be decoded before file lookup.
- Repeated clicks on the same wiki link add the target to the navigation history each time and allow Alt+← / Alt+→ to move through history.
- File watch reload after navigation must continue to watch the newly loaded file.
- Exported standalone HTML is out of scope for wiki link navigation; it may keep the original `href` as a normal link.

## Requirements

### Functional Requirements

- **FR-001**: The backend resolver MUST decode URL-encoded characters in a wiki link `href` before looking for the target file.
- **FR-002**: The backend resolver MUST accept wiki link targets ending with `.html` or `.htm` and resolve them to the corresponding `.md` file.
- **FR-003**: The backend resolver MUST accept wiki link targets without an extension and resolve them to a `.md` file.
- **FR-004**: The backend resolver MUST reject non-Markdown extensions and return an empty result so the frontend can show a "not found" message.
- **FR-005**: The frontend click handler MUST continue to treat external URLs, `mailto:` links, and in-page anchors as non-wiki links.
- **FR-006**: The frontend MUST display a user-friendly status message when a wiki link cannot be resolved, instead of leaving the UI in an error state.
- **FR-007**: Successful wiki link navigation MUST update the preview content, window title, and navigation history consistently with opening a file through the file dialog.
- **FR-008**: The implementation MUST NOT change the Markdown rendering pipeline, theme switching, file watching, footnote rendering, frontmatter display, Mermaid rendering, or auto-update behavior.

### Key Entities

- **Wiki Link**: An anchor element produced by the goldmark wikilink extension from `[[...]]` syntax.
- **Target File**: A local `.md` or `.markdown` file resolved relative to the currently previewed file.
- **Navigation History**: The stack of previously previewed file paths used for Alt+← / Alt+→ navigation.

## Success Criteria

### Measurable Outcomes

- **SC-001**: Clicking a wiki link with Chinese characters opens the target `.md` file in the preview within one second.
- **SC-002**: Clicking a wiki link with spaces opens the target `.md` file in the preview within one second.
- **SC-003**: Clicking a wiki link to a missing file shows a "Wiki link target not found" status message and leaves the current preview unchanged.
- **SC-004**: External URLs, `mailto:` links, and `#` anchors still behave with default browser-like behavior.
- **SC-005**: `go test ./...`, `npm --prefix frontend run build`, and `wails build` all complete successfully after the change.

## Assumptions

- Wiki link targets are located in the same directory as the currently previewed file.
- The goldmark wikilink extension renders `href` values that may be URL-encoded and may use a `.html` suffix.
- The desktop preview is the primary surface where wiki link navigation must work; exported HTML navigation is out of scope.
- File names in the working directory use a UTF-8 compatible filesystem.
