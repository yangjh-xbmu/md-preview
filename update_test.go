package main

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestUpdateSettingsDefaultEnabledAndPersistence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")

	settings := loadUpdateSettings(path)
	if !settings.AutoUpdateEnabled {
		t.Fatal("expected automatic updates enabled by default")
	}

	if err := saveUpdateSettings(path, UpdateSettings{AutoUpdateEnabled: false}); err != nil {
		t.Fatalf("saveUpdateSettings returned error: %v", err)
	}

	settings = loadUpdateSettings(path)
	if settings.AutoUpdateEnabled {
		t.Fatal("expected persisted automatic update setting to be disabled")
	}
}

func TestCompareStableVersions(t *testing.T) {
	cases := []struct {
		current string
		latest  string
		want    int
	}{
		{"0.1.1", "v0.1.2", -1},
		{"v0.2.0", "v0.1.9", 1},
		{"v1.0.0", "1.0.0", 0},
	}

	for _, tc := range cases {
		got, err := compareStableVersions(tc.current, tc.latest)
		if err != nil {
			t.Fatalf("compareStableVersions(%q, %q) returned error: %v", tc.current, tc.latest, err)
		}
		if got != tc.want {
			t.Fatalf("compareStableVersions(%q, %q) = %d, want %d", tc.current, tc.latest, got, tc.want)
		}
	}
}

func TestSelectCompatibleReleaseAsset(t *testing.T) {
	release := releaseInfo{
		TagName: "v0.1.2",
		Assets: []releaseAsset{
			{Name: "md-preview-v0.1.2-linux-amd64.tar.gz"},
			{Name: "md-preview-v0.1.2-windows-amd64.zip", BrowserDownloadURL: "https://example.test/windows.zip"},
		},
	}

	asset, ok := selectCompatibleReleaseAsset(release, "windows", "amd64")
	if !ok {
		t.Fatal("expected compatible Windows asset")
	}
	if asset.Name != "md-preview-v0.1.2-windows-amd64.zip" {
		t.Fatalf("unexpected asset selected: %s", asset.Name)
	}
}

func TestFetchLatestReleaseIgnoresDraftsAndPrereleases(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `{"tag_name":"v0.1.2","draft":true,"prerelease":false,"assets":[]}`)
	}))
	defer server.Close()

	_, err := fetchLatestRelease(context.Background(), server.URL, server.Client())
	if err == nil {
		t.Fatal("expected draft release to be rejected")
	}
	if !strings.Contains(err.Error(), "stable") {
		t.Fatalf("expected stable release error, got: %v", err)
	}
}

func TestCheckForUpdatesReportsUpToDate(t *testing.T) {
	app := newTestUpdateApp(t, "0.1.1")
	app.releaseAPIURL = testReleaseServer(t, `{
		"tag_name":"v0.1.1",
		"html_url":"https://github.com/yangjh-xbmu/md-preview/releases/tag/v0.1.1",
		"draft":false,
		"prerelease":false,
		"assets":[]
	}`)

	status := app.CheckForUpdates(true)
	if status.State != updateStateUpToDate {
		t.Fatalf("expected up-to-date status, got %#v", status)
	}
	if status.LatestVersion != "0.1.1" {
		t.Fatalf("expected latest version 0.1.1, got %q", status.LatestVersion)
	}
}

func TestCheckForUpdatesDownloadsAndStagesCompatibleAsset(t *testing.T) {
	zipBytes := makeTestUpdateZip(t, "new binary")
	sum := sha256.Sum256(zipBytes)
	digest := "sha256:" + hex.EncodeToString(sum[:])

	var assetURL string
	assetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(zipBytes)
	}))
	defer assetServer.Close()
	assetURL = assetServer.URL + "/md-preview-v0.1.2-windows-amd64.zip"

	app := newTestUpdateApp(t, "0.1.1")
	app.osName = "windows"
	app.archName = "amd64"
	app.releaseAPIURL = testReleaseServer(t, fmt.Sprintf(`{
		"tag_name":"v0.1.2",
		"html_url":"https://github.com/yangjh-xbmu/md-preview/releases/tag/v0.1.2",
		"draft":false,
		"prerelease":false,
		"assets":[{"name":"md-preview-v0.1.2-windows-amd64.zip","browser_download_url":%q,"digest":%q}]
	}`, assetURL, digest))

	status := app.CheckForUpdates(true)
	if status.State != updateStateReady {
		t.Fatalf("expected ready status, got %#v", status)
	}
	if status.DownloadedPath == "" {
		t.Fatal("expected staged downloaded path")
	}
	if _, err := os.Stat(status.DownloadedPath); err != nil {
		t.Fatalf("expected staged executable: %v", err)
	}
}

func TestCheckForUpdatesFailsDigestMismatch(t *testing.T) {
	zipBytes := makeTestUpdateZip(t, "new binary")
	assetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(zipBytes)
	}))
	defer assetServer.Close()

	app := newTestUpdateApp(t, "0.1.1")
	app.osName = "windows"
	app.archName = "amd64"
	app.releaseAPIURL = testReleaseServer(t, fmt.Sprintf(`{
		"tag_name":"v0.1.2",
		"html_url":"https://github.com/yangjh-xbmu/md-preview/releases/tag/v0.1.2",
		"draft":false,
		"prerelease":false,
		"assets":[{"name":"md-preview-v0.1.2-windows-amd64.zip","browser_download_url":%q,"digest":"sha256:0000"}]
	}`, assetServer.URL+"/asset.zip"))

	status := app.CheckForUpdates(true)
	if status.State != updateStateFailed {
		t.Fatalf("expected failed status, got %#v", status)
	}
	if !strings.Contains(strings.ToLower(status.Message), "checksum") {
		t.Fatalf("expected checksum failure message, got %q", status.Message)
	}
}

func TestAutoUpdateDisabledSkipsStartupButManualCheckStillRuns(t *testing.T) {
	app := newTestUpdateApp(t, "0.1.1")
	if err := saveUpdateSettings(app.updateSettingsPath, UpdateSettings{AutoUpdateEnabled: false}); err != nil {
		t.Fatal(err)
	}
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		_, _ = fmt.Fprint(w, `{"tag_name":"v0.1.1","draft":false,"prerelease":false,"assets":[]}`)
	}))
	defer server.Close()
	app.releaseAPIURL = server.URL

	app.checkForUpdatesOnStartup()
	if calls != 0 {
		t.Fatalf("expected startup check to skip network when disabled, got %d calls", calls)
	}
	if app.GetUpdateStatus().State != updateStateDisabled {
		t.Fatalf("expected disabled status, got %#v", app.GetUpdateStatus())
	}

	status := app.CheckForUpdates(true)
	if calls != 1 {
		t.Fatalf("expected manual check to call network once, got %d calls", calls)
	}
	if status.State != updateStateUpToDate {
		t.Fatalf("expected manual check to run while disabled, got %#v", status)
	}
}

func TestStartupCheckIsNonBlocking(t *testing.T) {
	app := newTestUpdateApp(t, "0.1.1")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(150 * time.Millisecond)
		_, _ = fmt.Fprint(w, `{"tag_name":"v0.1.1","draft":false,"prerelease":false,"assets":[]}`)
	}))
	defer server.Close()
	app.releaseAPIURL = server.URL
	app.updateHTTPClient = server.Client()

	start := time.Now()
	app.startUpdateCheckIfEnabled()
	if time.Since(start) > 50*time.Millisecond {
		t.Fatal("startup update check blocked caller")
	}
}

func TestSetAutoUpdateEnabledPersistsPreference(t *testing.T) {
	app := newTestUpdateApp(t, "0.1.1")

	settings := app.SetAutoUpdateEnabled(false)
	if settings.AutoUpdateEnabled {
		t.Fatal("expected disabled setting")
	}

	reloaded := loadUpdateSettings(app.updateSettingsPath)
	if reloaded.AutoUpdateEnabled {
		t.Fatal("expected disabled setting to persist")
	}
}

func makeTestUpdateZip(t *testing.T, exeContent string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create("md-preview.exe")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte(exeContent)); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func newTestUpdateApp(t *testing.T, version string) *App {
	t.Helper()
	app, err := NewApp(config{Watch: false})
	if err != nil {
		t.Fatal(err)
	}
	app.version = version
	app.updateSettingsPath = filepath.Join(t.TempDir(), "settings.json")
	app.updateStagingDir = filepath.Join(t.TempDir(), "stage")
	app.osName = "windows"
	app.archName = "amd64"
	app.updateHTTPClient = http.DefaultClient
	return app
}

func testReleaseServer(t *testing.T, body string) string {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, body)
	}))
	t.Cleanup(server.Close)
	return server.URL
}
