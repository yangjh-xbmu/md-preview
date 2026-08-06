# Data Model: macOS Full Screen and Quit Shortcuts

This feature introduces no persistent storage or backend entities. The following transient concepts are sufficient for implementation and validation.

## Window State

Represents whether the native md-preview window is in normal mode or native full screen.

| Field | Type | Rules |
| --- | --- | --- |
| `isFullscreen` | boolean | Mirrors the native window state after startup synchronization or a successful toggle. |
| `previousWindowState` | native window state | Managed by the existing desktop runtime and restored when full screen ends. |

### State transitions

```text
normal -- F11 or Option+F11 --> fullscreen
fullscreen -- F11 or Option+F11 --> normal
```

The menu action uses the same transitions. A runtime failure leaves the preview usable and reports the existing failure message.

## Keyboard Shortcut Event

Represents a transient browser key event received by the React shell.

| Shortcut | Modifier condition | Action |
| --- | --- | --- |
| `F11` | no required modifier | Toggle native full screen |
| `Option+F11` | `altKey` | Toggle native full screen |
| `Ctrl+Q` | `ctrlKey`, without `altKey` or `metaKey` | Quit md-preview and stop processing the event |

The feature does not persist shortcut preferences. Existing shortcuts keep their current dispatch behavior.
