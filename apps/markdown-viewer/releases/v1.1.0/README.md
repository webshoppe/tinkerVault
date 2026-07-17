# Markdown Viewer

![Release](https://img.shields.io/github/v/release/webshoppe/tinkerVault?filter=markdown-viewer*&label=release)

<img src="icon.svg" width="80" alt="Markdown Viewer logo icon">

**Version 1.1.0**, a free, offline Markdown viewer and light editor in a single HTML file.

This is an archived, self-contained snapshot of v1.1.0, preserved exactly as it shipped (it also happens to be the current version right now). Open it by **double-clicking `index.html`**. There is nothing to install, no account, and no server. Your files never leave your computer.

Part of the [tinkerVault](https://github.com/webshoppe/tinkerVault) collection of offline-first single-file apps. For the current version, full source, and license, see the [main app folder](https://github.com/webshoppe/tinkerVault/tree/main/apps/markdown-viewer) on GitHub.

---

## What it is

Markdown Viewer turns `.md` files into a clean, readable, GitHub-style page: tables, checklists, syntax-highlighted code, nested lists, blockquotes, a table of contents, and in-document search. Open several files at once and switch between them in tabs. Edit the rendered text directly, save your edits as a new file, copy the raw Markdown or a single code block, and export the current document as one portable standalone HTML file you can hand to someone else. It runs fully offline, no network calls at runtime, ever.

## How to open

1. Find `index.html` in this folder.
2. Double-click it (or drag it into Chrome, Edge, Brave, or Firefox).
3. Click **📂 Open**, or drag a `.md` file (or several at once) straight onto the page.

You can copy this whole folder to a USB drive or another machine and open it the same way.

**Want it to open in its own window instead of a browser tab?** See [Making an App Feel Like a Real Desktop App](https://github.com/webshoppe/tinkerVault/blob/main/DESKTOP-LAUNCHER.md).

## What it can do

- **Tabs.** Open several files at once, each gets its own tab, even if two files share the same name from different folders (disambiguated as `(2)`, `(3)`...).
- **Full GitHub-style rendering.** Tables, checklists, nested lists, blockquotes, and code blocks with real syntax highlighting.
- **Table of contents.** A list of the document's headings on the left, click any entry to jump there.
- **Search.** Highlights matches inside the document, jump between them.
- **Dark / light theme.**
- **Raw / rendered toggle**, see the original Markdown text or the formatted page.
- **Edit mode.** Edit the text directly in the rendered view; the formatted view updates when you switch back out of edit mode.
- **Copy.** Copy the raw Markdown for the whole document (Copy MD), or just a single code block.
- **Save.** Download your edits as a new file (a browser can't overwrite the original file it opened).
- **Export.** Save the current document as one standalone `.html` file you can share or open anywhere.
- **Keyboard shortcuts** for opening, closing, reopening the last closed tab, and switching between tabs.
- **Remembers where you left off.** Reopening the tool brings back your open tabs and scroll position.

Full walkthrough, in the order a first-time user meets each feature: [`docs/USER_GUIDE.md`](docs/USER_GUIDE.md) (included in this folder).

## What's new in v1.1.0

Fixed a real bug: two files with the same name from different folders used to collide into a single tab, each open file now gets its own tab, always. Added multi-file open, edit mode, Save, Copy raw Markdown, keyboard shortcuts, and per-tab scroll memory. Full history: [`CHANGELOG.md`](https://github.com/webshoppe/tinkerVault/blob/main/apps/markdown-viewer/CHANGELOG.md).

## Files in this snapshot

```
index.html          the app, exactly as it shipped for v1.1.0
icon.svg
VERSION
README.md            this file
docs/
  USER_GUIDE.md        every feature, in first-time-user order
```

Intentionally minimal, just what's needed to run this exact version standalone, no external dependencies. For source code, other versions, and the full build story, see the [main app folder](https://github.com/webshoppe/tinkerVault/tree/main/apps/markdown-viewer).

## License

MIT, see the repo's [LICENSE](https://github.com/webshoppe/tinkerVault/blob/main/LICENSE) file. See [CHANGELOG.md](https://github.com/webshoppe/tinkerVault/blob/main/apps/markdown-viewer/CHANGELOG.md) for what changed between versions.

## More detail

See [docs/USER_GUIDE.md](docs/USER_GUIDE.md) in this folder for a fuller walkthrough, or the current [`src/DEV_GUIDE.md`](https://github.com/webshoppe/tinkerVault/blob/main/apps/markdown-viewer/src/DEV_GUIDE.md) if you want to read or modify the source. For the story of how this got built, see [PROJECT_SUMMARY.md](https://github.com/webshoppe/tinkerVault/blob/main/apps/markdown-viewer/PROJECT_SUMMARY.md).
