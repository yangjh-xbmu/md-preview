package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
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
)

type previewPayload struct {
	FilePath   string `json:"filePath"`
	HTML       string `json:"html"`
	Version    string `json:"version"`
	RenderedAt string `json:"renderedAt"`
	Error      string `json:"error,omitempty"`
}

const watchInterval = time.Second

var checkboxPolicy = regexp.MustCompile(`^checkbox$`)

// App is bound into Wails and provides preview payloads to the frontend.
type App struct {
	ctx    context.Context
	cfg    config
	md     goldmark.Markdown
	policy *bluemonday.Policy

	lastStateKey string
	stateMu      sync.Mutex
	fileMu       sync.RWMutex
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
</style>
</head>
<body>
<article class="markdown-body theme-%s">%s</article>
</body>
</html>`

// NewApp creates a new App application struct.
func NewApp(cfg config) (*App, error) {
	if cfg.File == "" {
		return &App{
			cfg:    cfg,
			md:     newRenderer(),
			policy: markdownPolicy(),
		}, nil
	}

	absPath, err := filepath.Abs(cfg.File)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve file path: %w", err)
	}
	cfg.File = absPath

	return &App{
		cfg:    cfg,
		md:     newRenderer(),
		policy: markdownPolicy(),
	}, nil
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	go a.watchForChanges()
}

// LoadMarkdown loads the target file and returns rendered HTML.
func (a *App) LoadMarkdown() previewPayload {
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
func (a *App) SetFile(path string) previewPayload {
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

// OpenMarkdownFile shows a native file picker and loads the selected Markdown file.
func (a *App) OpenMarkdownFile() previewPayload {
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

func (a *App) emitIfChanged(payload previewPayload) {
	a.stateMu.Lock()
	defer a.stateMu.Unlock()

	current := stateSignature(payload)
	if current == a.lastStateKey {
		return
	}
	a.lastStateKey = current
	runtime.EventsEmit(a.ctx, "markdown-updated", payload)
}

func (a *App) renderMarkdown() previewPayload {
	filePath := a.currentFilePath()
	if filePath == "" {
		return errorPayload("", "No Markdown file path configured. Run with md-preview <file.md>.")
	}

	source, err := os.ReadFile(filePath)
	if err != nil {
		return errorPayload(filePath, fmt.Sprintf("cannot read Markdown file: %v", err))
	}

	var rendered strings.Builder
	if err := a.md.Convert(source, &byteBuffer{builder: &rendered}); err != nil {
		return errorPayload(a.cfg.File, fmt.Sprintf("cannot render Markdown: %v", err))
	}

		version, err := fileVersion(filePath)
	if err != nil {
		return errorPayload(filePath, err.Error())
	}

	return previewPayload{
		FilePath:   filePath,
		HTML:       a.policy.Sanitize(rendered.String()),
		Version:    version,
		RenderedAt: time.Now().Format(time.RFC3339),
	}
}

func (a *App) ExportHTML(path string, theme string) (string, error) {
	payload := a.renderMarkdown()
	if payload.Error != "" {
		return "", fmt.Errorf("cannot export Markdown: %s", payload.Error)
	}

	target, err := resolveExportPath(a.currentFilePath(), path)
	if err != nil {
		return "", err
	}

	body := payload.HTML
	themeClass := sanitizeTheme(theme)
	title := html.EscapeString(filepath.Base(a.currentFilePath()))
	output := fmt.Sprintf(exportHTMLTemplate, title, themeClass, body)

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

func stateSignature(state previewPayload) string {
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

func errorPayload(filePath, message string) previewPayload {
	return previewPayload{
		FilePath: filePath,
		Error:    message,
	}

}

func newRenderer() goldmark.Markdown {
	return goldmark.New(
		goldmark.WithExtensions(extension.GFM),
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
	return policy
}
