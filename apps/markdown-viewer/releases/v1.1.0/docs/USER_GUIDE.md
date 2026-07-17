# Markdown Viewer — User Guide

This guide is for someone opening the Markdown Viewer for the first time. It is written
from what the program actually does (the source in `src/`), not from assumptions.

---

## What it is

Markdown Viewer is a single `index.html` file that turns plain-text Markdown files
(`.md`, `.markdown`, or `.txt`) into a clean, formatted page **inside your own web
browser**. There is:

- **nothing to install,**
- **no account or sign-up,**
- **no server,** and
- **no internet connection required.**

Everything — the Markdown parser, the syntax highlighter, the application code, and all
styling — is bundled inside `index.html`. You can keep the file on a USB stick and open it
on any computer with a modern browser.

---

## How to open it

1. Open the `markdown-viewer-v2` folder and **double-click `index.html`**. It opens in your
   default browser. (You can also drag `index.html` onto an already-open browser window.)
2. The page shows a drop hint. **Open a Markdown file** in either of two ways:
   - Click **📂 Open** in the toolbar, pick one or more `.md` files, and confirm. Each file
     opens in its own tab.
   - **Drag a `.md` file** from your file manager and drop it anywhere on the page. You can
     drop several at once.

That is the entire setup. No preferences to set, no server to start.

> **Cross-platform note:** works on Windows, macOS, and Linux, in any reasonably recent
> Chrome, Edge, Firefox, or Safari.

---

## The toolbar

From left to right:

| Button | What it does |
|--------|--------------|
| 📂 Open | Opens the file picker. Supports selecting multiple files at once. |
| 🌙 Dark / ☀️ Light | Toggles the light/dark color theme. |
| ☰ TOC | Shows or hides the table of contents panel on the left. |
| 📝 Raw / 👁 Rendered | Toggles between the rendered (formatted) view and the raw Markdown source. |
| ✏️ Edit / 👁 View | Toggles **edit mode** (see below). |
| 📋 Copy MD | Copies the **raw Markdown of the active tab** to your clipboard. |
| 💾 Save | Downloads the active tab's **current content** (including any edits you've made) as a new `.md` file. |
| ⬇ Export | Exports the active tab's rendered view to a standalone, self-contained `.html` file. |
| Search box | Finds text inside the current document and highlights it. |

---

## Features

### Multiple tabs
Each file you open becomes its own tab at the top of the page. Click a tab to switch to it.
Close a tab with the **×** on its right edge.

- **Duplicate filenames are kept separate.** Two different files that happen to share a name
  (for example, two `README.md` from different folders) do **not** overwrite each other.
  The second one is labelled `README.md (2)`, the third `README.md (3)`, and so on. The
  real filename is still shown in the tab's hover tooltip.
- **Reopen last closed tab.** Closing a tab adds it to a history (up to 25 deep). Press
  **Ctrl/Cmd+Shift+T** to bring the most recently closed one back.

### Drag & drop
Drop one or many files anywhere on the page. Non-text files (images, PDFs, etc.) are
detected and **skipped with a short on-screen notice** instead of rendering garbage.

### Per-tab scroll memory
Each tab remembers how far you had scrolled. Switch away and back, and your place is
restored.

### Light / dark theme
The 🌙/☀️ button flips between GitHub-style light and dark themes. Your choice is remembered
across reloads.

### Table of contents
Headings in the document are collected into a clickable list on the left. Click an entry to
smooth-scroll to that section. The ☰ button hides or shows the panel.

### In-document search
Type in the search box (top-right). Matches inside the current document are highlighted;
press **Enter** to jump to the next/previous match (Shift+Enter goes backward). An `n/N`
counter shows your position.

### Copy buttons
- **Copy-code buttons** — every fenced code block gets a small **Copy** button (top-right of
  the block) that copies just that code to your clipboard.
- **📋 Copy MD** — copies the entire raw Markdown source of the *active tab* at once.

### Save (download your edits)
The 💾 **Save** button downloads the active tab's current text as a **new** `.md` file using
a standard client-side download — the default name is the tab's filename. This is how you
"get your edits out of the browser": any changes you made in edit mode are included. It
always creates a new file; it cannot overwrite the original (a `file://` page has no write
access to your disk).

### Export to standalone HTML
The ⬇ **Export** button produces a single self-contained `.html` file of the *rendered*
(already-formatted) view, including the current theme's syntax-highlighting colors. You can
open or share that file anywhere — no viewer required.

### Edit mode
The ✏️ **Edit** button switches the active tab into a raw-text editing area (a `<textarea>`).
While in edit mode the formatted view is hidden and you see only the source. As you type, the
app re-parses the Markdown internally, but **the visible rendered result only updates when
you switch back out of edit mode** (press ✏️ again, or 📝 Raw toggles the raw/rendered view).
Any edits you make persist on the tab and are included the next time you Copy MD, Save, or
Export.

### Keyboard shortcuts
| Shortcut | Action |
|----------|--------|
| Ctrl/Cmd+T | Open file picker |
| Ctrl/Cmd+W | Close the active tab |
| Ctrl/Cmd+Shift+T | Reopen the last closed tab |
| Ctrl/Cmd+Alt+Left / Right | Switch to previous / next tab (wraps around) |
| Ctrl/Cmd+PageUp / PageDown | Same as above: switch tabs |

### Browser tab title
The browser's own tab title shows the **bare filename of the active document** (e.g.
`README.md`). When no file is open it falls back to `Markdown Viewer`.

### What carries over between sessions
Open tabs (their names and content), the active theme, the TOC visibility, the raw/rendered
toggle, and edit mode are saved to your browser's `localStorage`, so reopening the viewer
restores your last session. (This is browser-local storage; it is not shared between
machines and is cleared if you clear site data for the `file://` origin.)

---

## Links inside a document
If a Markdown file contains a link to another `.md` file and you click it, the viewer tries
to load that file into the tab system. Under `file://`, browsers block local `fetch`, so the
link usually cannot be followed automatically; in that case a clear on-page notice tells you
to use **Open** or drag the file in instead.

---

## What it deliberately does NOT do

To keep the tool simple, offline, and dependency-free, these are intentionally out of scope:

- **No server, no install, no network.** It is one static file opened via `file://`.
- **No "open the same files next time by path."** Browsers won't let a `file://` page read
  arbitrary local file paths on its own, so the viewer restores *content* from `localStorage`
  but cannot re-open the *original files* automatically (that would need a server or the File
  System Access API).
- **No search across all open tabs.** The search box searches the *active* document only.
- **No Print / "Print to PDF" button.** Use your browser's own print dialog if you need a
  PDF — the Export feature produces a shareable HTML file instead.
- **No blank / "new document" tab button.** You open files; you don't start empty documents.
- **No syntax highlighting languages beyond the highlight.js "common" set** shipped in
  `vendor/highlight.min.js`.
- **No accounts, telemetry, or analytics.** Nothing leaves your machine.

---

## Troubleshooting

- **A dropped image / PDF did nothing.** That's expected — non-text files are skipped with a
  notice. Use a `.md` / `.markdown` / `.txt` file.
- **My edits didn't show while typing.** Expected: edit mode shows the raw textarea; the
  formatted view refreshes when you leave edit mode (or toggle raw/rendered).
- **Theme / tabs reset after clearing browser data.** That's `localStorage` being cleared —
  the viewer has nothing else to persist to.
- **A link to another `.md` showed a notice instead of opening.** Under `file://` the local
  fetch is blocked by the browser; open that file with the 📂 Open button or drag it in.
