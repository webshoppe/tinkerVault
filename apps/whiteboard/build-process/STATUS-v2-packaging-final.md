# Whiteboard App, Status (v2.4.0 packaging)

**Date:** 2026-07-12  
**Version:** 2.4.0  
**Working artifact:** `whiteboard.html`  
**Release package:** `releases/v2.4.0/`  
**Verification:** `verify-v2-package.mjs` → `verify-report-v2-package.json`

This pass is **packaging only** (no new product features). Mirrors the role of v1’s Tier 3 packaging after the full five-tier v2 feature arc.

---

## Packaging checklist

| Item | Status |
|------|--------|
| `releases/v2.4.0/whiteboard.html` = working copy (byte-identical) | **Done** |
| `releases/v2.4.0/VERSION` = 2.4.0 | **Done** |
| `releases/v2.4.0/icon.svg` refreshed from root | **Done** |
| `releases/v2.4.0/README.md` describes v2 feature set | **Done** |
| `releases/v2.4.0/docs/USER_GUIDE.md` full v2 guide | **Done** |
| Root `README.md` / `docs/USER_GUIDE.md` updated | **Done** |
| `RELEASE_NOTES.md` covers full v2 arc + two bugs | **Done** |
| In-app badge + About (**?**) reflect v2.4.0 & real features | **Done** |
| Cross-engine verify working + release | **Done** (Chromium + Firefox; WebKit skip) |

---

## Cross-engine results

| Engine | Working | Release | Notes |
|--------|---------|---------|--------|
| **Chromium** | 26/26 PASS | 26/26 PASS | Paint stroke, Kanban pointer-drag, About, trash, search/palette |
| **Firefox** | 26/26 PASS | 26/26 PASS | Same suite; Kanban grip drag **PASS** (regression class from Tier 4) |
| **WebKit** | SKIP | SKIP | Host missing shared libraries for Playwright MiniBrowser; no sudo on this WSL host |

**HTML identity:** working and release SHA-256 match (`verify-report-v2-package.json` → `identity`).

**Total:** 104 assertions pass, 0 fail, 2 engine skips (WebKit × 2 targets).

### Suite intentionally stresses engines
- Real **pointer-drag** on Kanban (not `moveCard()` / state rewrite)  
- Board menu → Time machine open  
- Soft-delete → trash  
- About content checks for Kanban, search, palette, paste, time machine, trash  
- No external network requests  

---

## Hard blockers

### WebKit automated run on this host (**blocked by environment**)
Playwright WebKit is installed under `~/.cache/ms-playwright/webkit-2311`, but launch fails without system packages (e.g. gstreamer/gtk/jxl/backtrace-related libs). This environment cannot `sudo apt-get install`.  

**Not a product code defect identified in Chromium/Firefox.**  
**Mitigation:** Flag clearly; re-run `node verify-v2-package.mjs` on macOS or a full Linux desktop with `npx playwright install-deps` for WebKit green. Manual Safari open of `releases/v2.4.0/whiteboard.html` recommended before wide iOS/macOS distribution.

No other hard blockers.

---

## What was not done (out of packaging scope)

- Sticky/Kanban “duplicate card” polish  
- Additional File menu items  
- Any new product features  

---

## How to ship

```
releases/v2.4.0/
  whiteboard.html
  icon.svg
  VERSION          # 2.4.0
  README.md
  docs/USER_GUIDE.md
```

Zip or copy that folder. End users only need `whiteboard.html`.

---

## Verify commands

```bash
# Packaging multi-engine
LD_LIBRARY_PATH=$HOME/.local/chrome-libs \
PLAYWRIGHT_SKIP_VALIDATE_HOST_REQUIREMENTS=1 \
  node verify-v2-package.mjs

# Optional: individual feature tiers still available
npm run verify:v2-tier5
```
