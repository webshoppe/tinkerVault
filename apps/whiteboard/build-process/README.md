# Build / verification process artifacts

This folder collects **scripts and reports actually used** while building Whiteboard across Tiers 1-3. Nothing here is required to **run** the app; only `whiteboard.html` (or `release/whiteboard.html`) is.

There was **no** `playwright.config.js` / `playwright.config.ts`. Checks are plain Node ESM scripts that import `playwright-core` and launch browsers with options in-code.

---

## Inventory

| Path | What it is | Provenance |
|------|------------|------------|
| `verify.mjs` | Automated UI suite | Evolved from Tier 1 Chrome checks; **final on-disk form is the Tier 2 suite** (multi-board, four templates, undo, import, etc.). Header/comments may still say “Tier 2”. |
| `verify-tier3.mjs` | Multi-engine smoke suite | Tier 3: Chromium + Firefox + WebKit, working copy + `release/` copy, version/favicon probes |
| `verify-report.json` | JSON results | **`tier: 2`**, 41 results, 0 failed (2026-07-11). This is **not** a preserved Tier 1 report; see below. |
| `verify-report-tier3.json` | JSON results | Tier 3 multi-engine matrix; all six runs PASS; `workingMatchesRelease: true` |
| `package.json` / `package-lock.json` | Optional Node deps | `playwright-core` only; `npm run verify` historically pointed at `verify.mjs` |
| `screenshots/01-paint.png` | PNG | From early verification (Tier 1 era naming) |
| `screenshots/02-notepad.png` | PNG | From early verification (Tier 1 era naming) |
| `screenshots/t2-01-paint.png` | PNG | Tier 2 suite |
| `screenshots/t2-02-dark.png` | PNG | Tier 2 suite (dark theme) |

Copies of the scripts also remain in the project root where they were originally run; this folder is the gathered archive.

---

## What was **not** saved on disk

| Missing artifact | Reality |
|------------------|---------|
| **Tier 1 JSON report** | Tier 1 was run and reported in session/STATUS as **39/39** on Chrome, but no `verify-report-tier1.json` (or equivalent) was kept. The filename `verify-report.json` was **reused and overwritten** by the Tier 2 run. **Do not treat `verify-report.json` as Tier 1 output.** |
| **Tier 1 machine-readable log** | Only narrative STATUS + screenshots `01-*.png` / `02-*.png` remain as on-disk evidence. |
| **Playwright config file** | Never created; not lost. |
| **Tier 3 screenshots** | Tier 3 suite did not write screenshot files (smoke/assertions only). |

No reconstructed Tier 1 report was invented after the fact.

---

## How verification was typically run

Environment used during the project: **Windows Node + Playwright browsers**, with HTML copied under a Windows path so `file://` loads correctly (WSL Linux Chromium lacked libraries).

Example (conceptual):

```text
# From a directory that has whiteboard.html + node_modules/playwright-core
node verify.mjs          # Tier 2 suite (Chrome via CHROME_PATH)
node verify-tier3.mjs    # Chromium / Firefox / WebKit × working + release
```

Chrome path default in scripts:  
`C:\Program Files\Google\Chrome\Application\chrome.exe`  
(overridable with `CHROME_PATH`).

---

## Relation to product packages

| Location | Role |
|----------|------|
| `../whiteboard.html` | Working source of the app, always the current version (2.4.0) |
| `../releases/v1.0.0/` | Clean 1.0.0 distribution tree (superseded `../release/`, kept as its own self-contained snapshot once v2.4.0 shipped) |
| `../releases/v2.4.0/` | Clean 2.4.0 distribution tree, same file as `../whiteboard.html` |
| `../build-process/` | How we verified; not shipped as a user requirement, covers both versions |

---

## v2 additions (2026-07-12), pulled and confirmed

All files below are now actually sitting in this folder, pulled from the live WSL2 build folder after the full five-tier arc plus packaging.

| Path | What it is | Result |
|------|------------|--------|
| `verify-v2-tier1.mjs` / `verify-report-v2-tier1.json` | Tier 1 suite (quota, corrupt-state, multi-tab, search, session restore) | 53/53 |
| `verify-v2-tier2.mjs` / `verify-report-v2-tier2.json` | Tier 2 suite (checklist, auto-color, tags, palette) | 57/57 |
| `verify-v2-tier3.mjs` / `verify-report-v2-tier3.json` | Tier 3 suite (paste/drop routing), **post-fix version**; the pre-fix 51/51 run that missed the routing bug was overwritten by the fix pass under the same filename, only the corrected 68/68 result survives on disk | 68/68 |
| `verify-v2-tier4.mjs` / `verify-report-v2-tier4.json` | Tier 4 suite (Kanban), **post-fix version**; same pattern, the pre-fix 31/31 run (the one that used a direct state call and missed the drag bug) was overwritten; only the corrected 39/39, real-pointer-event version survives | 39/39 |
| `verify-v2-tier5.mjs` / `verify-report-v2-tier5.json` | Tier 5 suite (time machine, trash), written DOM-driven from the start (menu → action → confirm), learned directly from Tiers 3 and 4 | 27/27 |
| `verify-v2-package.mjs` / `verify-report-v2-package.json` | Packaging cross-engine matrix, working copy + `release/` copy | Chromium 26/26, Firefox 26/26 (both copies); WebKit skipped, Playwright's MiniBrowser can't launch on this WSL host (missing system libs, no sudo to install them) |
| `STATUS-v2-packaging-final.md` | The only `STATUS.md` that survives on disk; same overwrite pattern v1's own inventory above already documented (only the final pass's version survives when every tier reuses the same filename). Covers the packaging pass specifically, not each individual tier | n/a |

**On the overwritten pre-fix reports:** this is the second time this exact failure mode has shown up in this project (v1's Tier 1 report was similarly overwritten by Tier 2's). The fuller human-readable account of what the *pre-fix* runs actually reported (51/51 and 31/31, both clean, both hiding a real bug) lives in this project's own `whiteboard-v2-run/README.md`, compiled from terminal output at the time rather than from a surviving file. Worth actually doing what 1.0.0's own retrospective already recommended and didn't happen again here either: version report filenames per attempt, not just per tier, if a fix pass might overwrite the interesting "here's what looked fine but wasn't" evidence.
