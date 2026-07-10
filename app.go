// INPUT: CLI config, Markdown files, Wails runtime, goldmark, bluemonday, YAML frontmatter, local image registry.
// OUTPUT: Wails-bound preview application, sanitized rendered HTML, local image assets, export and print actions.
// POS: Desktop app backend and Markdown rendering core for md-preview.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/microcosm-cc/bluemonday"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	gfmhtml "github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"go.abhg.dev/goldmark/wikilink"
	"gopkg.in/yaml.v3"
)

type PreviewPayload struct {
	FilePath    string `json:"filePath"`
	HTML        string `json:"html"`
	Version     string `json:"version"`
	RenderedAt  string `json:"renderedAt"`
	Error       string `json:"error,omitempty"`
	Frontmatter any    `json:"frontmatter,omitempty"`
}

const watchInterval = time.Second

var checkboxPolicy = regexp.MustCompile(`^checkbox$`)

var frontmatterRe = regexp.MustCompile(`^---\r?\n([\s\S]*?)\r?\n---\r?\n`)

var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

var footnoteIDPolicy = regexp.MustCompile(`^fn(ref[0-9]*)?:[^\s"'<>]+$`)

var footnoteClassPolicy = regexp.MustCompile(`^(footnotes|footnote-ref|footnote-backref)$`)

var footnoteRolePolicy = regexp.MustCompile(`^doc-(noteref|endnotes|backlink)$`)

func extractFrontmatter(source []byte) (any, []byte) {
	source = bytes.TrimPrefix(source, utf8BOM)

	matches := frontmatterRe.FindSubmatch(source)
	if len(matches) < 2 {
		return nil, source
	}

	var data map[string]any
	if err := yaml.Unmarshal(matches[1], &data); err != nil {
		return nil, source
	}

	return data, source[len(matches[0]):]
}

// App is bound into Wails and provides preview payloads to the frontend.
type App struct {
	ctx    context.Context
	cfg    config
	md     goldmark.Markdown
	policy *bluemonday.Policy

	lastStateKey string
	stateMu      sync.Mutex
	fileMu       sync.RWMutex
	theme        string
	themeMu      sync.RWMutex
	assetMu      sync.RWMutex
	localAssets  map[string]string

	updateMu             sync.Mutex
	updateStatus         UpdateStatus
	updateSettingsPath   string
	updateStagingDir     string
	updateHTTPClient     *http.Client
	releaseAPIURL        string
	version              string
	osName               string
	archName             string
	failedUpdateVersions map[string]bool
}

const exportHTMLTemplate = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>%s</title>
<style>
  :root {
    color-scheme: light dark;
  }

  * {
    box-sizing: border-box;
  }

  body {
    margin: 0;
    padding: 2rem;
    min-height: 100vh;
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
  }

  .markdown-body {
    max-width: 980px;
    margin: 0 auto;
    padding: 0;
  }

  .markdown-body,
  .markdown-body a,
  .markdown-body code,
  .markdown-body pre {
    color: #24292f;
    background: #fff;
  }

  .markdown-body p,
  .markdown-body ul,
  .markdown-body ol,
  .markdown-body blockquote,
  .markdown-body table {
    margin-top: 0;
    margin-bottom: 1rem;
  }

  .markdown-body a {
    color: #0969da;
    text-decoration: none;
  }

  .markdown-body a:hover {
    text-decoration: underline;
  }

  .markdown-body pre,
  .markdown-body code {
    font-family: ui-monospace, SFMono-Regular, SF Mono, Menlo, Monaco, Consolas, Liberation Mono, Courier New, monospace;
  }

  .markdown-body pre {
    padding: 1rem;
    border-radius: 0.5rem;
    overflow: auto;
    border: 1px solid #d0d7de;
    background: #f6f8fa;
    margin-bottom: 1rem;
  }

  .markdown-body pre[class*="language-"],
  .markdown-body code[class*="language-"],
  .markdown-body pre[class*="language-"] *,
  .markdown-body code[class*="language-"] * {
    text-shadow: none;
  }

  .markdown-body pre[class*="language-"] {
    background: #f6f8fa;
    border: 1px solid #d0d7de;
    box-shadow: none;
  }

  .markdown-body .line-numbers .line-numbers-rows {
    border-right-color: #d0d7de;
    letter-spacing: 0;
  }

  .markdown-body .line-numbers-rows > span:before {
    color: #6e7781;
    text-shadow: none;
  }

  .markdown-body hr {
    border: 0;
    border-top: 1px solid #d8dee4;
    margin: 1.5rem 0;
  }

  .markdown-body.theme-github-dark,
  .markdown-body.theme-github-dark a,
  .markdown-body.theme-github-dark code,
  .markdown-body.theme-github-dark pre {
    color: #c9d1d9;
    background: #0d1117;
  }

  .markdown-body.theme-github-dark {
    color: #c9d1d9;
  }

  .markdown-body.theme-github-dark a {
    color: #58a6ff;
  }

  .markdown-body.theme-github-dark pre {
    border-color: #30363d;
    background: #161b22;
  }

  .markdown-body.theme-github-dark pre[class*="language-"] {
    background: #161b22;
    border-color: #30363d;
  }

  .markdown-body.theme-github-dark .line-numbers .line-numbers-rows {
    border-right-color: #30363d;
  }

  .markdown-body.theme-github-dark .line-numbers-rows > span:before {
    color: #8b949e;
  }

  .markdown-body.theme-github-sepia,
  .markdown-body.theme-github-sepia a,
  .markdown-body.theme-github-sepia code,
  .markdown-body.theme-github-sepia pre {
    color: #5f4b32;
    background: #f7efdd;
  }

  .markdown-body.theme-github-sepia {
    color: #5f4b32;
  }

  .markdown-body.theme-github-sepia a {
    color: #8a5f1f;
  }

  .markdown-body.theme-github-sepia pre {
    border: 1px solid rgba(132, 102, 56, 0.35);
    background: #ead5a7;
  }

  .markdown-body.theme-github-sepia pre[class*="language-"] {
    background: #ead5a7;
    border-color: rgba(132, 102, 56, 0.35);
  }

  .markdown-body.theme-github-sepia .line-numbers .line-numbers-rows {
    border-right-color: rgba(132, 102, 56, 0.35);
  }

  .markdown-body.theme-github-sepia .line-numbers-rows > span:before {
    color: #7a6544;
  }

  .markdown-body table {
    border-collapse: collapse;
    width: 100%%;
    margin-bottom: 1rem;
  }

  .markdown-body th,
  .markdown-body td {
    border: 1px solid #d8dee4;
    padding: 0.5rem;
  }

  .markdown-body .footnotes {
    border-top: 1px solid #d0d7de;
    color: #4b5563;
    font-size: 0.875rem;
    line-height: 1.65;
    margin-top: 2rem;
    padding-top: 1.15rem;
  }

  .markdown-body .footnotes hr {
    display: none;
  }

  .markdown-body .footnotes ol {
    margin: 0;
    padding-left: 1.5rem;
  }

  .markdown-body .footnotes li {
    border-radius: 0.375rem;
    padding: 0.2rem 0.45rem;
  }

  .markdown-body .footnotes li::marker {
    color: #6b7280;
    font-weight: 600;
  }

  .markdown-body .footnotes li p {
    margin: 0;
  }

  .markdown-body .footnotes li:target {
    background: #edf5ff;
    box-shadow: inset 3px 0 0 #0969da;
    color: #1f2937;
    outline: none;
  }

  .markdown-body .footnote-ref,
  .markdown-body .footnote-backref {
    font-weight: 600;
    text-decoration: none;
  }

  .markdown-body.theme-github-dark .footnotes {
    border-top-color: #30363d;
    color: #adbac7;
  }

  .markdown-body.theme-github-dark .footnotes li::marker {
    color: #8b949e;
  }

  .markdown-body.theme-github-dark .footnotes li:target {
    background: #1f2937;
    box-shadow: inset 3px 0 0 #58a6ff;
    color: #dbeafe;
  }

  .markdown-body.theme-github-sepia .footnotes {
    border-top-color: #c4aa70;
    color: #5f4b32;
  }

  .markdown-body.theme-github-sepia .footnotes li::marker {
    color: #8a6f3c;
  }

  .markdown-body.theme-github-sepia .footnotes li:target {
    background: #efe0bd;
    box-shadow: inset 3px 0 0 #8a5f1f;
    color: #3f2f18;
  }

  .md-mermaid {
    margin: 1rem 0;
    text-align: center;
    overflow-x: auto;
    padding: 0.5rem 0;
  }

  .md-mermaid svg {
    max-width: 100%%;
    height: auto;
  }

  .md-mermaid-error {
    border: 1px dashed #d0d7de;
    border-radius: 0.375rem;
    background: #f6f8fa;
    color: #cf222e;
    font-family: ui-monospace, SFMono-Regular, SF Mono, Menlo, Monaco, Consolas, monospace;
    font-size: 0.85rem;
    padding: 0.75rem 1rem;
    text-align: left;
    white-space: pre-wrap;
  }

  .markdown-body.theme-github-dark .md-mermaid-error {
    border-color: #30363d;
    background: #161b22;
    color: #ff7b72;
  }

  .markdown-body.theme-github-sepia .md-mermaid-error {
    border-color: rgba(160, 132, 90, 0.45);
    background: #efe0c8;
    color: #9b2c2c;
  }

  .markdown-body.theme-github-sepia .md-mermaid {
    background: rgba(234, 213, 167, 0.35);
    border-radius: 0.5rem;
  }
</style>
<script defer src="https://cdn.jsdelivr.net/npm/mermaid@11/dist/mermaid.min.js"></script>
<script>
  window.addEventListener('DOMContentLoaded', function () {
    if (typeof window.mermaid === 'undefined') return;
    var blocks = document.querySelectorAll('pre > code.language-mermaid');
    if (!blocks.length) return;
    window.mermaid.initialize({
      startOnLoad: false,
      theme: '%s',
      securityLevel: 'strict'
    });
    blocks.forEach(function (codeBlock, index) {
      var pre = codeBlock.parentElement;
      if (!pre) return;
      var source = codeBlock.textContent || '';
      var container = document.createElement('div');
      container.className = 'md-mermaid';
      container.setAttribute('role', 'img');
      container.textContent = source;
      var id = 'md-mermaid-svg-' + index + '-' + Date.now();
      window.mermaid.render(id, source).then(function (result) {
        container.innerHTML = result.svg;
        pre.replaceWith(container);
      }).catch(function (err) {
        container.className = 'md-mermaid md-mermaid-error';
        container.setAttribute('role', 'alert');
        var msg = (err && (err.message || err.str)) || String(err);
        container.textContent = 'Mermaid render failed: ' + msg;
        pre.replaceWith(container);
      });
    });
  });
</script>
</head>
<body>
<article class="markdown-body theme-%s">%s</article>
</body>
</html>`

// NewApp creates a new App application struct.
func NewApp(cfg config) (*App, error) {
	if cfg.File == "" {
		app := &App{
			cfg:    cfg,
			md:     newRenderer(),
			policy: markdownPolicy(),
			theme:  "github-light",
		}
		app.initUpdateDefaults()
		return app, nil
	}

	absPath, err := filepath.Abs(cfg.File)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve file path: %w", err)
	}
	cfg.File = absPath

	app := &App{
		cfg:    cfg,
		md:     newRenderer(),
		policy: markdownPolicy(),
		theme:  "github-light",
	}
	app.initUpdateDefaults()
	return app, nil
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	go a.watchForChanges()
	a.startUpdateCheckIfEnabled()
}

// LoadMarkdown loads the target file and returns rendered HTML.
func (a *App) LoadMarkdown() PreviewPayload {
	return a.renderMarkdown()
}

// CurrentVersion returns file change fingerprint.
func (a *App) CurrentVersion() string {
	version, err := fileVersion(a.currentFilePath())
	if err != nil {
		return ""
	}
	return version
}

// SetFile updates the target markdown path and reloads the preview.
func (a *App) SetFile(path string) PreviewPayload {
	path = strings.TrimSpace(path)
	if path == "" {
		return errorPayload("", "No file path provided.")
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return errorPayload("", fmt.Sprintf("cannot resolve file path: %v", err))
	}

	if err := validateMarkdownFile(absPath); err != nil {
		return errorPayload("", err.Error())
	}

	a.fileMu.Lock()
	a.cfg.File = absPath
	a.fileMu.Unlock()

	payload := a.renderMarkdown()
	a.emitIfChanged(payload)
	return payload
}

// ResolveWikiLink resolves a wiki link href to an absolute .md file path.
// Returns empty string if the target file does not exist or is not a Markdown file.
func (a *App) ResolveWikiLink(href string) string {
	href = strings.TrimSpace(href)
	if href == "" {
		return ""
	}

	decoded, err := url.PathUnescape(href)
	if err != nil {
		decoded = href
	}

	ext := strings.ToLower(filepath.Ext(decoded))
	switch ext {
	case ".html", ".htm":
		decoded = decoded[:strings.LastIndex(decoded, ".")] + ".md"
	case "", ".md", ".markdown":
		if ext == "" {
			decoded += ".md"
		}
	default:
		return ""
	}

	current := a.currentFilePath()
	if current == "" {
		return ""
	}

	candidate := filepath.Join(filepath.Dir(current), decoded)
	absPath, err := filepath.Abs(candidate)
	if err != nil {
		return ""
	}

	if err := validateMarkdownFile(absPath); err != nil {
		return ""
	}

	return absPath
}

// SetTheme updates the active preview theme and notifies the frontend.
func (a *App) SetTheme(theme string) string {
	next := sanitizeTheme(theme)
	a.themeMu.Lock()
	a.theme = next
	a.themeMu.Unlock()

	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "theme-changed", next)
	}
	return next
}

func (a *App) CurrentTheme() string {
	return a.currentTheme()
}

// OpenMarkdownFile shows a native file picker and loads the selected Markdown file.
func (a *App) OpenMarkdownFile() PreviewPayload {
	if a.ctx == nil {
		return errorPayload("", "Application is not ready yet.")
	}

	current := a.currentFilePath()
	defaultDir := ""
	if current != "" {
		defaultDir = filepath.Dir(current)
	}

	selected, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title:            "Open Markdown File",
		DefaultDirectory: defaultDir,
		Filters: []runtime.FileFilter{
			{DisplayName: "Markdown Files (*.md;*.markdown)", Pattern: "*.md;*.markdown"},
		},
	})
	if err != nil {
		return errorPayload(current, fmt.Sprintf("cannot open file dialog: %v", err))
	}
	if strings.TrimSpace(selected) == "" {
		return a.renderMarkdown()
	}

	return a.SetFile(selected)
}

func (a *App) ExportHTMLWithDialog() (string, error) {
	if a.ctx == nil {
		return "", fmt.Errorf("application is not ready yet")
	}

	current := a.currentFilePath()
	if current == "" {
		return "", fmt.Errorf("open a Markdown file before exporting")
	}

	defaultDir := filepath.Dir(current)
	defaultName := strings.TrimSuffix(filepath.Base(current), filepath.Ext(current)) + "-preview.html"
	selected, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:                "Export HTML",
		DefaultDirectory:     defaultDir,
		DefaultFilename:      defaultName,
		CanCreateDirectories: true,
		Filters: []runtime.FileFilter{
			{DisplayName: "HTML Files (*.html)", Pattern: "*.html"},
		},
	})
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(selected) == "" {
		return "", nil
	}

	saved, err := a.ExportHTML(selected, a.currentTheme())
	if err != nil {
		return "", err
	}
	runtime.EventsEmit(a.ctx, "status-message", "Exported HTML to: "+saved)
	return saved, nil
}

func (a *App) watchForChanges() {
	if !a.cfg.Watch {
		payload := a.renderMarkdown()
		a.emitIfChanged(payload)
		return
	}

	ticker := time.NewTicker(watchInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			payload := a.renderMarkdown()
			a.emitIfChanged(payload)
		case <-a.ctx.Done():
			return
		}
	}
}

func (a *App) emitIfChanged(payload PreviewPayload) {
	a.stateMu.Lock()
	defer a.stateMu.Unlock()

	current := stateSignature(payload)
	if current == a.lastStateKey {
		return
	}
	a.lastStateKey = current
	runtime.EventsEmit(a.ctx, "markdown-updated", payload)
}

func (a *App) renderMarkdown() PreviewPayload {
	return a.renderMarkdownPayload(true)
}

func (a *App) renderMarkdownPayload(rewriteLocalImages bool) PreviewPayload {
	if rewriteLocalImages {
		a.replaceLocalAssets(nil)
	}

	filePath := a.currentFilePath()
	if filePath == "" {
		return errorPayload("", "No Markdown file path configured. Run with md-preview <file.md>.")
	}

	if a.ctx != nil {
		runtime.WindowSetTitle(a.ctx, filepath.Base(filePath))
	}

	source, err := os.ReadFile(filePath)
	if err != nil {
		return errorPayload(filePath, fmt.Sprintf("cannot read Markdown file: %v", err))
	}

	fm, body := extractFrontmatter(source)

	document := a.md.Parser().Parse(text.NewReader(body))
	localAssets := map[string]string(nil)
	if rewriteLocalImages {
		localAssets = rewriteLocalImageDestinations(document, filePath)
	}

	var rendered strings.Builder
	if err := a.md.Renderer().Render(&byteBuffer{builder: &rendered}, body, document); err != nil {
		return errorPayload(a.cfg.File, fmt.Sprintf("cannot render Markdown: %v", err))
	}

	version, err := fileVersion(filePath)
	if err != nil {
		return errorPayload(filePath, err.Error())
	}
	if rewriteLocalImages && a.currentFilePath() == filePath {
		a.replaceLocalAssets(localAssets)
	}

	return PreviewPayload{
		FilePath:    filePath,
		HTML:        a.policy.Sanitize(rendered.String()),
		Version:     version,
		RenderedAt:  time.Now().Format(time.RFC3339),
		Frontmatter: fm,
	}
}

func (a *App) ExportHTML(path string, theme string) (string, error) {
	payload := a.renderMarkdownPayload(false)
	if payload.Error != "" {
		return "", fmt.Errorf("cannot export Markdown: %s", payload.Error)
	}

	target, err := resolveExportPath(a.currentFilePath(), path)
	if err != nil {
		return "", err
	}

	body := payload.HTML
	themeClass := sanitizeTheme(theme)
	mermaidTheme := "default"
	if themeClass == "github-dark" {
		mermaidTheme = "dark"
	}
	title := html.EscapeString(filepath.Base(a.currentFilePath()))
	output := fmt.Sprintf(exportHTMLTemplate, title, mermaidTheme, themeClass, body)

	if err := os.WriteFile(target, []byte(output), 0o644); err != nil {
		return "", fmt.Errorf("failed to write export file: %w", err)
	}
	return target, nil
}

func (a *App) ExportMarkdown(path string) (string, error) {
	return a.ExportHTML(path, "github-light")
}

func (a *App) PrintPreview() {
	runtime.WindowPrint(a.ctx)
}

func resolveExportPath(sourcePath, requestedPath string) (string, error) {
	target := strings.TrimSpace(requestedPath)
	if target == "" {
		dir := filepath.Dir(sourcePath)
		name := filepath.Base(sourcePath)
		target = filepath.Join(dir, strings.TrimSuffix(name, filepath.Ext(name))+"-preview.html")
	}

	target = filepath.Clean(target)
	if filepath.Ext(target) == "" {
		target = target + ".html"
	}

	absPath, err := filepath.Abs(target)
	if err != nil {
		return "", fmt.Errorf("cannot resolve output path: %w", err)
	}

	if info, err := os.Stat(absPath); err == nil && info.IsDir() {
		return "", fmt.Errorf("output path is a directory: %s", absPath)
	}

	parent := filepath.Dir(absPath)
	if parentInfo, err := os.Stat(parent); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("output directory does not exist: %s", parent)
		}
		return "", fmt.Errorf("cannot access output directory %s: %w", parent, err)
	} else if !parentInfo.IsDir() {
		return "", fmt.Errorf("output directory is not a directory: %s", parent)
	}

	return absPath, nil
}

func sanitizeTheme(theme string) string {
	switch strings.ToLower(strings.TrimSpace(theme)) {
	case "github-dark", "github-sepia":
		return theme
	default:
		return "github-light"
	}
}

type byteBuffer struct {
	builder *strings.Builder
}

func (b *byteBuffer) Write(p []byte) (int, error) {
	return b.builder.WriteString(string(p))
}

func stateSignature(state PreviewPayload) string {
	key := struct {
		FilePath string `json:"filePath"`
		HTML     string `json:"html"`
		Version  string `json:"version"`
		Error    string `json:"error,omitempty"`
	}{
		FilePath: state.FilePath,
		HTML:     state.HTML,
		Version:  state.Version,
		Error:    state.Error,
	}
	b, err := json.Marshal(key)
	if err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	return string(b)
}

func (a *App) currentFilePath() string {
	a.fileMu.RLock()
	defer a.fileMu.RUnlock()
	return a.cfg.File
}

func (a *App) currentTheme() string {
	a.themeMu.RLock()
	defer a.themeMu.RUnlock()
	return a.theme
}

func errorPayload(filePath, message string) PreviewPayload {
	return PreviewPayload{
		FilePath: filePath,
		Error:    message,
	}

}

func newRenderer() goldmark.Markdown {
	return goldmark.New(
		goldmark.WithExtensions(extension.GFM, extension.Footnote, &wikilink.Extender{}),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		goldmark.WithRendererOptions(gfmhtml.WithXHTML()),
	)
}

func markdownPolicy() *bluemonday.Policy {
	policy := bluemonday.UGCPolicy()
	policy.AllowElements("input")
	policy.AllowAttrs("type").Matching(checkboxPolicy).OnElements("input")
	policy.AllowAttrs("checked", "disabled").OnElements("input")
	policy.AllowAttrs("class").Matching(regexp.MustCompile(`^language-[A-Za-z0-9_-]+$`)).OnElements("code")
	policy.AllowAttrs("id").Matching(footnoteIDPolicy).OnElements("sup", "li")
	policy.AllowAttrs("class").Matching(footnoteClassPolicy).OnElements("a", "div")
	policy.AllowAttrs("role").Matching(footnoteRolePolicy).OnElements("a", "div")
	return policy
}
