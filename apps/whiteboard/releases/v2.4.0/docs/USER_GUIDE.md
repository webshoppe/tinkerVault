# Whiteboard User Guide

**Version 2.4.0**

This guide covers every board type and the multi-board, save, search, paste, history, and trash features. The app is a single file: `whiteboard.html`. Double-click it to open. No install and no internet required.

---

## Getting started

1. Double-click `whiteboard.html` (Chrome, Edge, Firefox, or Safari).
2. You start with one **Paint** board by default (or whatever you last had open, including zoom and scroll).
3. Work is saved automatically into this browser’s **local storage** for this file.
4. Use **?** in the toolbar for About (version **v2.4.0** and a feature summary). The toolbar also shows a version badge and a **storage meter**.

> Tip: If you move the HTML file to a different folder or open it in a different browser, you may get separate storage. Use **File → Export all boards (JSON)** before moving, then import on the other side.

---

## Multi-board management

Boards are independent notebooks. Each has a **name**, a fixed **type**, and its own **content**.

| Action | How |
|--------|-----|
| Create | **+ New** → name + type → **Create** |
| Switch | **Board** dropdown |
| Rename / duplicate | **▾** menu next to **+ New** |
| Time machine | **▾** → **Time machine…** (last 10 snapshots) |
| Delete | **▾** → **Move to trash…** (recoverable for 7 days; cannot trash the last board) |

---

## Five board types

### Paint

Freehand drawing on a full-area canvas. Color swatches, brush S/M/L, Clear, Undo/Redo. Export as PNG via **File**. Large drawings may be auto-downscaled when storage is tight (see storage meter).

### Notepad (sticky notes)

| Action | How |
|--------|-----|
| Add | **+ Add note** |
| Edit | Click the note body and type |
| Move / resize | Header drag · corner handle |
| Color | 🎨 cycles colors (manual pick pauses auto-color until trigger text changes) |
| Checklists | Lines starting with `-`, `*`, or `[ ]` become checkboxes when not editing |
| Auto-color | `TODO`/`URGENT`/`ASAP` → red; `DONE`/`✓` → green; many `?` → yellow (editable under ⚙) |
| #tags | Tags appear in the tag bar; click a tag to filter |
| Send to Kanban | 📋 on the note header (creates a card; sticky stays) |
| Whole board → Kanban | Toolbar **→ Kanban board** (new board; stickies unchanged) |

### Kanban

Default columns: **To Do**, **In Progress**, **Done**. Add/rename/delete columns; **+ Card** / **+ Column** in the toolbar.  
**Drag cards from the “⋮⋮ drag” grip** at the top of each card (not from the text field). Drop on another column or another card to reorder. Column headers show live card counts (WIP). Card text uses the same checklist, auto-color, and #tag behavior as stickies.

### Annotate

Paste (`Ctrl+V`), drop, or **Open image…** a screenshot, then highlight, arrow, text label, redact, or crop. Undo/redo and PNG export. Screen capture is optional and may not work under `file://`.

### Wordpad

Rich text: bold/italic/underline, headings, lists, colors, highlights. Undo/redo. Export as plain text via **File**.

---

## Global search (`Ctrl/Cmd+K`)

Search box in the toolbar. Finds board names, sticky text, Wordpad text, Annotate text labels, and Kanban card text. Select a result to open that board (and highlight the note/card when possible).

---

## Command palette (`Ctrl/Cmd+/`)

Fuzzy list of **every board** and real actions (new board, theme, export, time machine, trash, settings, add note, new Kanban, etc.). Distinct from content search.

---

## Smart paste and drop

When focus is **not** inside a text field, paste/drop is routed:

| Content | Behavior |
|---------|----------|
| Image / image file | New **Annotate** board (images compressed like other large boards) |
| `.txt` / `.md` | New **Wordpad** |
| Single URL only | Link card sticky (domain label; no network fetch) |
| Short text (≤280 chars) | Sticky note |
| Long text (≥900 chars) | Wordpad |
| Medium text (281-899) | Toast: sticky vs Wordpad |
| Tabular text | Toast: Wordpad table option |
| `.json` | Toast: import as board(s); never silent |

Multiple image files → one Annotate board each. Click the board background before pasting if a field has focus.

---

## Time machine

Board menu **Time machine…** (also in the command palette). Lists up to **10** snapshots with time and a short content summary. Snapshots are taken when content meaningfully changes and at least **2 minutes** have passed (not every autosave). Restoring saves the current state first so you can undo the restore from the same list.

---

## Trash

Deleted boards go to **Trash** (File menu or palette), not permanent wipe. Restore within **7 days**; older trash is purged on load. Permanent purge is the only silent permanent delete in the app, and only after that window.

---

## Import & export

**File** menu:

- Export current / all boards as **JSON**
- **Import JSON** (appends boards; does not wipe the library)
- Export **PNG** (Paint / Annotate)
- Export **plain text** (Wordpad)
- **Trash…**

---

## Theme, zoom, settings

- **◐**, light / dark  
- Zoom controls in the toolbar (`Ctrl++` / `−` / `0`)  
- **⚙**, auto-color keyword map  
- Storage **%** meter, includes live boards, history, and trash  

---

## Keyboard reference

| Shortcut | Action |
|----------|--------|
| `Ctrl+Z` | Undo |
| `Ctrl+Shift+Z` or `Ctrl+Y` | Redo |
| `Ctrl+K` | Global search |
| `Ctrl+/` | Command palette |
| `Ctrl+V` | Smart paste (or native paste in a focused field) |
| `Ctrl++` / `−` / `0` | Zoom |

On Mac, use `Cmd` instead of `Ctrl`.

---

## Offline & privacy

No CDN, no accounts, no uploads. Data stays in **localStorage** under key `whiteboard-app-v2` for this browser and file path. Corrupt saved data is quarantined and the app opens cleanly with an export option for the raw blob.

---

## Known limitations

1. Board type is fixed at creation  
2. localStorage has a finite quota (large PNG history can fill it; meter + compression + snapshot trim help)  
3. Annotate “blur” is solid redaction  
4. Wordpad uses `document.execCommand`  
5. Kanban drag starts from the grip, not the card textarea  
6. JSON import appends; it does not replace the library  
7. Screen capture on Annotate is best-effort under `file://`  
