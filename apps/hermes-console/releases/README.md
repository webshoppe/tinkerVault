# Releases

Self-contained snapshots of each shipped version, exactly as they were packaged at the time (own `index.html`, `manifest.json`, `sw.js`, icons/favicons, `README.md`, `VERSION`, `docs/USER_GUIDE.md`; no external dependencies, each folder runs on its own).

| Folder | Version | Notes |
|--------|---------|-------|
| `v1.0.0/` | 1.0.0 | First shipped release, 2026-07-18. Connection screen, approve/deny of tool calls, run stop, session browse/open/create/fork, installable PWA shell. Run monitoring uses polling, not SSE, due to a CORS gap on the server's events endpoint. |
| `v1.1.0/` | 1.1.0 | 2026-07-18. Adds a per-message toolbar (Copy on any message, Rerun on user messages, resubmits the exact input as a new run) and an in-app version label next to the header title, read from `VERSION` at runtime. Also the first release folder created after the app root's `docs/` + `build-process/` restructuring, matching the rest of the monorepo. Same file as the one at the app root. |

The files at the app root (`index.html`, `manifest.json`, `sw.js`, icons, `README.md`, `VERSION`, `docs/`) always mirror the newest release (right now that's `1.1.0`). Grab the root copy if you just want the current app; use a specific version folder here if you need an older one, or want to compare versions directly. `PROJECT_SUMMARY.md` and `CHANGELOG.md` at the app root cover the full build history across versions and aren't duplicated per-release.
