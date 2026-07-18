# Hermes Console

[![Release](https://img.shields.io/github/v/release/webshoppe/tinkerVault?filter=hermes-console*&label=release&sort=semver)](https://github.com/webshoppe/tinkerVault/releases?q=hermes-console)

A single-file Progressive Web App for resolving **Hermes approval-gated tool calls**
and browsing sessions from your phone or any other device, something no existing
Hermes frontend (Open WebUI included) can do. It talks directly to an
already-running Hermes API server over plain HTTP using the Runs and Sessions APIs.

This is an archived, self-contained snapshot of **v1.0.0**, preserved exactly as
it shipped. For the current version, see the
[main app folder](https://github.com/webshoppe/tinkerVault/tree/main/apps/hermes-console)
on GitHub.

## What it does

- **Connect** to any Hermes server by host / port / bearer key (stored per-device
  in `localStorage`; nothing is hard-coded).
- **Chat**, send input, watch the assistant response stream in (via polling).
- **Approvals**, when a run pauses on an approval-gated tool call, an Approve /
  Deny card appears with the tool name and arguments. Tap a decision from
  anywhere; the run continues.
- **Stop** an in-flight run.
- **Sessions**, browse, open (with full history), create, and fork sessions.
- **Installable**, add to home screen on Android / install on desktop; loads
  offline (the app shell is cached; live API data never is).

## Configuring `start_url` / `scope` for your deploy

`manifest.json` ships with `start_url`/`scope` set to `/apps/hermes-console/`.
These are the only path-sensitive parts of this app; if you serve this snapshot
from a different location, update both fields to match, or the installed app
may launch to a 404.

## Files in this snapshot

```
index.html          the app, exactly as it shipped for v1.0.0
manifest.json       PWA manifest
sw.js               service worker
icon-192.png, icon-512.png, icon-maskable-512.png, apple-touch-icon.png
favicon.ico, favicon-16.png, favicon-32.png
VERSION
README.md            this file
docs/
  USER_GUIDE.md       every feature, from a first-time user's perspective
```

Intentionally minimal, just what's needed to run this exact version standalone,
no external dependencies. For source code, the changelog, and newer versions,
see the [main app folder](https://github.com/webshoppe/tinkerVault/tree/main/apps/hermes-console).

## License

MIT, see the repo's [LICENSE](https://github.com/webshoppe/tinkerVault/blob/main/LICENSE)
file. See this version's entry in
[CHANGELOG.md](https://github.com/webshoppe/tinkerVault/blob/main/apps/hermes-console/CHANGELOG.md)
for what changed since.

## Where to get help

- For a full walkthrough of every feature, see [`docs/USER_GUIDE.md`](docs/USER_GUIDE.md) (included in this folder).
- For technical details, architecture, and the release process, see the current [`docs/DEV_GUIDE.md`](https://github.com/webshoppe/tinkerVault/blob/main/apps/hermes-console/docs/DEV_GUIDE.md) in the main app folder.
- The full build story, phased build plan, and what got independently verified, is in [`PROJECT_SUMMARY.md`](https://github.com/webshoppe/tinkerVault/blob/main/apps/hermes-console/PROJECT_SUMMARY.md) in the main app folder.
