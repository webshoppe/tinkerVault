# Dossier 1.0.3; Icon paper art-direction; Status

**Date:** 2026-07-26  
**Version:** **1.0.3** (this file is for 1.0.3 only)  
**Binary:** `releases/1.0.3/Dossier.exe` (also `build/Dossier.exe`)  
**Scope:** Visual-only rework of the **document + sticky-note** treatment on the app icon. Folder body, alpha pipeline, and PE embedding wiring unchanged in intent.

---

## What changed

| Element | Before (1.0.2) | After (1.0.3) |
|---------|----------------|---------------|
| Sticky edge | Jagged/zigzag “torn” bottom | Clean rounded rectangle |
| Corners | Sharp rectangular paper stack | Soft rounded corners on both papers |
| Placement | Square-on stack into a boxy notch | Diagonal / offset, papers angled out from the folder |
| Tuck | Hard rectangular cutout feel | Softer “emerging from folder” tuck |
| Composition | White sheet + yellow sticky | **Kept** two layers, restyled |
| Folder body | Purple-blue large fill | **Unchanged** (color/shape/size intent) |

**Style reference only:** `assets/dossier-icon.source.jpg` (rounded sticky, angled papers). **Not** used as a pixel source or mask (flattened JPEG / prior alpha issues).

**Pipeline:** Image-edit on the 1.0.2 icon + style reference → checkerboard-key to real RGBA → multi-size `winres/` → go-winres → rebuild.

---

## What was not changed

- Folder body design language (large purple-blue rounded folder)  
- go-winres / `rsrc_windows_amd64.syso` embedding method  
- Context menus, Open-with prefs, Notes revert, FTS, etc.

---

## Verification (same bar as 1.0.1 / 1.0.2)

| Check | Result |
|-------|--------|
| PNG IHDR color type | **6 = RGBA** |
| `PixelFormat` | **Format32bppArgb** |
| Corner alpha | **A=0** |
| Opaque subject (256) | ~57% fill (~37k opaque px) |
| PE extract 32×32 | corner **A=0**, mid **A=255**, purple folder RGB, 588 opaque px |
| PE FileVersion / ProductVersion | **1.0.3** |
| In-app footer default | **1.0.3** |

Hand-check: open `releases/1.0.3/dossier-icon.png` for rounded/angled papers; taskbar/Explorer on light and dark for halo and small-size readability.

---

## Hard blockers

**None.**

---

## Package contents (`releases/1.0.3/`)

```
Dossier.exe
dossier-icon.png      # RGBA, papers restyled
VERSION               # 1.0.3
STATUS.md             # this file (1.0.3-specific)
README.md, DEV_GUIDE.md, USER_GUIDE.md, PROJECT_SUMMARY.md
sample-files/
```
