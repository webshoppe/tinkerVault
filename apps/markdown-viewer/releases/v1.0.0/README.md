# Markdown Viewer

![Release](https://img.shields.io/github/v/release/webshoppe/tinkerVault?filter=markdown-viewer*&label=release)

<img src="icon.svg" width="80" alt="Markdown Viewer logo icon">

A tiny, free, offline tool that turns Markdown files (`.md`) into a clean, readable page in your web browser, no installation, no internet, no account, no sign-up. You open a file and it just shows up, formatted, with tables, checklists, code, and more.

This is an archived, self-contained snapshot of **v1.0.0**, preserved exactly as it shipped. For the current version, see the [main app folder](https://github.com/webshoppe/tinkerVault/tree/main/apps/markdown-viewer) on GitHub.

## What you need

- A computer (Windows, macOS, or Linux).
- A web browser (Chrome, Edge, Firefox, Safari; anything reasonably recent).
- The `index.html` file from this folder.

That's it. There is nothing to install.

## How to use it (2 steps)

1. **Open the tool.** Find `index.html` in this folder and double-click it. It opens in your default web browser. (You can also drag `index.html` onto an open browser window.)
2. **Open a Markdown file.** Once the page is open:
   - Click the **"Choose a .md file"** text at the top-left, pick a `.md` file from your computer, and it appears formatted instantly; **or**
   - Drag a `.md` file from your file manager and drop it anywhere on the page.

That's the whole thing. No setup, no configuration.

## What it can do

Everything below works the moment a file is open:

- **Tables, checklists, and nested bullet lists** render properly.
- **Code blocks** are shown with real, colored syntax highlighting (not just a plain monospace font).
- **Copy button**, every code block has a small "Copy" button that copies the code to your clipboard so you can paste it elsewhere.
- **Dark mode**, click the 🌙 / ☀️ button in the toolbar to switch between light and dark.
- **Table of contents**, a list of the document's headings appears on the left; click any entry to jump there. Click the ☰ button to hide/show it.
- **Search**, type in the search box (top-right) to highlight matches inside the document; press Enter to jump between them.
- **Multiple files**, open several `.md` files at once; each gets its own tab at the top. Click a tab to switch.
- **Raw view**, click the 📝 button to see the original, unformatted Markdown text, and click again to go back to the formatted view.
- **Export**, click ⬇ to save the currently open, formatted document as a single standalone `.html` file you can share or open anywhere.
- **Remembers your file**, when you reopen the tool, the last file you were reading comes back automatically.

## Files in this snapshot

```
index.html          the app, exactly as it shipped for v1.0.0
icon.svg
VERSION
README.md            this file
docs/
  USER_GUIDE.md        every feature, in first-time-user order
```

Intentionally minimal, just what's needed to run this exact version standalone, no external dependencies. For source code, the changelog, and newer versions, see the [main app folder](https://github.com/webshoppe/tinkerVault/tree/main/apps/markdown-viewer).

## License

MIT, see the repo's [LICENSE](https://github.com/webshoppe/tinkerVault/blob/main/LICENSE) file. See this version's entry in [CHANGELOG.md](https://github.com/webshoppe/tinkerVault/blob/main/apps/markdown-viewer/CHANGELOG.md) for what changed since.

## Where to get help

- For a full walkthrough of every feature, see [`docs/USER_GUIDE.md`](docs/USER_GUIDE.md) (included in this folder).
- For technical details and how the tool is built, see the current [`src/DEV_GUIDE.md`](https://github.com/webshoppe/tinkerVault/blob/main/apps/markdown-viewer/src/DEV_GUIDE.md) in the main app folder.
- To try it out, use the sample file at [`test/test.md`](https://github.com/webshoppe/tinkerVault/blob/main/apps/markdown-viewer/test/test.md).
- The full build story, methods, decisions, and dead ends, is in [`PROJECT_SUMMARY.md`](https://github.com/webshoppe/tinkerVault/blob/main/apps/markdown-viewer/PROJECT_SUMMARY.md) in the main app folder.
