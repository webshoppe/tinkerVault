# Dossier User Guide

**Version 2.0.1**

This guide covers every surface in Dossier. Double-click `Dossier.exe` to open. No account and no internet required for core features. **Ask dossier** needs a local agent you configure yourself.

---

## Getting started

1. Launch **Dossier.exe**.
2. Choose **Create / open folder…**, open a **recent** workspace, open something from a **collection**, or **Open last**.
3. That folder is your **dossier** (documents, notes, boards, and database all live there).
4. Use the left sidebar to switch surfaces. The version shows next to the wordmark and in the footer (**v2.0.1**).
5. Optional first-open tips appear once per surface; dismiss with **Got it** (or turn them off in Settings).
6. Optional: Settings → **Open last workspace on launch** (default off). When on, Dossier opens the last valid dossier immediately; **Close** still returns to the launcher.
7. Optional: on **Recent workspaces**, **Rename…** sets a display name (does not rename the folder on disk).

> Tip: Copy the whole dossier folder to back up or move a project. App Settings stay in `%APPDATA%\Dossier` and are not inside the folder.

---

## Documents

Import files into `documents/` and browse them with sort and detail preview.

| Action | How |
|--------|-----|
| Import | **+ Import files…** or drag files onto the window |
| Sources / bookmarks | Left **Sources** panel: bookmark external folders, browse them, multi-select import into this dossier |
| Sort | Topbar **Sort**: Name A–Z / Z–A, Date added newest / oldest (remembers for this session) |
| Open detail | Click a row |
| External app | **Open externally** (preferred app for that extension if set, else Windows default) |
| Choose / remember app | **Open with…** — default, choose once, **Always open .ext with…**, or clear preferred |
| Delete | **Delete** in the detail header (confirm if enabled) |
| Rescan | **Rescan** re-indexes the folder |

**Supported kinds**

- **Markdown / text** — full preview + search  
- **PDF** — best-effort text extract; image-only pages get a warning badge  
- **Word / Excel (.docx / .xlsx)** — best-effort text for search; preview is lossy; use **Open externally**  
- **OpenDocument (.odt / .ods)** — same honesty as Office; use **Open externally** for Writer/Calc  
- **Other** — stored as attachments  

Office lock files (`~$…`) are skipped on import.

---

## Notes

Markdown notes as real `.md` files under `notes/`. Live preview uses offline marked + highlight.js (no network).

| Action | How |
|--------|-----|
| New | **+ New note** |
| Modes | **Edit** (source only), **Split** (default: source left + live preview right), **Preview** (read-only render) |
| Edit | Title + body; auto-saves after a short delay; preview updates while you type in Split/Preview |
| Splitter | Drag the vertical bar between panes in Split mode |
| Copy | **Copy** — title + body to the clipboard |
| Revert edits | **Revert edits** restores title+body to how they looked **when you opened** this note (one level; not full history) |
| Delete | **Delete** |
| Collapse list | **☰ List** on narrow layouts |
| Links in preview | Open in your system browser (never navigate the Dossier window away) |
| Images in preview | Local paths under the dossier only |

Task-list checkboxes in the preview are **visual only** (clicking does not rewrite the source).

Right-click in title/body for Cut / Copy / Paste / Select all.

---

## Sticky Notes

Freeform board of colorful notes.

| Action | How |
|--------|-----|
| Add | **+ Add sticky** (pick color/size in the topbar) |
| Move / resize | Drag the header; resize from the corner |
| Emoji / flags | Header controls; country flags show as real images on Windows |
| Due date | Set a due date on a sticky; it appears on **Agenda** |
| Link to Kanban | Link a sticky to a board/column; a card is created. Text and due date stay in sync **without** deleting either side when you unlink |
| Unlink | Remove the sticky↔card link; sticky and card remain independently |
| Delete | Trash on the sticky header |
| Escape (pickers) | Escape closes the emoji, color, or size picker and returns focus |

---

## Paint

Freehand canvases for quick sketches.

| Action | How |
|--------|-----|
| New | **+ New canvas** |
| Tools | Pen, eraser, colors, brush size, undo/redo |
| Save | Automatic when leaving the editor / switching canvas |

---

## Annotate

Paste, drop, or open an image, then mark it up.

| Tool | Use |
|------|-----|
| Highlight | Translucent box |
| Arrow | Point at something |
| Label | Text label; drag and resize after placing |
| Redact / blur | Blackout or soft blur region |
| Crop | Crop the canvas |

Drop an image on the Annotate nav item to start a new board. Ctrl+V pastes when the surface is active.

---

## Kanban

Track work across columns.

| Action | How |
|--------|-----|
| New board | **+ New board** |
| Columns | Add / rename; set a **WIP limit** (count/limit badge) |
| Cards | **+ Card**; edit text (auto-grows) |
| Due date | Optional due on a card; shows on **Agenda**; syncs with a linked sticky |
| Drag | Use the **⋮⋮** grip, not the text field |
| Move (keyboard) | **Move…** opens a column list; Tab or **Up/Down arrows** between options; Enter/click to move |
| Link from sticky | Created from Sticky Notes; see above |

---

## Decisions

Pin decisions on a chronological spine.

| Action | How |
|--------|-----|
| Pin | **+ Pin decision** |
| Detail | Click a spine item |
| Attach version | Freeze a copy of a document under `decision-versions/` |
| Open version | Open the frozen file externally |
| Agenda | Decision dates also appear on **Agenda** (historical pins, not “overdue”) |

---

## Agenda

Agenda is a **list of dated items**, not a full calendar grid.

| Piece | Behavior |
|-------|----------|
| Mini-month | Navigate months; click a day to filter the list |
| Agenda list | Sticky due dates, kanban card due dates, and decision dates |
| Overdue | Past sticky/kanban due dates are styled as overdue; decisions are not treated as “due” |
| Jump | Click a row to open that sticky, board card, or decision |
| Empty days | Days with no items stay off the list |

Set dates on stickies/cards or pin decisions to populate Agenda.

---

## Search

Topbar search (or Search nav). Finds documents, notes, stickies, kanban, and decisions. Click a result to jump there.

---

## Ask dossier (optional)

Hidden until Settings has a **non-empty host** and **port &gt; 0**.

1. Settings → **Agent host** (e.g. `127.0.0.1`) and **Agent port**.  
2. Optional bearer token and HTTP path (default `/v1/chat/completions`).  
3. **Save settings** — toast says panel **enabled** or **hidden**.  
4. **Clear the host field entirely and save** to hide the panel again.  

In the Ask panel: type a question → **Ask**; relevant snippets are selected via full-text search and sent as capped context. **Clear conversation** resets the chat without changing Settings.

Dossier does **not** ship or auto-detect a local model.

---

## Settings

App-wide preferences in `%APPDATA%\Dossier\config.json`:

- Confirm before deletes  
- Show first-open intros / reset intros  
- **Open last workspace on launch** (default off)  
- Note auto-save delay  
- Optional agent host / port / token / path  
- **Open externally** preferred apps per extension  

---

## Multiple workspaces and collections

| Control | Action |
|---------|--------|
| **Switch…** | Open another recent or browse |
| **+ Folder** | Create/open another dossier folder |
| **Close** | Return to the launcher without quitting |
| **Collections** (launcher) | Group related dossier folders; expand a collection; open a member; add/remove folders |
| **Collection…** (in shell) | Switch among members of the active collection when the open folder belongs to one |
| **Rename…** (recent/collection) | Display name only |

Each dossier folder is isolated (its own `dossier.db` and files). Collections live in app config, not inside a single folder.

---

## Keyboard and accessibility (highlights)

| Behavior | Notes |
|----------|--------|
| Left nav | Activates a view and moves focus to that view’s main heading |
| Sticky pickers | Tab trap inside; Escape closes and returns focus |
| Kanban Move… | Up/Down between columns; Tab still works |
| Cut/Copy/Paste | Right-click menus and standard shortcuts in fields |

Screen-reader support is a solid baseline (landmarks, names, live toasts, focus), not a full WCAG audit.

---

## Troubleshooting

| Issue | Try |
|-------|-----|
| Window won’t open | Install/update **WebView2 Runtime** |
| Ask panel missing | Settings: non-empty **host** + port; **Save** |
| Flags look like “CA”/“US” | Should not happen; flags are embedded images |
| PDF/Word empty in search | Image-only PDF or encrypted Office — use Open externally |
| Agenda empty | Set a due date on a sticky or kanban card, or pin a decision |
| Log file | `%LOCALAPPDATA%\Dossier\dossier.log` |
