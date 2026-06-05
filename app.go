package main

import (
	"context"
	"encoding/json"
	"fmt"
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
	ghtml "github.com/yuin/goldmark/renderer/html"
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
}

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
	version, err := fileVersion(a.cfg.File)
	if err != nil {
		return ""
	}
	return version
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
	if a.cfg.File == "" {
		return errorPayload("", "No Markdown file path configured. Run with md-preview <file.md>.")
	}

	source, err := os.ReadFile(a.cfg.File)
	if err != nil {
		return errorPayload(a.cfg.File, fmt.Sprintf("cannot read Markdown file: %v", err))
	}

	var rendered strings.Builder
	if err := a.md.Convert(source, &byteBuffer{builder: &rendered}); err != nil {
		return errorPayload(a.cfg.File, fmt.Sprintf("cannot render Markdown: %v", err))
	}

	version, err := fileVersion(a.cfg.File)
	if err != nil {
		return errorPayload(a.cfg.File, err.Error())
	}

	return previewPayload{
		FilePath:   a.cfg.File,
		HTML:       a.policy.Sanitize(rendered.String()),
		Version:    version,
		RenderedAt: time.Now().Format(time.RFC3339),
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
		goldmark.WithRendererOptions(ghtml.WithXHTML()),
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
