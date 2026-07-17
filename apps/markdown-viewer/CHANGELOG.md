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
- Browser tab t