# Contract: Auto Update Backend Bindings

## GetUpdateSettings

Returns the persistent update preference.

```json
{
  "autoUpdateEnabled": true
}
```

## SetAutoUpdateEnabled

Input:

```json
true
```

Returns the saved preference after persistence:

```json
{
  "autoUpdateEnabled": true
}
```

## GetUpdateStatus

Returns the current update status.

```json
{
  "state": "idle",
  "currentVersion": "0.1.1",
  "latestVersion": "",
  "message": "Automatic updates are enabled.",
  "downloadedPath": "",
  "releaseURL": "",
  "checkedAt": ""
}
```

## CheckForUpdates

Input:

```json
true
```

Returns the resulting status. Manual checks are allowed even when startup checks are disabled.

```json
{
  "state": "up-to-date",
  "currentVersion": "0.1.1",
  "latestVersion": "0.1.1",
  "message": "md-preview is up to date.",
  "downloadedPath": "",
  "releaseURL": "https://github.com/yangjh-xbmu/md-preview/releases/tag/v0.1.1",
  "checkedAt": "2026-06-30T10:30:00+08:00"
}
```

## InstallStagedUpdate

Starts the replacement handoff for a staged update and returns a final status before the current app quits.

```json
{
  "state": "ready",
  "currentVersion": "0.1.1",
  "latestVersion": "0.1.2",
  "message": "Update will be installed after restart.",
  "downloadedPath": "C:\\Users\\user\\AppData\\Local\\Temp\\md-preview-update\\md-preview.exe",
  "releaseURL": "https://github.com/yangjh-xbmu/md-preview/releases/tag/v0.1.2",
  "checkedAt": "2026-06-30T10:30:00+08:00"
}
```

## Events

The backend emits `update-status-changed` with the same `UpdateStatus` payload whenever status changes.
