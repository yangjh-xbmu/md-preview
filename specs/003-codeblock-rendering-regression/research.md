# Research: Code Block Rendering Regression

## Decision: Override Prism `text-shadow` at the Markdown preview boundary

**Rationale**: Prism's default theme applies `text-shadow` to `code[class*="language-"]` and `pre[class*="language-"]`. The cleanest fix is to override that behavior inside `.markdown-body` so the change affects rendered Markdown code blocks without changing global UI controls.

**Alternatives considered**:
- Replace Prism theme: too broad for a visual regression fix.
- Remove Prism CSS import: breaks syntax highlighting.
- Modify generated HTML classes: unnecessary and risks backend regressions.

## Decision: Match frontend CSS in exported HTML template

**Rationale**: Exported standalone HTML does not load the React app stylesheet. If only `frontend/src/App.css` changes, exported files can still show the Prism shadow. The export template must include matching no-shadow rules.

**Alternatives considered**:
- Leave export unchanged: fails feature parity.
- Inline every frontend style rule: too broad and already avoided by the existing template design.

## Decision: Preserve footnote tests and add style-specific regression checks

**Rationale**: Footnote rendering lives in the Go pipeline and sanitization policy. The reported failure was caused by previewing with a stale local binary, but tests should still guard the feature because the user experienced it as a regression.

**Alternatives considered**:
- Rely only on manual preview: insufficient after the stale binary issue.
- Add browser automation: useful but heavier than needed for a CSS/template regression.

## Decision: Fresh Wails build before visual preview

**Rationale**: `build/bin/md-preview.exe` was older than the current source and misrepresented feature support. The quickstart requires `wails build` before launching README preview.

**Alternatives considered**:
- Use `go run . README.md`: validates source behavior but not the packaged binary.
- Use an existing release asset: not suitable for validating local implementation.
