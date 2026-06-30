# Quickstart: Auto Update

## Prerequisites

- Go 1.23
- Node.js 20
- Wails CLI available on PATH
- Network access for manual GitHub release checks when testing live metadata

## Validation Commands

```bash
go test ./...
npm --prefix frontend run build
wails build -ldflags "-X main.appVersion=0.1.1"
```

Expected result: all commands complete successfully, and the built app contains a non-development version string.

## Manual Scenarios

### Startup Check Enabled

1. Ensure no local update settings file disables update checks.
2. Launch the freshly built app with `build/bin/md-preview.exe README.md`.
3. Open the app menu and inspect the update section.

Expected result: automatic update checks are enabled, Markdown preview loads normally, and the update status eventually reports current, available, unsupported, or failed without blocking preview interaction.

### Disable Startup Checks

1. Open the app menu.
2. Turn off automatic update checks.
3. Restart the app.

Expected result: the update section shows disabled, and no automatic startup check starts. Manual "Check for Updates" still works.

### Manual Check

1. Open the app menu.
2. Click the manual check action.

Expected result: the status changes through checking and then reports up to date, available, unsupported, or failed.

### Staged Update

1. Test with a controlled release metadata endpoint or a version older than the latest public release.
2. Let the update download and stage.
3. Trigger install or restart flow.

Expected result: the app prepares replacement safely. If automatic replacement is not supported in the current environment, it reports a clear non-blocking failure or unsupported status.
