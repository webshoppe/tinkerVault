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
- **Edit mode.** Edit the text directly in the rendered v