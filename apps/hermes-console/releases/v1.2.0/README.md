# Hermes Console v1.2.0

A single-file Progressive Web App for resolving **Hermes approval-gated tool
calls** and browsing sessions from your phone or any other device, something no
existing Hermes frontend (Open WebUI included) can do. It talks directly to an
already-running Hermes API server over HTTP.

This folder is a self-contained, independently-runnable snapshot of the v1.2.0
release. For docs, the live app, and the project history, see the repo:
`apps/hermes-console/`.

## What's in this snapshot

```
index.html          the entire app (HTML + inline CSS + JS, no build step)
manifest.json       PWA manifest
sw.js               service worker (cache-first app shell, network-only API)
icon-*.png          app icons (standard + maskable)
apple-touch-icon.png, favicon.*   tab/home-screen icons
docs/USER_GUIDE.md  end-user documentation
VERSION             this release's version (1.2.0)
```

## Run it

Serve this folder over HTTP/HTTPS (the service worker / install features only
activate over http(s), not `file://`):

    python -m http.server 8777
    # open http://localhost:8777/

or drop it onto any static host (GitHub Pages, a tailnet box, …). Then enter your
Hermes server's **host**, **port** (default `8642`), and **bearer key** on the
connection screen.

## v1.2.0 highlights

- **Footer Source / Docs links**: always-visible links (not gated on connection
  state) to the app's GitHub source tree and user guide.
- **Session auto-naming**: when a session has no server title, the console
  derives a display title from the session's first user message (via the
  server's `preview` field) — client-side only, no server writes.
- **Investigated** the refusal-vs-completed-action UI distinction; the server
  exposes no distinguishing field, so no UI change shipped — a one-time console
  warning documents the gap for debugging and flags it as an upstream ask.
- Everything from v1.1.0 and v1.0.0: remote approvals, chat, sessions
  (browse/open/create/fork), installable PWA, polling-based run monitoring.

Full source and developer docs live at `apps/hermes-console/` in the repo.
