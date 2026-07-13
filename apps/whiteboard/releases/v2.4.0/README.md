# Whiteboard

<img src="icon.svg" width="48" height="48" alt="Whiteboard icon" />

**Version 2.4.0**, a portable, offline whiteboard in a single HTML file.

Open it by **double-clicking `whiteboard.html`**. There is nothing to install, no account, and no server. Your work stays in the browser on your computer.

---

## What it is

Whiteboard is a multi-board workspace with five board types, automatic saving, smart paste, global search, a command palette, board history (time machine), and a 7-day trash. It runs fully offline from one double-clicked file.

## How to open

1. Find `whiteboard.html` (project root for development, or the `release/` folder for distribution).
2. Double-click it (or open it in Chrome, Edge, Firefox, or Safari).
3. Start working. Changes save automatically to this browser’s local storage.

You can copy the file to a USB drive or another machine and open it the same way. Use **File → Export all boards (JSON)** before moving machines so you can import elsewhere.

## Board types

| Type | What it’s for |
|------|----------------|
| **Paint** | Freehand drawing with colors and brush sizes |
| **Notepad** | Sticky notes (drag, resize, checklists, auto-color, #tags) |
| **Kanban** | Columns and cards; drag from the **⋮⋮ drag** grip |
| **Annotate** | Paste or drop a screenshot/image, then mark it up |
| **Wordpad** | Simple rich-text document (bold, lists, headings, colors) |

When you create a board you choose its type; that type stays fixed for that board.

## Highlights (v2)

- **Multi-board**, create, switch, rename, duplicate; delete moves to **Trash** (7 days)
- **Global search** (`Ctrl/Cmd+K`), text across boards and templates  
- **Command palette** (`Ctrl/Cmd+/`), boards and actions  
- **Smart paste/drop**, images → Annotate; text routed by content; JSON import toast  
- **Time machine**, last 10 content snapshots per board  
- **Storage meter**, localStorage usage; large images compress when needed  
- **Session restore**, last board, zoom, scroll, tools  
- Light/dark theme · import/export (JSON, PNG, text)

## Managing boards

- **Board** dropdown, switch  
- **+ New**, name + type  
- **▾** menu, rename, duplicate, **Time machine…**, move to trash  
- **File**, export/import, **Trash…**

## Keyboard shortcuts

| Shortcut | Action |
|----------|--------|
| `Ctrl+Z` / `Ctrl+Shift+Z` or `Ctrl+Y` | Undo / Redo (Paint, Annotate, Wordpad) |
| `Ctrl+K` | Global content search |
| `Ctrl+/` | Command palette |
| `Ctrl+V` | Smart paste (when not typing in a field) |
| `Ctrl++` / `−` / `0` | Zoom in / out / reset |

On a Mac, use `Cmd` instead of `Ctrl`.

## Theme & help

- **◐**, light / dark theme  
- **?**, About panel (version and feature summary)  
- Toolbar badge shows **v2.4.0**

## Distribution package

For shipping, use the **`release/`** folder:

```
release/
  whiteboard.html   ← the app (this is all you need to run it)
  icon.svg
  VERSION           ← 2.4.0
  README.md
  docs/USER_GUIDE.md
```

## Development / verification

Optional Node + Playwright checks (not required to run the app):

```bash
npm run verify:v2-tier5          # last feature tier
node verify-v2-package.mjs       # cross-engine packaging suite
```

## License

License is intentionally left for the project owner to choose; no LICENSE file is included here.

## More detail

See [docs/USER_GUIDE.md](docs/USER_GUIDE.md) for a full walkthrough. See [RELEASE_NOTES.md](RELEASE_NOTES.md) for the v2.4.0 release arc.
