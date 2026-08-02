# Dossier User Guide

**Version 1.0.3**

This guide covers every surface in Dossier. Double-click `Dossier.exe` to open. No account and no internet required for core features. **Ask dossier** needs a local agent you configure yourself.

---

## Getting started

1. Launch **Dossier.exe**.
2. Choose **Create / open folder…** (or **Open last** / a recent workspace).
3. That folder is your **dossier**; documents, notes, and boards all live there.
4. Use the left sidebar to switch surfaces. Footer shows **v1.0.3**.
5. Optional first-open tips appear once per surface; dismiss with **Got it** (or turn them off in Settings).

> Tip: Copy the whole dossier folder to back up or move a project. App Settings (agent, intros) stay in `%APPDATA%\Dossier` and are not inside the folder.

---

## Documents

Import files into `documents/` and browse them with sort and detail preview.

| Action | How |
|--------|-----|
| Import | **+ Import files…** or drag files onto the window |
| Sort | Topbar **Sort**: Name A–Z / Z–A, Date added newest / oldest (remembers for this session) |
| Open detail | Click a row |
| External app | **Open externally**; uses Dossier’s preferred app for that file type if set, otherwise the Windows default |
| Choose / remember app | **Open with…**; Windows default, choose once, **Always open .ext with…**, or clear preferred |
| Delete | **Delete** in the detail header (confirm if enabled) |
| Rescan | **Rescan** re-indexes the folder |

**Supported kinds**

- **Markdown / text**; full preview + search  
- **PDF**; best-effort text extract; image-only pages get a warning badge  
- **Word / Excel (.docx / .xlsx)**; best-effort text for search; preview is lossy; use **Open externally** for the real file  
- **Other**; stored as attachments  

Office lock files (`~$…`) are skipped on import.

---

## Notes

Markdown notes stored as real `.md` files under `notes/`.

| Action | How |
|--------|-----|
| New | **+ New note** |
| Edit | Title + body; auto-saves after a short delay (Settings) |
| Copy | **Copy** next to Delete; puts title + body on the clipboard |
| Revert edits | **Revert edits** appears after you change the note; restores title+body to how they looked **when you opened** this note (one level; not full history) |
| Delete | **Delete** |
| Collapse list | **☰ List** on narrow layouts |

Right-click in title/body for Cut / Copy / Paste / Select all (standard edit menu).

---

## Sticky Notes

Freeform board of colorful notes.

| Action | How |
|--------|-----|
| Add | **+ Add sticky** (pick color/size in the topbar) |
| Move | Drag the header |
| Resize | Corner handle |
| Emoji / flags | Header controls; country flags show as real flag images on Windows |
| Delete | Trash control on the sticky header |

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
| Redact | Solid blackout |
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
| Drag | Use the **⋮⋮** grip; not the text field |

---

## Decisions

Pin decisions on a chronological spine.

| Action | How |
|--------|-----|
| Pin | **+ Pin decision** |
| Detail | Click a spine item |
| Attach version | Freeze a copy of a document under `decision-versions/` at decision time |
| Open version | Open the frozen file externally |

---

## Search

Topbar search box (or the Search nav item). Finds documents, notes, stickies, kanban, and decisions via full-text search. Click a result to jump to that surface.

---

## Ask dossier (optional)

Hidden until Settings has a **non-empty host** and **port &gt; 0**.

1. Settings → set **Agent host** (e.g. `127.0.0.1`) and **Agent port**.  
2. Optional bearer token and HTTP path (default `/v1/chat/completions`).  
3. **Save settings**; toast says panel **enabled** or **hidden**.  
4. **Clear the host field entirely and save** to hide the panel again.  

In the Ask panel:

- Type a question → **Ask**  
- Relevant dossier snippets are selected via full-text search and sent as context (capped)  
- **Clear conversation** resets the chat view without changing Settings  

Dossier does **not** ship or auto-detect a local model. You run the agent yourself.

---

## Settings

App-wide preferences in `%APPDATA%\Dossier\config.json`:

- Confirm before deletes  
- Show first-open intros / reset intros  
- Note auto-save delay  
- Optional agent host / port / token / path  
- **Open externally** preferred apps per extension (clear from Settings; set from a document’s **Open with…**) 

---

## Multiple workspaces

- **Switch…**; open another recent or browse  
- **+ Folder**; create/open another dossier folder  
- **Close**; return to the launcher without quitting  

Each dossier folder is isolated (its own `dossier.db` and files).

---

## Keyboard & clipboard

| Shortcut | Action |
|----------|--------|
| Ctrl+C / X / V | Copy / cut / paste in text fields (also via right-click menu) |
| Enter (Ask box) | Send question (Shift+Enter for newline) |

---

## Troubleshooting

| Issue | Try |
|-------|-----|
| Window won’t open | Install/update **WebView2 Runtime** |
| Ask panel missing | Settings: non-empty **host** + port; **Save** |
| Flags look like “CA”/“US” letters | Should not happen in 1.0; flags are embedded images |
| PDF/Word empty in search | Image-only PDF or encrypted Office; use Open externally |
| Log file | `%LOCALAPPDATA%\Dossier\dossier.log` |
