# Changelog, Markdown Viewer

All notable changes to this app are listed here, newest first. See [PROJECT_SUMMARY.md](PROJECT_SUMMARY.md) for the full build story behind each entry.

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
