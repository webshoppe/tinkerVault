# Changelog, tinkerVault (monorepo)

The build-out history of the **repo itself**, day by day: structure, conventions, the landing page, licensing, and cross-app housekeeping. Each app also keeps its own feature-level changelog for just that app's version history: [`apps/hermes-console/CHANGELOG.md`](apps/hermes-console/CHANGELOG.md), [`apps/markdown-viewer/CHANGELOG.md`](apps/markdown-viewer/CHANGELOG.md), [`apps/whiteboard/CHANGELOG.md`](apps/whiteboard/CHANGELOG.md).

## 2026-07-18, Hermes Console 1.1.0

- Added a per-message toolbar to Hermes Console: Copy on any message, Rerun on user messages (resubmits the exact input as a new run through the existing send path, no history replay).
- Added a version label next to the header title, read from `VERSION` at runtime rather than hardcoded, so it can't drift from the CHANGELOG/README again.
- Investigated the footer's "Powered by" line for a possible version-field gap; confirmed the live server exposes no version-like field anywhere in `/v1/capabilities`, not a bug, no code change, documented so it isn't relitigated.
- Archived v1.0.0 as `apps/hermes-console/releases/v1.0.0/`, added `apps/hermes-console/releases/v1.1.0/` as the new current snapshot.
- Updated `apps/hermes-console/PROJECT_SUMMARY.md` to also cover the v1.1.0 build (the file itself already existed from the v1.0.0 release; this pass adds a v1.1.0 section, it isn't a new file).

## 2026-07-18, Hermes Console 1.0.0

- Added a new app, Hermes Console (`apps/hermes-console/`), tag `hermes-console-v1.0.0`, a single-file PWA for resolving Hermes approval-gated tool calls and browsing sessions from a phone or any other device, talking directly to a running Hermes API server over HTTP.
- Ships a connection screen (host/port/bearer, stored per-device), capability-gated UI, chat, remote approve/deny of tool calls, run stop, and session browse/open/create/fork.
- Installable as a PWA: manifest, service worker (cache-first shell, network-only API), custom install button, maskable and standard icons.
- Run monitoring uses polling rather than SSE, since the server's `/v1/runs/{id}/events` endpoint is missing its CORS header; the SSE reader is retained behind a flag for when that's fixed.
- Per-message toolbar (copy, rerun, etc.) is a planned v1.1 follow-up, not in this release.
- First app in the repo to use `PROJECT_SUMMARY.md`, `docs/`, and a separate `build-process/` folder for dev-only artifacts from day one, rather than folding those in on a later version bump.

## 2026-07-17, Markdown Viewer v1.1.0 folded in, repo hygiene pass

- Removed two stray duplicate files (`apps/README.md`, `apps/CHANGELOG.md`) left over from an earlier commit, sitting one directory too high, not referenced from anywhere.
- Added a real `apps/README.md` afterward, a short overview linking to both app folders, filling the gap the stray-duplicate cleanup left behind.
- Archived the outgoing Markdown Viewer v1.0.0 build as a `releases/v1.0.0/` snapshot before replacing it, matching the pattern already established for Whiteboard.
- Shipped Markdown Viewer v1.1.0: rebuilt its source pipeline into `src/template.html` + `src/app.js` + `src/vendor/*`, retiring the old flat `src/` layout (`github-dark.css`, `github-light.css`, and root-level `marked.min.js`/`highlight.min.js` all moved into `src/vendor/`).
- Caught and fixed a handful of docs (`README.md`, `CHANGELOG.md`, `CONTRIBUTING.md`, `PROJECT_SUMMARY.md`) that had been silently truncated mid-write during the v1.1.0 commit, and restored `index.html` after it drifted to Linux line endings during a build-pipeline check.
- Added Markdown Viewer to the GitHub Pages landing page (`index.html`) with proper version cards, it previously showed a single unversioned card; now matches Whiteboard's archived+current two-card pattern. Added a small "Latest" badge to both current-version cards (Markdown Viewer v1.1.0, Whiteboard v2.4.0).
- Archived Markdown Viewer v1.1.0 itself as a `releases/v1.1.0/` snapshot, mirroring how Whiteboard keeps a redundant copy of its current version inside `releases/`; added the matching row to `releases/README.md`.

## 2026-07-15, Privacy fix

- Redacted a local file path (including a personal identifier) that had leaked into Markdown Viewer's `PROJECT_SUMMARY.md`, in its cold-read follow-up section, without losing the finding it was describing.

## 2026-07-13, Repo created

- Initial scaffolding: root `README.md` and the GitHub Pages landing page (`index.html`), a build philosophy and hardware-notes doc (`docs/PHILOSOPHY.md`, later corrected once, an early claim that the WSL2 build tooling ran on "a second machine in the fleet" was wrong, it's an isolated WSL2 distro on the same physical rig, fixed everywhere it appeared).
- Added Markdown Viewer v1.0.0 and Whiteboard v2.4.0. Whiteboard's original four-board v1.0.0 build was preserved as a release snapshot from day one, `releases/v1.0.0/`.
- Added a release badge to each app's README, scoped to that app's own release tags (`filter=markdown-viewer*` / `filter=whiteboard*`) so one app's badge never reflects the other's release status.
- Whiteboard's README gained a "Getting Started" section (download the zip, unzip, run it) and a "Privacy-focused browsers" section documenting an HTML5 canvas permission prompt some browsers show (Mullvad, Tor, strict Firefox tracking protection) and a known striped-placeholder display quirk while that permission is unresolved, verified live in Firefox and Mullvad Browser.
- Added a shared root-level [`DESKTOP-LAUNCHER.md`](DESKTOP-LAUNCHER.md) guide (how to make any app open in its own chromeless window via the browser's `--app` flag, plus optional profile isolation). Deliberately generic and written once, linked from Whiteboard's Getting Started section rather than duplicated per app.
- Added a repo-root MIT `LICENSE`. Both apps' READMEs and the root README's License section were switched from placeholder/TBD text to point at it, with a short "fork it, no attribution required" line added to the root README and the landing page footer.
- Fixed the landing page's Source buttons, they'd pointed at the same relative path as Open (the running app) instead of the actual GitHub folder; both now open the real `github.com/.../tree/main/apps/...` view in a new tab.
- Made Whiteboard's v1.0.0 build reachable straight from the landing page, in three steps over the course of the day: first as a small "Also available: v1.0.0" text link under the v2.4.0 card, then upgraded to a proper button in the same row as Open/Source, then finally split into a fully separate peer card (Markdown Viewer, Whiteboard v1.0.0, Whiteboard v2.4.0, three equal listings), the shape the landing page has kept ever since. The temporary `.versions` CSS from the first step was cleaned up once the final layout landed.
- Added a per-app `CHANGELOG.md` to both apps, each linked from that app's own README.
