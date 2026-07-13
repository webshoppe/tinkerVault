# Changelog, Whiteboard

All notable changes to this app are listed here, newest first. See [PROJECT_SUMMARY.md](PROJECT_SUMMARY.md) for the full build story behind each entry.

## v2.4.0, 2026-07-12

### Added
- Fifth board type, **Kanban**: editable columns (three defaults: To Do / In Progress / Done), live WIP counts, and two non-destructive promotion paths (a single sticky note, or a whole Notepad board) into a new Kanban board without touching the original
- Global search (`Ctrl/Cmd+K`) across every board's text
- Command palette (`Ctrl/Cmd+/`) for actions and board switching, currently the only way to reach Trash
- Smart paste/drop routing by content type and length (images, `.txt`/`.md`, URLs, short text, long text), with a dismissable toast for genuinely ambiguous cases
- Auto-checklist detection, keyword-driven auto-color (user-editable keyword map), and `#tag` extraction on sticky notes and Kanban cards
- Board time machine: up to 10 rolling snapshots per board, forced snapshot before any restore or trash action
- Trash: 7-day soft-delete for boards with auto-purge; the last remaining board can't be removed
- Visible storage quota meter, with automatic downscaling of large Paint/Annotate images before they'd hit the browser's storage limit
- Schema-validated corrupt-state recovery (quarantines bad data to a downloadable export instead of losing it or white-screening)
- Multi-tab-safety banner when the same file is open in two tabs at once
- Full session restore: last board, zoom, scroll position, and last tool per template

### Fixed
- Paste/drop routing looked broken at everyday lengths despite correct 280/900-character thresholds in the code; root cause was a stale toast not clearing between pastes, plus a tabular-text false positive on ordinary prose with one comma per line
- Kanban drag-and-drop didn't work at all by hand despite a clean automated pass; the card's draggable area was mostly a `<textarea>`, which blocks native HTML5 drag in Firefox, and the original verify script moved cards by calling internal state directly instead of simulating a real pointer gesture. Rebuilt on a dedicated `⋮⋮` grip handle with real `pointerdown`→`pointermove`→`pointerup` events

## v1.0.0, 2026-07-11

### Added
- Initial release: four board types (Paint, Notepad, Annotate, Wordpad), each pluggable through a shared `register → create → setState / mount / renderTools / getState / unmount` contract
- Multi-board management (create, rename, duplicate, delete), each board storing its own type and state independently
- Automatic saving to `localStorage`, with a v1-to-v2 storage schema migration built in ahead of the next release
- JSON import (appends, never silently replaces) and export, per-board or full library
- PNG export for Paint and Annotate; plain-text export for Wordpad
- Undo/redo (`Ctrl+Z` / `Ctrl+Shift+Z`) for Paint, Annotate, and Wordpad

### Fixed
- A WebKit-only crash (`TypeError: null is not an object (evaluating 'c2d.save')`) when switching boards unmounted Paint while a pending `ResizeObserver` callback still touched a null canvas context; caught only by cross-engine testing, invisible in Chromium. Fixed with an `alive` flag, canceled pending resize timers, and null-guards on canvas operations
