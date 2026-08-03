# Tier 1 Test Fixture Pack

Generated 2026-07-23 for manual smoke testing of Dossier Tier 1, ahead of Tier 2. Each file targets a specific untested (or under-tested) piece of the Tier 1 scope. Verified programmatically (via `pdfplumber`) before handoff, not just assumed to work.

| File | Tests | Expected result |
|---|---|---|
| `01-real-text.pdf` | PDF text extraction | Extracts cleanly (462 chars confirmed). Search keyword: `FALCONBRIDGE9142` |
| `02-image-only.pdf` | Image-only page flagging | Zero extractable text (confirmed via pdfplumber). Should show the "image-only, not extracted" flag in the UI, not silently appear empty or attempt OCR |
| `03-notes-source.md` | Markdown import (untested until now, only .txt/.pdf had been tried) | Imports as markdown, renders correctly. Search keyword: `MARIGOLD7731` |
| `04-plain-text.txt` | Plain text import, alongside md/pdf for a clean three-way comparison | Imports as text. Search keyword: `ZEPHYRQUILL204`. Ships in its clean, unedited state; make a scratch copy before doing the Rescan test in step 5 below, so this fixture stays reusable |
| `05-generic-attach.png` | Non-first-class file type (anything besides md/txt/pdf) | Should attach with filename, size, and an "open externally" action, per the Tier 1 brief, not be silently dropped or misclassified |

Renamed to a numbered convention (`01`-`05`) for this v2.0.0 upload, matching the app's `sample-files/` naming; content is unchanged from the original Tier 1 pass. Two undocumented stragglers that had accumulated alongside these five during actual test runs (an already-edited copy of the plain-text fixture, and an unrelated highlight-opacity PNG from a different test) were left out of this pack.

## Suggested test order

1. Import all 5 via "+ Import files..." (or test drag-and-drop separately, known gap, see project memory)
2. Confirm the PNG shows "open externally," not an in-app preview
3. Open `02-image-only.pdf` in the Documents list, confirm the image-only flag actually surfaces in the UI
4. Search `FALCONBRIDGE9142`, `MARIGOLD7731`, and `ZEPHYRQUILL204` individually in the Search tab, each should return exactly one hit, its matching source file
5. Copy `04-plain-text.txt` somewhere outside the dossier, edit the copy, re-import it externally, then hit "Rescan" to confirm incremental reindex picks up the change without a full re-import
