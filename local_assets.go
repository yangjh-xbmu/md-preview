// INPUT: Goldmark image nodes, the active Markdown path, and Wails asset requests.
// OUTPUT: Opaque local image URLs and safely served allowlisted image files.
// POS: Security boundary between rendered Markdown and local filesystem image assets.
package main

import (
	"crypto/sha256"
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark/ast"
)

const localAssetPrefix = "/__md_asset/"

var localImageExtensions = map[string]struct{}{
	".gif":  {},
	".jpeg": {},
	".jpg":  {},
	".png":  {},
	".svg":  {},
	".webp": {},
}

func rewriteLocalImageDestinations(document ast.Node, markdownPath string) map[string]string {
	assets := make(map[string]string)
	baseDir := filepath.Dir(markdownPath)

	_ = ast.Walk(document, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering || node.Kind() != ast.KindImage {
			return ast.WalkContinue, nil
		}

		image := node.(*ast.Image)
		path, ok := resolveLocalImagePath(baseDir, string(image.Destination))
		if !ok {
			return ast.WalkSkipChildren, nil
		}

		token := localAssetToken(path)
		assets[token] = path
		image.Destination = []byte(localAssetPrefix + token)
		return ast.WalkSkipChildren, nil
	})

	return assets
}

func resolveLocalImagePath(baseDir, destination string) (string, bool) {
	destination = strings.TrimSpace(destination)
	if destination == "" {
		return "", false
	}

	parsed, err := url.Parse(destination)
	if err != nil || parsed.IsAbs() || parsed.Host != "" || strings.HasPrefix(parsed.Path, "/") {
		return "", false
	}

	decoded, err := url.PathUnescape(parsed.Path)
	if err != nil || decoded == "" || filepath.IsAbs(decoded) {
		return "", false
	}

	path, err := filepath.Abs(filepath.Join(baseDir, filepath.FromSlash(decoded)))
	if err != nil || !isSupportedLocalImage(path) {
		return "", false
	}

	info, err := os.Stat(path)
	if err != nil || !info.Mode().IsRegular() {
		return "", false
	}
	return path, true
}

func isSupportedLocalImage(path string) bool {
	_, ok := localImageExtensions[strings.ToLower(filepath.Ext(path))]
	return ok
}

func localAssetToken(path string) string {
	sum := sha256.Sum256([]byte(filepath.Clean(path)))
	return fmt.Sprintf("%x", sum)
}

func (a *App) replaceLocalAssets(assets map[string]string) {
	a.assetMu.Lock()
	a.localAssets = assets
	a.assetMu.Unlock()
}

func (a *App) localAssetPath(token string) (string, bool) {
	a.assetMu.RLock()
	path, ok := a.localAssets[token]
	a.assetMu.RUnlock()
	return path, ok
}

type localAssetHandler struct {
	app *App
}

func newLocalAssetHandler(app *App) http.Handler {
	return &localAssetHandler{app: app}
}

// ServeHTTP serves only image files registered by the current Markdown render.
func (h *localAssetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if !strings.HasPrefix(r.URL.Path, localAssetPrefix) {
		http.NotFound(w, r)
		return
	}

	token := strings.TrimPrefix(r.URL.Path, localAssetPrefix)
	if token == "" || strings.Contains(token, "/") {
		http.NotFound(w, r)
		return
	}

	path, ok := h.app.localAssetPath(token)
	if !ok || !isSupportedLocalImage(path) {
		http.NotFound(w, r)
		return
	}

	file, err := os.Open(path)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		http.NotFound(w, r)
		return
	}

	contentType := mime.TypeByExtension(strings.ToLower(filepath.Ext(path)))
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'; sandbox")
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	http.ServeContent(w, r, filepath.Base(path), info.ModTime(), file)
}
