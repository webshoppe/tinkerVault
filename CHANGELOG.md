# Changelog, tinkerVault (monorepo)

The build-out history of the **repo itself**, day by day: structure, conventions, the landing page, licensing, and cross-app housekeeping. Each app also keeps its own feature-level changelog for just that app's version history: [`apps/hermes-console/CHANGELOG.md`](apps/hermes-console/CHANGELOG.md), [`apps/markdown-viewer/CHANGELOG.md`](apps/markdown-viewer/CHANGELOG.md), [`apps/whiteboard/CHANGELOG.md`](apps/whiteboard/CHANGELOG.md), [`apps/dossier/CHANGELOG.md`](apps/dossier/CHANGELOG.md).

## 2026-08-12, Hermes Console updated to 1.2.1, plus a missed root-sync catch-up for 1.2.0

- Root-level housekeeping catch-up: the 1.2.0 release (2026-08-07) shipped its own app-level `CHANGELOG.md` entry, `PROJECT_SUMMARY.md` section, and `releases/v1.2.0/` snapshot correctly, but never got its root README table row, archived-versions bullet, structure tree, or a root `CHANGELOG.md` entry updated to match, caught during this pass rather than at the time.
- Updated the Hermes Console app (`apps/hermes-console/`) from 1.2.0 to 1.2.1, tag `hermes-console-v1.2.1`. Two fixes, both in the approval-card path: approvals always failed with HTTP 400 (`submitApproval()` sent `{decision, approved, approval_id?}`; the server only ever reads `choice`), so every approve/deny silently never resolved and just re-showed the same pending card, this was the original loop bug the whole investigation started from. Separately, the card displayed misleading "(unknown)" / "null" for dangerous-command approvals, which only ever carry `command`/`description`, never `name`/`arguments`; replaced with an honest message.
- Archived 1.2.0 remains exactly as originally shipped at `apps/hermes-console/releases/v1.2.0/`; added `apps/hermes-console/releases/v1.2.1/` as the new current snapshot.
- Updated the landing page (`index.html`) and root `README.md` to move the "Latest" badge from the 1.1.0 card/row to 1.2.1, with the archived-versions list now correctly showing 1.2.0 and 1.1.0 both as archived, not just 1.0.0.
- Verified before closing out: `node --check` on the patched script block in all three touched files (root, v1.1.0, the new v1.2.1 snapshot), byte-identical confirmation between root `index.html` and its `releases/v1.2.1/` snapshot, and a live local test against a running Hermes API server confirming the card renders the honest message and the run actually resolves after approve/deny, not just displays cleanly.

## 2026-08-07, Dossier updated to 2.0.1

- Updated the Dossier app (`apps/dossier/`) from 2.0.0 to 2.0.1, tag `dossier-v2.0.1`: gap-fix release, surgical CSS only, no new features. Fixes the Notes and Decisions two-pane layout squishing to the right side at a maximized window: the list/spine pane used a soft, shrinkable flex basis left over from an earlier v1 containment fix that only addressed horizontal scroll and stacking, not the actual width ratio; replaced with a hard-pinned width (matching the pattern Documents' sources rail already used) on both panes.
- Archived the outgoing 2.0.0 build as `apps/dossier/releases/2.0.0/` and added `apps/dossier/releases/2.0.1/` as the new current snapshot, extending the same archived + current pattern used for the 1.0.3 → 2.0.0 transition.
- Published `dossier-v2.0.1` as a GitHub Release with a lean end-user `.zip` asset (`Dossier.exe`, icon, sample files, standalone README), reusing the existing release-scoped badge (`filter=dossier*`).
- Updated the landing page (`index.html`) and root `README.md` to move the "Latest" badge from the 2.0.0 card/row to the new 2.0.1 card/row, with 2.0.0 picking up a "preserved as a self-contained snapshot" line the same way 1.0.3 did when 2.0.0 shipped.
- Verified the 2.0.1 build before closing out: dual rebuild (both exes byte-matched at 12,342,784 bytes), PE version fields confirmed via PowerShell `FileVersionInfo`, an automated headless-Chrome layout harness across four viewports, and a real-hardware hand-check on Windows (Explorer Properties, in-app layout at default/maximized/narrow window sizes for both Notes and Decisions).

## 2026-08-03, Dossier updated to 2.0.0

- Updated the Dossier app (`apps/dossier/`) from 1.0.3 to 2.0.0, tag `dossier-v2.0.0`: adds .odt/.ods import with an import-bookmarks Sources panel, a Notes Edit/Split/Preview Markdown workbench, due dates on Sticky Notes and Kanban cards with a non-destructive sticky-to-Kanban link, a new Agenda view (mini-month navigator plus a due-date list), Collections and workspace display names on the launcher, and an accessibility pass. Also fixes a three-layer Sticky Notes Escape-key bug (a struct-layout ABI mismatch plus a KEY_UP/KEY_DOWN event-kind gap).
- Archived the outgoing 1.0.3 build as `apps/dossier/releases/1.0.3/` and added `apps/dossier/releases/2.0.0/` as the new current snapshot, matching the archived + current pattern already used by Whiteboard and Markdown Viewer.
- Preserved 12 undocumented dev/QA test fixtures turned up by the sanitization sweep as an organized `apps/dossier/build-process/test-fixtures/` folder (Tier 1 and Tier 4 manual regression packs) instead of dropping them as sweep noise, at JP's request, for possible future regression-testing value.
- Published `dossier-v2.0.0` as a GitHub Release with a lean end-user `.zip` asset (`Dossier.exe`, icon, sample files incl. two new odt/ods examples, a standalone README), reusing the existing release-scoped badge (`filter=dossier*`).
- Caught and fixed two live em dashes that had slipped into `apps/dossier/README.md`'s "How to run" section despite an earlier sweep that had reported it clean (the sweep's `grep -P` escaped-byte pattern gave false negatives; switched to a reliable bash/Python method for all sweeps going forward), plus a wrong icon path (`assets/dossier-icon.png` referenced from inside a `releases/` snapshot folder, which has no `assets/` subfolder of its own) in both `releases/2.0.0/README.md` and `releases/1.0.3/README.md`.
- Added Dossier's version history to the landing page (`index.html`) and root `README.md` for the first time: it had only ever appeared as a single unpaired "Latest" card/row for 1.0.3, with no archived-version counterpart; now follows the same two-card, archived-plus-current pattern as the other three apps.
- Verified the 2.0.0 build on real hardware before closing out: a fresh `git clone`, `verify-build.sh` in WSL2, and `verify-windows.ps1` in Windows PowerShell all passed clean.

## 2026-08-02, Dossier 1.0.3 added

- Added a new app, Dossier (`apps/dossier/`), tag `dossier-v1.0.3`, a portable offline dossier workspace for Windows: documents, notes, sticky notes, paint, kanban, annotate, and a decision timeline, plus full-text search and an optional local agent panel.
- First app in this repo that isn't a single-file browser app: Dossier is a compiled native Windows `.exe` (Go + WebView2), so it has no GitHub Pages "Open" link; the landing page and root README link to its GitHub Release download instead.
- Ships the same doc set as Whiteboard/Markdown Viewer (`README.md`, `CHANGELOG.md`, `PROJECT_SUMMARY.md`, `docs/DEV_GUIDE.md`, `docs/USER_GUIDE.md`, `build-process/`, `releases/`), plus a release-scoped badge (`filter=dossier*`).
- Only the current 1.0.3 build is published as a release folder; three earlier internal versions (1.0.0-1.0.2) were icon-only packaging fixes on the same feature set and aren't kept as separate snapshots, see the app's own `CHANGELOG.md`/`PROJECT_SUMMARY.md`.
- Added `dossier-v1.0.3` as a GitHub Release with a lean end-user `.zip` asset (`Dossier.exe`, icon, short README, sample files), matching the Whiteboard/Markdown Viewer release pattern.
- Swept ~110 stray em dashes out of the app's docs, source comments, and UI copy across two follow-up passes (house style is semicolon/comma/parentheses, not em dash); one hand fix was needed for a UI dropdown placeholder that would have looked broken under the mechanical replace.

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
