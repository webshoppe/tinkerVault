# Whiteboard User Guide

**Version 2.4.0**

This guide covers every board type and the multi-board, search, history, and export features. The app is a single file: `whiteboard.html`. Double-click it to open. No install and no internet required.

---

## Getting started

1. Double-click `whiteboard.html` (Chrome, Edge, Brave, or Firefox recommended).
2. You start with one board by default (or whatever you last had open, the app restores your last board, zoom, scroll position, and last-used tool).
3. Work is saved automatically into this browser's **local storage** for this file.
4. Use **?** in the toolbar for a short About panel, including the version number (**v2.4.0**).
5. There is no File menu in this app. Reach import/export, Trash, and most non-toolbar actions through the **command palette** (`Ctrl/Cmd+/`) instead, see below.

> Tip: If you move the HTML file to a different folder or open it in a different browser, you may get a separate empty storage. Use the command palette's **Export all boards (JSON)** before moving, then import on the other side.

---

## Multi-board management

Boards are independent notebooks. Each has:

- A **name**
- A fixed **type** (Paint, Notepad, Annotate, Wordpad, or Kanban)
- Its own **content** and settings

### Create

Click **+ New**, enter a name, pick a type, then **Create**.

### Switch

Use the **Board** dropdown, or `Ctrl/Cmd+K` to search for it by name, or `Ctrl/Cmd+/` and pick it from the palette.

### Rename / duplicate / delete

Open the **▾** menu next to **+ New**:

- **Rename…**, change the display name
- **Duplicate**, full copy of the board and its content
- **Delete board…**, confirmation required; you cannot delete the last remaining board. Deleting doesn't destroy the board immediately; it moves to **Trash** (see below) for 7 days.

Switching boards briefly fades the workspace while the previous board is saved and the next one loads.

---

## Finding things fast

### Global search, `Ctrl/Cmd+K`

Searches across every board's text at once: sticky note text, Wordpad content, Annotate text labels, Kanban card text, and board names. Select a result to jump straight to it.

### Command palette, `Ctrl/Cmd+/`

A separate, fuzzy-searchable list of every board (by name) and every available action, switch template, toggle theme, export, open the time machine, open Trash, promote a board to Kanban, and more. Deliberately a different shortcut from search, so the two don't collide: `Ctrl/Cmd+K` is always content search, `Ctrl/Cmd+/` is always the action/board palette.

---

## Paint

Freehand drawing on a full-area canvas.

| Control | Purpose |
|---------|---------|
| Color swatches | Black, red, green, blue, orange, white |
| Brush S / M / L | Stroke thickness |
| Clear | Wipe the canvas (asks for confirmation) |
| Undo / Redo | Step through drawing history |

**Tips**

- Mouse or touch both work.
- `Ctrl+Z` / `Ctrl+Shift+Z` undo and redo.
- Export as PNG from the command palette.
- Dense drawings stored as PNG can use a fair amount of browser storage; the quota meter (see below) auto-downscales large images before that becomes a problem.

---

## Notepad

A board of sticky notes on a grid background, with three "auto" behaviors layered on top of plain text; none of them need to be turned on, they just work as you type.

| Action | How |
|--------|-----|
| Add a note | **+ Add note** |
| Edit text | Click inside the note and type |
| Move | Drag the note header |
| Resize | Drag the corner handle |
| Change color | 🎨 on the note header, cycles colors, and manually picking one turns off auto-color for that note until its text changes again |
| Delete one | ✕ on the note header |
| Send to Kanban | 📋 on the note header, pick or create a Kanban board/column; the original sticky is never modified or deleted |
| Delete all | **Delete all** in the toolbar (confirmation) |

### Auto-checklist

Lines starting with `-`, `*`, `[ ]`, or `[x]` render as real checkboxes instead of plain text. Checking one strikes that line through and updates the underlying text.

### Auto-color

Keywords in a note's text set its background color automatically: TODO/URGENT/ASAP → red, DONE/✓ → green, question-heavy text → yellow. The keyword-to-color map is editable in Settings (⚙) if you want different triggers.

### Auto-tags

Any `#tag` written in a note's text is collected into a tag bar for the board. Click a tag to filter the board to just that tag's notes; click it again to clear the filter.

Notes are saved with position, size, text, color, and tags.

---

## Annotate (image markup)

Built for screenshots and pictures you already have; not for requiring a server or special install.

### Loading an image (primary, reliable offline path)

Any of these work from a double-clicked `file://` page:

1. **Paste**, copy a screenshot (or image) and press `Ctrl+V` while the Annotate board is open, or anywhere on any board (smart paste routes it here automatically)
2. **Drag and drop**, drop an image file onto the board area (also works from anywhere, smart routing sends it to a new Annotate item)
3. **Open image…**, pick a file from disk

Until an image is loaded, a short hint explains paste and drop.

### Annotation tools

| Tool | Behavior |
|------|----------|
| **Highlight** | Semi-transparent yellow freehand marker |
| **Arrow** | Click-drag to place an arrow |
| **Text** | Click, then type a label in the prompt |
| **Blur/Redact** | Drag a rectangle; region is solidly redacted (privacy-safe overlay) |
| **Crop** | Drag a rectangle; canvas is cropped to that region |

Color swatches apply to arrows and text labels. **Undo** / **Redo** walk annotation history. **Clear image** removes the picture and markup (with confirmation).

Export as PNG from the command palette saves the annotated result.

### Screen capture (bonus only)

The toolbar includes **Capture screen…** when the browser exposes the `getDisplayMedia` API.

| Expectation | Reality |
|-------------|---------|
| Ideal use | Pick a window or screen, capture a frame into the Annotate canvas |
| Under `file://` | API may be present (especially in Chrome), but permission UI, policy, or headless environments often block or reject the call |
| If it fails | Status message explains the error and points you back to **paste or drop** |
| Design intent | Capture is optional. **Paste / drop / open** are the supported everyday workflow |

Do not rely on screen capture for critical work without verifying it in your browser. After a successful OS screenshot, paste is usually faster and more reliable.

---

## Wordpad

A simple rich-text page for notes and write-ups.

| Control | Purpose |
|---------|---------|
| B / I / U | Bold, italic, underline |
| H1 / H2 / P | Heading 1, heading 2, paragraph |
| • List / 1. List | Bulleted / numbered lists |
| Text colors | Foreground color of selection |
| HL colors | Highlight / background of selection |
| Undo / Redo | Document history |

Export as text (plain `.txt`, formatting flattened) from the command palette.

Shortcuts: `Ctrl+Z` / `Ctrl+Shift+Z` for undo/redo while editing.

**Note:** pasting while your cursor is inside an open Wordpad document types normally there; smart paste routing (see below) intentionally steps aside when you're actively editing text, rather than second-guessing what you're typing.

---

## Kanban

Columns and draggable cards, for tracking work through stages.

| Action | How |
|--------|-----|
| Add a card | **+ Card** in the top toolbar (lands in the first column) |
| Move a card | Drag it by the **`⋮⋮` handle** at the top of the card; not the text body, which stays reserved for clicking into to edit |
| Reorder within a column | Same handle, drop it above/below another card in the same column |
| Add / rename / delete a column | Controls on the column header |
| Edit card text | Click into the card body | 

New boards start with three columns: To Do, In Progress, Done. Card text gets the same auto-color, auto-checklist, and auto-tag treatment as sticky notes.

### Promoting stickies to Kanban

- **One sticky:** the 📋 icon on a sticky note's header. Pick an existing Kanban board/column, or create a new one on the spot. The original sticky is untouched.
- **A whole Notepad board:** from that board's toolbar, or the command palette ("Promote stickies to new Kanban…"). Every sticky on the board becomes a card on a brand-new Kanban board, in the first column. The original board and its stickies are completely untouched; this creates a new board, it doesn't convert the existing one.

---

## Smart paste and drop

Paste (`Ctrl/Cmd+V`) anywhere on a board, or drop a file, and the app figures out what to do with it:

**Routed silently, no prompt:**

| Input | Destination |
|---|---|
| Image (paste or file) | New Annotate item |
| `.txt` / `.md` file | New Wordpad page |
| Just a URL, alone | Link-card sticky showing the domain (no network fetch for a page title) |
| Text ≤280 characters | New sticky note |
| Text ≥900 characters | New Wordpad page |

**Offered as a choice, small popup, dismissable:**

| Input | Offer |
|---|---|
| Text 281-899 characters | Sticky note vs. Wordpad page |
| Tabular-looking text (consistent tab/comma columns) | Add as a table in a new Wordpad page (there's no separate spreadsheet template) |
| `.json` file dropped | Import as a new board; never done silently |

If you're estimating whether a paste will land as a sticky or a Wordpad page, go by character count, not how long it looks; an ordinary paragraph clears 900 characters faster than it feels like it should.

---

## History and safety

### Time machine

Every board keeps up to 10 rolling snapshots, taken periodically as you work (not on every single autosave, so the window stays useful rather than filling up with near-duplicates). Open the picker from the board's menu or the command palette to see a timestamped list and restore an earlier state. Restoring is itself safe: your current state gets snapshotted first, so restoring an older version never silently throws away what you had.

### Trash

Deleting a board moves it to Trash instead of destroying it. Reach Trash through the command palette (`Ctrl/Cmd+/`, search "trash"). From there you can restore any trashed board back to normal. Trashed boards auto-purge after 7 days. You can never delete the very last remaining board.

### Corrupt-state recovery

If saved data fails validation when the app loads, it doesn't lose your work silently or show a blank white screen; the bad data is quarantined into a downloadable export, and the app opens with a clean default board so you're never locked out.

### Multi-tab safety

If the same file is open in two browser tabs at once, editing in one shows a banner in the other: reload to see the other tab's changes, or keep editing here (autosave pauses until you choose, so nothing gets silently overwritten).

### Quota meter

A storage-usage bar in the toolbar. It accounts for live boards, time-machine snapshots, and Trash together, not just what's currently open. Large Paint/Annotate images auto-downscale before they'd push you over the browser's storage limit; if things are ever tight, trimming older snapshots is tried before anything more aggressive.

---

## Import and export

Reach these from the **command palette** (`Ctrl/Cmd+/`):

| Action | Result |
|-----------|--------|
| Export current board (JSON) | One board, including type and content |
| Export all boards (JSON) | Full library backup |
| Import JSON… | Adds boards from a previous export (does **not** delete existing boards) |
| Export as PNG… | Paint or Annotate only |
| Export as text… | Wordpad only |

JSON exports use format `whiteboard-export`. Older single-template saves from earlier app versions can still be imported when recognized.

---

## Theme

**◐** toggles light and dark themes. The choice is remembered with your boards.

---

## Automatic saving

- Saves after edits (short debounce), when switching boards, periodically, and when closing the tab.
- If storage is full (often from large Paint/Annotate images), the status line shows a save error; the quota meter and auto-downscale/auto-trim behavior above exist specifically to keep this from happening in normal use.

---

## Keyboard reference

| Shortcut | Where | Action |
|----------|-------|--------|
| `Ctrl/Cmd+K` | Anywhere | Global search |
| `Ctrl/Cmd+/` | Anywhere | Command palette |
| `Ctrl+Z` | Paint, Annotate, Wordpad | Undo |
| `Ctrl+Shift+Z` | same | Redo |
| `Ctrl+Y` | same | Redo |
| `Ctrl+V` | Anywhere | Paste (smart-routed by content, see above) |
| `Esc` | Modals | Close dialog |

Use `Cmd` on macOS.

---

## Privacy and offline behavior

- No network calls at runtime; the HTML file does not load scripts or styles from the internet, and pasted URLs never get fetched for a title/preview.
- Content stays in the browser unless you export a file.
- Double-click / `file://` is the intended way to run the app.

### A note for privacy-focused browsers

If you're using Mullvad Browser, Tor Browser, or Firefox with strict tracking protection, you may see a prompt asking to allow "HTML5 canvas image data." That's your browser's anti-fingerprinting protection, not a problem with the app: Whiteboard reads canvas data only to save your own drawings, and never sends anything anywhere, this app makes zero network calls of any kind. Allow is safe to click.

**Known issue:** part of the board area may briefly show a colored striped placeholder pattern (green in Firefox, red in Mullvad Browser) instead of rendering normally while this permission is unresolved. In Firefox it clears up once the prompt appears; in Mullvad Browser specifically, it has been observed to persist even after the prompt is dismissed. The two are likely separate bugs rather than the same root cause. Being investigated.

---

## Troubleshooting

| Problem | What to try |
|---------|-------------|
| Work "disappeared" after moving the file | Different path/browser = different storage. Import a JSON backup if you have one |
| Can't find "File" menu | There isn't one; use the command palette (`Ctrl/Cmd+/`) for import/export, Trash, and most non-toolbar actions |
| Annotate capture fails | Use paste or drop instead |
| PNG/text export does nothing useful | Wrong board type, PNG needs Paint/Annotate; text needs Wordpad |
| Cannot delete a board | You must keep at least one board |
| Deleted a board by mistake | Command palette → Trash → restore, within 7 days |
| Save failed | Storage full; the quota meter shows how close you are; export important boards, or check the time machine for older snapshots you can trim |
| Paste looks like it went to the wrong place | If your cursor was already inside a Wordpad document, paste there always types normally rather than smart-routing, on purpose. Otherwise, check the actual character count if it's a length-based routing question, see Smart paste above |
| Kanban card won't drag | Grab it by the `⋮⋮` handle at the top, not the text body |

---

## Version

This guide matches **Whiteboard 2.4.0**. The version appears in the toolbar (`v2.4.0`), the About panel (**?**), and the `VERSION` file in the release package.
