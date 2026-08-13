# Changelog: Hermes Console

All notable changes to this app are documented here. Format follows
[Keep a Changelog](https://keepachangelog.com/), and this project adheres to
[Semantic Versioning](https://semver.org/). Release tags are app-prefixed
(`hermes-console-v1.0.0`) because tags are repo-wide across the tinkerVault
monorepo.

## [1.2.1] - 2026-08-12

### Fixed
- **Approval POST always failed with HTTP 400, approvals never actually
  resolved.** `submitApproval()` sent `{decision, approved, approval_id?}`.
  The server's approval endpoint only ever reads a `choice` field
  (`"once"|"session"|"always"|"deny"`, with `"approve"`/`"approved"`/`"allow"`
  aliased to `"once"`), keyed on `run_id` alone, no `approval_id` needed. Since
  `choice` was never sent, every approval decision was rejected, so the run
  never actually resumed, it just re-polled back into the same pending card,
  looking like a UI loop. Now sends `{choice: decision}` directly.
- **Approval card showed misleading "(unknown)" / "null".** Dangerous-command
  approvals only ever carry `command`/`description`, never a `name` or
  `arguments` pair, so `handleApprovalPending()` now shows an honest message
  explaining tool details aren't available over this polling transport
  (SSE is disabled server-side), instead of implying missing data.

## [1.2.0] - 2026-08-07

Three scoped changes: a footer link row, session auto-naming, and a completed
investigation into distinguishing refusals from completed actions in the UI.

### Added
- **Footer Source / Docs links.** A new `.footer-links` row under the
  `powered-by` line links to the app's GitHub source tree and user guide.
  Both links are `target="_blank" rel="noopener"`, separated by a middot, and
  use the existing muted/accent palette. They are **always visible**, not
  gated on connection state — the footer renders regardless of whether the
  console is connected.

- **Session auto-naming (client-side, display-only).** When a session has no
  server-provided title, the console now derives a display title from the
  session's first user message. The server already exposes that text via the
  `preview` field on each session object, so `sessionDisplayTitle()` takes its
  first line and uses it as the display name in both the session list and the
  active-session header; otherwise it falls back to the raw session id. This
  is purely a render-time derivation — there is **no server write** (no PATCH),
  so it cannot clobber an existing manual/server title and needs no extra
  round-trip.

### Investigated
- **Refusal-vs-completed-action UI distinction.** Probed the live API
  (`GET /api/sessions/{id}/messages` across 40 sessions / 299 assistant
  messages, plus the `GET /v1/runs/{id}` poll object) for any field separating
  a refusal from a normal completed answer. Confirmed the server exposes **no**
  such field: a refusal is structurally identical to an ordinary answer
  (`role:"assistant"`, `finish_reason:"stop"`, `tool_calls:null`, with the
  only difference living in the free-text content). There is no refusal
  boolean / `finish_reason`, no `tool_name` signal, and the run object carries
  none of `output`/`content`/`tool_calls`/`finish_reason` either. A faithful UI
  distinction is therefore impossible from API fields alone, and per the brief
  it must not be faked with text-pattern matching. No UI change shipped for
  this; instead a **one-time `console.warn("[gap] ...")`** in `renderHistory()`
  documents the gap for anyone debugging later, and flags it as a real upstream
  ask: add a `refused:true` or `finish_reason:"refusal"` field to the message
  object so the client can render it cleanly.

[1.2.0]: https://github.com/webshoppe/tinkerVault/releases/tag/hermes-console-v1.2.0

## [1.1.0] - 2026-07-18

Two scoped fast-follows on top of the shipped v1.0.0 app:
a per-message toolbar, plus a footer version-field investigation.

### Added
- **Per-message toolbar.** Every rendered chat message now carries a small
  action row:
  - **Copy** (all messages): copies that message's raw text via the Clipboard
    API (`navigator.clipboard.writeText`), with a `document.execCommand`
    fallback for non-secure contexts. The button flashes "Copied" on success.
  - **Rerun** (user messages only): resubmits that exact input as a brand-new
    run through the existing `sendMessage(text)` path in the current session;
    no history replay, no inline edit (both out of scope for this release).
  - The toolbar is hidden until the message is hovered (desktop) and stays
    inside the message bubble; system/status/error lines are unaffected because
    they don't go through `addMessage()`.

- **App version label.** Hermes Console now displays its own version next to
  the header title (e.g. "Hermes Console v1.1.0"), read from `./VERSION` at
  runtime rather than hardcoded, so it can't drift from this file or
  `README.md`. Falls back to a static placeholder if the fetch can't run
  (e.g. opened via `file://`). Distinct from, and does not touch, the
  unrelated `powered-by` footer line below.

### Changed
- `sendMessage()` now accepts an optional `text` argument (defaults to the
  composer input) so Rerun can reuse the identical send/stream/poll flow.
- Service worker precache list now includes `VERSION`; cache version bumped
  `v2` → `v3` so returning visitors pick up both changes.

### Investigated
- **Footer "Powered by" version field.** Probed the full live
  `GET /v1/capabilities` response (every key) for a version-like field. None
  exists: top-level keys are `object`, `platform`, `model`, `auth`, `runtime`,
  `features`, `endpoints`; no `version`/`build`/`release`/`agent_version`
  anywhere. The footer already renders correctly from `model` (`"Powered by:
  hermes-agent"`), so this was confirmed a non-issue, not a bug. No code change.
  (Noted here so a future reader doesn't re-litigate it.)

[1.1.0]: https://github.com/webshoppe/tinkerVault/releases/tag/hermes-console-v1.1.0

## [1.0.0] - 2026-07-18

First public release. A single-file PWA for resolving Hermes approval-gated
tool calls and browsing sessions from a phone or any other device, talks
directly to a running Hermes API server over HTTP.

### Added
- **Connection screen**: host / port (default `8642`) / bearer key, stored
  per-device in `localStorage` (`hermesConsoleConfig`). Nothing hard-coded.
- **Health + capability check** on connect (`GET /health`,
  `GET /v1/capabilities`); UI features are gated on the reported `features`
  flags (`run_submission`, `run_approval`, `run_events_sse`, `run_stop`).
- **Chat**: `POST /v1/runs` with `{input, session_id}`; assistant output
  rendered as it arrives.
- **Approval flow**: pending approval-gated tool calls surface an Approve /
  Deny card (tool name + arguments); the decision POSTs to
  `/v1/runs/{id}/approval` and the run resumes. Detection is isolated in a
  single `isApprovalPending()` function for easy correction.
- **Stop**: `POST /v1/runs/{id}/stop` for in-flight runs.
- **Sessions**: browse (`GET /api/sessions`), open with full history
  (`GET /api/sessions/{id}/messages`), create (`POST /api/sessions`), and fork
  (`POST /api/sessions/{id}/fork`). Selecting a session threads all subsequent
  messages into it via `session_id`.
- **PWA shell**: `manifest.json`, service worker (cache-first app shell,
  network-only for API traffic, versioned cache), custom Install button
  (`beforeinstallprompt`), maskable + standard icons, favicons.
- **"Powered by" footer**: shows the agent/model identity reported by
  `GET /v1/capabilities` (read-only, no new endpoint).

### Fixed
- **Install button / PWA layer was silently broken.** `dom.installBtn` was
  referenced in the event wiring but never added to the `dom` element map, so
  `dom.installBtn.addEventListener(...)` threw at load; which aborted the rest
  of the boot script, including service-worker registration. Caught during the
  v1.0.0 packaging test pass (the sandbox that first exercised the SW had masked
  it with an opaque exception). Added the missing `installBtn: el("installBtn")`
  reference; install prompt and SW registration now both initialize cleanly
  (verified: `[pwa] install available` and `[pwa] service worker registered`
  with zero JS errors).

### Known limitations
- **Run monitoring uses polling, not SSE.** The events endpoint
  (`GET /v1/runs/{run_id}/events`) is missing its `Access-Control-Allow-Origin`
  header on the target server, so the browser blocks the cross-origin read.
  The app polls `GET /v1/runs/{run_id}` (~1.5 s) instead. The fetch-based SSE
  reader is retained behind `const USE_SSE = false;` and can be re-enabled if
  the server ever sends the CORS header on that endpoint.
- Session **rename** (`PATCH /api/sessions/{id}`) is available server-side but
  not yet wired into the UI.

### Planned for v1.1
- **Per-message toolbar** (copy, rerun, etc.). Deliberately out of scope for
  the v1.0.0 close-out; tracked as the primary v1.1 follow-up.

[1.0.0]: https://github.com/webshoppe/tinkerVault/releases/tag/hermes-console-v1.0.0
