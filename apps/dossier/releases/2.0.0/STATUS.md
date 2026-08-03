# Dossier 2.0.0 — Packaging tier

**Version:** **2.0.0** (final product string, no suffix)  
**PE fixed quad:** **2.0.0.23**  
**Binaries:** `build/Dossier.exe` + `build/Dossier-console.exe` (size **12342272** each)  
**Ship:** `releases/2.0.0/` (same exes + docs + sample-files)  
**Date:** 2026-08-01  

This is the packaging ship for v2 after all feature tiers (1–5), accessibility sub-passes, polish, and Tier 6 close-out. Scope is **docs + release folder + verify**, not new product features.

---

## Items

### 1. README.md rework

Full pass for 2.0.0: full surface table (Notes split-pane, due dates, sticky↔kanban link, Agenda, Collections, bookmarks, ODF), run paths for `releases/2.0.0/`, dual exes, build-from-source with `make build` / OriginalFilename note, data layout including `bookmarks.json`, docs index.

### 2. DEV_GUIDE.md rework

Current architecture, module layout, **working** build walkthrough (Go + go-winres paths, `make build` dual-exe, version field table PE 2.0.0.23), go-webview2 Escape/BOOL patches, icon history (kept, not duplicated elsewhere), schema notes (v3, due_at, links, Agenda, collections/bookmarks), limitations including blank-window root cause, verify/debug.

### 3. USER_GUIDE.md rework

End-user coverage for Notes Edit/Split/Preview, sticky/kanban due dates and non-destructive link, Agenda (mini-month + list), Collections, import bookmarks, ODF, auto-resume, Move… arrows, Escape pickers, a11y highlights.

### 4. PROJECT_SUMMARY.md refresh

Header **2.0.0**. Tier history extended through v2 T1–T4, a11y Escape 3.1–3.6 with real root cause (BOOL ABI + KEY_UP gate), T6/T7 close-out, Packaging 2.0.0.  

**Cold-read section:** left **empty/reserved** for a separate fresh session (not filled here).

### 5. Dev build + release folder + verify script

| Deliverable | Location |
|-------------|----------|
| Dev binaries | `build/Dossier.exe`, `build/Dossier-console.exe` (same version as ship) |
| How to build from source | README + DEV_GUIDE walkthroughs |
| Ship folder | `releases/2.0.0/` (structure like 1.0.3, plus console exe) |
| `build-process/verify-build.sh` | Version from `VERSION` file only; builds both exes + smoke; greps `$VERSION` |

`make release` now copies **both** exes and STATUS when present.

---

## Version agreement

| Field | Value |
|-------|--------|
| `VERSION` / `version.go` / UI / winres strings | `2.0.0` |
| PE fixed | `2.0.0.23` |
| OriginalFilename | Dossier.exe / Dossier-console.exe (per binary) |

---

## Verification

| Check | Result |
|-------|--------|
| `bash build-process/verify-build.sh` | **PASS** → `VERIFY_BUILD_OK version=2.0.0` |
| Unit tests | PASS (via verify-build) |
| PE both exes | FileVersion **2.0.0**, PE **2.0.0.23**, OriginalFilename each correct |
| `releases/2.0.0/` | Dossier.exe + Dossier-console.exe + four docs + STATUS + sample-files |
| Docs only (no intentional UI feature changes) | Surgical version string update in `ui/index.html` only |

---

## Hard blockers

**None.**

---

## Files touched

- `README.md`, `DEV_GUIDE.md`, `USER_GUIDE.md`, `PROJECT_SUMMARY.md`
- `VERSION`, `internal/app/version.go`, `winres/winres.json`, `ui/index.html` (version strings only)
- `Makefile` (release includes console + STATUS)
- `build-process/verify-build.sh`, `build-process/verify-windows.ps1`
- `build/Dossier.exe`, `build/Dossier-console.exe`
- `releases/2.0.0/*`
- `STATUS.md` (this file)
