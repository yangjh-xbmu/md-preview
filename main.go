package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

var errHelp = errors.New("help requested")

type config struct {
	File  string
	Watch bool
	// Compatibility flag from the original implementation.
	// It is ignored in Wails mode, which is now the default desktop path.
	UseBrowser bool
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

	if cfg.File != "" {
		if err := validateMarkdownFile(cfg.File); err != nil {
			fmt.Fprintln(stderr, err)
			cfg.File = ""
		}
	}

	if cfg.UseBrowser {
		fmt.Fprintln(stderr, "warning: --browser is retained for compatibility and will run desktop mode")
	}

	app, err := NewApp(cfg)
	if err != nil {
		return err
	}
	return runDesktopApp(app)
}

func parseArgs(args []string, output io.Writer) (config, error) {
	cfg := config{
		Watch: true,
	}

	fs := flag.NewFlagSet("md-preview", flag.ContinueOnError)
	fs.SetOutput(output)
	fs.BoolVar(&cfg.Watch, "watch", cfg.Watch, "watch file changes")
	fs.BoolVar(&cfg.UseBrowser, "browser", false, "kept for compatibility, still launches desktop app")
	fs.Usage = func() {
		_, _ = fmt.Fprintln(output, "Usage: md-preview [--browser] [--watch=false] <file.md>")
	}

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return cfg, errHelp
		}
		return cfg, err
	}

	if fs.NArg() != 1 {
		if fs.NArg() == 0 {
			return cfg, nil
		}
		fs.Usage()
		return cfg, fmt.Errorf("expected exactly one Markdown file")
	}

	cfg.File = fs.Arg(0)
	return cfg, nil
}

func validateMarkdownFile(path string) error {
	if path == "" {
		return nil
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("cannot resolve file path: %w", err)
	}
	info, err := os.Stat(abs)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("file does not exist: %s", path)
		}
		return fmt.Errorf("cannot read file %s: %w", path, err)
	}
	if info.IsDir() {
		return fmt.Errorf("expected a Markdown file, got directory: %s", path)
	}

	ext := strings.ToLower(filepath.Ext(abs))
	if ext != ".md" && ext != ".markdown" {
		return fmt.Errorf("unsupported file extension %q, expected .md or .markdown", ext)
	}
	return nil
}

func runDesktopApp(app *App) error {
	return wails.Run(&options.App{
		Title:  "md-preview",
		Width:  1280,
		Height: 840,
		Menu:   buildApplicationMenu(app),
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 14, G: 17, B: 22, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})
}

func buildApplicationMenu(app *App) *menu.Menu {
	appMenu := menu.NewMenu()

	fileMenu := appMenu.AddSubmenu("File")
	fileMenu.AddText("Open Markdown...", keys.CmdOrCtrl("o"), func(*menu.CallbackData) {
		payload := app.OpenMarkdownFile()
		if payload.Error != "" {
			emitStatus(app, payload.Error)
		}
	})
	fileMenu.AddText("Export HTML...", keys.CmdOrCtrl("s"), func(*menu.CallbackData) {
		if _, err := app.ExportHTMLWithDialog(); err != nil {
			emitStatus(app, err.Error())
		}
	})
	fileMenu.AddText("Print / Export PDF...", keys.CmdOrCtrl("p"), func(*menu.CallbackData) {
		app.PrintPreview()
	})
	fileMenu.AddSeparator()
	fileMenu.AddText("Quit", keys.CmdOrCtrl("q"), func(*menu.CallbackData) {
		if app.ctx != nil {
			runtime.Quit(app.ctx)
		}
	})

	viewMenu := appMenu.AddSubmenu("View")
	viewMenu.AddRadio("GitHub Light", true, nil, func(*menu.CallbackData) {
		app.SetTheme("github-light")
	})
	viewMenu.AddRadio("GitHub Dark", false, nil, func(*menu.CallbackData) {
		app.SetTheme("github-dark")
	})
	viewMenu.AddRadio("GitHub Sepia", false, nil, func(*menu.CallbackData) {
		app.SetTheme("github-sepia")
	})

	return appMenu
}

func emitStatus(app *App, message string) {
	if app.ctx != nil {
		runtime.EventsEmit(app.ctx, "status-message", message)
	}
}

func fileVersion(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("cannot stat Markdown file: %w", err)
	}
	return fmt.Sprintf("%d:%d", info.ModTime().UnixNano(), info.Size()), nil
}
