# Dossier 2.0.1 — Notes/Decisions two-pane width gap-fix

**Version:** **2.0.1**  
**PE fixed quad:** **2.0.1.24**  
**Binaries:** `build/Dossier.exe` + `build/Dossier-console.exe` (size **12342784** each)  
**Date:** 2026-08-07  
**Scope:** Surgical CSS only (Notes + Decisions list/detail split). No feature work, no bulk rewrite of `ui/index.html`.

---

## Diagnosis

### What was broken

At a maximized / very wide window (~1920 CSS px), **Notes** and **Decisions** showed the list/spine column taking most of the main area, with the detail pane (note editor / decision form) squeezed into a narrow strip on the right. Empty elevated-panel space read as a large dead gap between list content and the detail strip. No horizontal scrollbar (content stayed inside the window). **Documents** (fixed `flex: 0 0 220px` sources rail) and canvas surfaces did not show this.

### Rules governing the split (before fix)

| Surface | List / spine | Detail |
|---------|----------------|--------|
| Notes | `#notes-list`: `width: min(260px, 34vw); min-width: 140px; max-width: 320px; flex: 0 1 260px` | `#note-editor`: `flex: 1; min-width: 0` |
| Decisions | `#decision-spine`: `width: min(300px, 36vw); min-width: 160px; max-width: 360px; flex: 0 1 300px` | `#decision-detail`: `flex: 1; min-width: 0` |
| Stack breakpoint | `@media (max-width: 1320px)` → column stack, list/spine `width: 100%`, `max-width: none` | same media query |

Both views put list + detail as **direct flex children** of `.view.active` with `#notes-view` / `#decisions-view { flex-direction: row !important }`. Documents avoids that by nesting a `.docs-layout` row under a column view.

### Root cause (why the ratio failed)

1. **List/spine used a soft flex basis, not a hard pin.**  
   `flex: 0 1 260px` (grow 0, **shrink 1**, basis 260) plus viewport `min()` widths was introduced in the earlier gap-fix that raised the stack threshold to 1320px to kill the default-width horizontal scrollbar. That pass treated overflow by stacking earlier; it did not lock the side-by-side width ratio the way Documents does with `flex: 0 0 220px`.

2. **Flex min-content blowout reproduces the exact symptom.**  
   Headless Chrome measurement at 1920 with long unbroken list titles and list forced to `min-width: auto` / no max-width:

   - list **~93.7%** of main, editor **~6.3%** (narrow strip on the right), gap between boxes **0**  
   - The "dead gap" is empty elevated background **inside** the oversized list column, not a third flex item.

   Pre-fix CSS tried to cap with `max-width: 320px`, but that is still a soft constraint stacked on shrinkable basis + `width: min(..., vw)` + `.view { overflow: auto }` on the row flex container. Documents never relies on that combination.

3. **Secondary: intro banners as a third column.**  
   `showIntroIfNeeded` inserts `.intro-banner` as the first child of the view. In **row** mode (viewport &gt; 1320), the intro became a horizontal flex item. Measured at 1920 with full app CSS: intro **~54%**, list **~16%**, editor **~28%**. That also squeezes detail whenever intros are still visible. Documents is unaffected because its view stays column and intro sits above `.docs-layout`.

4. **Why only Notes and Decisions:** only those two use this list+detail **row on the view root**. Documents uses a nested row; Stickies/Kanban/Paint/Annotate are canvas or list-only patterns.

### Evidence

- Source comparison: original pre-gap-fix CSS (`build/blank-repro-fixed.html`) used `width: 260px; flex-shrink: 0` / `width: 320px; flex-shrink: 0` (hard pin). Gap-fix replaced that with `min()` + `flex: 0 1 …px` + stack media query.
- Controlled layout harness (real extracted `<style>` from `ui/index.html` + real DOM shape) in headless Chrome at 1920 / 1600 / 1280 / 1100.
- Min-content stress case above matches hand-reported "list almost full width, detail strip on the right."

---

## Fix (surgical CSS only)

### Notes (`#notes-view` / `#notes-list` / `#note-editor`)

- List: **`width: 260px; flex: 0 0 260px; min-width: 0`** (no grow, no shrink; long titles ellipsis + list `overflow-y: auto`, `overflow-x: hidden`).
- Editor + placeholder: **`flex: 1 1 0%; min-width: 0`** so remaining width always goes to detail.
- View: `overflow: hidden` in side-by-side (children scroll); `flex-wrap: wrap` so intro can take a full first row.
- Intro: `#notes-view > .intro-banner { flex: 0 0 100% }` (own row, does not grow).
- **Stack `@media (max-width: 1320px)` kept:** column, list `width: 100%`, `max-height: 34%`, intro reset to `flex: 0 0 auto`, collapse class unchanged.

### Decisions (mirror)

- Spine: **`width: 280px; flex: 0 0 280px; min-width: 0`**.
- Detail + placeholder: **`flex: 1 1 0%; min-width: 0`**.
- Same intro wrap + stack media query + `☰ Spine` collapse.

No JS changes. No bulk find-replace across `ui/index.html`.

---

## Version agreement

| Field | Value |
|-------|--------|
| `VERSION` | `2.0.1` |
| `internal/app/version.go` | `2.0.1` |
| UI footer / brand / launcher (`ui/index.html`) | `2.0.1` |
| `winres/winres.json` FileVersion / ProductVersion strings | `2.0.1` |
| PE fixed file/product version | `2.0.1.24` |
| OriginalFilename | `Dossier.exe` / `Dossier-console.exe` (per binary) |

---

## Verification performed (agent)

| Check | Result |
|-------|--------|
| `ui/index.html` structure after edit | **OK** — ends with `</html>`; all 11 views present; List/Spine toggles present; stack media query present; not truncated |
| Unit tests `go test ./internal/...` | **PASS** |
| Dual rebuild | `build/Dossier.exe` + `build/Dossier-console.exe`, both **12342784** bytes |
| PE versions (PowerShell `FileVersionInfo`) | Both **FileVersion=2.0.1**, **ProductVersion=2.0.1**, raw **2.0.1.24**; OriginalFilename correct per binary |
| Embedded UI | Both exes contain `2.0.1` and `flex: 0 0 260px`; no leftover product `2.0.0` string in binary search |

### Layout harness (headless Chrome, full extracted CSS)

| Viewport | Notes | Decisions |
|----------|-------|-----------|
| **1920** (wide / maximized stand-in) | row; list **260px (~16%)**, editor **~84%**, gap **0**; long titles stay clipped inside list; collapse → editor **100%** | row; spine **280px (~17%)**, detail **~83%** |
| **1600** | same pin; editor fills remainder | same |
| **1280** (default-ish, below 1320 stack) | **column** stack; list full width, max-height band; editor below full width; no horizontal squeeze | column stack; spine full width band; detail below |
| **1100** (narrower stacked) | still column; List collapse hides list | still column |

Intro full-width row at 1920 (not a side column); intro content-height at 1280 (not eating the column).

### Hand-verification (JP, real Windows, 2026-08-07)

Confirmed on real hardware after the agent's automated checks: Explorer Properties reports File version **2.0.1.24** / Product version **2.0.1** / Copyright webShoppe on both `Dossier.exe` and `Dossier-console.exe`; in-app Notes and Decisions at maximized (~1920px) show the fixed narrow rail with the detail/editor pane filling the rest, no dead gap; default and narrower window sizes still stack correctly with working `☰ List` / `☰ Spine` toggles.

---

## Hard blockers

**None.** Both the agent's automated verification and JP's real-hardware hand-check are complete and passing.

---

## Files touched

- `ui/index.html` — Notes + Decisions two-pane CSS only (+ version display strings)
- `VERSION`
- `internal/app/version.go`
- `winres/winres.json`
- `build/Dossier.exe`, `build/Dossier-console.exe`
- `STATUS.md` (this file)

Temporary harness files under `build/*repro*.html` were used for measurement only; not part of the product surface.
