# Quickstart: macOS Full Screen and Quit Shortcuts

## Prerequisites

- macOS desktop environment
- Go toolchain and Wails CLI available on `PATH`
- Node.js and npm installed for the frontend build
- A local Markdown file, such as `README.md`

## Automated validation

Run from the repository root:

```bash
shortcut_test_dir="$(mktemp -d)"
(cd frontend && ./node_modules/.bin/tsc --target ES2020 --module commonjs --strict --outDir "$shortcut_test_dir" src/keyboard.ts tests/keyboard.test.ts)
node "$shortcut_test_dir/tests/keyboard.test.js"
go test ./...
npm --prefix frontend run build
wails build
```

Expected results:

- Go tests pass.
- The standalone shortcut predicate test prints `keyboard shortcuts: pass`.
- TypeScript compilation and Vite production build pass.
- Wails produces the desktop binary under `build/bin/`.

## macOS manual smoke test

1. Build the application with `wails build` and launch `./build/bin/md-preview README.md`.
2. Confirm the preview window is visible in a normal window.
3. Press `F11`. Confirm the window enters native macOS full screen and the menu label changes to `Exit Full Screen`.
4. Press `F11` again. Confirm the previous window dimensions and position are restored.
5. Relaunch the app and press `Option+F11`. Confirm it enters native full screen.
6. With the preview focused, press `Option+F11` again. Confirm it exits full screen.
7. Relaunch the app, open the menu, focus a menu control, and press `Ctrl+Q`. Confirm the window closes and the process exits.
8. Relaunch the app, focus the preview, press `Ctrl+O`, and confirm the existing Open Markdown dialog still appears. Cancel the dialog and close the app normally.

## Regression boundaries

Do not expect this feature to change Markdown rendering, file watching, navigation, themes, export, print, Mermaid, or update behavior. Any regression in those areas is outside the intended shortcut change and must be investigated before completion.
