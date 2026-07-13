# Developer guide, Whiteboard 2.4.0

This document is for people who want to **read or change** `whiteboard.html`.  
For end-user instructions, see [USER_GUIDE.md](./USER_GUIDE.md) and the root [README.md](../README.md).

**Current app version is 2.4.0** (7,357 lines, up from v1.0.0's ~2,840). Everything up to the "v2 additions" section near the end describes v1.0.0's architecture and is still accurate; the shell, the template contract, and the original four templates were extended in place across all five v2 tiers, not rewritten. The "v2 additions" section has been verified directly against the real v2.4.0 source (confirmed line ranges, confirmed storage schema, confirmed Kanban state shape), not just narrated from build summaries.

Everything that runs in production is **one self-contained HTML file**. CSS and JavaScript are inlined. There is no build step for the app itself. Optional Node scripts under `build-process/` only exist for automated browser checks.

Line numbers in the first part of this doc refer to `whiteboard.html` as of v1.0.0 (~2840 lines) and will have drifted after five more tiers of changes; the shell/template sections listed there moved to new (larger) line numbers in v2.4.0; see the confirmed v2 file map near the end for their current locations.

---

## Big picture

```
┌─────────────────────────────────────────────────────────┐
│  Shell (toolbar + #board-root)                          │
│  App object: boards list, open/save, File menu, theme   │
└───────────────────────────┬─────────────────────────────┘
                            │ openBoard(id)
                            ▼
┌─────────────────────────────────────────────────────────┐
│  TemplateRegistry.get(templateId).create(ctx)           │
│  → instance: setState → mount → renderTools             │
│  On leave / save: getState → board.state                 │
└─────────────────────────────────────────────────────────┘
                            │
        ┌───────────────────┼───────────────────┐
        ▼                   ▼                   ▼
   Paint / Notepad / Annotate (snip) / Wordpad
```

The shell never draws strokes or sticky notes itself. It only:

1. Knows which **board** is active (`currentBoardId`).
2. Looks up that board’s **`templateId`**.
3. Creates a template **instance**, restores **state**, mounts into `#board-root`, and fills `#template-tools`.
4. On change (or on switch/unload), calls **`getState()`** and writes the whole store to **`localStorage`**.

---

## File map (single file)

| Region (approx. lines) | What lives there |
|------------------------|------------------|
| Head: CSS | Design tokens, light/dark theme, layout, template-specific styles (paint, notepad, snip, wordpad, modals) |
| Body | Toolbar chrome (`#board-select`, File menu, theme, About), `#workspace` → `#board-root`, hidden file inputs |
| ~506-583 | IIFE start, `APP_VERSION`, keys, utilities, `HistoryStack` |
| ~586-684 | Storage load/save, **`migrateLegacy`**, `defaultStore`, `loadStore` |
| ~687-710 | **`TemplateRegistry`** + contract comment |
| ~712-827 | Modal helper (`showModal`) |
| ~829-1383 | **`App`** shell (boards, open/save, import/export, undo keys, About) |
| ~1386-1723 | **`PaintTemplate`** (`id: 'paint'`) |
| ~1725-2015 | **`NotepadTemplate`** (`id: 'notepad'`) |
| ~2018-2583 | **`SnipTemplate`** (`id: 'snip'`, UI name **Annotate**) |
| ~2585-2814 | **`WordpadTemplate`** (`id: 'wordpad'`) |
| ~2817-end | Boot: `register(...)` ×4, `App.init()`, `window.__whiteboard` |

Debug/test hook after boot:

```js
window.__whiteboard = {
  App, TemplateRegistry, STORAGE_KEY, STORAGE_KEY_LEGACY, VERSION
};
```

---

## Template registry and contract

### Registry API

```js
TemplateRegistry.register(templateDef);
TemplateRegistry.get(id);   // or null
TemplateRegistry.list();    // array of defs, insertion order depends on object key order
```

A **template definition** is a plain object:

| Field | Required | Meaning |
|-------|----------|---------|
| `id` | yes | Stable string key stored on each board (`paint`, `notepad`, `snip`, `wordpad`) |
| `name` | yes | Label in UI (type badge, new-board picker) |
| `description` | no | Shown in the new-board template picker cards |
| `kind` | no | Hint string: `'raster'`, `'text'`, or `'other'` (not heavily used by shell) |
| `create(ctx)` | yes | Factory that returns a **fresh instance** every time a board of this type is opened |

### Context (`ctx`)

Passed into `create(ctx)`:

| Method | When the template should call it |
|--------|----------------------------------|
| `ctx.onChange()` | After any user edit that must be persisted (debounced save follows) |
| `ctx.setStatus(msg, kind?)` | Optional status-bar feedback (`kind`: `'saved'` \| `'error'` or omit) |

### Instance methods

**Required (in practice for a usable board):**

| Method | Role |
|--------|------|
| `setState(state)` | Apply a plain JSON-serializable object (call **before** `mount` when opening a board) |
| `getState()` | Return a plain object to store on `board.state` |
| `mount(containerEl)` | Build DOM inside the element (usually `#board-root`), attach listeners |
| `unmount()` | Detach listeners, release resources; shell will clear the container HTML next |
| `renderTools(toolsEl)` | Fill `#template-tools` with this board’s toolbar controls |

**Optional (shell uses `typeof` checks):**

| Method | Used for |
|--------|----------|
| `undo()` / `redo()` / `canUndo()` / `canRedo()` | Global `Ctrl+Z` / `Ctrl+Shift+Z` / `Ctrl+Y` and template Undo buttons |
| `exportPng()` | File → Export as PNG (returns a `data:image/png…` URL or null) |
| `exportText()` | File → Export as text (returns a string or null) |

### Lifecycle when opening a board

`App.openBoard` (simplified):

1. `snapshotCurrent()` → previous instance `getState()` into previous board.
2. Fade: `#board-root` gets class `switching`.
3. After ~90ms: previous `unmount()`, clear root and tools.
4. `instance = TemplateRegistry.get(board.templateId).create(ctx)`.
5. `instance.setState(board.state || {})`.
6. `instance.mount(root)`.
7. `instance.renderTools(toolsEl)`.
8. Remove `switching`, schedule save.

**Order matters:** restore state first, then mount, so the first paint/layout can use restored data (e.g. canvas image URL, notes array, HTML).

**Important:** each open creates a **new** instance. Do not keep long-lived singleton instances on the template def.

---

## Multi-board storage

### Keys

| Constant | Value | Role |
|----------|--------|------|
| `STORAGE_KEY` | `whiteboard-app-v2` | Current store |
| `STORAGE_KEY_LEGACY` | `whiteboard-app-v1` | Read once for migration |

### v2 shape (what `flushSave` writes)

```json
{
  "version": 2,
  "theme": "light",
  "currentBoardId": "board_…",
  "boards": [
    {
      "id": "board_…",
      "name": "Paint board",
      "templateId": "paint",
      "state": { /* template-specific */ },
      "createdAt": "ISO-8601",
      "updatedAt": "ISO-8601"
    }
  ],
  "savedAt": "ISO-8601"
}
```

- **`templateId` is fixed per board** at create time. The shell does not convert types.
- **`state`** is opaque to the shell; only the matching template understands it.
- Save is debounced (~400ms) via `scheduleSave` → `flushSave`, plus `beforeunload` and a 30s interval.

### Per-template `state` shapes (as implemented)

| Template | Typical `state` fields |
|----------|-------------------------|
| Paint | `{ color, sizeId, imageDataUrl }` |
| Notepad | `{ notes: [{ id, x, y, w, h, text, color }, …] }` |
| Annotate (`snip`) | `{ tool, strokeColor, imageDataUrl }` |
| Wordpad | `{ html }` |

### v1 → v2 migration

Tier 1 stored roughly:

```json
{
  "version": 1,
  "currentTemplateId": "paint",
  "templates": {
    "paint": { … },
    "notepad": { … }
  }
}
```

On load (`loadStore`):

1. If `whiteboard-app-v2` exists and looks valid → use it.
2. Else if `whiteboard-app-v1` exists → `migrateLegacy(v1)`:
   - For each entry in `templates`, create one board with `templateId` = key and `state` = value.
   - Prefer `currentTemplateId` when choosing `currentBoardId`.
   - Write the result to v2 and return it.
3. Else → `defaultStore()` (one empty Paint board).

Export JSON uses a related envelope:

```json
{
  "format": "whiteboard-export",
  "version": 2,
  "exportedAt": "…",
  "boards": [ /* same board objects */ ]
}
```

Import **appends** boards (new ids); it does not wipe the existing library.

---

## Where each template’s code lives

### Paint, `PaintTemplate` (~1386-1723)

- Canvas freehand, brush colors/sizes, clear, undo history of PNG snapshots (`HistoryStack`).
- `exportPng` via `canvas.toDataURL`.
- Uses an `alive` flag so async resize/draw after `unmount` is a no-op (important for WebKit).

### Notepad, `NotepadTemplate` (~1725-2015)

- Sticky notes as absolute-positioned DOM; drag/resize via pointer events.
- No undo stack; structural edit via add/delete.

### Annotate, `SnipTemplate` (~2018-2583)

- Display name **Annotate**, id **`snip`** (historical snipping-tool naming).
- Primary input: paste, drag-drop, file picker; optional `getDisplayMedia` capture button.
- Tools: highlight, arrow, text, blur/redact, crop; state is mainly a composite PNG `imageDataUrl`.

### Wordpad, `WordpadTemplate` (~2585-2814)

- `contenteditable` editor; formatting via `document.execCommand`.
- Debounced HTML history for undo; `exportText` flattens HTML to plain text.

Registration order at boot (~2819-2822):

```js
TemplateRegistry.register(PaintTemplate);
TemplateRegistry.register(NotepadTemplate);
TemplateRegistry.register(SnipTemplate);
TemplateRegistry.register(WordpadTemplate);
```

That order is what the new-board picker walks when listing templates.

---

## How to add a fifth template

Follow the existing pattern; you do **not** need a bundler.

### 1. Implement a template definition

Near the other templates (before Boot), add something like:

```js
const ChecklistTemplate = {
  id: 'checklist',
  name: 'Checklist',
  description: 'Simple checkboxes',
  kind: 'other',
  create: function (ctx) {
    let items = [];
    let root = null;

    return {
      setState: function (state) {
        if (state && Array.isArray(state.items)) items = state.items.slice();
      },
      getState: function () {
        return { items: items };
      },
      mount: function (container) {
        root = document.createElement('div');
        // build UI from `items`, call ctx.onChange() on edits
        container.appendChild(root);
      },
      unmount: function () {
        // remove listeners; shell clears container
        root = null;
      },
      renderTools: function (toolsEl) {
        // e.g. "Add item" button
      }
      // optional: undo/redo, exportText, exportPng
    };
  }
};
```

Rules of thumb:

- **Only serializable data** in `getState()` (no DOM nodes, no functions).
- **Detach global listeners** in `unmount` (e.g. `window` key/mouse handlers).
- Call **`ctx.onChange()`** after user edits so multi-board save stays correct.
- Prefer guarding async callbacks after unmount if you use timers, `Image.onload`, or `ResizeObserver`.

### 2. Register it

In the Boot section:

```js
TemplateRegistry.register(ChecklistTemplate);
```

### 3. CSS (optional)

Add a `.checklist-root { … }` (or similar) block in the `<style>` section so the board matches the rest of the UI.

### 4. What you do **not** need to change

- Board create/rename/duplicate/delete UI (it uses `TemplateRegistry.list()`).
- Save/load path (state is opaque).
- Import/export of boards (as long as `templateId` is known at import time).

### 5. Migration note

`migrateLegacy` includes a preferred order list `['paint','notepad','snip','wordpad']` and then any extra keys in `templates`. A new template only appears in v1 migration if something already wrote it under `templates[id]`. New installs only need the register call.

### 6. Optional product hooks

- If the board is image-like, implement `exportPng`.
- If text-like, implement `exportText`.
- If undoable, implement `undo` / `redo` / `canUndo` / `canRedo` so the global shortcut handler picks them up.

### 7. Verify

Open `whiteboard.html`, create a board of the new type, reload, confirm state survives. If you maintain automated checks, extend `build-process/verify*.mjs` similarly.

---

## Shell pieces worth knowing

| Concern | Where |
|---------|--------|
| Board CRUD UI | `App.promptNewBoard`, `promptRename`, `duplicateBoard`, `deleteBoard` |
| Theme | `App.applyTheme`, `data-theme` on `<html>` |
| Import/export | `exportBoardJson`, `importData`, `exportPng`, `exportText` |
| Undo shortcuts | `App.onGlobalKey` |
| Version / About | `APP_VERSION`, `#version-badge`, `App.showAbout` |

---

## Constraints that shape the design

1. **`file://` first**, no CDN, no server, no build required to run.
2. **localStorage**, simple, but quota-limited (especially PNG data URLs).
3. **One HTML file**, templates are modules-by-convention, not separate files.
4. **Template type fixed per board**, avoids messy mid-board type conversion.

---

## Related docs and tooling

| Path | Purpose |
|------|---------|
| `docs/USER_GUIDE.md` | End-user feature walkthrough |
| `README.md` | Short user-facing overview |
| `PROJECT_SUMMARY.md` | Build retrospective (process / tradeoffs) |
| `build-process/` | Verification scripts and saved reports (not needed at runtime) |

---

## v2 additions (2026-07-12), verified against the real source

**Update:** the section below was originally written unverified, from tier-completion summaries only, with an explicit warning not to trust specifics from it. That's no longer true; the real `whiteboard.html` v2.4.0 (7,357 lines, up from v1.0.0's ~2,840) has since been pulled into this bundle and read directly. Everything below is confirmed against the actual source, not inferred from narration.

### File map, v2 additions

| Region | What lives there |
|---|---|
| ~line 1053-1084 | `APP_VERSION = '2.4.0'`, `STORAGE_KEY` / `STORAGE_KEY_LEGACY` (unchanged keys, see below), `HISTORY_MAX_SNAPSHOTS`, `HISTORY_MIN_INTERVAL_MS`, `TRASH_MAX_AGE_MS`, `QUOTA_SNAPSHOT_TRIM_RATIO`, `STORAGE_SOFT_CHARS`, `STORAGE_QUOTA_ESTIMATE` |
| ~line 1131-1161 | `HistoryStack`, unchanged from v1.0.0, still backs Paint/Annotate/Wordpad undo |
| ~line 1945-2053 | `purgeExpiredTrash`, `normalizeHistoryMap`, `trimHistoryForQuota`, Tier 5's history/trash/quota-integration logic |
| ~line 2447-2453 | `updateQuotaMeter`, reads `STORAGE_QUOTA_ESTIMATE`, writes the toolbar quota bar |
| ~line 2851 | `snapshotCurrent`, used both for autosave *and* the pre-restore/pre-trash forced snapshot Tier 5 added |
| ~line 4940 | `PaintTemplate` (unchanged position/contract) |
| ~line 5288 | `NotepadTemplate`, extended in place with checklist/auto-color/tag logic, contract unchanged |
| ~line 5588-5637 | Pointer-based drag handlers on Notepad sticky headers (pre-existing from v1, same pattern Kanban's card drag reuses) |
| ~line 5872 | `SnipTemplate` (Annotate, unchanged position/contract) |
| ~line 6461 | `WordpadTemplate` (unchanged position/contract) |
| ~line 6693 | **`KanbanTemplate`**, new, `id: 'kanban'`, follows the exact same `create(ctx) → {setState, getState, mount, unmount, renderTools}` contract as the other four. Nothing about the shell's template registry changed to accommodate it. |
| ~line 6754-6871 | Kanban's pointer-drag session (`pointerDrag` state var, `onPointerDragMove`/`onPointerDragUp`, drop-target resolution via `document.elementsFromPoint`); confirms the Tier 4 fix's own account: genuinely pointer-event based, not native HTML5 `draggable`/`dragstart` |
| ~line 7185-7205 | Kanban `setState`/`getState` |
| line 7301-7305 | Boot registration: `TemplateRegistry.register(PaintTemplate)`, `NotepadTemplate`, `SnipTemplate`, `WordpadTemplate`, `KanbanTemplate`, same pattern as v1, one line added |

### Storage schema: confirmed additive, no version bump

`STORAGE_KEY` is still `'whiteboard-app-v2'`, `STORAGE_KEY_LEGACY` is still `'whiteboard-app-v1'`; **no `v3` key was introduced.** History and trash data live alongside the existing `boards[]` array in the same store object, keyed separately (`history`, keyed by board id, and `trash`, a list). No new migration path was needed since the v2 key's shape simply grew rather than changing incompatibly. `version: 2` still appears in the relevant places in the source.

### Kanban's actual state shape (confirmed)

```json
{
  "columns": [
    { "id": "col_…", "name": "To Do", "cardIds": ["card_…", "card_…"] }
  ],
  "cards": {
    "card_…": { "id": "card_…", "text": "…", /* plus color/tag/checklist fields, same shape family as Notepad's notes */ }
  }
}
```
Columns hold ordered `cardIds` arrays (cards aren't nested inside columns directly), and `cards` is a flat id-keyed map, closer to a normalized store than Notepad's flat array of note objects. `normalizeKanbanState()` is the entry point `setState` calls to defend against malformed/missing fields, in the same spirit as the shell's own corrupt-state recovery.

### History/trash/quota constants (confirmed, worth knowing if you touch this code)

| Constant | Value | Role |
|---|---|---|
| `HISTORY_MAX_SNAPSHOTS` | 10 | Rolling cap per board |
| `HISTORY_MIN_INTERVAL_MS` | 2 minutes | Minimum gap between auto-snapshots (this is the "not every autosave tick" cadence limit from the Tier 5 prompt) |
| `TRASH_MAX_AGE_MS` | 7 days | Auto-purge threshold |
| `QUOTA_SNAPSHOT_TRIM_RATIO` | 0.72 | Pressure-relief trim point, old snapshots get trimmed before Tier 1's more aggressive image compression kicks in |
| `STORAGE_SOFT_CHARS` | ~2.5M UTF-16 chars | Soft warning threshold feeding the quota meter |
| `STORAGE_QUOTA_ESTIMATE` | 5MB | What the quota meter's percentage is computed against |

### Paste routing and search/palette

Not yet individually traced to exact line numbers the way the above was; the confirmed facts from the tier fix reports (a real `pasteRouteCharCount()` doing a trimmed `.length` check, toast state cleared on every new route, search and palette as genuinely separate `Ctrl/Cmd+K` / `Ctrl/Cmd+/` implementations per the Tier 2 prompt's explicit instruction) held up under a spot-check of the surrounding code but weren't traced function-by-function the way the sections above were. Treat those as reliable but slightly lower-confidence than the file map above.
