# Changelog, Dossier

All notable changes to this app are listed here. See [PROJECT_SUMMARY.md](PROJECT_SUMMARY.md) for the full tier-by-tier build story behind each entry.

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
