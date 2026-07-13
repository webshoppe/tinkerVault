# Releases

Self-contained snapshots of each shipped version, exactly as they were packaged at the time (own `whiteboard.html`, `icon.svg`, `README.md`, `VERSION`, `docs/USER_GUIDE.md`; no external dependencies, each folder runs on its own).

| Folder | Version | Boards | Notes |
|--------|---------|--------|-------|
| `v1.0.0/` | 1.0.0 | 4 (Paint, Sticky Notes, Annotate, Wordpad) | First shipped release, 2026-07-11. |
| `v2.4.0/` | 2.4.0 | 5 (adds Kanban) | Five-tier feature pass plus packaging, 2026-07-12. Same file as the one at the repo root. |

The files at the repository root (`whiteboard.html`, `icon.svg`, `README.md`, `VERSION`, `docs/`) always mirror the newest release (right now that's `v2.4.0/`). Grab the root copy if you just want the current app; use a specific version folder here if you need an older one, or want to compare versions directly. `PROJECT_SUMMARY.md` and `build-process/` at the repo root cover the full build history across both versions and aren't duplicated per-release.
