# Releases

Self-contained snapshots of each shipped version, exactly as they were packaged at the time (own `index.html`, `icon.svg`, `README.md`, `VERSION`, `docs/USER_GUIDE.md`; no external dependencies, each folder runs on its own).

| Folder | Version | Notes |
|--------|---------|-------|
| `v1.0.0/` | 1.0.0 | First shipped release, 2026-07-11. Single-file-at-a-time, no edit mode. |
| `v1.1.0/` | 1.1.0 | Fixes the same-name tab collision bug; adds edit mode, Save, Copy MD, keyboard shortcuts, per-tab scroll memory. 2026-07-16. Same file as the one at the repo root. |

The files at the app root (`index.html`, `icon.svg`, `README.md`, `VERSION`, `docs/`, `src/`) always mirror the newest release (right now that's `1.1.0`). Grab the root copy if you just want the current app; use a specific version folder here if you need an older one, or want to compare versions directly. `PROJECT_SUMMARY.md` and `CHANGELOG.md` at the app root cover the full build history across versions and aren't duplicated per-release.
