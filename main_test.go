// INPUT: md-preview package functions and temporary Markdown fixtures.
// OUTPUT: Unit coverage for CLI parsing, rendering, file validation, and HTTP handlers.
// POS: Focused regression tests for the md-preview CLI implementation.
package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseArgsDefaults(t *testing.T) {
	var output bytes.Buffer
	cfg, err := parseArgs([]string{"README.md"}, &output)
	if err != nil {
		t.Fatalf("parseArgs returned error: %v", err)
	}

	if cfg.File != "README.md" {
		t.Fatalf("expected file README.md, got %q", cfg.File)
	}
	if cfg.Host != "127.0.0.1" {
		t.Fatalf("expected default host, got %q", cfg.Host)
	}
	if cfg.Port != defaultPort {
		t.Fatalf("expected default port %d, got %d", defaultPort, cfg.Port)
	}
	if !cfg.Open {
		t.Fatalf("expected browser auto-open by default")
	}
	if !cfg.Watch {
		t.Fatalf("expected watch enabled by default")
	}
}

func TestParseArgsFlags(t *testing.T) {
	var output bytes.Buffer
	cfg, err := parseArgs([]string{"--host", "0.0.0.0", "--port", "0", "--no-open", "--watch=false", "doc.markdown"}, &output)
	if err != nil {
		t.Fatalf("parseArgs returned error: %v", err)
	}

	if cfg.Host != "0.0.0.0" || cfg.Port != 0 || cfg.Open || cfg.Watch {
		t.Fatalf("unexpected config: %+v", cfg)
	}
}

func TestValidateMarkdownFile(t *testing.T) {
	dir := t.TempDir()
	md := filepath.Join(dir, "doc.md")
	if err := os.WriteFile(md, []byte("# ok"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := validateMarkdownFile(md); err != nil {
		t.Fatalf("expected valid markdown file: %v", err)
	}

	if err := validateMarkdownFile(filepath.Join(dir, "missing.md")); err == nil || !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("expected missing file error, got %v", err)
	}
	if err := validateMarkdownFile(dir); err == nil || !strings.Contains(err.Error(), "directory") {
		t.Fatalf("expected directory error, got %v", err)
	}

	txt := filepath.Join(dir, "doc.txt")
	if err := os.WriteFile(txt, []byte("plain"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := validateMarkdownFile(txt); err == nil || !strings.Contains(err.Error(), "unsupported file extension") {
		t.Fatalf("expected unsupported extension error, got %v", err)
	}
}

func TestRenderFileSupportsGFMAndSanitizesHTML(t *testing.T) {
	dir := t.TempDir()
	md := filepath.Join(dir, "doc.md")
	source := strings.Join([]string{
		"# Title",
		"",
		"- [x] done",
		"",
		"| A | B |",
		"|---|---|",
		"| 1 | 2 |",
		"",
		"```go",
		"func main() {}",
		"```",
		"",
		"<script>alert(1)</script>",
	}, "\n")
	if err := os.WriteFile(md, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}

	server := newPreviewServer(config{File: md, Watch: true})
	html, err := server.renderFile()
	if err != nil {
		t.Fatalf("renderFile returned error: %v", err)
	}

	for _, want := range []string{"<h1", "<table>", "checkbox", "language-go", "func main()"} {
		if !strings.Contains(html, want) {
			t.Fatalf("rendered HTML missing %q:\n%s", want, html)
		}
	}
	if strings.Contains(strings.ToLower(html), "<script") {
		t.Fatalf("rendered HTML should not contain script tags:\n%s", html)
	}
}

func TestHTTPHandlers(t *testing.T) {
	dir := t.TempDir()
	md := filepath.Join(dir, "doc.md")
	if err := os.WriteFile(md, []byte("# Served\n\nhello"), 0o644); err != nil {
		t.Fatal(err)
	}

	server := newPreviewServer(config{File: md, Watch: true})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	server.routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d with body %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "<h1") || !strings.Contains(rec.Body.String(), "Served") {
		t.Fatalf("expected rendered markdown in response:\n%s", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "setInterval") {
		t.Fatalf("expected watch polling script in response")
	}

	statusRec := httptest.NewRecorder()
	statusReq := httptest.NewRequest(http.MethodGet, "/status", nil)
	server.routes().ServeHTTP(statusRec, statusReq)
	if statusRec.Code != http.StatusOK {
		t.Fatalf("expected status endpoint 200, got %d", statusRec.Code)
	}

	var payload struct {
		Version string `json:"version"`
		OK      bool   `json:"ok"`
	}
	if err := json.Unmarshal(statusRec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("invalid status JSON: %v", err)
	}
	if !payload.OK || payload.Version == "" {
		t.Fatalf("unexpected status payload: %+v", payload)
	}
}

func TestHTTPHandlerOmitsWatchScriptWhenDisabled(t *testing.T) {
	dir := t.TempDir()
	md := filepath.Join(dir, "doc.md")
	if err := os.WriteFile(md, []byte("# Manual"), 0o644); err != nil {
		t.Fatal(err)
	}

	server := newPreviewServer(config{File: md, Watch: false})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	server.routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "setInterval") {
		t.Fatalf("did not expect watch polling script when watch is disabled")
	}
}
