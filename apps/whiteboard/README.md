# Whiteboard

![Release](https://img.shields.io/github/v/release/webshoppe/tinkerVault?filter=whiteboard*&label=release)

<img src="icon.svg" width="48" height="48" alt="Whiteboard icon" />

**Version 2.4.0**, a portable, offline whiteboard in a single HTML file.

Open it by **double-clicking `whiteboard.html`**. There is nothing to install, no account, and no server. Your work stays in the browser on your computer.

---

## What it is

Whiteboard is a clipboard-style board app with five board types, multiple named boards, cross-board search, a command palette, content-aware paste/drop routing, undo-grade history (per-board snapshots plus a soft-delete trash), and import/export for backups. It runs fully offline from a double-clicked file, no network calls at runtime, ever.

## Getting Started

**1. Download it**

Go to the [Whiteboard release page](https://github.com/webshoppe/tinkerVault/releases/tag/whiteboard-v2.4.0) and click `whiteboard-v2.4.0.zip` under Assets. It downloads like any other file.

**2. Unzip it**

Find the downloaded file (usually in your Downloads folder), right-click it, choose **Extract All**, and pick somewhere easy to find, like your Desktop. You'll end up with a folder called `Whiteboard` containing `whiteboard.html` and a couple of other files.

**3. Use it**

Double-click `whiteboard.html`. It opens in your browser and you're ready to go. Everything you draw or write saves automatically, right there on your computer, nothing is uploaded anywhere.

**Want it to open in its own window instead of a browser tab?**

See [Making an App Feel Like a Real Desktop App](../../DESKTOP-LAUNCHER.md). When it asks for the app's folder and filename, use: `Desktop/Whiteboard/whiteboard.html` (or wherever you actually put it).

## How to open

If you'd rather work from the source in this repo instead of a downloaded zip:

1. Find `whiteboard.html` on your computer.
2. Double-click it (or drag it into Chrome, Edge, Brave, or Firefox).
3. Start working. Changes save automatically.

You can copy the file to a USB drive or another machine and open it the same way.

### Optional: a taskbar/desktop icon instead of a browser tab

Point a Chromium-based browser (Brave, Chrome, Edge; not Firefox, it doesn't support this) at the file with the `--app` flag to get a chromeless app window instead of a normal tab, then pin the shortcut:

```
"C:\path\to\brave.exe" --app="file:///C:/path/to/whiteboard.html"
```

Set the shortcut's icon via Properties → Change Icon to a `.ico` version of `icon.svg` (Explorer's icon picker doesn't accept raw `.svg`). Optional extra isolation: add `--user-data-dir="C:\path\to\a\dedicated\profile"` so this app's local storage never shares an origin with anything else you might open as `file://` in the same browser profile.

## Board types

| Type | What it's for |
|------|----------------|
| **Paint** | Freehand drawing with colors and brush sizes |
| **Notepad** | Sticky notes you can drag, resize, and edit, with live checklists, auto-color, and `#tags` |
| **Annotate** | Paste or drop a screenshot/image, then highlight, arrow, label, redact, or crop |
| **Wordpad** | A simple rich-text document (bold, lists, headings, colors) |
| **Kanban** | Columns and draggable cards, with WIP counts. Promote a sticky note (or a whole Notepad board) into a Kanban board without touching the original |

When you create a board you choose its type; that type stays fixed for that board.

## Managing boards

- **Board** dropdown, switch between saved boards
- **+ New**, create a named board and pick a type
- **▾** menu, rename, duplicate, or delete (deleted boards go to Trash, not gone immediately, see below)

Each board keeps its own content independently.

## Finding things

- **`Ctrl/Cmd+K`**, global search across every board's text (sticky notes, Wordpad, Annotate labels, Kanban cards, board names). Jump straight to a result.
- **`Ctrl/Cmd+/`**, command palette: fuzzy-search every board by name and every available action (switch template, toggle theme, export, open time machine, open trash, promote to Kanban, and more). This is currently the only way to reach some actions (Trash, for one); there's no separate File menu in the app.

## Smart paste and drop

Paste (`Ctrl/Cmd+V`) or drop a file anywhere on a board and the app routes it automatically:

| You give it | It becomes |
|---|---|
| An image (paste or file) | A new Annotate item |
| A `.txt` or `.md` file | A new Wordpad page |
| Just a URL, nothing else | A clickable link card showing the domain (no page-title fetch; that would need network access, which this app never uses) |
| Short text (≤280 characters) | A new sticky note |
| Long text (≥900 characters) | A new Wordpad page |
| Text in between (281-899 characters), or text that looks like a table | A small popup asks which you'd prefer, rather than guessing |
| A `.json` file | A popup offers to import it as a board; never done silently |

Note: pasting while your cursor is already inside an open Wordpad document types normally there rather than triggering smart routing; that's intentional, not a bug.

## Notepad extras (auto-detected, no setup)

- Lines starting with `-`, `*`, or `[ ]` become live checkboxes; checking one strikes the line through.
- Keywords color a note automatically: TODO/URGENT/ASAP → red, DONE/✓ → green, question-heavy text → yellow. Manually changing a note's color turns off auto-color for it until the text changes again. The keyword-to-color map is editable in Settings.
- Any `#tag` in a note's text is collected into a per-board tag bar; click a tag to filter, click again to clear.

These same detectors also apply to Kanban card text.

## Kanban

- New boards start with three columns (To Do / In Progress / Done); add, rename, or delete columns after that.
- Drag a card by its **`⋮⋮` handle** at the top of the card; not the text body, which stays reserved for editing.
- **Send a sticky to Kanban:** the 📋 icon on a sticky note's header lets you pick (or create) a Kanban board and column. The original sticky is never modified or deleted.
- **Promote a whole Notepad board:** from the board's toolbar or the command palette, turn every sticky on the current board into cards on a brand-new Kanban board. The original board is untouched.

## History and safety

- **Time machine:** every board keeps up to 10 rolling snapshots (not one per autosave, spaced out so the window stays useful). Open it from the board's menu or the command palette to restore an older state. Restoring always snapshots your *current* state first, so a restore can't silently erase what you had.
- **Trash:** deleting a board moves it to Trash instead of destroying it immediately. Reach Trash via the command palette (`Ctrl/Cmd+/` → "trash"). Trashed boards restore in one click, and auto-purge after 7 days. You can't remove the last remaining board.
- **Corrupt-state recovery:** if saved data ever fails validation on load, it's quarantined to a downloadable export instead of losing it or white-screening, and the app falls back to a clean board so it still opens.
- **Multi-tab safety:** if you have the same file open in two tabs, a banner appears letting you choose to reload the other tab's changes or keep editing here; autosave pauses until you decide.
- **Quota meter:** a visible storage indicator in the toolbar. Large Paint/Annotate images auto-downscale before they'd hit the browser's storage limit; the meter also counts snapshot and trash storage, not just live boards.

## Import & export

Reach these via the command palette (`Ctrl/Cmd+/`):

| Action | Result |
|-----------|--------|
| Export current board (JSON) | One board, including type and content |
| Export all boards (JSON) | Full library backup |
| Import JSON… | Adds boards from a previous export (does **not** delete existing boards) |
| Export as PNG… | Paint or Annotate only |
| Export as text… | Wordpad only |

## Keyboard shortcuts

| Shortcut | Action |
|----------|--------|
| `Ctrl/Cmd+K` | Global search |
| `Ctrl/Cmd+/` | Command palette |
| `Ctrl+Z` | Undo (Paint, Annotate, Wordpad) |
| `Ctrl+Shift+Z` or `Ctrl+Y` | Redo |
| `Ctrl+V` | Paste (routes automatically by content, see Smart paste above) |

On a Mac, use `Cmd` instead of `Ctrl`.

## Theme & help

- **◐**, light / dark theme
- **?**, About panel (version and short help)
- Version also shows as **v2.4.0** in the toolbar

## Privacy-focused browsers

Using Mullvad Browser, Tor Browser, or Firefox with strict tracking protection? You may see a prompt asking to allow "HTML5 canvas image data," that's expected, not a bug. See [docs/USER_GUIDE.md](docs/USER_GUIDE.md#a-note-for-privacy-focused-browsers) for what it means and a known display quirk on some of those browsers.

## Files in this package

```
whiteboard.html      ← the app (this is all you need to run it), always the current version, 2.4.0
icon.svg             ← icon for docs / shortcuts
VERSION               ← 2.4.0
README.md            ← this file
docs/USER_GUIDE.md
docs/DEV_GUIDE.md
PROJECT_SUMMARY.md    ← the build story, in two voices, covering v1 and v2
build-process/         ← verification scripts, reports, and screenshots
releases/              ← self-contained snapshot of every shipped version, including v1.0.0
  v1.0.0/
  v2.4.0/              ← same file as the root copy above
```

Everything at the root of this package always tracks the newest release. If you want an older version, or want both side by side, see [releases/](releases/).

## License

License is intentionally left for the project owner to choose; no LICENSE file is included here.

## More detail

See [docs/USER_GUIDE.md](docs/USER_GUIDE.md) for a fuller walkthrough of every feature, or [docs/DEV_GUIDE.md](docs/DEV_GUIDE.md) if you want to read or modify the source.

For the story of how this got built, both what the agent that built it was thinking and what an outside observer verified along the way, across both the original four-template build and the five-tier v2 pass, see [PROJECT_SUMMARY.md](PROJECT_SUMMARY.md).
