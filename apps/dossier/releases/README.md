# Releases

Self-contained snapshots of each shipped version, exactly as they were packaged at the time (own `README.md`, `docs/`, `PROJECT_SUMMARY.md`, `STATUS.md`, `VERSION`, `Dossier.exe`, `dossier-icon.png`, `sample-files/`).

| Folder | Version | Notes |
|--------|---------|-------|
| `1.0.3/` | 1.0.3 | First public release. Five feature tiers (documents, notes, sticky notes, paint, kanban, annotate, decisions, search, optional agent panel) plus a packaging pass and three icon/UX gap-fixes (1.0.0-1.0.2, not published as separate release folders; see the app's `CHANGELOG.md` and `PROJECT_SUMMARY.md` for that history). |

The files at the repo root (`README.md`, `VERSION`, `docs/`, etc.) are meant to always mirror the newest **shipped** release (right now that's `1.0.3/`), the same convention Whiteboard and Markdown Viewer use. Grab the root copy if you just want the current app; use a specific version folder here if you need an older one. `PROJECT_SUMMARY.md` and `build-process/` at the app root cover the full build history and aren't duplicated per-release.

Unlike Whiteboard/Markdown Viewer, Dossier is a compiled native Windows app, not a single HTML file, so there's no "Open" link, just the packaged `Dossier.exe` in each version folder plus a matching GitHub Release with a lean end-user `.zip` attached.
