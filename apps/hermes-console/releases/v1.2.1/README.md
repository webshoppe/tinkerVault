# Hermes Console v1.2.1

A single-file Progressive Web App for resolving **Hermes approval-gated tool
calls** and browsing sessions from your phone or any other device, something no
existing Hermes frontend (Open WebUI included) can do. It talks directly to an
already-running Hermes API server over HTTP.

This folder is a self-contained, independently-runnable snapshot of the v1.2.1
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
VERSION             this release's version (1.2.1)
```

## Run it

Serve this folder over HTTP/HTTPS (the service worker / install features only
activate over http(s), not `file://`):

    python -m http.server 8777
    # open http://localhost:8777/

or drop it onto any static host (GitHub Pages, a tailnet box, …). Then enter your
Hermes server's **host**, **port** (default `8642`), and **bearer key** on the
connection screen.

## v1.2.1 highlights

Two fixes, both in the approval-card path:

- **Approvals actually resolve now.** `submitApproval()` was sending
  `{decision, approved, approval_id?}`, but the server's `_handle_run_approval`
  only ever reads `choice`. Every approve/deny had been failing with HTTP 400
  since v1.0.0, silently, so the card just re-showed the same pending state and
  looked like a UI loop rather than a rejected request. Now sends
  `{choice: decision}` directly, matching the server contract.
- **Honest tool-detail messaging.** Dangerous-command approvals only ever carry
  `command`/`description`, never `name`/`arguments`, so the card used to show a
  misleading "(unknown)" / "null". It now shows a plain message explaining that
  tool details aren't available over this transport, and you can still approve
  or deny normally.
- Everything from v1.2.0, v1.1.0, and v1.0.0: footer Source/Docs links, session
  auto-naming, per-message toolbar (copy, rerun), remote approvals, chat,
  sessions (browse/open/create/fork), installable PWA, polling-based run
  monitoring.

Full source and developer docs live at `apps/hermes-console/` in the repo.
