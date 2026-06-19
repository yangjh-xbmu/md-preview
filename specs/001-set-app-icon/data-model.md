# Data Model: Set md-preview App Icon

This feature does not introduce application data entities.

## Asset: Supplied Icon Artwork

- **Source**: `D:\Users\yangjh\Desktop\Inbox\md-preview-app-icon-final-left-layout.svg`
- **Meaning**: Final md-preview app icon artwork selected by the user.
- **Validation**: File exists, is readable, and declares an SVG root.

## Asset: Windows Application Icon

- **Target**: `build/windows/icon.ico`
- **Meaning**: Wails Windows packaging icon consumed during desktop app build.
- **Validation**: File is recognized as a Windows icon resource and contains multiple icon sizes.

## Asset: Common Application Icon Source

- **Target**: `build/appicon.png`
- **Meaning**: Wails common app icon source used when platform-specific icons need to be regenerated.
- **Validation**: File is a readable PNG generated from the supplied SVG.
