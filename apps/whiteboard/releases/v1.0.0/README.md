# Whiteboard

<img src="icon.svg" width="48" height="48" alt="Whiteboard icon" />

**Version 1.0.0**, a portable, offline whiteboard in a single HTML file.

Open it by **double-clicking `whiteboard.html`**. There is nothing to install, no account, and no server. Your work stays in the browser on your computer.

---

## What it is

Whiteboard is a small clipboard-style board app with four board types, multiple named boards, automatic saving, and import/export for backups. It runs fully offline from a double-clicked file.

## How to open

1. Find `whiteboard.html` on your computer.
2. Double-click it (or drag it into Chrome, Edge, or Firefox).
3. Start working. Changes save automatically.

You can copy the file to a USB drive or another machine and open it the same way.

## Board types

| Type | What it’s for |
|------|----------------|
| **Paint** | Freehand drawing with colors and brush sizes |
| **Notepad** | Sticky notes you can drag, resize, and edit |
| **Annotate** | Paste or drop a screenshot/image, then highlight, arrow, label, redact, or crop |
| **Wordpad** | A simple rich-text document (bold, lists, headings, colors) |

When you create a board you choose its type; that type stays fixed for that board.

## Managing boards

- **Board** dropdown, switch between saved boards  
- **+ New**, create a named board and pick a type  
- **▾** menu, rename, duplicate, or delete (delete asks for confirmation)  

Each board keeps its own content independently.

## Import & export

Open **File** in the toolbar:

- Export the current board, or all boards, as **JSON** (backup / move to another computer)
- **Import JSON** to bring boards back in (adds them; does not wipe what you already have)
- Export **PNG** from Paint or Annotate boards
- Export **plain text** from Wordpad boards

## Keyboard shortcuts

| Shortcut | Action |
|----------|--------|
| `Ctrl+Z` | Undo (Paint, Annotate, Wordpad) |
| `Ctrl+Shift+Z` or `Ctrl+Y` | Redo |
| `Ctrl+V` | Paste an image (on an Annotate board) |

On a Mac, use `Cmd` instead of `Ctrl`.

## Theme & help

- **◐**, light / dark theme  
- **?**, About panel (version and short help)  
- Version also shows as **v1.0.0** in the toolbar  

## Files in this package

```
whiteboard.html   ← the app (this is all you need to run it)
icon.svg          ← icon for docs / shortcuts
VERSION           ← 1.0.0
README.md         ← this file
docs/USER_GUIDE.md
```

## License

License is intentionally left for the project owner to choose; no LICENSE file is included here.

## More detail

See [docs/USER_GUIDE.md](docs/USER_GUIDE.md) for a fuller walkthrough of every feature.
