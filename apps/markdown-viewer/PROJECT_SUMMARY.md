# Project Summary, Markdown Viewer

This is the story behind the tool, not the tool itself. If you just want to use it, `README.md` is one folder over. This doc is for anyone poking through the source because they're on a similar quest: building small, genuinely self-contained tools with an AI agent driving, and wanting to know what actually happened along the way, not just the finished result.

## What this was really testing

The Markdown Viewer wasn't originally about markdown. It was a calibration test: a cheap, low-stakes way to find out whether a specific free model could genuinely carry a multi-session, autonomous, self-verifying build task before that model's free access route closed. It ran inside an isolated agent profile with no credentials, no git access, and no account-level auth of any kind, deliberately walled off, so whatever happened, the blast radius was "a folder of files," nothing more.

If you're reading this because you're about to hand a similarly open-ended build to an agent you don't have a track record with yet, that framing is the actual reusable lesson: give it a task where failure is cheap and recoverable before you trust it with something that isn't.

## Build, tier by tier

**Tier 1, GFM rendering, syntax highlighting, copy button.** Verified in a real browser run, not by reading its own code back to itself: headings with anchor links, bold/italic/strikethrough/inline code, a real table with a header row, a task list with genuinely checked/unchecked boxes, nested lists, a blockquote, and a code block with real `highlight.js` output. The one honest gap: the agent's own sandboxed browser tool blocked `fetch()` and clipboard *readback*, so the copy button's write path was verified by spying on the captured string and the fallback logic, not by confirming an actual OS paste. It also hit its framework's tool-call iteration cap one step before writing its own summary to disk, a hard ceiling distinct from the API's own rate limit. A single one-line follow-up prompt in the same session was enough to get it to finish writing the file. Worth remembering: an autonomous run stopping just short of the finish line isn't always a real failure, sometimes it's just a nudge away from done.

**Tier 2, theme toggle, table of contents, search, persistence.** Dark/light toggle confirmed via the actual DOM attribute and computed background color, not just a visual glance. TOC collapse, in-document search with real highlight-and-jump behavior, and reload persistence (theme, open file, and TOC state all survived a real reload via `localStorage`) all verified live.

**Tier 3, multi-file tabs, raw/rendered toggle, export.** Two files open and switchable via tabs, raw markdown view toggling correctly, and export producing a genuinely standalone, self-contained HTML file. One honest limitation flagged rather than papered over: the exported file's structural validity was confirmed (closes properly, embeds its own styling and content), but it wasn't actually opened in a second browser session to confirm the visual result matched. That gap got closed later by a real human opening it directly.

## Packaging: what I created

Packaged the already-built, already-verified Markdown Viewer into a distributable `release/` folder (what's now this repo's project root):

```
release/
├── index.html          # byte-identical copy of the verified deliverable
├── README.md           # plain-language user doc (no jargon, no setup, license placeholder)
├── CONTRIBUTING.md     # where features slot in + the known unverified edge case
├── docs/
│   └── USER_GUIDE.md   # every verified feature, in first-time-user discovery order
├── src/
│   ├── template.html
│   ├── assemble.py
│   ├── marked.min.js
│   ├── highlight.min.js
│   ├── github-light.css
│   ├── github-dark.css
│   └── DEV_GUIDE.md     # architecture, placeholder table, exact rebuild command, test note
└── test/
    └── test.md           # the existing fixture, copied as-is
```

12 files total. No git commands, no credentials, no tokens, no network operations beyond reading local files.

**What got verified, climbing tier by tier again on the packaged copy:**

- `release/index.html` confirmed byte-identical to the original verified `index.html` (MD5-matched). Re-checked in a real browser (served over local HTTP): loads with no JS errors, toolbar present, renders correctly. Full feature re-run was deliberately skipped since the file itself was unchanged from the previously verified build.
- **`assemble.py` actually run fresh from `src/`**, not just claimed to work: it read `template.html` plus the four library files and produced a working `index.html` with all six placeholders replaced and zero external `src`/`href` URLs. The build output was then removed from `src/` so that folder holds only the intended source files, confirming `src/` genuinely regenerates the app from scratch, by execution, not assertion.
- `docs/USER_GUIDE.md` covers every feature in the order a first-time user meets them. `src/DEV_GUIDE.md` explains the file layout, the placeholder-to-library mapping, and the exact rebuild command.
- `CONTRIBUTING.md` documents where new features would slot in, plus the one known unverified edge case (export's visual styling).
- Every feature described in the docs is grounded in what was actually verified, nothing is presented as working that wasn't confirmed. The Copy button, for instance, is described as "copies to clipboard" because the write path and button state were confirmed, while the note about clipboard *readback* not being independently confirmable by the automated tool is carried through honestly rather than smoothed over.
- The test-only scaffolding (`harness.html` / `build_harness.py`) built earlier to work around the agent's own sandbox limitations was excluded from `src/` and the shipped tree, and mentioned in `DEV_GUIDE.md` in exactly one line, framed as internal build-session scaffolding, not part of the normal developer workflow. It's preserved separately in [`build-process/`](build-process/) rather than thrown away, see that folder's own README for why.

## Two follow-up passes, and a fourth pass that reversed one of them

**Second pass, internal `.md` link interception, favicon, versioning.** Added a capture-phase click handler in `template.html` that intercepts left-clicks on same-origin relative links ending in `.md`/`.markdown`/`.txt` and loads the target via the existing `openTab()` multi-file-tab system, instead of navigating away. On a failed fetch it catches the error and shows an on-page warning banner instead of falling through to raw-text navigation. External links and in-page anchors are explicitly excluded and keep normal browser behavior. Also added an inline SVG favicon and a `VERSION` file that `assemble.py` reads and stamps into a footer, one source of truth for the version number. This pass ended with an honestly flagged unknown: the agent's own sandbox couldn't confirm whether `fetch()` of a local sibling `.md` file actually succeeds under `file://` in a real, unrestricted browser (Chrome typically blocks it, Firefox is inconsistent). The code was written to handle either outcome gracefully, but resolving which branch real browsers actually take needed a human with a real browser.

**Third pass, removed the clickable cross-file links.** Live testing in two separate real browsers confirmed the `fetch()`-under-`file://` restriction is real and inconsistent across engines. Since a link that reliably fails on every click reads as broken to a first-time user regardless of why, the three cross-file references in `README.md` were converted from clickable links to plain inline-code text, rather than standing up a local server, which would have defeated the entire point of a zero-install, double-click tool. The in-app click-interception code itself was deliberately left in place, since it still has one genuinely working use case: dragging a `.md` file onto an already-open instance of the viewer.

**Fourth pass, reversed the third pass, restored the links, and explained why in the docs themselves.** Re-weighing the actual risk and benefit surfaced something the third pass had missed: the primary way these docs actually get read isn't inside the running app at all, it's on GitHub, in a code editor, or any standard markdown renderer, contexts where these are ordinary relative links that resolve normally with zero JavaScript involved. The `file://`/`fetch()` constraint only applies to the one narrow case of loading `README.md` *through the running app itself* and clicking a link from inside it, and for that one case, the click-interception code from the second pass was never removed, so a click there still shows the graceful on-page message rather than failing silently. The three links in `README.md` were restored to real markdown links, with a short note added directly beneath them:

> These links work normally on GitHub or in a code editor. If you're viewing this file inside the running tool itself, clicking one will show a short message instead of loading it, that's a browser security limit on local files, not a bug. Use the **Open** button or drag the file in instead.

Net effect: correct, standard behavior in the majority of real-world reading contexts, with an already-tested graceful fallback in the one context where the limitation genuinely applies, now clearly explained rather than left implicit. This particular reversal was reasoned through in documentation rather than re-verified in a fresh build session, since it's a markdown-syntax revert of something already built and tested two passes earlier, not new functionality, worth a quick manual visual check that the links render correctly before treating it as final.

## What's still an open, known limitation

The export feature's visual styling, specifically, was never confirmed end-to-end in a from-scratch second browser session as part of the automated build. If you touch the export logic, open a real exported file in a real browser before trusting it.

## v1.1.0: the tab collision bug, an editing pass, and a rebuilt pipeline

A second build round, run months after the v1.0.0 packaging above, started from a real bug report rather than a feature wishlist: two open files that happened to share a filename (say, two different `README.md` from different project folders) silently collapsed into a single tab, one clobbering the other. Root cause was exactly what it sounds like, the tab system keyed tabs by filename string instead of anything unique. The fix keys every tab by a fresh per-load id instead, confirmed live by opening two same-named fixture files (kept in the repo as `test/fixtures/projectA/` and `test/fixtures/projectB/`, marked `ALPHA-123` and `BRAVO-456` so a mixup would be obvious) and checking both rendered with distinct content, independently closeable.

With the collision fixed, four more real-usage requests came in across three follow-up sessions: a tab-close button that could get clipped behind a long filename (fixed by moving the ellipsis truncation off the whole tab row and onto just the name label), a browser-tab title that tracks the active document, a "Save" action to get edited content out of the browser as a new file, and a rebuilt template+assemble pipeline so the app could keep being hand-edited rather than patched in place inside one 200KB file. A deeper pass through the original Tier 2 wishlist added multi-file open, edit mode, per-tab scroll memory, keyboard shortcuts, reopen-last-closed-tab, and Copy MD. Each item was verified against the actual running code (real `openTab`/`closeTab`/`setActive` calls, synthetic keyboard events, DOM measurements for the clipped-button fix), not asserted from reading the diff.

One gap got flagged, then genuinely closed rather than left open: the first pass couldn't verify the localStorage round-trip (save state, reload, restore) because the sandboxed test harness blocks direct storage access. That gap was closed by an actual human opening two same-named files, reloading the page, and confirming both tabs came back, exactly the kind of manual verification this project treats as non-optional for anything a sandboxed tool can't reach on its own.

A later cold-read pass, someone reading the shipped docs fresh rather than the code, caught three documentation-only issues: a stale example of a title format that had since been simplified away, two real shipped features (Copy MD, Save) missing from the README's feature list, and an editing description that implied a live on-screen preview when the actual behavior is textarea-then-refresh-on-exit. All three were doc corrections, no code changed.

**Folding this into the monorepo** (this pass, 2026-07-17): the verified v1.1.0 build replaced the app's root files, the previous v1.0.0 build was preserved unmodified as a snapshot in `releases/v1.0.0/` (mirroring the pattern already established for Whiteboard), and `CHANGELOG.md`/`CONTRIBUTING.md`/this file were updated to match. One incidental finding during the fold: rerunning `assemble.py` in a Linux shell instead of the original Windows build environment produces a byte-for-byte identical page with different line endings only (CRLF vs LF), a build-environment quirk, not a functional regression, now documented in `src/DEV_GUIDE.md` so a future contributor doesn't mistake it for one.

## Why this doc exists at all

The honest answer: because countless other people decided documenting the "simple," "obvious" stuff, the trials, the wrong turns, the constraints that turned out to be more (or less) real than expected, was worth their time, even when it felt like overkill for something this small. That's a large part of why an agent could help build this at all. This doc is that same bet, paid forward.
