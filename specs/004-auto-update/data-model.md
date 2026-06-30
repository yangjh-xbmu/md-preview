# Data Model: Auto Update

## UpdateSettings

- **Meaning**: Persistent user preference controlling startup update checks.
- **Fields**:
  - `autoUpdateEnabled`: boolean, defaults to true when no settings file exists.
- **Validation**:
  - Missing or unreadable settings file falls back to enabled.
  - Writes are atomic enough for a small single-user config file.

## UpdateStatus

- **Meaning**: User-visible state of update checking and installation.
- **Fields**:
  - `state`: one of `idle`, `disabled`, `checking`, `up-to-date`, `available`, `downloading`, `ready`, `failed`.
  - `currentVersion`: installed application version.
  - `latestVersion`: newest compatible stable release version when known.
  - `message`: concise user-facing status text.
  - `downloadedPath`: staged update path when ready.
  - `releaseURL`: release page URL when known.
  - `checkedAt`: timestamp for the last completed status transition when applicable.
- **State transitions**:
  - `idle` -> `disabled` when startup checks are off.
  - `idle` -> `checking` -> `up-to-date` when no newer release exists.
  - `checking` -> `available` -> `downloading` -> `ready` when an update is staged.
  - Any active state -> `failed` on metadata, compatibility, download, checksum, extraction, or install-preparation failure.

## ReleaseInfo

- **Meaning**: Parsed latest stable release metadata.
- **Fields**:
  - `tagName`: release tag such as `v0.1.2`.
  - `name`: release display name.
  - `htmlURL`: browser URL for release details.
  - `draft`: must be false.
  - `prerelease`: must be false.
  - `assets`: list of release assets.
- **Validation**:
  - Draft and prerelease releases are ignored.
  - Tags must parse as stable semantic versions.

## ReleaseAsset

- **Meaning**: Downloadable package for a supported platform and architecture.
- **Fields**:
  - `name`: release asset filename.
  - `browserDownloadURL`: direct download URL.
  - `digest`: optional `sha256:<hex>` digest.
  - `size`: asset size in bytes.
- **Validation**:
  - Asset name must match the current platform and architecture naming convention.
  - If a digest is present, the downloaded bytes must match it before staging.
