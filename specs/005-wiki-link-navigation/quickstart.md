# Quickstart: Wiki Link Navigation Fix

## Prerequisites

- Run commands from the repository root.
- Ensure `frontend/node_modules` is installed (`npm --prefix frontend install`).
- Create a temporary sibling Markdown file for manual testing.

## Sample Markdown for Manual Verification

Create a temporary file `specs/005-wiki-link-navigation/sample.md` with the following content (do not commit it):

```markdown
# Wiki Link Sample

- ASCII link: [[README]]
- Chinese link: [[Git基本概念与常用命令]]
- Space link: [[My Note]]
- Missing link: [[Does Not Exist]]
- External link: [GitHub](https://github.com)
```

Also create a sibling file `specs/005-wiki-link-navigation/My Note.md` (do not commit it):

```markdown
# My Note

You arrived via a space-encoded wiki link.
```

## Validation

Run the unit tests and frontend build:

```bash
go test ./...
npm --prefix frontend install
npm --prefix frontend run build
```

Expected result: all commands complete successfully.

Run the full desktop build:

```bash
wails build
```

Expected result: `build/bin/md-preview.exe` is rebuilt.

Manual smoke test:

```bash
wails dev
```

Open `specs/005-wiki-link-navigation/sample.md` in the running app. Expected observations:

1. Clicking `README` loads `README.md`.
2. Clicking `Git基本概念与常用命令` loads the matching `.md` file if it exists in the same directory.
3. Clicking `My Note` loads `My Note.md`.
4. Clicking `Does Not Exist` shows a status message "Wiki link target not found: ..." and leaves the current file loaded.
5. Clicking `GitHub` opens the external URL in the default browser.
6. Alt+← and Alt+→ move through the navigation history as before.
