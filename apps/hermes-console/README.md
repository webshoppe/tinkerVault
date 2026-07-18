# Hermes Console

[![Release](https://img.shields.io/github/v/release/webshoppe/tinkerVault?filter=hermes-console*&label=release&sort=semver)](https://github.com/webshoppe/tinkerVault/releases?q=hermes-console)

**Current version: 1.0.0**

A single-file Progressive Web App for resolving **Hermes approval-gated tool calls**
and browsing sessions from your phone or any other device — something no existing
Hermes frontend (Open WebUI included) can do. It talks directly to an
already-running Hermes API server over plain HTTP using the Runs and Sessions APIs.

Built as part of the `tinkerVault` monorepo (`apps/hermes-console/`).

---

## What it does

- **Connect** to any Hermes server by host / port / bearer key (stored per-device
  in `localStorage`; nothing is hard-coded).
- **Chat** — send input, watch the assistant response stream in (via polling).
- **Approvals** — when a run pauses on an approval-gated tool call, an Approve /
  Deny card appears with the tool name and arguments. Tap a decision from
  anywhere; the run continues.
- **Stop** an in-flight run.
- **Sessions** — browse, open (with full history), create, and fork sessions.
  Messages thread into one real conversation instead of scattering.
- **Installable** — add to home screen on Android / install on desktop; loads
  offline (the app shell is cached; live API data never is).

---

## Files

| File | Role |
|------|------|
| `index.html` | The entire app — inline CSS + JS, no build step, no runtime dependencies. |
| `manifest.json` | PWA manifest (name, icons, colors, `standalone` display). |
| `sw.js` | Service worker: cache-first app shell, network-only API. |
| `icon-192.png`, `icon-512.png` | App icons (`purpose: any`). |
| `icon-maskable-512.png` | Maskable icon for Android adaptive launchers. |
| `apple-touch-icon.png` | 180×180 iOS home-screen icon. |
| `favicon.ico`, `favicon-16.png`, `favicon-32.png` | Browser-tab favicons. |
| `VERSION` | Semver version string (`1.0.0`). |
| `CHANGELOG.md` | App-level changelog. |
| `PROJECT_SUMMARY.md` | The build story: why this app exists, the phased build, what got verified. |
| `docs/USER_GUIDE.md` | End-user documentation. |
| `docs/DEV_GUIDE.md` | Developer / contributor documentation. |
| `build-process/make_icons.py` | Regenerates all icon PNGs (requires Pillow). Dev-only, not deployed. |
| `build-process/preview.png` | Contact sheet of the three main icons. Dev-only, not deployed. |
| `releases/v1.0.0/` | Self-contained snapshot of the 1.0.0 release. |

Everything except `build-process/`, `docs/`, `PROJECT_SUMMARY.md`, and
`releases/` is part of the deployed shell.

## Documentation

- **[docs/USER_GUIDE.md](docs/USER_GUIDE.md)** — how to use the app.
- **[docs/DEV_GUIDE.md](docs/DEV_GUIDE.md)** — architecture, local dev, release process.
- **[CHANGELOG.md](CHANGELOG.md)** — version history.
- **[PROJECT_SUMMARY.md](PROJECT_SUMMARY.md)** — the build story.

---

## Architecture notes (important)

### Polling, not SSE
The Runs API has an events stream (`GET /v1/runs/{run_id}/events`), but on the
target server **that one endpoint is missing its `Access-Control-Allow-Origin`
header**, so a browser blocks the cross-origin read. Every other endpoint sends
CORS correctly. The app therefore **polls** `GET /v1/runs/{run_id}` every ~1.5 s
instead of opening an event stream.

The fetch-based SSE reader is still in `index.html`, gated behind
`const USE_SSE = false;`. If the server ever starts sending the CORS header on
the events endpoint, flip that to `true` to prefer real-time streaming.

The same assumption applies to `POST /api/sessions/{id}/chat/stream` — it's a
streaming endpoint, so the app avoids it and uses plain GET/POST/PATCH for
sessions.

### Shell vs. API caching
The service worker splits traffic by origin:

- **Same-origin GET** (the static HTML/CSS/JS/icons — the "app shell") →
  **cache-first**, so the console opens instantly and works offline.
- **Cross-origin** (the Hermes server, on a different host:port) →
  **never intercepted**, always live network. Runs and sessions must never be
  served stale.
- **Non-GET** (all API writes) → passed straight through.

The cache is versioned (`SHELL_VERSION` in `sw.js`). **Bump it on every deploy**
so clients drop the old shell instead of getting stuck on it. Old caches are
purged automatically on the service worker's `activate` event.

### No push notifications
Deliberate non-goal — public push infrastructure conflicts with the tailnet-only
constraint.

---

## Configuring `start_url` / `scope` for your deploy

`manifest.json` ships with:

```json
"start_url": "/apps/hermes-console/",
"scope": "/apps/hermes-console/"
```

These two fields are the **only** path-sensitive parts (all other references in
the HTML, manifest, and SW are relative). Set them to match where the app is
actually served:

| Deploy location | `start_url` / `scope` |
|-----------------|-----------------------|
| `https://<user>.github.io/apps/hermes-console/` | `/apps/hermes-console/` (default) |
| `https://<user>.github.io/hermes-console/` (repo named `hermes-console`) | `/hermes-console/` |
| Custom domain root (`https://console.example/`) | `/` |

If the paths don't match, the installed app may launch to a 404 or the service
worker scope won't cover the page.

---

## Deploying to GitHub Pages

Service workers require **HTTPS or localhost**. GitHub Pages is HTTPS, so it works
out of the box.

1. Put these files where Pages will serve them at the path in `start_url`.
   For the default `/apps/hermes-console/`, that means the repo serving Pages has
   the files under `apps/hermes-console/`.
2. Enable Pages (Settings → Pages → deploy from branch, usually `main` / root).
3. Visit `https://<user>.github.io/apps/hermes-console/`.
4. First load registers the service worker; check **DevTools → Application →
   Service Workers** to confirm it's activated, and **Application → Manifest** to
   confirm the icons and colors load.
5. On Android Chrome you'll get an **Install** button in the header (Add to Home
   Screen); on desktop Chrome/Edge, the same button plus the address-bar install
   icon. iOS Safari installs manually via **Share → Add to Home Screen** (no
   `beforeinstallprompt` event exists there — the button just stays hidden).

> **Redeploying:** bump `SHELL_VERSION` in `sw.js` whenever you change any shell
> file, or returning visitors may keep the cached old version until the SW
> updates on its own.

---

## Local testing

Serve over `localhost` (service workers are allowed there) from the **directory
that reproduces your deploy path**:

```bash
# reproduce the /apps/hermes-console/ path
mkdir -p /tmp/site/apps/hermes-console
cp -r * /tmp/site/apps/hermes-console/
cd /tmp/site
python -m http.server 8777
# open http://localhost:8777/apps/hermes-console/
```

Then point the connection screen at your Hermes server's host, port (default
`8642`), and bearer key.

> Opening `index.html` via `file://` will load the UI but the service worker
> won't register (needs http/https). Use a local server for full PWA testing.

---

## Regenerating icons

Icons are generated programmatically so they're reproducible and easy to restyle:

```bash
pip install Pillow
python build-process/make_icons.py
```

Edit the palette constants or the `draw_h_with_wings()` routine at the top of
`build-process/make_icons.py` to change the mark, then re-run. To also refresh the favicons,
regenerate them from `icon-192.png` (see the project history) or add them to the
script.

---

## API endpoints used

All bearer-authenticated (`Authorization: Bearer <key>`), all plain GET/POST/PATCH.

**Runs API**
- `GET /health` — expects `{"status":"ok"}`
- `GET /v1/capabilities` — feature flags gate the UI
- `POST /v1/runs` — `{"input": "...", "session_id": "..."}`
- `GET /v1/runs/{run_id}` — polled for status/output/approval state
- `POST /v1/runs/{run_id}/approval` — approve / deny
- `POST /v1/runs/{run_id}/stop`
- ~~`GET /v1/runs/{run_id}/events`~~ — disabled (missing CORS header)

**Sessions API**
- `GET /api/sessions` — paginated list
- `POST /api/sessions` — create
- `GET /api/sessions/{id}/messages` — history
- `POST /api/sessions/{id}/fork` — branch
- `PATCH /api/sessions/{id}` — update title (available; not yet wired to UI)
- ~~`POST /api/sessions/{id}/chat/stream`~~ — avoided (streaming/CORS assumption)
