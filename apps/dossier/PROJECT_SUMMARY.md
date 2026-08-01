# Dossier — Project Summary

**Current version:** 1.0.3 (packaging + 3 gap-fix passes)  
**Binary:** single Windows GUI `.exe` (Go + WebView2), cross-built from WSL2  
**Repo path:** `dossier-app/`

This document consolidates the tier-by-tier history that previously lived across an internal `STATUS.md` and several `STATUS-GAPFIX-*.md` working files (dev-only, not included in this public copy); this is the narrative map.

---

## Product intent

A **portable dossier workspace**: one folder per project, real files on disk, local full-text search, sticky/paint/kanban/annotate boards, decision spine, optional “ask your local agent” client — as a **native window**, not a browser tab.

---

## Tier history

### Tier 1 — Foundation

- Native Win32 window via **go-webview2** (no CGO)  
- Dossier folder marker + **SQLite FTS5**  
- Import markdown/text/PDF (best-effort) + attachments  
- Notes (`notes/*.md`), sticky notes, search  

### Tier 2 — Boards & multi-workspace

- Drag-and-drop import  
- Paint, Kanban, Annotate  
- Multi-dossier: create/open/switch/recents  
- Decisions + document version snapshots  
- Richer document list (date + preview)  

### Tier 3 / gap-fix 3.1 — Polish

- Twemoji offline flag images (Windows cannot glyph RI pairs)  
- Sticky ink contrast, mini delete, blur/highlight fixes  
- First-open intros, Settings shell, sample-files, generated icon **asset** (not yet PE-embedded)  
- Root cause of the Tier 3 blank-window regression: a bulk Python find-replace mid-pass truncated `ui/index.html` (lost the annotate/kanban sections); restored from the Tier-3-fixed backup and gap-fix edits re-applied surgically (see `STATUS-GAPFIX.md`, "Process note"). Surfaced independently during the v2 self-directed reflection pass, 2026-07-26.  

### Tier 4 / 4.1 / 4.2 — Agent & Office

- Optional **Ask dossier** (host/port/token; hidden when unconfigured)  
- FTS context: stopword OR → **bm25** rank → **top 4**  
- Host first-save UX refinements; `~$` lock skip; toast glyph fix  
- **.docx / .xlsx** extract for search + open externally  

### Tier 5 — Small UX

- Ask **Clear conversation**  
- Documents **sort** (name / date; session `state.docSort`)  

### Packaging → 1.0.0

- Settings **host** as true hide/enable gate (help + toast + behavior aligned)  
- **Context menus** on without enabling DevTools (local go-webview2 patch)  
- **PE icon + FileVersion** via go-winres  
- Notes **Copy** button  
- Docs suite + `build-process/` + `releases/1.0.0/`  

### Packaging gap-fix 1.1 → 1.0.1

- `dossier-icon.png` alpha-channel regression fixed: source had been converted from a flattened JPEG (`jpeg.Decode` + `png.Encode`, no alpha) instead of true RGBA; re-derived via subject mask with real alpha, re-embedded via go-winres
- Per-dossier icon variant explored again, skipped with a concrete reason: one process, one Win32 window, switching dossiers creates no extra taskbar/Alt-Tab entry to badge; icon is set once from the PE resource at window-class creation
- **Open externally** now supports a Dossier-remembered preferred app per file extension (`config.json` → `settings.openWith`), independent of the OS default; "Open with…" menu offers preferred / Windows default / choose-once / always-use / clear
- Notes **auto-save safety net**: shallow one-open-session snapshot with a "Revert edits" control; explicitly not multi-version history, resets on reopening a note

### Packaging gap-fix 1.2 → 1.0.2

- Icon regenerated fresh with image-generation tooling rather than re-masking the degraded JPEG source (no pre-flattening original existed to recover); subject fills ~67% of canvas vs. 1.1's sparse mask-derived art; no alpha regression

### Packaging gap-fix 1.3 → 1.0.3

- Icon's paper/sticky-note art direction reworked: rounded corners on both pieces, jagged/torn sticky edge removed, diagonal/angled placement instead of a square stack, softer tuck into the folder; folder body itself unchanged from 1.2; ~57% opaque fill (down from 1.2's 67% due to the diagonal recompose, not a regression)

---

## Architecture (one-liner)

`main.go` embeds UI → binds `internal/app.API` → `internal/dossier.Store` on the open folder; WebView2 renders `ui/index.html`.

---

## Verification ethos

Each tier was **hand-verified on a real Windows `.exe`**, not only unit tests. Packaging re-checks icon (Explorer/taskbar), version Properties, context menu, Notes Copy, host-gate toast/nav.

---

## Cold read from the agent that built this

_Written 2026-07-26, from a genuinely fresh Grok Build session (real terminal restart, no memory of any prior session) reading only README.md, DEV_GUIDE.md, USER_GUIDE.md, this file, and every STATUS.md across releases/1.0.0/ through releases/1.0.3/. Reproduced verbatim below; this is the actual output, not a paraphrase._

I expected something small. What I found is a full multi-surface Windows productivity app: documents, notes, stickies, paint, annotate, kanban, decisions, FTS search, settings, and an optional "Ask dossier" HTTP client, all as one portable workspace folder with real files on disk plus SQLite. Shipping truth, per root VERSION, internal/app/version.go, winres/winres.json, root STATUS.md, and releases/1.0.3/STATUS.md, is 1.0.3. The binary shape matches the story: GOOS=windows / CGO_ENABLED=0, go-webview2, embedded ui/index.html and offline Twemoji flags, PE icon and version via go-winres, roughly twelve-megabyte releases/1.0.3/Dossier.exe. Architecture in the guides lines up with the source: main.go binds a large Go API surface into WebView2; internal/dossier owns the store and FTS; app prefs live under %APPDATA%\Dossier, workspace data stays in the folder. That portable-folder model is the product idea that actually holds the thing together.

What stands out as clever, and verifiable in tree, is how much packaging grit got treated as product work. Upstream go-webview2 coupled context menus and DevTools to one Debug flag; the local replace under third_party/go-webview2 turns menus on always and leaves DevTools on DOSSIER_DEBUG=1 only. That is a real, focused fork of a dependency rather than a UI workaround. The agent path is also deliberately thin: no model in the binary, host+port gate, FTS loose search with stopwords, OR match, bm25, top N. Icon history is almost comically over-documented for a sticky-and-folder glyph: JPEG-misnamed-as-PNG flattened alpha in 1.0.0, subject-mask RGBA fix in 1.0.1, full regenerate for fill in 1.0.2, art-direction-only paper restyle in 1.0.3. I can verify current assets are true RGBA (color type 6) with corner transparency claims repeated in the 1.0.3 STATUS. The decision to skip per-dossier taskbar icons is written down with a concrete process-model reason, not a vague "later." That honesty is better than shipping a half-broken badge.

What I would flag hard is documentation and verify-script drift after the version number moved. Root VERSION, version.go, and winres say 1.0.3. Root README.md still advertises Version 1.0.0 and tells people to run releases/1.0.0/Dossier.exe. DEV_GUIDE.md titles itself 1.0.1 and still says Version = "1.0.0". USER_GUIDE.md says 1.0.1 at the top and "Footer shows v1.0.0" in getting started. The same stale README/DEV/USER files are copied into releases/1.0.3/, so the ship folder's own docs disagree with that folder's VERSION and STATUS.md. Worse for automation: build-process/verify-build.sh still greps the PE and source for hardcoded 1.0.0 and const Version = "1.0.0", while the real constant is "1.0.3". So either that script is stale relative to the current tree, or it would fail if re-run as written. verify-windows.ps1 is more flexible (it reads VERSION), which makes the bash side look like packaging-era glue that never got generalized. Project summary even leaves a "cold read" section empty on purpose; the release STATUS files are careful, but the living docs did not keep pace with three icon patch ships on the same calendar day (all STATUS dates are 2026-07-26).

Unusual or fragile areas I would not present as bugs without running Windows, but they are called out in the artifacts: WebView2 coordinate automation is repeatedly described as flaky for context menus, Notes Copy clipboard, and interactive smoke; unit tests plus hand-check is the real gate. UI is essentially one ~3.3k-line ui/index.html, which is fine for a single-exe ship and painful for the next person who wants to change one surface without regression. PDF and Office text extract are explicitly best-effort. Notes "Revert edits" is a one-session snapshot, not history. Preferred Open-with is a path map, not full Windows AppX association. There is no git repo at the project root according to the 1.0.2 STATUS (so the icon source recovery story is partly "we had no history to recover"). The build/ directory is a museum of screenshot probes and trial PNGs; useful archaeology, noisy as a workspace.

If I were taking ownership tomorrow, I would treat 1.0.3 STATUS + VERSION + version.go + winres + exe as truth, fix the user-facing and developer docs so they no longer point at 1.0.0/1.0.1, and make verify-build.sh assert the current version from VERSION instead of a frozen string. I would also ask whether three icon-only patch releases were the right cadence, or whether icon work should have stayed unreleased until docs and verify scripts moved with it. That last bit is opinion. What is not opinion is that the product design (folder is the backup, one native window, optional local agent) is coherent in the source, while the packaging narrative after 1.0.0 is currently split across truthful STATUS files and older README/guide text that still sounds like the first ship.

_Follow-up, same day: the doc drift and verify-build.sh findings above were confirmed and fixed. README.md/DEV_GUIDE.md/USER_GUIDE.md version strings corrected (root and releases/1.0.3/ mirrored), and verify-build.sh's two hardcoded "1.0.0" checks (the PE-string grep and the `const Version = "1.0.0"` source gate) were switched to read `$VERSION` instead, then verified by actually running the script end to end: `VERIFY_BUILD_OK version=1.0.3`. See root STATUS.md / release notes for the fix itself; this section stays as the original unedited reflection._

---

## Related status files

| File | Scope |
|------|--------|
| [`releases/1.0.3/STATUS.md`](./releases/1.0.3/STATUS.md) | Verification notes for the current shipped version |

The tier-by-tier development history (Tiers 1-5, Packaging, and gap-fix passes 1.1-1.3) that would otherwise live across a root `STATUS.md` and several `STATUS-GAPFIX-*.md` files, plus three superseded `1.0.0`-`1.0.2` release snapshots, is consolidated into the Tier history above instead; those are internal working files and aren't part of this public copy.

For day-to-day development, prefer **DEV_GUIDE.md** + this summary over re-reading every gap-fix log.
