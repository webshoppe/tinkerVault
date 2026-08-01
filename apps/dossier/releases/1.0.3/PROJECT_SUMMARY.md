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

<!-- Reserved for a future reflective note by the implementing agent / human maintainer.
     When filling this in, cover: what surprised you, what you'd do differently,
     fragile areas (WebView2 input automation, go-webview2 Debug coupling), and
     advice for the next person shipping a patch. -->

_(Not filled in this packaging pass — left open on purpose.)_

---

## Related status files

| File | Scope |
|------|--------|
| [`STATUS.md`](./STATUS.md) (this folder) | Verification notes for this shipped version |

The tier-by-tier development history (Tiers 1-5, Packaging, and gap-fix passes 1.1-1.3) that would otherwise live across a root `STATUS.md` and several `STATUS-GAPFIX-*.md` files, plus three superseded `1.0.0`-`1.0.2` release snapshots, is consolidated into the Tier history above instead; those are internal working files and aren't part of this public copy.

For day-to-day development, prefer **DEV_GUIDE.md** + this summary over re-reading every gap-fix log.
