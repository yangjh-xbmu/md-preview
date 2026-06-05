package main

import (
	"bytes"
	"encoding/json"
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
	if !cfg.Watch {
		t.Fatal("expected watch enabled by default")
	}
}

func TestParseArgsFlags(t *testing.T) {
	var output bytes.Buffer
	cfg, err := parseArgs([]string{"--watch=false", "doc.markdown"}, &output)
	if err != nil {
		t.Fatalf("parseArgs returned error: %v", err)
	}
	if cfg.Watch {
		t.Fatalf("expected watch disabled")
	}
}

func TestParseArgsWithoutFileReturnsNoErrorForTooling(t *testing.T) {
	var output bytes.Buffer
	cfg, err := parseArgs([]string{}, &output)
	if err != nil {
		t.Fatalf("parseArgs with no file returned error: %v", err)
	}
	if cfg.File != "" {
		t.Fatalf("expected empty file for no-arg mode, got %q", cfg.File)
	}
}

func TestValidateMarkdownFile(t *testing.T) {
	dir := t.TempDir()
	md := filepath.Join(dir, "doc.md")
	if err := os.WriteFile(md, []byte("# hello"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := validateMarkdownFile(md); err != nil {
		t.Fatalf("expected valid file: %v", err)
	}

	if err := validateMarkdownFile(filepath.Join(dir, "missing.md")); err == nil {
		t.Fatalf("expected missing file error")
	}

	if err := validateMarkdownFile(dir); err == nil {
		t.Fatalf("expected directory error")
	}

	txt := filepath.Join(dir, "doc.txt")
	if err := os.WriteFile(txt, []byte("plain"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := validateMarkdownFile(txt); err == nil {
		t.Fatalf("expected unsupported extension error")
	}
}

func TestLoadMarkdownRendersAndSanitizes(t *testing.T) {
	dir := t.TempDir()
	md := filepath.Join(dir, "doc.md")
	source := strings.Join([]string{
		"# Title",
		"- [x] done",
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

	app, err := NewApp(config{File: md, Watch: false})
	if err != nil {
		t.Fatal(err)
	}
	payload := app.LoadMarkdown()
	if payload.Error != "" {
		t.Fatalf("expected success, got error %q", payload.Error)
	}
	if !strings.Contains(payload.HTML, "<h1") {
		t.Fatalf("expected rendered heading, got: %s", payload.HTML)
	}
	if !strings.Contains(payload.HTML, "language-go") {
		t.Fatalf("expected language class, got: %s", payload.HTML)
	}
	if strings.Contains(strings.ToLower(payload.HTML), "<script") {
		t.Fatalf("script tag was not sanitized: %s", payload.HTML)
	}
	if !strings.Contains(payload.FilePath, md) {
		t.Fatalf("expected file path in payload, got: %s", payload.FilePath)
	}
}

func TestCurrentVersion(t *testing.T) {
	dir := t.TempDir()
	md := filepath.Join(dir, "doc.md")
	if err := os.WriteFile(md, []byte("v1"), 0o644); err != nil {
		t.Fatal(err)
	}

	app, err := NewApp(config{File: md})
	if err != nil {
		t.Fatal(err)
	}
	current := app.CurrentVersion()
	if current == "" {
		t.Fatalf("expected non-empty version")
	}
}

func TestExportHTMLWritesFileWithThemeAndSanitization(t *testing.T) {
	dir := t.TempDir()
	md := filepath.Join(dir, "note.md")
	content := strings.Join([]string{
		"# Note",
		"",
		"```go",
		"fmt.Println(\"x\")",
		"```",
		"",
		"<script>console.log(\"x\")</script>",
	}, "\n")
	if err := os.WriteFile(md, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	app, err := NewApp(config{File: md, Watch: false})
	if err != nil {
		t.Fatal(err)
	}

	outputPath, err := app.ExportHTML("", "github-dark")
	if err != nil {
		t.Fatalf("expected export success: %v", err)
	}
	defer os.Remove(outputPath)

	raw, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("expected exported file: %v", err)
	}

	exported := string(raw)
	if !strings.Contains(exported, `class="markdown-body theme-github-dark"`) {
		t.Fatalf("expected exported theme class, got: %s", exported)
	}
	if !strings.Contains(exported, "<h1") {
		t.Fatalf("expected rendered heading in export")
	}
	if strings.Contains(strings.ToLower(exported), "<script") {
		t.Fatalf("expected script to be sanitized")
	}

	expected, err := filepath.Abs(filepath.Join(dir, "note-preview.html"))
	if err != nil {
		t.Fatal(err)
	}
	if outputPath != expected {
		t.Fatalf("expected default export path %q, got %q", expected, outputPath)
	}
}

func TestResolveExportPathRejectsBadDirectory(t *testing.T) {
	dir := t.TempDir()
	md := filepath.Join(dir, "note.md")
	if err := os.WriteFile(md, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	badDir := filepath.Join(dir, "missing", "path.html")
	if _, err := resolveExportPath(md, badDir); err == nil {
		t.Fatalf("expected error for missing output directory")
	}
}

func TestStateSignatureIsDeterministic(t *testing.T) {
	payload := previewPayload{
		FilePath:   "/tmp/a.md",
		HTML:       "<h1>a</h1>",
		Version:    "123",
		RenderedAt: "2026-06-05",
	}
	a := stateSignature(payload)
	b := stateSignature(payload)
	if a != b {
		t.Fatalf("state signature changed unexpectedly: %s != %s", a, b)
	}
}

func TestConfigCanMarshalForFrontend(t *testing.T) {
	payload := previewPayload{FilePath: "a", HTML: "<h1>a</h1>", Version: "1", RenderedAt: "x"}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("unable to marshal payload: %v", err)
	}
	if len(raw) == 0 {
		t.Fatalf("expected serialized payload")
	}
}
