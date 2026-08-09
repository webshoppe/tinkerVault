# Changelog, Dossier

All notable changes to this app are listed here. See [PROJECT_SUMMARY.md](PROJECT_SUMMARY.md) for the full tier-by-tier build story behind each entry.

## v2.0.1, 2026-08-07

Gap-fix; surgical CSS only, no feature work.

### Fixed
- **Notes and Decisions: two-pane layout no longer breaks at a maximized window.** At wide widths (~1920px), the list/spine column ballooned to consume nearly the full width, squeezing the note editor / decision detail pane into a narrow strip on the right with a dead gap in between. Root cause: the list/spine pane used a soft, shrinkable flex basis (`flex: 0 1 <n>px` plus a `min()`-based viewport width) left over from the earlier v1 containment fix that only addressed horizontal scroll and column stacking, not the side-by-side width ratio. Documents' sources rail never had this problem because it uses a hard-pinned `flex: 0 0 220px`. Fixed by hard-pinning Notes' list to `width: 260px; flex: 0 0 260px` and Decisions' spine to `width: 280px; flex: 0 0 280px` (both with `min-width: 0`), and giving the editor/detail panes `flex: 1 1 0%; min-width: 0` so they always claim the remaining width. First-open intro banners also now force their own full-width row instead of becoming a third flex column. The existing stack breakpoint at 1320px and narrower is unchanged.

## v2.0.0, 2026-08-01

Current shipped version. Built up over five v2 feature tiers plus an accessibility pass and a Tier 6 close-out/packaging pass (internal versions 2.0.0-t1 through 2.0.0-t7, not published as their own release folders; see PROJECT_SUMMARY.md for the full tier history).

### Added
- Documents: **.odt / .ods** import alongside the existing md/txt/pdf/docx/xlsx support, plus a Sources panel with **import bookmarks** (saved shortcut folders)
- Notes: **Edit / Split / Preview** live Markdown workbench (offline `marked` + `highlight.js`, vendored)
- Due dates on Sticky Notes and Kanban cards, with a **non-destructive link** between a sticky and a Kanban card (text and due date stay in sync; unlinking keeps both cards intact)
- **Agenda**: a mini-month navigator plus an agenda list aggregating sticky/kanban due dates and decision dates (not a full grid calendar)
- **Collections**: group related dossier folders together on the launcher, alongside recents
- **Workspace display names**: rename how a folder appears in the launcher without renaming it on disk
- **Open last workspace on launch** as an opt-in Settings toggle (the existing `DOSSIER_AUTO_OPEN` env-var override remains separate, for debug/verify use)
- Accessibility baseline: landmarks, accessible names, focus-visible states, keyboard support on custom widgets (accordions, Kanban tab order, picker focus), live-region toast announcements, route-change focus-to-heading, dark-mode-aware scrollbars
- Kanban "Move to column" menu now supports arrow-key navigation between options, in addition to Tab
- App version now visible in the launcher/sidebar title area, not just the footer and Explorer Properties
- Console binary (`Dossier-console.exe`) now reports its own filename in Explorer Properties → Details, instead of `Dossier.exe`

### Fixed
- **Sticky Notes Emoji/Color/Size picker: Escape key now closes the picker.** Root cause took three diagnostic passes to isolate: a struct-layout ABI bug (`COREWEBVIEW2_PHYSICAL_KEY_STATUS`'s `WasKeyDown` field was a Go `bool` instead of the real 4-byte Windows `BOOL`, so the value could misread), plus a `KEY_UP`-vs-`KEY_DOWN` event-kind gap (ordinary keys arrive as `KEY_DOWN`; Escape arrives as `KEY_UP` on this host, and the accelerator-key callback's gate only fired on `KEY_DOWN`). Fixed by correcting the struct field type and extending the gate to also fire on `KEY_UP`/`SYSTEM_KEY_UP` for Escape specifically; hand-verified with NVDA across multiple pickers.
- Launcher "Rename workspace" button visual overlay bug
- Renaming a workspace no longer hides the real folder path in the sidebar (both display name and path now show)
- User-facing "Calendar" terminology renamed to "Agenda" throughout, to match what the surface actually is (a due-date list with a mini-month navigator, not a full grid calendar)

## v1.0.3, 2026-07-26

Current shipped version. Built up over five feature tiers, then a packaging pass and three packaging gap-fixes (internal versions 1.0.0-1.0.2, not published separately as their own release folders; see PROJECT_SUMMARY.md for why).

### Added
- Native Win32 desktop window (go-webview2), not a browser tab
- Dossier folder concept: one self-contained workspace per folder; backup is just copying the folder
- Documents: import markdown/text/PDF (best-effort)/docx/xlsx plus other file attachments, drag-and-drop onto the window, sort, detail preview, and "open externally" with a per-file-type preferred-app override
- Full-text search (SQLite FTS5) across documents, notes, stickies, Kanban, and decisions
- Notes: markdown notes stored as real files, auto-save, a one-click Copy button, and a shallow "Revert edits" safety net
- Sticky Notes: colors, sizes, and emoji including flag glyphs (offline Twemoji images, so flags render correctly without relying on the OS font)
- Paint and Kanban boards (configurable per-column WIP limits, drag-and-drop cards)
- Annotate: highlight, arrow, text label, redact, and crop tools
- Decisions: a chronological spine with optional frozen document-version snapshots
- Multiple dossiers/workspaces, with recents and quick switching
- Optional "Ask dossier" panel: sends FTS-selected, relevance-ranked (bm25) context to a local HTTP agent you configure yourself; off by default, no embedded model, never assumes any specific local AI stack
- Settings panel, first-open intros per surface, and a bundled sample-files folder for first-time onboarding
- Real Windows packaging: a PE-embedded icon (taskbar, Explorer, and window), a matching file version resource, and right-click Cut/Copy/Paste restored without re-enabling WebView2 DevTools

### Fixed (packaging hardening, pre-1.0.3)
- The app icon shipped with no alpha channel at all (a solid dark box behind the artwork instead of a transparent background); regenerated with a real RGBA icon, then reworked twice more for size and art direction
- The Ask panel's Settings host/port gate didn't reliably enable on the very first save
- The Ask panel's FTS context selection was too literal on natural-language questions at first, then overcorrected and pulled in too many documents; tuned to relevance-ranked (bm25) top matches only
- Microsoft Office `~$` lock/temp files were being imported as real documents during a folder Rescan while the source file was open elsewhere
