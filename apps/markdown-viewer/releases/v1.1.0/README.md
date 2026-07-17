# Markdown Viewer

![Release](https://img.shields.io/github/v/release/webshoppe/tinkerVault?filter=markdown-viewer*&label=release)

<img src="icon.svg" width="80" alt="Markdown Viewer logo icon">

**Version 1.1.0**, a free, offline Markdown viewer and light editor in a single HTML file.

Open it by **double-clicking `index.html`**. There is nothing to install, no account, and no server. Your files never leave your computer.

Part of the [tinkerVault](../../) collection of offline-first single-file apps. See the [root README](../../README.md) for what else lives here.

---

## What it is

Markdown Viewer turns `.md` files into a clean, readable, GitHub-style page: tables, checklists, syntax-highlighted code, nested lists, blockquotes, a table of contents, and in-document search. Open several files at once and switch between them in tabs. Edit the rendered text directly, save your edits as a new file, copy the raw Markdown or a single code block, and export the current document as one portable standalone HTML file you can hand to someone else. It runs fully offline, no network calls at runtime, ever.

## Getting Started

**1. Download it**

Go to the [Markdown Viewer release page](https://github.com/webshoppe/tinkerVault/releases/tag/markdown-viewer-v1.1.0) and download the zip under Assets. It downloads like any other file.

**2. Unzip it**

Find the downloaded file (usually in your Downloads folder), right-click it, choose **Extract All**, and pick somewhere easy to find, like your Desktop. You'll end up with a folder containing `index.html` and a couple of other files.

**3. Use it**

Double-click `index.html`. It opens in your browser and you're ready to go.

**Want it to open in its own window instead of a browser tab?**

See [Making an App Feel Like a Real Desktop App](../../DESKTOP-LAUNCHER.md). When it asks for the app's folder and filename, use: `Desktop/MarkdownViewer/index.html` (or wherever you actually put it).

## How to open

If you'd rather work from the source in this repo instead of a downloaded zip:

1. Find `index.html` on your computer.
2. Double-click it (or drag it into Chrome, Edge, Brave, or Firefox).
3. Click **📂 Open**, or drag a `.md` file (or several at once) straight onto the page.

You can copy the file to a USB drive or another machine and open it the same way.

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

Full walkthrough, in the order a first-time user meets each feature: [`docs/USER_GUIDE.md`](docs/USER_GUIDE.md).

## What's new in v1.1.0

Fixed a real bug: two files with the same name from different folders used to collide into a single tab, each open file now gets its own tab, always. Added multi-file open, edit mode, Save, Copy raw Markdown, keyboard shortcuts, and per-tab scroll memory. Full history: [`CHANGELOG.md`](CHANGELOG.md).

## Older versions

Every past release stays available as a complete, independently runnable snapshot. Nothing is ever deleted.

- [`releases/v1.0.0/`](releases/v1.0.0/index.html), the original single-file-at-a-time release

## Files in this package

```
index.html          the app (this is all you need to run it), always the current version, 1.1.0
icon.svg
VERSION
README.md            this file
CHANGELOG.md          what changed between versions
CONTRIBUTING.md
docs/
  USER_GUIDE.md        every feature, in first-time-user order
src/
  DEV_GUIDE.md          how to rebuild from source
  (template.html, app.js, assemble.py, and vendored libraries under vendor/)
test/
  test.md               sample file exercising Tier 1 rendering features
  fixtures/              same-name-collision test fixtures (projectA/, projectB/)
PROJECT_SUMMARY.md      the build story
releases/
  v1.0.0/                 the original release, preserved as-is
```

Everything at the root of this package always tracks the newest release. If you want an older version, see [releases/](releases/).

## License

MIT, see the repo's [LICENSE](../../LICENSE) file. See also [CHANGELOG.md](CHANGELOG.md) for what changed between versions.

## More detail

See [docs/USER_GUIDE.md](docs/USER_GUIDE.md) for a fuller walkthrough of every feature, or [src/DEV_GUIDE.md](src/DEV_GUIDE.md) if you want to read or modify the source. For the story of how this got built, see [PROJECT_SUMMARY.md](PROJECT_SUMMARY.md).
