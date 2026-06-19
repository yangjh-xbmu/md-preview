# Quickstart: Set md-preview App Icon

## Prerequisites

- Run commands from the repository root.
- Ensure the supplied SVG exists at `D:\Users\yangjh\Desktop\Inbox\md-preview-app-icon-final-left-layout.svg`.

## Validation

Inspect the icon file:

```bash
file build/appicon.png
file build/windows/icon.ico
```

Expected result: output identifies `build/appicon.png` as a PNG image and `build/windows/icon.ico` as an MS Windows icon resource with multiple embedded icon sizes.

Run project validation:

```bash
go test ./...
npm --prefix frontend run build
wails build
```

Expected result: all commands complete successfully, and `build/bin/md-preview.exe` is rebuilt with the updated packaged icon.
