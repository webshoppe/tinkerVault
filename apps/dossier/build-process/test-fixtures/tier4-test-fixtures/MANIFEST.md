# Tier 4 Test Fixture Pack (docx/xlsx)

Generated 2026-07-25 for manual smoke testing of Dossier Tier 4 item 4 ("open in..." hooks for .docx/.xlsx), same pattern as `tier1-test-fixtures/`. Each file targets a specific piece of the item 4 scope: import, pure-Go text extraction for FTS, the partial-extraction badge/banner, and "open externally."

| File | Tests | Expected result |
|---|---|---|
| `01-docx-simple.docx` | Clean/full .docx extraction path | Plain paragraphs only, no table, no image. Should extract fully with no partial-extraction flag. Search keyword: `THISTLEVAPOR8823` |
| `02-docx-table-image.docx` | Lossy .docx extraction path (table + image) | Contains a body-text keyword, a keyword that lives only inside a table cell, and an image with a caption keyword. Should trigger the partial-extraction badge/banner. Body keyword: `COBALTLANTERN3391`. Table-only keyword: `TABLECELL5510` (search for this specifically to see whether table text reaches the index at all). Caption keyword: `IMAGECAPTION7749` |
| `03-xlsx-simple.xlsx` | Clean/full .xlsx extraction path | One sheet, plain text/number values, no formulas. Should extract cleanly, no partial-extraction flag expected. Search keyword: `EMBERFOXGLOVE6604` |
| `04-xlsx-formulas.xlsx` | Multi-sheet + formula handling | Two sheets ("Data", "Calc"). `Calc!B1` holds a SUM formula referencing `Data!B2:B4`. **Caveat**: this file was built with openpyxl, not real Excel, so the formula cell likely has no cached computed value stored in the file itself (openpyxl doesn't evaluate formulas). If the extractor shows blank or the literal formula text instead of `60` for that cell, treat it as a fixture limitation, not automatically an app bug, check what the app actually surfaces before concluding either way. Search keyword: `GRANITEWHISPER2258` |

Renamed to a numbered convention (`01`-`04`) for this v2.0.0 upload, matching the app's `sample-files/` naming; content is unchanged from the original Tier 4 pass.

## Suggested test order

1. Import all 4 via "+ Import files..." (or drag-and-drop), screenshot the Documents list afterward.
2. Check which files show a partial-extraction badge/banner and which don't, compare against the "expected result" column above.
3. Search each keyword individually in the Search tab: `THISTLEVAPOR8823`, `COBALTLANTERN3391`, `TABLECELL5510`, `IMAGECAPTION7749`, `EMBERFOXGLOVE6604`, `GRANITEWHISPER2258`. Note which ones hit and which don't, the table/image/formula keywords are the interesting ones, a miss there tells you exactly what's not making it into the index.
4. Click "Open externally" on each of the 4 files, confirm real Word/Excel actually opens them (not an in-app viewer, not an error).
5. If time allows, try re-exporting `04-xlsx-formulas.xlsx` from real Excel first (so it has a genuine cached value for the formula), re-import, and re-check, that isolates the fixture caveat above from real app behavior.
