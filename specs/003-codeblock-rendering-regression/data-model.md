# Data Model: Code Block Rendering Regression

## Rendered Markdown Document

Represents the sanitized HTML produced from a Markdown file.

**Relevant fields**:
- `html`: sanitized HTML payload
- `theme`: preview theme applied by the frontend or export template
- `containsCodeBlock`: whether the document includes normal Prism-highlighted code
- `containsFootnotes`: whether the document includes footnote references and endnotes
- `containsMermaid`: whether the document includes Mermaid source blocks

**Validation rules**:
- Footnote ids, classes, roles, and links remain allowed by sanitization.
- Mermaid source code blocks remain present for frontend transformation.
- Normal code blocks retain `language-*` classes for Prism highlighting.

## Preview Theme

Represents the active visual mode.

**States**:
- `github-light`
- `github-dark`
- `github-sepia`

**Validation rules**:
- Each theme provides readable code block background and border colors.
- Each theme suppresses Prism text-shadow and line-number text-shadow.

## Exported HTML Document

Represents the standalone HTML file generated from the rendered Markdown.

**Relevant fields**:
- `inlineStyles`: CSS embedded in the export template
- `contentHtml`: sanitized Markdown content
- `mermaidRuntime`: script used only for Mermaid DSL rendering

**Validation rules**:
- Inline styles include footnote presentation rules.
- Inline styles include code block no-shadow rules.
- User-provided script content remains sanitized.
