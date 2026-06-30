package main

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"strconv"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

var appVersion = "dev"

const (
	defaultReleaseAPIURL = "https://api.github.com/repos/yangjh-xbmu/md-preview/releases/latest"
	updateEventName      = "update-status-changed"
	updateCheckTimeout   = 12 * time.Second

	updateStateIdle        = "idle"
	updateStateDisabled    = "disabled"
	updateStateChecking    = "checking"
	updateStateUpToDate    = "up-to-date"
	updateStateAvailable   = "available"
	updateStateDownloading = "downloading"
	updateStateReady       = "ready"
	updateStateFailed      = "failed"
)

type UpdateSettings struct {
	AutoUpdateEnabled bool `json:"autoUpdateEnabled"`
}

type updateSettingsFile struct {
	AutoUpdateEnabled *bool `json:"autoUpdateEnabled,omitempty"`
}

type UpdateStatus struct {
	State          string `json:"state"`
	CurrentVersion string `json:"currentVersion"`
	LatestVersion  string `json:"latestVersion"`
	Message        string `json:"message"`
	DownloadedPath string `json:"downloadedPath"`
	ReleaseURL     string `json:"releaseURL"`
	CheckedAt      string `json:"checkedAt"`
}

type releaseInfo struct {
	TagName    string         `json:"tag_name"`
	Name       string         `json:"name"`
	HTMLURL    string         `json:"html_url"`
	Draft      bool           `json:"draft"`
	Prerelease bool           `json:"prerelease"`
	Assets     []releaseAsset `json:"assets"`
}

type releaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Digest             string `json:"digest"`
	Size               int64  `json:"size"`
}

func defaultUpdateSettings() UpdateSettings {
	return UpdateSettings{AutoUpdateEnabled: true}
}

func defaultUpdateSettingsPath() string {
	dir, err := os.UserConfigDir()
	if err != nil || strings.TrimSpace(dir) == "" {
		dir = os.TempDir()
	}
	return filepath.Join(dir, "md-preview", "settings.json")
}

func defaultUpdateStagingDir() string {
	return filepath.Join(os.TempDir(), "md-preview-update")
}

func loadUpdateSettings(path string) UpdateSettings {
	settings := defaultUpdateSettings()
	raw, err := os.ReadFile(path)
	if err != nil {
		return settings
	}

	var stored updateSettingsFile
	if err := json.Unmarshal(raw, &stored); err != nil {
		return settings
	}
	if stored.AutoUpdateEnabled != nil {
		settings.AutoUpdateEnabled = *stored.AutoUpdateEnabled
	}
	return settings
}

func saveUpdateSettings(path string, settings UpdateSettings) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("cannot create settings directory: %w", err)
	}
	raw, err := json.MarshalIndent(updateSettingsFile{AutoUpdateEnabled: &settings.AutoUpdateEnabled}, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot encode update settings: %w", err)
	}
	return os.WriteFile(path, raw, 0o644)
}

func newUpdateStatus(version string) UpdateStatus {
	return UpdateStatus{
		State:          updateStateIdle,
		CurrentVersion: normalizeVersion(version),
		Message:        "Automatic updates are enabled.",
	}
}

func normalizeVersion(version string) string {
	version = strings.TrimSpace(version)
	version = strings.TrimPrefix(version, "v")
	if version == "" {
		return "dev"
	}
	return version
}

func compareStableVersions(current, latest string) (int, error) {
	currentParts, err := parseStableVersion(current)
	if err != nil {
		return 0, err
	}
	latestParts, err := parseStableVersion(latest)
	if err != nil {
		return 0, err
	}

	for i := range currentParts {
		if currentParts[i] < latestParts[i] {
			return -1, nil
		}
		if currentParts[i] > latestParts[i] {
			return 1, nil
		}
	}
	return 0, nil
}

func parseStableVersion(version string) ([3]int, error) {
	var parts [3]int
	version = normalizeVersion(version)
	chunks := strings.Split(version, ".")
	if len(chunks) != 3 {
		return parts, fmt.Errorf("unsupported version %q", version)
	}
	for i, chunk := range chunks {
		if chunk == "" || strings.ContainsAny(chunk, "-+") {
			return parts, fmt.Errorf("unsupported version %q", version)
		}
		value, err := strconv.Atoi(chunk)
		if err != nil || value < 0 {
			return parts, fmt.Errorf("unsupported version %q", version)
		}
		parts[i] = value
	}
	return parts, nil
}

func fetchLatestRelease(ctx context.Context, url string, client *http.Client) (releaseInfo, error) {
	if client == nil {
		client = http.DefaultClient
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return releaseInfo{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "md-preview/"+normalizeVersion(appVersion))

	resp, err := client.Do(req)
	if err != nil {
		return releaseInfo{}, fmt.Errorf("cannot check latest release: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return releaseInfo{}, fmt.Errorf("release service returned %s", resp.Status)
	}

	var release releaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return releaseInfo{}, fmt.Errorf("cannot parse release metadata: %w", err)
	}
	if release.Draft || release.Prerelease || strings.TrimSpace(release.TagName) == "" {
		return releaseInfo{}, errors.New("latest release is not a stable public release")
	}
	if _, err := parseStableVersion(release.TagName); err != nil {
		return releaseInfo{}, err
	}
	return release, nil
}

func selectCompatibleReleaseAsset(release releaseInfo, osName, archName string) (releaseAsset, bool) {
	version := normalizeVersion(release.TagName)
	ext := ".tar.gz"
	if osName == "windows" {
		ext = ".zip"
	}
	expected := fmt.Sprintf("md-preview-v%s-%s-%s%s", version, osName, archName, ext)
	for _, asset := range release.Assets {
		if asset.Name == expected {
			return asset, true
		}
	}
	return releaseAsset{}, false
}

func (a *App) initUpdateDefaults() {
	a.version = normalizeVersion(appVersion)
	a.releaseAPIURL = defaultReleaseAPIURL
	a.updateSettingsPath = defaultUpdateSettingsPath()
	a.updateStagingDir = defaultUpdateStagingDir()
	a.updateHTTPClient = &http.Client{Timeout: updateCheckTimeout}
	a.osName = goruntime.GOOS
	a.archName = goruntime.GOARCH
	a.failedUpdateVersions = map[string]bool{}
	a.updateStatus = newUpdateStatus(a.version)
}

func (a *App) GetUpdateSettings() UpdateSettings {
	return loadUpdateSettings(a.updateSettingsPath)
}

func (a *App) SetAutoUpdateEnabled(enabled bool) UpdateSettings {
	settings := UpdateSettings{AutoUpdateEnabled: enabled}
	if err := saveUpdateSettings(a.updateSettingsPath, settings); err != nil {
		a.setUpdateStatus(UpdateStatus{
			State:          updateStateFailed,
			CurrentVersion: a.version,
			Message:        "Failed to save automatic update setting: " + err.Error(),
			CheckedAt:      time.Now().Format(time.RFC3339),
		})
		return a.GetUpdateSettings()
	}

	if !enabled {
		a.setUpdateStatus(UpdateStatus{
			State:          updateStateDisabled,
			CurrentVersion: a.version,
			Message:        "Automatic update checks are disabled.",
			CheckedAt:      time.Now().Format(time.RFC3339),
		})
		return settings
	}

	a.setUpdateStatus(UpdateStatus{
		State:          updateStateIdle,
		CurrentVersion: a.version,
		Message:        "Automatic updates are enabled.",
		CheckedAt:      time.Now().Format(time.RFC3339),
	})
	return settings
}

func (a *App) GetUpdateStatus() UpdateStatus {
	a.updateMu.Lock()
	defer a.updateMu.Unlock()
	return a.updateStatus
}

func (a *App) setUpdateStatus(status UpdateStatus) UpdateStatus {
	if status.CurrentVersion == "" {
		status.CurrentVersion = a.version
	}
	a.updateMu.Lock()
	a.updateStatus = status
	a.updateMu.Unlock()
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, updateEventName, status)
	}
	return status
}

func (a *App) startUpdateCheckIfEnabled() {
	settings := a.GetUpdateSettings()
	if !settings.AutoUpdateEnabled {
		a.setUpdateStatus(UpdateStatus{
			State:          updateStateDisabled,
			CurrentVersion: a.version,
			Message:        "Automatic update checks are disabled.",
			CheckedAt:      time.Now().Format(time.RFC3339),
		})
		return
	}
	go a.CheckForUpdates(false)
}

func (a *App) checkForUpdatesOnStartup() {
	settings := a.GetUpdateSettings()
	if !settings.AutoUpdateEnabled {
		a.setUpdateStatus(UpdateStatus{
			State:          updateStateDisabled,
			CurrentVersion: a.version,
			Message:        "Automatic update checks are disabled.",
			CheckedAt:      time.Now().Format(time.RFC3339),
		})
		return
	}
	_ = a.CheckForUpdates(false)
}

func (a *App) CheckForUpdates(manual bool) UpdateStatus {
	if !manual && !a.GetUpdateSettings().AutoUpdateEnabled {
		return a.setUpdateStatus(UpdateStatus{
			State:          updateStateDisabled,
			CurrentVersion: a.version,
			Message:        "Automatic update checks are disabled.",
			CheckedAt:      time.Now().Format(time.RFC3339),
		})
	}

	a.setUpdateStatus(UpdateStatus{
		State:          updateStateChecking,
		CurrentVersion: a.version,
		Message:        "Checking for updates...",
	})

	if _, err := parseStableVersion(a.version); err != nil {
		return a.failUpdate("", "", "Automatic updates require a release build.")
	}

	ctx, cancel := context.WithTimeout(context.Background(), updateCheckTimeout)
	defer cancel()

	release, err := fetchLatestRelease(ctx, a.releaseAPIURL, a.updateHTTPClient)
	if err != nil {
		return a.failUpdate("", "", err.Error())
	}

	latest := normalizeVersion(release.TagName)
	cmp, err := compareStableVersions(a.version, latest)
	if err != nil {
		return a.failUpdate(latest, release.HTMLURL, err.Error())
	}
	if cmp >= 0 {
		return a.setUpdateStatus(UpdateStatus{
			State:          updateStateUpToDate,
			CurrentVersion: a.version,
			LatestVersion:  latest,
			Message:        "md-preview is up to date.",
			ReleaseURL:     release.HTMLURL,
			CheckedAt:      time.Now().Format(time.RFC3339),
		})
	}
	if a.hasFailedUpdateVersion(latest) {
		return a.failUpdate(latest, release.HTMLURL, "Skipping repeated automatic download for this version during this session.")
	}

	asset, ok := selectCompatibleReleaseAsset(release, a.osName, a.archName)
	if !ok {
		return a.failUpdate(latest, release.HTMLURL, "No compatible update asset is available for this system.")
	}

	a.setUpdateStatus(UpdateStatus{
		State:          updateStateAvailable,
		CurrentVersion: a.version,
		LatestVersion:  latest,
		Message:        "Update " + latest + " is available.",
		ReleaseURL:     release.HTMLURL,
	})

	a.setUpdateStatus(UpdateStatus{
		State:          updateStateDownloading,
		CurrentVersion: a.version,
		LatestVersion:  latest,
		Message:        "Downloading update " + latest + "...",
		ReleaseURL:     release.HTMLURL,
	})

	staged, err := a.downloadAndStageAsset(ctx, latest, asset)
	if err != nil {
		a.markFailedUpdateVersion(latest)
		return a.failUpdate(latest, release.HTMLURL, err.Error())
	}

	return a.setUpdateStatus(UpdateStatus{
		State:          updateStateReady,
		CurrentVersion: a.version,
		LatestVersion:  latest,
		Message:        "Update " + latest + " is ready. Restart to install.",
		DownloadedPath: staged,
		ReleaseURL:     release.HTMLURL,
		CheckedAt:      time.Now().Format(time.RFC3339),
	})
}

func (a *App) failUpdate(latest, releaseURL, message string) UpdateStatus {
	return a.setUpdateStatus(UpdateStatus{
		State:          updateStateFailed,
		CurrentVersion: a.version,
		LatestVersion:  latest,
		Message:        message,
		ReleaseURL:     releaseURL,
		CheckedAt:      time.Now().Format(time.RFC3339),
	})
}

func (a *App) hasFailedUpdateVersion(version string) bool {
	a.updateMu.Lock()
	defer a.updateMu.Unlock()
	return a.failedUpdateVersions[version]
}

func (a *App) markFailedUpdateVersion(version string) {
	a.updateMu.Lock()
	defer a.updateMu.Unlock()
	a.failedUpdateVersions[version] = true
}

func (a *App) downloadAndStageAsset(ctx context.Context, version string, asset releaseAsset) (string, error) {
	if a.osName != "windows" {
		return "", errors.New("automatic installation is not supported on this platform yet")
	}
	if strings.TrimSpace(asset.BrowserDownloadURL) == "" {
		return "", errors.New("release asset has no download URL")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, asset.BrowserDownloadURL, nil)
	if err != nil {
		return "", err
	}
	resp, err := a.updateHTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot download update: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("update download returned %s", resp.Status)
	}

	dir := filepath.Join(a.updateStagingDir, "v"+normalizeVersion(version))
	if err := os.RemoveAll(dir); err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	archivePath := filepath.Join(dir, asset.Name)
	out, err := os.Create(archivePath)
	if err != nil {
		return "", err
	}
	hash := sha256.New()
	if _, err := io.Copy(io.MultiWriter(out, hash), resp.Body); err != nil {
		_ = out.Close()
		return "", err
	}
	if err := out.Close(); err != nil {
		return "", err
	}

	if err := verifyAssetDigest(asset.Digest, hash.Sum(nil)); err != nil {
		return "", err
	}
	return extractWindowsExecutable(archivePath, dir)
}

func verifyAssetDigest(digest string, actual []byte) error {
	digest = strings.TrimSpace(digest)
	if digest == "" {
		return nil
	}
	expected, ok := strings.CutPrefix(digest, "sha256:")
	if !ok {
		return fmt.Errorf("unsupported asset checksum format")
	}
	if !strings.EqualFold(expected, hex.EncodeToString(actual)) {
		return fmt.Errorf("update checksum mismatch")
	}
	return nil
}

func extractWindowsExecutable(archivePath, targetDir string) (string, error) {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return "", fmt.Errorf("cannot open update archive: %w", err)
	}
	defer reader.Close()

	for _, file := range reader.File {
		if filepath.Base(file.Name) != "md-preview.exe" {
			continue
		}
		src, err := file.Open()
		if err != nil {
			return "", err
		}
		defer src.Close()
		target := filepath.Join(targetDir, "md-preview.exe")
		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
		if err != nil {
			return "", err
		}
		if _, err := io.Copy(out, src); err != nil {
			_ = out.Close()
			return "", err
		}
		if err := out.Close(); err != nil {
			return "", err
		}
		return target, nil
	}
	return "", errors.New("update archive does not contain md-preview.exe")
}

func (a *App) InstallStagedUpdate() UpdateStatus {
	status := a.GetUpdateStatus()
	if status.State != updateStateReady || status.DownloadedPath == "" {
		return a.failUpdate(status.LatestVersion, status.ReleaseURL, "No staged update is ready to install.")
	}
	if a.osName != "windows" {
		return a.failUpdate(status.LatestVersion, status.ReleaseURL, "Automatic installation is not supported on this platform yet.")
	}

	exePath, err := os.Executable()
	if err != nil {
		return a.failUpdate(status.LatestVersion, status.ReleaseURL, "Cannot locate current executable: "+err.Error())
	}
	scriptPath := filepath.Join(filepath.Dir(status.DownloadedPath), "install-md-preview-update.cmd")
	script := buildWindowsInstallScript(os.Getpid(), status.DownloadedPath, exePath, os.Args[1:])
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		return a.failUpdate(status.LatestVersion, status.ReleaseURL, "Cannot write update installer: "+err.Error())
	}

	if err := exec.Command("cmd", "/C", "start", "", scriptPath).Start(); err != nil {
		return a.failUpdate(status.LatestVersion, status.ReleaseURL, "Cannot start update installer: "+err.Error())
	}

	final := a.setUpdateStatus(UpdateStatus{
		State:          updateStateReady,
		CurrentVersion: a.version,
		LatestVersion:  status.LatestVersion,
		Message:        "Installing update after restart...",
		DownloadedPath: status.DownloadedPath,
		ReleaseURL:     status.ReleaseURL,
		CheckedAt:      time.Now().Format(time.RFC3339),
	})
	if a.ctx != nil {
		go func() {
			time.Sleep(200 * time.Millisecond)
			runtime.Quit(a.ctx)
		}()
	}
	return final
}

func buildWindowsInstallScript(pid int, stagedPath, exePath string, args []string) string {
	quotedArgs := make([]string, 0, len(args))
	for _, arg := range args {
		quotedArgs = append(quotedArgs, `"`+strings.ReplaceAll(arg, `"`, `""`)+`"`)
	}
	return fmt.Sprintf(`@echo off
setlocal
set PID=%d
set "STAGED=%s"
set "TARGET=%s"
:wait
tasklist /FI "PID eq %%PID%%" | find "%%PID%%" >nul
if not errorlevel 1 (
  timeout /t 1 /nobreak >nul
  goto wait
)
copy /Y "%%STAGED%%" "%%TARGET%%" >nul
start "" "%%TARGET%%" %s
endlocal
`, pid, stagedPath, exePath, strings.Join(quotedArgs, " "))
}
