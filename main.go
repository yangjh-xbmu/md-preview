// INPUT: Command-line arguments and a local Markdown file.
// OUTPUT: A local HTTP preview server that renders GitHub-style Markdown HTML.
// POS: CLI entry point and preview server implementation for md-preview.
package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	ghtml "github.com/yuin/goldmark/renderer/html"
)

const defaultPort = 17776

var errHelp = errors.New("help requested")

type config struct {
	File  string
	Host  string
	Port  int
	Open  bool
	Watch bool
}

type previewServer struct {
	cfg      config
	markdown goldmark.Markdown
	policy   *bluemonday.Policy
	page     *template.Template
}

type pageData struct {
	Title       string
	FilePath    string
	Content     template.HTML
	Version     string
	Watch       bool
	RenderError string
	GeneratedAt string
}

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		if errors.Is(err, errHelp) {
			return
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	cfg, err := parseArgs(args, stdout)
	if err != nil {
		if errors.Is(err, errHelp) {
			return err
		}
		return err
	}

	if err := validateMarkdownFile(cfg.File); err != nil {
		return err
	}

	listener, err := net.Listen("tcp", net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port)))
	if err != nil {
		return fmt.Errorf("port %d on host %s is not available: %w", cfg.Port, cfg.Host, err)
	}

	srv := newPreviewServer(cfg)
	url := "http://" + listener.Addr().String() + "/"
	fmt.Fprintf(stdout, "Previewing %s at %s\n", cfg.File, url)

	if cfg.Open {
		if err := openBrowser(url); err != nil {
			fmt.Fprintf(stderr, "Could not open browser automatically. Open this URL manually: %s\n", url)
		}
	}

	return http.Serve(listener, srv.routes())
}

func parseArgs(args []string, output io.Writer) (config, error) {
	cfg := config{
		Host:  "127.0.0.1",
		Port:  defaultPort,
		Open:  true,
		Watch: true,
	}

	fs := flag.NewFlagSet("md-preview", flag.ContinueOnError)
	fs.SetOutput(output)
	fs.StringVar(&cfg.Host, "host", cfg.Host, "HTTP bind host")
	fs.IntVar(&cfg.Port, "port", cfg.Port, "HTTP bind port")
	noOpen := fs.Bool("no-open", false, "print URL without opening the browser")
	fs.BoolVar(&cfg.Watch, "watch", cfg.Watch, "reload the preview after file changes")
	fs.Usage = func() {
		fmt.Fprintln(output, "Usage: md-preview [--host 127.0.0.1] [--port 17776] [--no-open] [--watch=false] <file.md>")
	}

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return cfg, errHelp
		}
		return cfg, err
	}
	if fs.NArg() != 1 {
		fs.Usage()
		return cfg, fmt.Errorf("expected exactly one Markdown file")
	}
	if cfg.Port < 0 || cfg.Port > 65535 {
		return cfg, fmt.Errorf("port must be between 0 and 65535")
	}

	cfg.Open = !*noOpen
	cfg.File = fs.Arg(0)
	return cfg, nil
}

func validateMarkdownFile(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("file does not exist: %s", path)
		}
		return fmt.Errorf("cannot read file %s: %w", path, err)
	}
	if info.IsDir() {
		return fmt.Errorf("expected a Markdown file, got directory: %s", path)
	}
	ext := strings.ToLower(filepath.Ext(path))
	if ext != ".md" && ext != ".markdown" {
		return fmt.Errorf("unsupported file extension %q, expected .md or .markdown", ext)
	}
	return nil
}

func newPreviewServer(cfg config) *previewServer {
	return &previewServer{
		cfg: cfg,
		markdown: goldmark.New(
			goldmark.WithExtensions(extension.GFM),
			goldmark.WithParserOptions(parser.WithAutoHeadingID()),
			goldmark.WithRendererOptions(ghtml.WithXHTML()),
		),
		policy: markdownPolicy(),
		page:   template.Must(template.New("page").Parse(pageTemplate)),
	}
}

func markdownPolicy() *bluemonday.Policy {
	policy := bluemonday.UGCPolicy()
	policy.AllowElements("input")
	policy.AllowAttrs("type").Matching(regexp.MustCompile(`^checkbox$`)).OnElements("input")
	policy.AllowAttrs("checked", "disabled").OnElements("input")
	policy.AllowAttrs("class").Matching(regexp.MustCompile(`^language-[A-Za-z0-9_-]+$`)).OnElements("code")
	return policy
}

func (s *previewServer) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/status", s.handleStatus)
	return mux
}

func (s *previewServer) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	content, renderErr := s.renderFile()
	version, versionErr := fileVersion(s.cfg.File)
	if versionErr != nil && renderErr == nil {
		renderErr = versionErr
	}

	status := http.StatusOK
	errText := ""
	if renderErr != nil {
		status = http.StatusInternalServerError
		errText = renderErr.Error()
	}

	abs, _ := filepath.Abs(s.cfg.File)
	data := pageData{
		Title:       filepath.Base(s.cfg.File),
		FilePath:    abs,
		Content:     template.HTML(content),
		Version:     version,
		Watch:       s.cfg.Watch,
		RenderError: errText,
		GeneratedAt: time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if err := s.page.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *previewServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	version, err := fileVersion(s.cfg.File)
	response := struct {
		Version string `json:"version"`
		OK      bool   `json:"ok"`
		Error   string `json:"error,omitempty"`
	}{
		Version: version,
		OK:      err == nil,
	}
	if err != nil {
		response.Error = err.Error()
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	_ = json.NewEncoder(w).Encode(response)
}

func (s *previewServer) renderFile() (string, error) {
	source, err := os.ReadFile(s.cfg.File)
	if err != nil {
		return "", fmt.Errorf("cannot read Markdown file: %w", err)
	}

	var html bytes.Buffer
	if err := s.markdown.Convert(source, &html); err != nil {
		return "", fmt.Errorf("cannot render Markdown: %w", err)
	}
	return s.policy.Sanitize(html.String()), nil
}

func fileVersion(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("cannot stat Markdown file: %w", err)
	}
	return strconv.FormatInt(info.ModTime().UnixNano(), 10) + ":" + strconv.FormatInt(info.Size(), 10), nil
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}

const pageTemplate = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{{.Title}} · md-preview</title>
  <style>
    :root {
      color-scheme: light;
      --canvas: #ffffff;
      --canvas-subtle: #f6f8fa;
      --border: #d0d7de;
      --border-muted: #d8dee4;
      --fg: #24292f;
      --fg-muted: #57606a;
      --accent: #0969da;
      --danger: #cf222e;
      --code-bg: rgba(175, 184, 193, 0.2);
    }

    * {
      box-sizing: border-box;
    }

    body {
      margin: 0;
      color: var(--fg);
      background: var(--canvas);
      font: 16px/1.5 -apple-system, BlinkMacSystemFont, "Segoe UI", "Noto Sans", Helvetica, Arial, sans-serif;
    }

    .topbar {
      position: sticky;
      top: 0;
      z-index: 1;
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 16px;
      min-height: 48px;
      padding: 10px 24px;
      background: rgba(246, 248, 250, 0.92);
      border-bottom: 1px solid var(--border);
      backdrop-filter: blur(8px);
    }

    .file {
      min-width: 0;
      font-size: 14px;
      color: var(--fg-muted);
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    .status {
      flex: 0 0 auto;
      font-size: 12px;
      color: var(--fg-muted);
    }

    .markdown-body {
      max-width: 980px;
      margin: 0 auto;
      padding: 32px;
      overflow-wrap: break-word;
    }

    .markdown-body > *:first-child {
      margin-top: 0 !important;
    }

    .markdown-body > *:last-child {
      margin-bottom: 0 !important;
    }

    .markdown-body a {
      color: var(--accent);
      text-decoration: none;
    }

    .markdown-body a:hover {
      text-decoration: underline;
    }

    .markdown-body h1,
    .markdown-body h2,
    .markdown-body h3,
    .markdown-body h4,
    .markdown-body h5,
    .markdown-body h6 {
      margin-top: 24px;
      margin-bottom: 16px;
      font-weight: 600;
      line-height: 1.25;
    }

    .markdown-body h1,
    .markdown-body h2 {
      padding-bottom: 0.3em;
      border-bottom: 1px solid var(--border-muted);
    }

    .markdown-body h1 {
      font-size: 2em;
    }

    .markdown-body h2 {
      font-size: 1.5em;
    }

    .markdown-body h3 {
      font-size: 1.25em;
    }

    .markdown-body h4 {
      font-size: 1em;
    }

    .markdown-body h5 {
      font-size: 0.875em;
    }

    .markdown-body h6 {
      font-size: 0.85em;
      color: var(--fg-muted);
    }

    .markdown-body p,
    .markdown-body blockquote,
    .markdown-body ul,
    .markdown-body ol,
    .markdown-body dl,
    .markdown-body table,
    .markdown-body pre,
    .markdown-body details {
      margin-top: 0;
      margin-bottom: 16px;
    }

    .markdown-body blockquote {
      padding: 0 1em;
      color: var(--fg-muted);
      border-left: 0.25em solid var(--border);
    }

    .markdown-body ul,
    .markdown-body ol {
      padding-left: 2em;
    }

    .markdown-body li + li {
      margin-top: 0.25em;
    }

    .markdown-body table {
      display: block;
      width: max-content;
      max-width: 100%;
      overflow: auto;
      border-spacing: 0;
      border-collapse: collapse;
    }

    .markdown-body th,
    .markdown-body td {
      padding: 6px 13px;
      border: 1px solid var(--border);
    }

    .markdown-body tr {
      background-color: var(--canvas);
      border-top: 1px solid var(--border-muted);
    }

    .markdown-body tr:nth-child(2n) {
      background-color: var(--canvas-subtle);
    }

    .markdown-body code,
    .markdown-body tt {
      padding: 0.2em 0.4em;
      margin: 0;
      font-size: 85%;
      background-color: var(--code-bg);
      border-radius: 6px;
      font-family: ui-monospace, SFMono-Regular, SF Mono, Consolas, Liberation Mono, Menlo, monospace;
    }

    .markdown-body pre {
      padding: 16px;
      overflow: auto;
      font-size: 85%;
      line-height: 1.45;
      background-color: var(--canvas-subtle);
      border-radius: 6px;
    }

    .markdown-body pre code {
      display: inline;
      padding: 0;
      margin: 0;
      overflow: visible;
      font-size: 100%;
      word-break: normal;
      white-space: pre;
      background: transparent;
      border: 0;
    }

    .markdown-body img {
      max-width: 100%;
      background-color: var(--canvas);
    }

    .markdown-body hr {
      height: 0.25em;
      padding: 0;
      margin: 24px 0;
      background-color: var(--border);
      border: 0;
    }

    .markdown-body input[type="checkbox"] {
      margin: 0 0.2em 0.25em -1.4em;
      vertical-align: middle;
    }

    .error {
      max-width: 980px;
      margin: 24px auto 0;
      padding: 12px 16px;
      color: var(--danger);
      background: #fff8f8;
      border: 1px solid #ffebe9;
      border-radius: 6px;
      font-size: 14px;
    }

    @media (max-width: 720px) {
      .topbar {
        align-items: flex-start;
        flex-direction: column;
        gap: 4px;
        padding: 10px 16px;
      }

      .status {
        white-space: normal;
      }

      .markdown-body {
        padding: 24px 16px;
      }
    }
  </style>
</head>
<body>
  <header class="topbar">
    <div class="file" title="{{.FilePath}}">{{.FilePath}}</div>
    <div class="status">Generated {{.GeneratedAt}}</div>
  </header>
  {{if .RenderError}}<div class="error">{{.RenderError}}</div>{{end}}
  <main class="markdown-body">
    {{.Content}}
  </main>
  {{if .Watch}}
  <script>
    (() => {
      let version = {{printf "%q" .Version}};
      const status = document.querySelector(".status");
      setInterval(async () => {
        try {
          const response = await fetch("/status", { cache: "no-store" });
          const data = await response.json();
          if (data.version && data.version !== version) {
            window.location.reload();
          }
          if (status && data.ok === false && data.error) {
            status.textContent = data.error;
          }
        } catch {
          if (status) {
            status.textContent = "Waiting for preview server";
          }
        }
      }, 1000);
    })();
  </script>
  {{end}}
</body>
</html>
`
