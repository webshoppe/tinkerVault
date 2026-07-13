# Whiteboard User Guide

**Version 1.0.0**

This guide covers every board type and the multi-board, save, and export features. The app is a single file: `whiteboard.html`. Double-click it to open. No install and no internet required.

---

## Getting started

1. Double-click `whiteboard.html` (Chrome, Edge, or Firefox recommended).
2. You start with one **Paint** board by default (or whatever you last had open).
3. Work is saved automatically into this browser’s **local storage** for this file.
4. Use **?** in the toolbar for a short About panel, including the version number (**v1.0.0**).

> Tip: If you move the HTML file to a different folder or open it in a different browser, you may get a separate empty storage. Use **File → Export all boards (JSON)** before moving, then import on the other side.

---

## Multi-board management

Boards are independent notebooks. Each has:

- A **name**
- A fixed **type** (Paint, Notepad, Annotate, or Wordpad)
- Its own **content** and settings

### Create

Click **+ New**, enter a name, pick a type, then **Create**.

### Switch

Use the **Board** dropdown. The type badge next to it shows the current board’s type.

### Rename / duplicate / delete

Open the **▾** menu next to **+ New**:

- **Rename…**, change the display name  
- **Duplicate**, full copy of the board and its content  
- **Delete board…**, confirmation required; you cannot delete the last remaining board  

Switching boards briefly fades the workspace while the previous board is saved and the next one loads.

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
- **File → Export as PNG** downloads the canvas image.
- Dense drawings stored as PNG can use a fair amount of browser storage.

---

## Notepad

A board of sticky notes on a grid background.

| Action | How |
|--------|-----|
| Add a note | **+ Add note** |
| Edit text | Click inside the note and type |
| Move | Drag the note header (“drag”) |
| Resize | Drag the corner handle |
| Change color | 🎨 on the note header (cycles yellow, pink, blue, green, orange) |
| Delete one | ✕ on the note header |
| Delete all | **Delete all** in the toolbar (confirmation) |

Notes are saved with position, size, text, and color.

---

## Annotate (image markup)

Built for screenshots and pictures you already have; not for requiring a server or special install.

### Loading an image (primary, reliable offline path)

Any of these work from a double-clicked `file://` page:

1. **Paste**, copy a screenshot (or image) and press `Ctrl+V` while the Annotate board is open  
2. **Drag and drop**, drop an image file onto the board area  
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

**File → Export as PNG** saves the annotated result.

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

**File → Export as text…** writes a plain `.txt` file (formatting flattened to text).

Shortcuts: `Ctrl+Z` / `Ctrl+Shift+Z` for undo/redo while editing.

---

## Import and export (File menu)

| Menu item | Result |
|-----------|--------|
| Export current board (JSON) | One board, including type and content |
| Export all boards (JSON) | Full library backup |
| Import JSON… | Adds boards from a previous export (does **not** delete existing boards) |
| Export as PNG… | Paint or Annotate only |
| Export as text… | Wordpad only |

JSON exports use format `whiteboard-export` version 2. Older single-template saves from earlier app versions can still be imported when recognized.

---

## Theme

**◐** toggles light and dark themes. The choice is remembered with your boards.

---

## Automatic saving

- Saves after edits (short debounce), when switching boards, periodically, and when closing the tab.
- Storage key: `whiteboard-app-v2` in `localStorage`.
- If storage is full (often from large Paint/Annotate images), the status line shows a save error; export JSON/PNG and clear unused boards.

---

## Keyboard reference

| Shortcut | Where | Action |
|----------|-------|--------|
| `Ctrl+Z` | Paint, Annotate, Wordpad | Undo |
| `Ctrl+Shift+Z` | same | Redo |
| `Ctrl+Y` | same | Redo |
| `Ctrl+V` | Annotate | Paste image from clipboard |
| `Esc` | Modals | Close dialog |

Use `Cmd` on macOS.

---

## Privacy and offline behavior

- No network calls at runtime; the HTML file does not load scripts or styles from the internet.
- Content stays in the browser unless you export a file.
- Double-click / `file://` is the intended way to run the app.

---

## Troubleshooting

| Problem | What to try |
|---------|-------------|
| Work “disappeared” after moving the file | Different path/browser = different storage. Import a JSON backup if you have one |
| Annotate capture fails | Use paste or drop instead |
| PNG/text export does nothing useful | Wrong board type, PNG needs Paint/Annotate; text needs Wordpad |
| Cannot delete a board | You must keep at least one board |
| Save failed | Storage full; export, then delete large boards |

---

## Version

This guide matches **Whiteboard 1.0.0**. The version appears in the toolbar (`v1.0.0`), the About panel (**?**), and the `VERSION` file in the release package.
