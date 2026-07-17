# Changelog, Markdown Viewer

All notable changes to this app are listed here, newest first. See [PROJECT_SUMMARY.md](PROJECT_SUMMARY.md) for the full build story behind each entry.

## v1.1.0, 2026-07-16

### Fixed
- Tab collision bug: two open files that happened to share a name (e.g. two different `README.md` from different folders) used to overwrite each other into a single tab. Tabs are now keyed by a unique per-load id rather than filename, so same-named files stay separate and are labelled `name`, `name (2)`, `name (3)`...
- Tab-close button (`×`) could be clipped or unreachable behind a long filename. Text truncation now applies only to the name label, so the close button always stays visible and clickable.

### Added
- Open multiple files at once from the file picker (previously only the first selected file loaded)
- Edit mode: switch the active tab to a raw-text editor; the formatted view refreshes when you leave edit mode
- Save: download the active tab's current content, including live edits, as a new `.md` file
- Copy MD: copy the active tab's raw Markdown source in one click
- Per-tab scroll position memory, restored when you switch back to a tab
- Keyboard shortcuts: open, close active tab, reopen last closed tab, switch between tabs
- Browser tab title now shows the active document's filename
- Drag-and-drop skips non-text files (images, PDFs, etc.) with an on-screen notice instead of rendering them as garbage
- Rebuilt, human-editable build pipeline: `src/template.html` + `src/app.js` + `src/vendor/*`, assembled by `src/assemble.py`

## v1.0.0, 2026-07-11

### Added
- Initial release: GitHub-Flavored Markdown rendering (headings with anchors, bold/italic/strikethrough/inline code, tables, task lists, nested lists, blockquotes) with real syntax-highlighted code blocks
- Copy button on every code block
- Dark/light theme toggle
- Collapsible table of contents generated from headings
- In-document search with highlight and jump-between-matches
- Multiple files open at once, each in its own tab
- Raw/rendered view toggle
- Export to a standalone, self-contained `.html` file
- Remembers the last file opened, plus theme and TOC state, via `localStorage`
- Internal `.md`/`.markdown`/`.txt` link interception: clicking a same-origin relative link while viewing a file inside the running app loads it via the tab system, or shows a graceful on-page message instead of failing silently if `fetch()` is blocked under `file://` (browser-dependent; the links themselves work normally everywhere else, GitHub, editors, any standard renderer)
- Inline SVG favicon and a `VERSION` file read by the build script and stamped into the footer
