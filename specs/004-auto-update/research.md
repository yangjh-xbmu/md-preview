# Research: Auto Update

## Decision: Backend-owned update preference

**Rationale**: Startup checks must be decided before or during app startup, so the preference cannot live only in frontend localStorage. A backend config file under the user config directory gives startup code access and lets the frontend menu call backend methods to read and change the setting.

**Alternatives considered**: Frontend localStorage would be easy to wire but cannot reliably suppress startup checks before the frontend loads. A CLI flag would not satisfy the persistent menu toggle requirement.

## Decision: Use the existing GitHub release channel

**Rationale**: The project already publishes versioned release assets through GitHub Releases. The current asset naming is deterministic, including `md-preview-v<VERSION>-windows-amd64.zip`, which is enough for platform asset selection without adding a new release service.

**Alternatives considered**: A custom update manifest would add another publication step. Third-party updater services would add dependencies and operational overhead.

## Decision: Standard library implementation instead of an updater dependency

**Rationale**: The feature needs JSON metadata fetch, semantic version comparison for simple `vMAJOR.MINOR.PATCH` tags, archive extraction, SHA-256 verification, and a replacement handoff. These are small enough to implement with the Go standard library while preserving the dependency-light constraint.

**Alternatives considered**: A general-purpose self-update library could handle more platforms but would add dependency and behavior surface. The current app's release workflow and Windows-first target do not require that extra complexity.

## Decision: Asynchronous startup check with bounded network timeout

**Rationale**: Startup preview must remain usable even if the release service is unreachable. Running the check in a background goroutine and publishing status events keeps the UI responsive and satisfies the failure-safe requirement.

**Alternatives considered**: Blocking startup until version metadata returns would make slow or offline networks visible as app startup failures.

## Decision: Windows-first install handoff, safe fallback elsewhere

**Rationale**: The current user environment and primary build target are Windows desktop. Windows cannot overwrite a running executable directly, so the update flow downloads and verifies the asset, stages the new executable, then launches a short handoff script that waits for the app to exit, replaces the executable, and restarts it. Non-Windows builds can still check metadata and report unsupported automatic installation until a platform-specific installer path is added.

**Alternatives considered**: Replacing files in-process is unreliable on Windows. Requiring a full NSIS installer would require release workflow changes and a heavier install experience.
