// INPUT: md-preview backend functions, temporary Markdown and image files, HTTP requests, JSON payloads.
// OUTPUT: Unit coverage for argument parsing, rendering, local assets, sanitization, export and state helpers.
// POS: Regression test suite for the md-preview Go backend.
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

func TestLoadMarkdownRendersHeadingWithUTF8BOM(t *testing.T) {
	dir := t.TempDir()
	md := filepath.Join(dir, "doc.md")
	if err := os.WriteFile(md, []byte("\ufeff# Title\n\nBody"), 0o644); err != nil {
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
	if !strings.Contains(payload.HTML, "<h1") || strings.Contains(payload.HTML, "\ufeff# Title") {
		t.Fatalf("expected BOM-prefixed heading to render, got: %s", payload.HTML)
	}
}

func TestLoadMarkdownRendersFootnotes(t *testing.T) {
	dir := t.TempDir()
	md := filepath.Join(dir, "doc.md")
	source := strings.Join([]string{
		"正文引用。[^FN-WEI-WENXIAN-BAIMA-2019]",
		"",
		"[^FN-WEI-WENXIAN-BAIMA-2019]: 这里是脚注内容。",
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
	if strings.Contains(payload.HTML, "[^FN-WEI-WENXIAN-BAIMA-2019]") {
		t.Fatalf("expected footnote marker to be rendered, got: %s", payload.HTML)
	}
	if !strings.Contains(payload.HTML, `class="footnote-ref"`) {
		t.Fatalf("expected footnote reference class, got: %s", payload.HTML)
	}
	if !strings.Contains(payload.HTML, `href="#fn:1"`) {
		t.Fatalf("expected footnote reference link, got: %s", payload.HTML)
	}
	if !strings.Contains(payload.HTML, `id="fn:1"`) {
		t.Fatalf("expected footnote target id, got: %s", payload.HTML)
	}
	if !strings.Contains(payload.HTML, `class="footnotes"`) {
		t.Fatalf("expected footnotes block, got: %s", payload.HTML)
	}
	if !strings.Contains(payload.HTML, "这里是脚注内容") {
		t.Fatalf("expected footnote content, got: %s", payload.HTML)
	}
	if !strings.Contains(payload.HTML, `role="doc-noteref"`) {
		t.Fatalf("expected accessible footnote role, got: %s", payload.HTML)
	}
}

func TestLoadMarkdownRendersWikiLinks(t *testing.T) {
	dir := t.TempDir()
	md := filepath.Join(dir, "doc.md")
	source := strings.Join([]string{
		"See [[Foo Bar]] and [[baz.pdf]] also [[Image|display text]]",
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
	if !strings.Contains(payload.HTML, `href="Foo%20Bar.html"`) {
		t.Fatalf("expected wikilink with space-encoded href, got: %s", payload.HTML)
	}
	if !strings.Contains(payload.HTML, ">Foo Bar<") {
		t.Fatalf("expected wikilink display text, got: %s", payload.HTML)
	}
	if !strings.Contains(payload.HTML, `href="baz.pdf"`) {
		t.Fatalf("expected wikilink with extension preserved, got: %s", payload.HTML)
	}
	if !strings.Contains(payload.HTML, `href="Image.html"`) {
		t.Fatalf("expected wikilink with alias href, got: %s", payload.HTML)
	}
	if !strings.Contains(payload.HTML, ">display text<") {
		t.Fatalf("expected wikilink alias display text, got: %s", payload.HTML)
	}
}

func TestLoadMarkdownServesLocalImagesThroughOpaqueAssets(t *testing.T) {
	dir := t.TempDir()
	imagePath := filepath.Join(dir, "研究 框架图.svg")
	imageBody := `<svg xmlns="http://www.w3.org/2000/svg"><text>framework</text></svg>`
	if err := os.WriteFile(imagePath, []byte(imageBody), 0o644); err != nil {
		t.Fatal(err)
	}

	md := filepath.Join(dir, "doc.md")
	source := "![研究框架图](./研究%20框架图.svg)\n\n![remote](https://example.com/image.png)"
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

	prefix := `src="` + localAssetPrefix
	start := strings.Index(payload.HTML, prefix)
	if start < 0 {
		t.Fatalf("expected local asset URL, got: %s", payload.HTML)
	}
	start += len(`src="`)
	end := strings.Index(payload.HTML[start:], `"`)
	if end < 0 {
		t.Fatalf("expected terminated image source, got: %s", payload.HTML)
	}
	assetURL := payload.HTML[start : start+end]
	if !strings.Contains(payload.HTML, `src="https://example.com/image.png"`) {
		t.Fatalf("expected remote image URL to remain unchanged, got: %s", payload.HTML)
	}

	recorder := httptest.NewRecorder()
	newLocalAssetHandler(app).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, assetURL, nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected local asset response 200, got %d", recorder.Code)
	}
	if recorder.Body.String() != imageBody {
		t.Fatalf("expected SVG body %q, got %q", imageBody, recorder.Body.String())
	}
	if recorder.Header().Get("Content-Type") != "image/svg+xml" {
		t.Fatalf("expected SVG content type, got %q", recorder.Header().Get("Content-Type"))
	}
	if recorder.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatal("expected nosniff header")
	}

	unknown := httptest.NewRecorder()
	newLocalAssetHandler(app).ServeHTTP(unknown, httptest.NewRequest(http.MethodGet, localAssetPrefix+strings.Repeat("0", 64), nil))
	if unknown.Code != http.StatusNotFound {
		t.Fatalf("expected unknown asset to return 404, got %d", unknown.Code)
	}
}

func TestResolveLocalImagePathRejectsUnsupportedSources(t *testing.T) {
	dir := t.TempDir()
	textPath := filepath.Join(dir, "secret.txt")
	if err := os.WriteFile(textPath, []byte("secret"), 0o644); err != nil {
		t.Fatal(err)
	}

	cases := []string{
		"secret.txt",
		"missing.png",
		".",
		"https://example.com/image.png",
		"//example.com/image.png",
		filepath.ToSlash(textPath),
	}
	for _, destination := range cases {
		if path, ok := resolveLocalImagePath(dir, destination); ok {
			t.Fatalf("expected %q to be rejected, got %q", destination, path)
		}
	}
}

func TestLocalImageAssetsExpireWhenDocumentChanges(t *testing.T) {
	dir := t.TempDir()
	firstImage := filepath.Join(dir, "first.png")
	if err := os.WriteFile(firstImage, []byte("first"), 0o644); err != nil {
		t.Fatal(err)
	}
	firstMarkdown := filepath.Join(dir, "first.md")
	if err := os.WriteFile(firstMarkdown, []byte("![first](first.png)"), 0o644); err != nil {
		t.Fatal(err)
	}
	secondMarkdown := filepath.Join(dir, "second.md")
	if err := os.WriteFile(secondMarkdown, []byte("# no images"), 0o644); err != nil {
		t.Fatal(err)
	}

	app, err := NewApp(config{File: firstMarkdown, Watch: false})
	if err != nil {
		t.Fatal(err)
	}
	first := app.LoadMarkdown()
	start := strings.Index(first.HTML, localAssetPrefix)
	if start < 0 {
		t.Fatalf("expected first document asset URL, got: %s", first.HTML)
	}
	end := strings.Index(first.HTML[start:], `"`)
	assetURL := first.HTML[start : start+end]

	app.fileMu.Lock()
	app.cfg.File = secondMarkdown
	app.fileMu.Unlock()
	second := app.LoadMarkdown()
	if second.Error != "" {
		t.Fatalf("expected second document to load, got %q", second.Error)
	}
	recorder := httptest.NewRecorder()
	newLocalAssetHandler(app).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, assetURL, nil))
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected expired asset to return 404, got %d", recorder.Code)
	}
}

func TestExportHTMLKeepsRelativeLocalImagePath(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "figure.png"), []byte("image"), 0o644); err != nil {
		t.Fatal(err)
	}
	md := filepath.Join(dir, "note.md")
	if err := os.WriteFile(md, []byte("![figure](figure.png)"), 0o644); err != nil {
		t.Fatal(err)
	}

	app, err := NewApp(config{File: md, Watch: false})
	if err != nil {
		t.Fatal(err)
	}
	outputPath, err := app.ExportHTML("", "github-light")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(outputPath)

	exported, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(exported), `src="figure.png"`) {
		t.Fatalf("expected relative image path in export, got: %s", exported)
	}
	if strings.Contains(string(exported), localAssetPrefix) {
		t.Fatalf("export must not contain desktop asset URLs, got: %s", exported)
	}
}

func TestResolveWikiLink(t *testing.T) {
	dir := t.TempDir()
	md := filepath.Join(dir, "doc.md")
	if err := os.WriteFile(md, []byte("# doc"), 0o644); err != nil {
		t.Fatal(err)
	}

	files := []string{
		"Foo Bar.md",
		"Git基本概念与常用命令.md",
		"README.md",
		"Note.md",
	}
	for _, name := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("# "+name), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	app, err := NewApp(config{File: md, Watch: false})
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		href       string
		wantSuffix string
	}{
		{"Foo%20Bar.html", "Foo Bar.md"},
		{"Foo Bar.html", "Foo Bar.md"},
		{"Foo%20Bar.md", "Foo Bar.md"},
		{"Git%E5%9F%BA%E6%9C%AC%E6%A6%82%E5%BF%B5%E4%B8%8E%E5%B8%B8%E7%94%A8%E5%91%BD%E4%BB%A4.html", "Git基本概念与常用命令.md"},
		{"README", "README.md"},
		{"README.md", "README.md"},
		{"Note.html", "Note.md"},
	}

	for _, c := range cases {
		resolved := app.ResolveWikiLink(c.href)
		if resolved == "" {
			t.Fatalf("expected resolved path for href %q", c.href)
		}
		if !strings.HasSuffix(resolved, c.wantSuffix) {
			t.Fatalf("href %q: expected path ending with %q, got: %s", c.href, c.wantSuffix, resolved)
		}
	}

	// non-existent target returns empty
	if app.ResolveWikiLink("Missing.html") != "" {
		t.Fatal("expected empty string for missing wiki link target")
	}

	// non-markdown extension returns empty
	if app.ResolveWikiLink("invalid.txt") != "" {
		t.Fatal("expected empty string for non-markdown extension")
	}

	// external URL returns empty
	if app.ResolveWikiLink("https://example.com") != "" {
		t.Fatal("expected empty string for external URL")
	}

	// empty href returns empty
	if app.ResolveWikiLink("") != "" {
		t.Fatal("expected empty string for empty href")
	}
}

func TestExportHTMLIncludesFootnoteStyles(t *testing.T) {
	dir := t.TempDir()
	md := filepath.Join(dir, "note.md")
	content := strings.Join([]string{
		"正文引用。[^note]",
		"",
		"[^note]: 导出的脚注。",
	}, "\n")
	if err := os.WriteFile(md, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	app, err := NewApp(config{File: md, Watch: false})
	if err != nil {
		t.Fatal(err)
	}

	outputPath, err := app.ExportHTML("", "github-light")
	if err != nil {
		t.Fatalf("expected export success: %v", err)
	}
	defer os.Remove(outputPath)

	raw, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("expected exported file: %v", err)
	}

	exported := string(raw)
	if !strings.Contains(exported, ".markdown-body .footnotes") {
		t.Fatalf("expected exported footnote styles, got: %s", exported)
	}
	if !strings.Contains(exported, `class="footnotes"`) {
		t.Fatalf("expected exported footnotes block, got: %s", exported)
	}
}

func TestExportHTMLDisablesPrismCodeBlockTextShadow(t *testing.T) {
	dir := t.TempDir()
	md := filepath.Join(dir, "note.md")
	content := strings.Join([]string{
		"```yaml",
		"proxies:",
		"  - name: usa",
		"```",
	}, "\n")
	if err := os.WriteFile(md, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	app, err := NewApp(config{File: md, Watch: false})
	if err != nil {
		t.Fatal(err)
	}

	outputPath, err := app.ExportHTML("", "github-light")
	if err != nil {
		t.Fatalf("expected export success: %v", err)
	}
	defer os.Remove(outputPath)

	raw, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("expected exported file: %v", err)
	}

	exported := string(raw)
	requiredRules := []string{
		`.markdown-body pre[class*="language-"],`,
		`.markdown-body code[class*="language-"],`,
		`.markdown-body pre[class*="language-"] *,`,
		`.markdown-body code[class*="language-"] *`,
		`text-shadow: none;`,
		`.markdown-body .line-numbers-rows > span:before`,
	}
	for _, rule := range requiredRules {
		if !strings.Contains(exported, rule) {
			t.Fatalf("expected exported HTML to include no-shadow rule %q, got: %s", rule, exported)
		}
	}
}

func TestFrontendCSSDisablesPrismCodeBlockTextShadow(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("frontend", "src", "App.css"))
	if err != nil {
		t.Fatalf("expected App.css to be readable: %v", err)
	}
	css := string(raw)
	requiredRules := []string{
		`.markdown-body pre[class*="language-"],`,
		`.markdown-body code[class*="language-"],`,
		`.markdown-body pre[class*="language-"] *,`,
		`.markdown-body code[class*="language-"] *`,
		`text-shadow: none;`,
		`.markdown-body .line-numbers-rows > span:before`,
		`.markdown-body.theme-github-dark .line-numbers-rows > span:before`,
		`.markdown-body.theme-github-sepia .line-numbers-rows > span:before`,
	}
	for _, rule := range requiredRules {
		if !strings.Contains(css, rule) {
			t.Fatalf("expected frontend CSS to include no-shadow rule %q", rule)
		}
	}
}

func TestFrontendCSSSeparatesCodeCopyControlFromContent(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("frontend", "src", "App.css"))
	if err != nil {
		t.Fatalf("expected App.css to be readable: %v", err)
	}
	css := string(raw)
	requiredRules := []string{
		`.markdown-body pre.md-code-block {`,
		`padding-top: 2.75rem;`,
		`padding-right: 7.25rem;`,
		`.markdown-body pre.md-code-block::before`,
		`height: 2.15rem;`,
		`border-bottom: 1px solid rgba(208, 215, 222, 0.8);`,
		`.md-preview-root.theme-shell-github-dark .markdown-body pre.md-code-block::before`,
		`.md-preview-root.theme-shell-github-sepia .markdown-body pre.md-code-block::before`,
	}
	for _, rule := range requiredRules {
		if !strings.Contains(css, rule) {
			t.Fatalf("expected frontend CSS to separate copy control with rule %q", rule)
		}
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
	if strings.Contains(strings.ToLower(exported), "<script>console.log") {
		t.Fatalf("expected user-provided script to be sanitized")
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
	payload := PreviewPayload{
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
	payload := PreviewPayload{FilePath: "a", HTML: "<h1>a</h1>", Version: "1", RenderedAt: "x"}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("unable to marshal payload: %v", err)
	}
	if len(raw) == 0 {
		t.Fatalf("expected serialized payload")
	}
}

func TestLoadMarkdownPreservesMermaidCodeBlock(t *testing.T) {
	dir := t.TempDir()
	md := filepath.Join(dir, "doc.md")
	source := strings.Join([]string{
		"# Diagram",
		"",
		"```mermaid",
		"flowchart LR",
		"  A --> B",
		"```",
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
	if !strings.Contains(payload.HTML, `class="language-mermaid"`) {
		t.Fatalf("expected mermaid code block to survive sanitization, got: %s", payload.HTML)
	}
	if !strings.Contains(payload.HTML, "flowchart LR") {
		t.Fatalf("expected mermaid source to survive sanitization, got: %s", payload.HTML)
	}
}

func TestExportHTMLIncludesMermaidRuntime(t *testing.T) {
	dir := t.TempDir()
	md := filepath.Join(dir, "note.md")
	content := strings.Join([]string{
		"# Note",
		"",
		"```mermaid",
		"flowchart LR",
		"  A --> B",
		"```",
		"",
		"<script>alert(1)</script>",
	}, "\n")
	if err := os.WriteFile(md, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	app, err := NewApp(config{File: md, Watch: false})
	if err != nil {
		t.Fatal(err)
	}

	outputPath, err := app.ExportHTML("", "github-light")
	if err != nil {
		t.Fatalf("expected export success: %v", err)
	}
	defer os.Remove(outputPath)

	raw, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("expected exported file: %v", err)
	}

	exported := string(raw)
	if !strings.Contains(exported, "cdn.jsdelivr.net/npm/mermaid@") {
		t.Fatalf("expected exported HTML to include mermaid CDN script")
	}
	if !strings.Contains(exported, "mermaid.render(") {
		t.Fatalf("expected exported HTML to include mermaid initializer")
	}
	if !strings.Contains(exported, "theme: 'default'") {
		t.Fatalf("expected exported HTML to initialize mermaid with default theme for light export, got: %s", exported)
	}
	if !strings.Contains(exported, `class="language-mermaid"`) {
		t.Fatalf("expected exported HTML to retain mermaid code block")
	}
	if strings.Contains(strings.ToLower(exported), "<script>alert(1)</script>") {
		t.Fatalf("expected inline script tag to be sanitized in export")
	}
}

func TestExportHTMLMermaidThemeMatchesPreviewTheme(t *testing.T) {
	dir := t.TempDir()
	md := filepath.Join(dir, "note.md")
	if err := os.WriteFile(md, []byte("```mermaid\nflowchart LR\n  A --> B\n```\n"), 0o644); err != nil {
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
	if !strings.Contains(exported, "theme: 'dark'") {
		t.Fatalf("expected exported HTML to initialize mermaid with dark theme for dark export")
	}
	if !strings.Contains(exported, `class="markdown-body theme-github-dark"`) {
		t.Fatalf("expected exported HTML to retain dark preview theme class")
	}
}
