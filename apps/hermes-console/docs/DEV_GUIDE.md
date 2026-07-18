# Hermes Console — Developer Guide

Everything you need to understand, modify, and re-release this app. It's a
single self-contained `index.html` (inline CSS + JS, no build step, no runtime
dependencies) plus a small PWA shell.

---

## 1. Layout

```
apps/hermes-console/
├── index.html              # the entire app (HTML + inline CSS + inline JS)
├── manifest.json           # PWA manifest
├── sw.js                   # service worker (shell cache + API passthrough)
├── icon-192.png            # app icon (purpose: any)
├── icon-512.png            # app icon (purpose: any)
├── icon-maskable-512.png   # maskable icon (Android adaptive)
├── apple-touch-icon.png    # 180x180 iOS home-screen icon
├── favicon.ico             # tab favicon (16/32/48)
├── favicon-16.png
├── favicon-32.png
├── VERSION                 # semver, single line
├── CHANGELOG.md            # app-level changelog
├── PROJECT_SUMMARY.md      # the build story
├── README.md               # overview + deploy notes
├── docs/
│   ├── USER_GUIDE.md       # end-user documentation
│   └── DEV_GUIDE.md        # this file
├── build-process/
│   ├── make_icons.py       # regenerates all icons (needs Pillow) — dev only
│   └── preview.png         # icon contact sheet — dev only
└── releases/
    ├── README.md           # index of archived versions
    └── v1.0.0/             # self-contained snapshot of this release
```

`build-process/` and the `releases/` snapshot are developer artifacts; the
deployed *shell* is everything else in the app folder.

---

## 2. Running locally

Service workers require **HTTPS or localhost**, so use a local server (not
`file://`) for full PWA testing. Serve from a directory that reproduces the
deploy path so `start_url`/`scope` resolve:

```bash
mkdir -p /tmp/site/apps/hermes-console
cp index.html manifest.json sw.js *.png *.ico /tmp/site/apps/hermes-console/
cd /tmp/site
python -m http.server 8777
# open http://localhost:8777/apps/hermes-console/
```

Point the connection screen at your Hermes server. Opening `index.html` directly
via `file://` loads the UI but the service worker won't register.

There is **no build step and no test framework**. Validate before shipping with:

```bash
node --check sw.js                 # service worker syntax
node -e "JSON.parse(require('fs').readFileSync('manifest.json','utf8'))"
# and confirm the inline app script parses:
node -e "const h=require('fs').readFileSync('index.html','utf8'); new Function(h.match(/<script>([\s\S]*)<\/script>/)[1]); console.log('ok')"
```

Then load it in a real browser and watch the console — the app logs every run
event, poll, and API request/response.

---

## 3. Architecture

### State
A single `state` object holds config, capabilities/features, the active run id,
polling flags, streaming assistant text, approval state, and the active session
id. `dom` caches element references. There's no framework — plain DOM APIs.

### Connection
`connect()` runs `GET /health` then `GET /v1/capabilities`, stores the response,
and calls `renderCapabilities()` + `renderPoweredBy()`. `featureOn(name)` gates
UI on the reported `features` booleans.

### Sending + monitoring
`sendMessage()` POSTs to `/v1/runs` with `{input, session_id?}`. Because the SSE
events endpoint is CORS-blocked (see below), it then **polls**
`GET /v1/runs/{run_id}` on `POLL_INTERVAL_MS` (1500 ms) via `pollOnce()` until a
terminal status or an approval-pending state.

### Approval
`isApprovalPending(event)` is the single, isolated detector — deliberately one
function so the real event/field shape can be corrected in one place after a
live test. When it fires, `handleApprovalPending()` renders the Approve/Deny
card and `submitApproval()` POSTs to `/v1/runs/{id}/approval` (logging the full
request and response both ways). Polling resumes after the decision.

### Sessions
`loadSessions()`, `selectSession()`, `createSession()`, `forkSession()` cover
the Sessions API. Field-name access goes through defensive helpers
(`sessionId`, `sessionTitle`, `sessionUpdatedAt`) that probe several plausible
keys, because the exact response shape isn't contractually fixed. The first raw
response from list/poll is logged **and** dropped into an on-screen debug panel
so real field names are easy to confirm.

### PWA
`sw.js` precaches the shell on `install`, purges old caches on `activate`, and
routes fetches: same-origin GET → cache-first; cross-origin → not intercepted
(always live); non-GET → passthrough. Install is a custom button driven by
`beforeinstallprompt`.

---

## 4. Why polling instead of SSE (important)

The Runs API exposes `GET /v1/runs/{run_id}/events` (SSE), but on the target
server **that endpoint alone is missing its `Access-Control-Allow-Origin`
header**, so a browser blocks the cross-origin read (`200`, but no CORS header).
Every other endpoint sends CORS correctly.

The app therefore polls `GET /v1/runs/{run_id}`. The full fetch-based SSE reader
(`streamEvents()`, which can send an `Authorization` header — native
`EventSource` can't) is **retained** behind:

```js
const USE_SSE = false;
```

If the server is fixed to send the CORS header on the events endpoint, flip that
to `true` to prefer real-time streaming. Same caution applies to
`POST /api/sessions/{id}/chat/stream` — a streaming endpoint, so it's avoided in
favor of plain GET/POST/PATCH.

---

## 5. Field-name assumptions to verify against a live server

These are read defensively but were not fully pinned by documentation. Confirm
against the logged raw responses if behavior looks off:

- **Run object** (`GET /v1/runs/{id}`): `status`, and output under
  `output`/`text`/`content`/`messages[]`.
- **Approval shape**: `isApprovalPending()` looks for `status`
  `requires_approval`/`pending_approval` or a type/status containing
  `"approval"`, including nested `data`.
- **Approval body**: sent as `{decision, approved, approval_id?}` — a defensive
  superset; trim once the server's expected shape is confirmed.
- **Capabilities identity** (`renderPoweredBy()`): probes
  `agent`/`name`/`model` and `version`/`build` at top level and under
  `agent`/`model`/`server`/`info`.
- **Sessions**: list envelope (bare array or `sessions`/`items`/`results`/`data`),
  and per-session `id`/`title`/`updated_at` via the helper accessors.

---

## 6. Icons

Generated programmatically for reproducibility:

```bash
pip install Pillow
python build-process/make_icons.py
```

Edit the palette constants or `draw_h_with_wings()` at the top of the script,
re-run, and (if you change the shell asset list) bump `SHELL_VERSION` in `sw.js`.

---

## 7. Deploy & release process

Hosting and release notes are in `README.md`. Repo-wide conventions for the
tinkerVault monorepo:

- **App folder** lives at `apps/hermes-console/` in the repo.
- **`start_url`/`scope`** in `manifest.json` are the only path-sensitive fields;
  they default to `/apps/hermes-console/` for a Pages project site at
  `https://<user>.github.io/apps/hermes-console/`. Adjust if the deploy path
  differs.
- **Release tags are app-prefixed** (`hermes-console-v1.0.0`) because tags are
  repo-wide and a bare `v1.0.0` would collide with other apps.
- **README release badge** is scoped `filter=hermes-console*` so it never
  reflects another app's release.
- **Versioned snapshot**: each release is copied, complete and independently
  runnable, into `releases/v<version>/`, separate from the live working copy.
- **Bump `SHELL_VERSION`** in `sw.js` on any shell change so returning visitors
  drop the cached old version.

### Human-in-the-loop publishing

All GitHub operations — commit, push, tag, PR, Release publish — are done **by a
human (webShoppe) through the GitHub web UI**. This build environment is intentionally
barred from git. Changes are staged as upload-ready file batches grouped by
target folder, each with a prepared commit message; a human uploads them via
**Add file → Upload files** and clicks Publish. Nothing is pushed automatically.

### Verifying a change landed

Because uploads go through the web UI (not a local push), "uploaded" and "live
and correct" are worth checking separately. Mounted-folder shell reads and raw
GitHub mirrors can show stale cached content for a few minutes after an edit or
push — if a cached read contradicts what was just done, re-check with a fresh
direct read or the actual hosted page before concluding anything.
