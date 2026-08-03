# Test fixtures

Manual regression packs used during Dossier's own development to hand-verify specific import/extraction behavior. These are **developer/QA resources, not end-user onboarding content**; end-user sample files live separately in [`sample-files/`](../../sample-files/) at the app root.

Each fixture set has its own `MANIFEST.md` describing what each file targets, the expected result, a unique FTS5 search keyword per file (so a Search-tab hit is unambiguous about which file matched), and a suggested test order.

| Folder | Built during | Covers |
|--------|--------------|--------|
| [`tier1-test-fixtures/`](./tier1-test-fixtures/MANIFEST.md) | v2 Tier 1 | Markdown/plain-text/PDF import, image-only PDF flagging, non-first-class attachment handling |
| [`tier4-test-fixtures/`](./tier4-test-fixtures/MANIFEST.md) | v2 Tier 4 (Office "open in…" hooks) | Clean vs. lossy `.docx`/`.xlsx` extraction, partial-extraction badge/banner, table/image/formula edge cases, "open externally" round-trip |

Useful for re-running the same manual pass after a change touching import/extraction/search, or as a starting point for a future automated regression suite. Not wired into `verify-build.sh` or `verify-windows.ps1` today; those check packaging (version, PE resources, build success), not import/extraction behavior.
