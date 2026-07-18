# Project Summary, Hermes Console

This is the story behind the tool, not the tool itself. If you just want to use it, `README.md` is right here in the same folder.

## Why this exists

Every existing Hermes frontend, including Open WebUI, talks to the API server-to-server. None of them can do the one thing this project actually needed: surface an approval-gated tool call and let a person resolve it from a phone or any second device. That gap wasn't theoretical. Earlier mobile-access testing (`mobile-access-plan.md`'s test matrix) hit it live: a delete fired without approval because the client had no way to even display the decision, let alone act on it. Checking the full list of frontends Hermes's own docs test against confirmed none of them speak the Runs API either. That's not a client bug anywhere, it's a missing feature category, and building a narrow tool for exactly that gap, not a general-purpose chat client, is the entire scope of this app.

## Build engine and constraint

Built entirely by `cn-fleet`, the isolated Hermes profile running `tencent/hy3:free` via OpenRouter, ahead of that model's free-access window closing (~2026-07-21). Same isolation design as everything else that profile touches: no git access, no credentials beyond its own scoped OpenRouter key, no shared history with the main profile. Whatever happened during the build, the blast radius was a working folder, nothing more.

## Build plan: five phases, ordered so the reason for existing lands first

**Phase 1, approval flow end to end.** This is the entire reason to build this instead of just using Open WebUI. One real unknown going in: the docs describe the approval endpoint's existence and purpose but don't enumerate the exact SSE event shape a run emits when it pauses waiting on a decision. Rather than let the build invent a plausible-looking event name, the plan called for probing a real approval-gated command against a live server first and building the parser against what actually came back, not a guess.

**What that probe actually found.** The target server's `GET /v1/runs/{run_id}/events` endpoint is missing its `Access-Control-Allow-Origin` header, every other endpoint sends CORS correctly, just not that one. A browser blocks the cross-origin read outright. Rather than ship a feature that silently fails in exactly the environment it's meant for (a browser, on a phone, talking cross-origin to a Tailscale host), the app polls `GET /v1/runs/{run_id}` every ~1.5s instead. The fetch-based SSE reader is retained in the code behind `const USE_SSE = false`, not deleted, so flipping one flag is all that's needed if the server-side CORS gap ever closes. The same reasoning applied to `/api/sessions/{id}/chat/stream`, a streaming endpoint avoided for the same cross-origin reason.

**Phase 2, connection and capability handling.** Host/port/key entry, stored per-device in `localStorage` (a real standalone browser app, not a Claude artifact, so this is the correct persistence choice here). `GET /health` and `GET /v1/capabilities` checked on connect, with the UI gated on advertised feature flags rather than assuming a given server build supports everything. This is the documented pattern the Hermes API docs describe for exactly this situation, so external UIs can detect support and fall back safely, not an invented workaround.

**Phase 3, session browser.** List, open, create, and fork sessions via the Sessions API, with defensive field-name access (`sessionId`, `sessionTitle`, `sessionUpdatedAt` helpers probing several plausible keys) since the exact response shape isn't contractually fixed by the docs. The first raw response from list/poll is logged and dropped into an on-screen debug panel specifically so real field names are easy to confirm against, rather than assumed once and forgotten.

**Phase 4, PWA shell and polish.** Manifest, service worker, install prompt, visual pass. Mid-build, an earlier working assumption from Phase 4 planning, that service worker registration was being blocked by the build agent's own sandbox, turned out to be incomplete. A real bug was hiding behind that assumption: `dom.installBtn` was referenced in the boot script before it was ever registered, which threw at load and aborted the rest of the boot sequence, including SW registration, and the resulting error was opaque enough in the sandbox to look like an environment limitation rather than a code bug. Found and fixed in Phase 5, verified clean in a real browser. Worth remembering: a plausible-sounding sandbox excuse is still worth re-checking once a fresh angle is available.

**Phase 5, release packaging.** Matching the precedent already set by Markdown Viewer v1.1.0: a `releases/` snapshot, `USER_GUIDE.md` and `DEV_GUIDE.md`, `VERSION` starting at 1.0.0, root README and CHANGELOG entries.

## What got independently verified, not just self-reported

Approval gating specifically was checked against live screenshots during the actual Phase 4 build, real Run/Reject/Command approval UI appearing mid-build for a Python/PIL dependency check and for an `rm -rf` cleanup step, confirming the gating mechanism works correctly in practice. That mattered because an earlier, unrelated finding (pre-dating this build) had suggested some tool calls went through ungated; this build's own evidence shows gating is command-dependent and working as intended, not uniformly broken.

## Self-improvement, observed twice during this build

The build agent's local `offline-html-tools` skill patched itself during this arc, twice. First, mid-Phase-3, it created an entirely new skill, `browser-api-client`, logging the change inline rather than silently. Second, during Phase 4, it patched that same new skill again, meaning it was iterating on something it had created for itself, not only refining the older skill. Both patches were read in full afterward and judged legitimate, no red flags, but `browser-api-client` specifically was kept local to the cn-fleet profile rather than promoted anywhere else, pending a fuller audit. Worth noting for anyone reading this cold: these review log entries surface several minutes after the actual work finishes, not inline with it, a quiet stretch during a build isn't the same as "nothing happened."

## One naming fix caught before shipping

`DEV_GUIDE.md` originally described the human publishing step as "a human (JP) through the GitHub web UI." Per this project's public-facing naming convention (`webshoppe`/`webShoppe` in anything repo-bound or generated for/by a build agent, "JP" reserved for private/internal notes), that line was corrected to "webShoppe" directly in the source before this release folder was cloned from it, so both the working copy and this shipped copy are clean.

## What shipped in v1.0.0

Connection screen, capability-gated UI, chat, remote approve/deny of tool calls, run stop, session browse/open/create/fork, and an installable PWA shell (manifest, service worker, custom install button, maskable and standard icons). Per-message toolbar (copy, rerun) was scoped and deliberately deferred to v1.1, recorded in this app's own `CHANGELOG.md` so it isn't forgotten. The footer's version field is currently blank on the live server, confirmed as graceful degradation working as designed (`renderPoweredBy()` skips fields it can't find), because the live `/v1/capabilities` response doesn't expose version in any shape the function checks yet, not a bug, just a low-priority v1.1 item.

## Human-in-the-loop publishing

All GitHub operations, commit, push, tag, release, are done by hand by webShoppe through the GitHub web UI. `cn-fleet` never touches git, by design, same as every other app in this repo. This build environment stages upload-ready file batches, grouped by target folder, each with a prepared commit message; nothing is pushed automatically.

## What's carried forward, not part of this app

A display-only bug in Hermes Agent's own dashboard (session titles falling back to raw `run_` ids instead of a message preview) surfaced during this same window. Confirmed by request/response shape (`GET /api/sessions`, port 9119) to live entirely in `NousResearch/hermes-agent`'s own upstream code, not in `hermes-console` and not in any fleet-built code. Deliberately not touched here, live shared infrastructure and someone else's codebase isn't something to hand to an unattended build agent. Logged separately, not yet fixed, two paths identified (a local patch that would get wiped on the next Hermes update, or an upstream PR), neither started.
