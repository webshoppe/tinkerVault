# Dossier — Project Summary

**Current version:** 2.0.0 (v2 packaging ship; last v1 ship was 1.0.3)
**Binary:** single Windows GUI `.exe` + console twin (Go + WebView2), cross-built from WSL2
**Repo path:** `dossier-app/`

This document consolidates tier-by-tier history that previously lived across an internal `STATUS.md` and several `STATUS-GAPFIX-*.md` working files (dev-only, not included in this public copy); this is the narrative map. Root `STATUS.md` is overwritten each ship pass; `releases/<ver>/STATUS.md` preserves a snapshot.

---

## Product intent

A **portable dossier workspace**: one folder per project, real files on disk, local full-text search, sticky/paint/kanban/annotate boards, decision spine, Agenda of due dates, multi-folder Collections, optional "ask your local agent" client — as a **native window**, not a browser tab.

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

### v2 Tier 1 → 2.0.0-t1

- Opt-in **Open last workspace on launch**; `DOSSIER_AUTO_OPEN` remains separate
- Best-effort **.odt / .ods** extract + import + rescan
- Schema **v3**: `stickies.due_at` (later UI)

### v2 Tier 2 → 2.0.0-t2

- Workspace display names; color picker presets; Notes **Edit/Split/Preview** + vendored marked/hljs
- Annotate handle polish; launcher arrow tagline

### v2 Tier 3 → due dates & sticky↔kanban

- Sticky and kanban card due dates
- Non-destructive sticky-to-kanban link (sync text/due; unlink keeps both)
- Agenda surface prep / list aggregation

### v2 Tier 4 → Collections & bookmarks

- **Collections** of dossier folders (launcher portfolio)
- Import **bookmarks** (`bookmarks.json` per dossier)
- Switcher/collection UX refinements (t4.1–t4.2)

### v2 Tier 5 → Accessibility baseline & keyboard

- Landmarks, names, focus-visible, labels, contrast
- Keyboard widgets; accordion Space fix; launcher scroll containment
- Dark scrollbars; toast live region; Kanban tab order; picker focus

### v2 Accessibility Escape (3.1–3.6) → sticky picker

Three failed attempts, then a real root cause:

1. **ABI layout:** `COREWEBVIEW2_PHYSICAL_KEY_STATUS` used Go `bool` for Windows `BOOL` (4-byte), so `WasKeyDown` mis-read and the host accelerator callback body often never ran.
2. **Event kind:** ordinary keys arrive as **KEY_DOWN (0)**; Escape arrives as **KEY_UP (1)** on this host. A KEY_DOWN-only gate silently dropped Escape.

**Fix (kept in 2.0.0):** correct BOOL to `int32`; invoke `AcceleratorKeyCallback` for Escape on KEY_UP/SYSTEM_KEY_UP as well as the normal KEY_DOWN path; page handlers + host Eval close open sticky popovers. Hand-verified with NVDA after the KEY_UP fix, confirmed four separate times across the session including once on the Color picker in addition to Emoji. One honest gap: no screenshot of the literal Escape keypress closing a picker; this rests on a checked hand-verification checklist item plus the NVDA log evidence, not a smoking-gun screenshot.

### Polish wishlist / Tier 6 close-out (t6–t7)

- Em dash cleanup; copyright **webShoppe**; user-facing Calendar → **Agenda**
- Strip Escape diag logging; version in launcher/sidebar; route-change focus to view title
- Kanban Move… arrow keys; console **OriginalFilename** fix; icon history in DEV_GUIDE

### Packaging → **2.0.0**

- All living docs refreshed to match full v2 feature set
- `releases/2.0.0/` ship folder (both exes + docs)
- `verify-build.sh` asserts version from `VERSION` file, not a hardcoded literal

---

## Architecture (one-liner)

`main.go` embeds UI → binds `internal/app.API` → `internal/dossier.Store` on the open folder; WebView2 renders `ui/index.html`.

---

## Verification ethos

Each tier was **hand-verified on a real Windows `.exe`**, not only unit tests. Agent UI automation (PrintWindow/SendKeys) is often flaky; STATUS files record what was automated vs. hand-checked. Packaging re-checks icon (Explorer/taskbar/Task Manager), version Properties, context menu, Notes Copy, host-gate toast/nav.

---

## Cold read from the agent that built this — v1.0.3 (2026-07-26)

_Written 2026-07-26, from a genuinely fresh Grok Build session (real terminal restart, no memory of any prior session) reading only README.md, DEV_GUIDE.md, USER_GUIDE.md, this file, and every STATUS.md across releases/1.0.0/ through releases/1.0.3/. Reproduced verbatim below; this is the actual output, not a paraphrase._

I expected something small. What I found is a full multi-surface Windows productivity app: documents, notes, stickies, paint, annotate, kanban, decisions, FTS search, settings, and an optional "Ask dossier" HTTP client, all as one portable workspace folder with real files on disk plus SQLite. Shipping truth, per root VERSION, internal/app/version.go, winres/winres.json, root STATUS.md, and releases/1.0.3/STATUS.md, is 1.0.3. The binary shape matches the story: GOOS=windows / CGO_ENABLED=0, go-webview2, embedded ui/index.html and offline Twemoji flags, PE icon and version via go-winres, roughly twelve-megabyte releases/1.0.3/Dossier.exe. Architecture in the guides lines up with the source: main.go binds a large Go API surface into WebView2; internal/dossier owns the store and FTS; app prefs live under %APPDATA%\Dossier, workspace data stays in the folder. That portable-folder model is the product idea that actually holds the thing together.

What stands out as clever, and verifiable in tree, is how much packaging grit got treated as product work. Upstream go-webview2 coupled context menus and DevTools to one Debug flag; the local replace under third_party/go-webview2 turns menus on always and leaves DevTools on DOSSIER_DEBUG=1 only. That is a real, focused fork of a dependency rather than a UI workaround. The agent path is also deliberately thin: no model in the binary, host+port gate, FTS loose search with stopwords, OR match, bm25, top N. Icon history is almost comically over-documented for a sticky-and-folder glyph: JPEG-misnamed-as-PNG flattened alpha in 1.0.0, subject-mask RGBA fix in 1.0.1, full regenerate for fill in 1.0.2, art-direction-only paper restyle in 1.0.3. I can verify current assets are true RGBA (color type 6) with corner transparency claims repeated in the 1.0.3 STATUS. The decision to skip per-dossier taskbar icons is written down with a concrete process-model reason, not a vague "later." That honesty is better than shipping a half-broken badge.

What I would flag hard is documentation and verify-script drift after the version number moved. Root VERSION, version.go, and winres say 1.0.3. Root README.md still advertises Version 1.0.0 and tells people to run releases/1.0.0/Dossier.exe. DEV_GUIDE.md titles itself 1.0.1 and still says Version = "1.0.0". USER_GUIDE.md says 1.0.1 at the top and "Footer shows v1.0.0" in getting started. The same stale README/DEV/USER files are copied into releases/1.0.3/, so the ship folder's own docs disagree with that folder's VERSION and STATUS.md. Worse for automation: build-process/verify-build.sh still greps the PE and source for hardcoded 1.0.0 and const Version = "1.0.0", while the real constant is "1.0.3". So either that script is stale relative to the current tree, or it would fail if re-run as written. verify-windows.ps1 is more flexible (it reads VERSION), which makes the bash side look like packaging-era glue that never got generalized. Project summary even leaves a "cold read" section empty on purpose; the release STATUS files are careful, but the living docs did not keep pace with three icon patch ships on the same calendar day (all STATUS dates are 2026-07-26).

Unusual or fragile areas I would not present as bugs without running Windows, but they are called out in the artifacts: WebView2 coordinate automation is repeatedly described as flaky for context menus, Notes Copy clipboard, and interactive smoke; unit tests plus hand-check is the real gate. UI is essentially one ~3.3k-line ui/index.html, which is fine for a single-exe ship and painful for the next person who wants to change one surface without regression. PDF and Office text extract are explicitly best-effort. Notes "Revert edits" is a one-session snapshot, not history. Preferred Open-with is a path map, not full Windows AppX association. There is no git repo at the project root according to the 1.0.2 STATUS (so the icon source recovery story is partly "we had no history to recover"). The build/ directory is a museum of screenshot probes and trial PNGs; useful archaeology, noisy as a workspace.

If I were taking ownership tomorrow, I would treat 1.0.3 STATUS + VERSION + version.go + winres + exe as truth, fix the user-facing and developer docs so they no longer point at 1.0.0/1.0.1, and make verify-build.sh assert the current version from VERSION instead of a frozen string. I would also ask whether three icon-only patch releases were the right cadence, or whether icon work should have stayed unreleased until docs and verify scripts moved with it. That last bit is opinion. What is not opinion is that the product design (folder is the backup, one native window, optional local agent) is coherent in the source, while the packaging narrative after 1.0.0 is currently split across truthful STATUS files and older README/guide text that still sounds like the first ship.

_Follow-up, same day: the doc drift and verify-build.sh findings above were confirmed and fixed. README.md/DEV_GUIDE.md/USER_GUIDE.md version strings corrected (root and releases/1.0.3/ mirrored), and verify-build.sh's two hardcoded "1.0.0" checks (the PE-string grep and the `const Version = "1.0.0"` source gate) were switched to read `$VERSION` instead, then verified by actually running the script end to end: `VERIFY_BUILD_OK version=1.0.3`. See root STATUS.md / release notes for the fix itself; this section stays as the original unedited reflection._

---

## Cold read from the agent that built this — v2.0.0

_Written 2026-08-03, from a genuinely fresh Grok Build session (headless `grok -p`, real `git clone` of the published repo into an empty scratch directory, no memory of any prior session) reading only apps/dossier/README.md, docs/DEV_GUIDE.md, docs/USER_GUIDE.md, PROJECT_SUMMARY.md, and every STATUS.md across releases/1.0.3/ and releases/2.0.0/. Reproduced verbatim below; this is the actual output, not a paraphrase._

I opened the docs expecting a small utility. What I got is a finished multi-surface Windows workspace: documents, notes, stickies, paint, annotate, kanban, decisions, Agenda, FTS search, collections, and an optional local-agent "Ask" panel, all bound to one portable folder with real files on disk plus SQLite. README, USER_GUIDE, DEV_GUIDE, PROJECT_SUMMARY, and both release STATUS files agree that shipping truth is **2.0.0**, with PE fixed version **2.0.0.23**, dual binaries (`Dossier.exe` and `Dossier-console.exe`), and a packaging scope that is docs + release folder + verify rather than new product features. The product idea that holds the rest together is stated clearly and consistently: the folder is the backup; app prefs live under `%APPDATA%\Dossier`; one process, one window.

What stands out as careful engineering, not just feature volume, is how often packaging and host quirks were treated as first-class work. Upstream go-webview2 gated context menus and DevTools together; the local patch splits them. Escape on sticky pickers failed three times until the docs name a concrete root cause: Windows `BOOL` mis-mapped as Go `bool`, and Escape arriving as KEY_UP on this host. That is unusual to document this precisely in a DEV_GUIDE, and it reads like hard-won truth rather than theory. The agent path is deliberately thin (no embedded model; host+port gate; FTS stopwords, OR, bm25, top 4). Icon history is almost comic in its thoroughness: JPEG-as-PNG alpha breakage, mask fix, regenerate, paper art-direction only. The decision against per-dossier taskbar badges is justified by process model, not "later." Verification ethos is honest: unit tests plus real Windows hand-check, because WebView2 automation is described as flaky.

What I would flag or question, still only from the text:

**Verifiable in these files.** Version agreement tables are coherent across README, DEV_GUIDE, and `releases/2.0.0/STATUS.md` (`VERSION` / version.go / winres strings = `2.0.0`, PE = `2.0.0.23`). The 2.0.0 STATUS claims verify-build PASS and docs-only ship (plus surgical UI version strings). Known limitations are explicit: best-effort PDF/Office/ODF extract; Notes Revert is one open-session snapshot; Collections live in app config and do not travel with a copied folder; `ui/index.html` is a single large file where a bulk replace once truncated the UI; WebView2 automation is unreliable. PROJECT_SUMMARY reserves the "Cold read … v2.0.0" section empty on purpose, and still includes a full cold-read from a 1.0.3-era session plus a long monorepo-upload incident (source briefly ahead of docs, winres cosmetic vs RT_VERSION, stale syso). That incident is presented as closed before the 2.0.0 upload.

**Inferred or unsure (not claiming as fact).** I did not run the binary, tests, or verify scripts, so I cannot confirm PASS results, PE fields, or binary size myself; I only see them asserted. I did not read source outside these docs, so architecture claims (module roles, schema v3, patch details) are trusted as written, not re-proved. The monorepo drift story sounds complete in PROJECT_SUMMARY, but I cannot independently verify the timeline or that the checked-in 1.0.3 exe stayed byte-stable. Whether three icon-only 1.x patch ships were the right cadence is opinion. Whether a ~5k-line single HTML UI is a long-term liability is a judgment call the docs themselves already warn about.

If I were taking ownership tomorrow from these pages alone, I would treat 2.0.0 STATUS + VERSION agreement + the DEV_GUIDE limitations as the contract, re-run verify-build and a real Windows hand-check before trusting ship claims, and be careful not to "simplify" the Escape path or casually rewrite `ui/index.html`. The design (portable folder, optional local agent, honest extract limits) looks coherent. The unusual strength is the paper trail of what broke and why; the open question for me is how much of that rigor still has to live as human discipline, given how much of the product sits in one embedded UI file and how much verification still depends on a person on real Windows.

_Usage note: weekly SuperGrok pool measured 0% before this run and 0% after, no measurable cost, consistent with every prior read-only reflection task logged in `grok-usage-log.md`._

---

## Warm read from the agent that built this — v2.0.0

_Written 2026-08-04, from a genuinely fresh Grok Build session (headless `grok -p`, real `git clone` of the published repo into an empty scratch directory, no memory of any prior session) reading only apps/dossier/README.md, docs/DEV_GUIDE.md, docs/USER_GUIDE.md, PROJECT_SUMMARY.md, and every STATUS.md across releases/1.0.3/ and releases/2.0.0/. Same wiped-memory discipline as the cold read above, but forward-looking: proposes a wishlist rather than catching bugs. Reproduced verbatim below; this is a proposal only, not authorization to build anything._

**Method:** Fresh `git clone` of `https://github.com/webshoppe/tinkerVault.git` into `/home/jp/warm-read-scratch/tinkerVault-fresh`. Worked only under `apps/dossier/`. Read: `README.md`, `docs/DEV_GUIDE.md`, `docs/USER_GUIDE.md`, `PROJECT_SUMMARY.md`, `releases/1.0.3/STATUS.md`, `releases/2.0.0/STATUS.md`. Did not run binaries, tests, or verify scripts. Did not open source outside those docs. Proposal only; no builds or edits.

### Verified (from the files above)

| Topic | What the docs state |
|--------|---------------------|
| Ship truth | **2.0.0** product string; PE fixed **2.0.0.23**; dual exes (`Dossier.exe`, `Dossier-console.exe`); last v1 **1.0.3** |
| Shape | Single Windows GUI `.exe` + console twin; Go + WebView2; `CGO_ENABLED=0` cross-build from WSL; folder = workspace (real files + SQLite); app prefs in `%APPDATA%\Dossier` |
| Surfaces | Documents, Notes (Edit/Split/Preview), Stickies, Paint, Annotate, Kanban, Decisions, Agenda (list + mini-month, not a full grid), Search, optional Ask (HTTP client only), Settings, Collections, bookmarks |
| Explicit limits | Best-effort PDF/Office/ODF extract; Notes Revert = one open-session snapshot; Collections live in app config (do not travel with a copied folder); `ui/index.html` is one large file (bulk replace once truncated UI); WebView2 automation flaky; no embedded model; one process / one window (no per-dossier taskbar badge) |
| Packaging ethos | Ship = portable folder; verify + hand-check on real Windows; local go-webview2 patches (menus vs DevTools, Escape KEY_UP + BOOL ABI) |

**Not verified here:** PE fields, binary sizes, `VERIFY_BUILD_OK`, unit tests, runtime behavior, or any claim that requires opening Go/HTML outside these docs. Architecture details (module roles, schema v3, agent top-4 bm25, etc.) are taken as the guides state them, not re-proved from source.

### Inferred (wishlist only)

The table below is **proposal**, not a bug list and not a commitment that these gaps are defects. "Fits current shape" means:

- **As-is**: small prefs/UX/API work that matches existing surfaces and portable-folder model
- **Contained**: real work, still one exe + portable folder, no new runtime model
- **Structural**: new dependency class, different UI stack, build-pipeline redesign, multi-process/sync service, or anything that breaks the portable-folder / single-exe story

Effort: **small** / **medium** / **real new tier**.

| Item | Fits current shape | Effort | Notes |
|------|--------------------|--------|-------|
| Persist document sort and other session-only prefs across launches | As-is | Small | USER_GUIDE says sort is remembered for the session only; Settings/`config.json` already hold app prefs. |
| Task-list checkboxes in Notes preview rewrite the markdown source | Contained | Small | Docs say checkboxes are visual only; fits Notes files + auto-save pattern. |
| Notes version history (multiple named snapshots, not one open-session Revert) | Contained | Medium | Explicit current limit is one-level Revert; store + UI, still real files under `notes/`. |
| Portable Collections (membership travels with a folder or export/import pack) | Contained | Medium | Collections are app-config only today; portable model is the product idea, so this closes a coherence gap without leaving the folder story. |
| Cross-dossier search within a Collection (or open recents) | Contained | Medium | FTS is per open store today; multi-folder query stays offline and local if carefully scoped. |
| Agenda as optional full month grid + list (recurring due dates optional follow-on) | Contained | Medium | Agenda is deliberately a list of dated items; grid/recurrence are product depth on existing due fields. |
| Search filters (surface type, date range, overdue-only) and saved queries | Contained | Small to medium | Builds on FTS surfaces already indexed; UI + query params. |
| Decision links to stickies/kanban (not only frozen document versions) | Contained | Medium | Decisions already pin chronology; linking work items would stitch spine to boards without a new product genre. |
| Split `ui/index.html` into modules with a ship-time concat/embed (surgical-edit safety) | Structural (build pipeline) | Medium to real new tier | Docs warn bulk replace once blanked the UI; keeps single-exe embed if the build still produces one HTML blob. |
| Stable automated smoke beyond flaky SendKeys (e.g. host-scripted JS API + checklist, not pixel coordinates) | Contained | Medium | Verification ethos already prefers unit tests + hand-check; better gates without changing the product binary model. |
| Optional "portable mode": config beside the exe or inside a chosen root so Settings can travel on a USB stick | Contained | Medium | Today app-wide vs dossier split is fixed; portable-mode would extend the same "folder is backup" idea without cloud. |
| Prefer Open-with remembered as friendly app names / AppX aliases where possible | Contained | Medium | Preferred apps are a path map; richer OS association is Windows-dialog work, still local. |
| Better extract honesty UX (per-file "no text / image-only / encrypted" badges + one-click open externally) | As-is | Small | Extract limits are already documented; surfacing them more consistently is UX polish. |
| OCR path for scanned PDFs (optional external tool or bundled engine) | Structural | Real new tier | New heavy dependency or shell-out; may break "small portable zip" unless strictly optional and external. |
| Rich Office/ODF in-app preview (layout fidelity) | Structural | Real new tier | Far beyond best-effort text extract; new engines, size, and license surface. |
| Ship or bundle a local agent (or one-click install of a known local runtime) | Structural | Real new tier | Ask is intentionally HTTP-only with no embedded model; bundling a model changes size, support, and trust model. |
| Multi-window or multi-process open dossiers (and/or per-dossier taskbar identity) | Structural | Real new tier | Explicitly rejected for v1/v2 (one process, one window, icon set once); would reverse a core host constraint. |
| Non-Windows ports (Linux/mac) | Structural | Real new tier | WebView2 + Win32 dialogs are the documented host; other OS needs a different webview/dialog stack. |
| Cloud sync / multi-device merge for a dossier | Structural | Real new tier | Conflicts with "backup = copy the folder" unless sync is a pure external layer; CRDT/merge is a new product tier. |
| Plugin or extension API for custom surfaces | Structural | Real new tier | New security boundary and packaging story beyond the fixed surface set. |
| Dark/light (or high-contrast) theme fully controllable in Settings | Contained | Small to medium | A11y baseline and dark scrollbars are documented; a full theme toggle is contained UI + prefs if contrast is kept honest. |
| Attachments on stickies, cards, or decisions (files living under the dossier folder) | Contained | Medium | Matches "real files on disk"; needs store + open-externally path, not a new architecture. |
| Workspace export/import as a single zip (dossier folder pack/unpack) | Contained | Small to medium | Pure convenience over "copy the folder"; still portable-folder native. |
| Soft multi-user / network-share caution (read-only open, lock file, or "DB busy" recovery) | Contained | Medium | Portable SQLite folders get copied and shared; no claim in docs that concurrent multi-user is supported today. |

### Rough priority sense (opinion, not verified)

If ownership continued with **no constraints**, the highest leverage near the current design would be: (1) Notes history + portable Collections (finish the "folder is the backup" story), (2) safer UI modularization without abandoning the single-exe embed, (3) Agenda/search depth that uses dates and FTS already present. The largest "new product" bets (OCR, rich Office preview, embedded agent, multi-window, cross-OS, cloud sync) are real new tiers and would change what "portable Dossier" means.

_Usage note: weekly SuperGrok pool measured 0% before this run and 0% after, no measurable cost, same as the cold read above._

---

## Landing page & root docs: Dossier added to the archived+current history pattern for the first time (2026-08-03, v2.0.0 era)

Through 1.0.3, `index.html` and the root `README.md` listed Dossier as a single unpaired "Latest" card/row, no archived-version counterpart, since 1.0.3 was Dossier's first public release with nothing earlier to pair it with. This pass (staged alongside the 2.0.0 upload) extends Dossier to the same two-card, archived+current pattern the other three apps already use: the existing 1.0.3 card lost its "Latest" badge and picked up a "preserved as a self-contained snapshot" line (Source link repointed at `releases/1.0.3/`), and a new 2.0.0 card was added with the "Latest" badge and a Download link to the `dossier-v2.0.0` tag. Root README.md's Apps table row, its archived-versions bullet list, and its Structure tree all got the matching update in the same pass; root CHANGELOG.md picked up a new 2026-08-03 entry. `apps/README.md` was left untouched by design, JP's call: it's a short one-liner-per-app overview and doesn't need per-version detail.

One false alarm along the way, worth recording so it isn't relitigated: a first pass at reading the live `index.html`/`README.md` via `raw.githubusercontent.com` (no cache-bust query param) showed no Dossier entry at all, which looked like it had never been added to either file, not even for 1.0.3. Re-fetching with a `?v=` cache-busting param showed the real, already-correct 1.0.3 state; this was the project's known CDN staleness gotcha, not an actual gap.

---

## Monorepo-upload incident: root source briefly drifted ahead of docs (found + fixed 2026-08-02/03, v1.0.3 era)

While staging Dossier's first upload into `tinkerVault`, the packaged doc set (`README.md`, `CHANGELOG.md`, `docs/`) always described 1.0.3 correctly. But somewhere during the manual, 100+-file GitHub upload, the source batch (`main.go`, `internal/app/`, `internal/dossier/`, `internal/dialog/`) got pulled from the live WSL2 working copy after v2 Tier 1 work had already started there, instead of from the pre-v2-tagged staging copy. Confirmed directly: `VERSION` read `2.0.0-t1`, `internal/app/version.go` matched, `winres/winres.json` said `2.0.0`, `internal/dossier/store.go`'s `SchemaVersion` was `"3"` (v1.0.3 shipped schema `"2"`), and `main.go` already had the v2 `AutoOpenLastFromSettings` logic, while every doc at the same root still described 1.0.3-only features. `releases/1.0.3/Dossier.exe` itself was unaffected throughout, verified by checksum against the original 1.0.3 build; this was a checked-in-source/docs mismatch, not a broken release.

A true byte-exact 1.0.3-only source snapshot turned out not to be recoverable (no git history at any point in this project, and the local staging copy used to prep the GitHub upload had itself been re-synced from the same already-drifted live repo during a later cleanup pass). Given the v2 additions were additive and confirmed non-breaking, this side quest of finding a "better" fix landed on: restore `VERSION`/`version.go`/`winres.json` to `1.0.3` so nothing actively lied, leave the (harmless, opt-in, migration-safe) v2 code in place rather than trying to hand-strip it back out, and document the real state here plus in DEV_GUIDE.md's Known Limitations rather than pretend it didn't happen.

`winres.json` carries version data in two separate places: the `RT_MANIFEST` block's `identity.version` (cosmetic) and the `RT_VERSION` block's `fixed.file_version`/`fixed.product_version`/`info.0409.FileVersion`/`info.0409.ProductVersion` (the data that actually gets compiled into the .exe's real PE version resource). The first fix only touched the cosmetic field; the `RT_VERSION` block was left at `2.0.0.1`/`2.0.0-t1` until a follow-up. **Confirmed fixed on a real Windows box, 2026-08-03:** both fields corrected to `1.0.3.0`/`1.0.3`, a stale `rsrc_windows_amd64.syso` (verify-build.sh only regenerates it if one doesn't already exist) deleted to force a clean rebuild, then a fresh clone built and run for real: `VERIFY_BUILD_OK version=1.0.3` and `VERIFY_WINDOWS_OK version=1.0.3` (real GUI launch, real PE `FileVersion=1.0.3 ProductVersion=1.0.3`). Both incidents are closed; this history is kept here as the lesson for future version bumps (both `winres.json` blocks, every time), which is why this v2.0.0 pass double-checked both blocks up front rather than assuming the cosmetic fix was sufficient.

---

## Related status files (historical)

| File | Scope |
|------|--------|
| `STATUS.md` (repo root) | Latest verification (overwritten each ship pass); internal working file, not part of this public copy |
| [`releases/1.0.3/STATUS.md`](./releases/1.0.3/STATUS.md) | Last v1 packaging snapshot |
| [`releases/2.0.0/STATUS.md`](./releases/2.0.0/STATUS.md) | This 2.0.0 packaging snapshot |
| `STATUS-GAPFIX*.md` | Historical gap-fix notes; internal working files, not part of this public copy |

The tier-by-tier development history (all tiers, packaging passes, and gap-fixes across v1 and v2) that would otherwise live across a root `STATUS.md` and several `STATUS-GAPFIX-*.md` files, plus three superseded `1.0.0`-`1.0.2` release snapshots, is consolidated into the Tier history above instead.

For day-to-day development, prefer **DEV_GUIDE.md** + this summary over re-reading every gap-fix log.
