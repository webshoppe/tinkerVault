# Developer Guide

This document is for a future contributor who wants to read, modify, or rebuild the Markdown Viewer from its source.

## What this folder contains

```
src/
  template.html       HTML skeleton + placeholder tokens; library code and app.js
                       are injected at build time via those placeholders.
  assemble.py          Build script. Reads template.html + the vendored libs below
                       and writes a single self-contained index.html one level up,
                       at the app root.
  app.js               The entire application script (human-editable source).
  vendor/
    base.css           The viewer's own styles (theme variables + layout). Not a
                       third-party file despite living in vendor/, see vendor/README.md.
    marked.min.js       Markdown parser (marked v12.0.2, MIT).
    highlight.min.js    Syntax highlighter (highlight.js v11.9.0, BSD-3-Clause).
    README.md           Provenance notes for the vendored libraries.
  DEV_GUIDE.md          This file.

VERSION                Single source of truth for the version string.
index.html              The SHIPPED deliverable, output of assemble.py, never hand-edited.
```

The `index.html` you ship is the *output* of `assemble.py`. The source of truth is `template.html` + `app.js` + `vendor/*` + the `VERSION` file.

## Architecture in one paragraph

The viewer is a single static page. `marked` parses Markdown into HTML; `highlight.js` colorizes `<code>` blocks. The app's own JavaScript (in `app.js`, inlined into `template.html` at build time) drives the UI: file loading (picker + drag/drop via `FileReader`), tab management (keyed by a unique per-load id, never by filename, see the `openTab`/`genTabId` functions), theme toggle, table-of-contents generation, in-document search/highlight, raw/rendered toggle, edit mode, clipboard copy, Save (download), and Export to a standalone HTML file. All of it is inlined, no CDN, no network, so the page works from `file://` with no server.

## Placeholders

`template.html` contains fixed tokens that `assemble.py` replaces with the bundled pieces:

| Placeholder | Replaced with |
|---|---|
| `/*__STYLE__*/` | contents of `vendor/base.css` |
| `/*__MARKED__*/` | contents of `vendor/marked.min.js` |
| `/*__HLJS__*/` | contents of `vendor/highlight.min.js` |
| `/*__APP__*/` | contents of `app.js` |
| `--VERSION--` | contents of the `VERSION` file at the app root |

The script asserts that no token remains unreplaced and that the vendored `.js` files do not contain a literal `</script>` (which would prematurely close the inlined `<script>` block) before writing `index.html`.

## Version stamp (single source of truth)

The version string lives in exactly one place: the `VERSION` file at the app root (currently `1.1.0`). `assemble.py` reads that file and substitutes it for the `--VERSION--` placeholder in `template.html` (used in the footer, e.g. `MD Viewer v1.1.0`). To bump the version, edit `VERSION` only, never hand-edit the footer string in the template or the built file.

## How to rebuild

From the app root:

```
python src/assemble.py
```

This writes `index.html` at the app root.

Requirements: Python 3 only. No pip packages, no network.

### Typical edit-and-rebuild loop

1. Edit `src/template.html` (structure/placeholders), `src/app.js` (behavior), or swap a vendored lib in `src/vendor/`.
2. If you change the version, edit the `VERSION` file.
3. Run `python src/assemble.py`.
4. Open the generated `index.html` in a browser to verify.

## A note on line endings when rebuilding cross-platform

The shipped `index.html` was assembled on Windows and carries CRLF line endings throughout. Re-running `assemble.py` on Linux/macOS reproduces byte-for-byte identical *content* but writes LF line endings instead, so a raw `diff`/`md5sum` against the shipped file will show every line as changed even though nothing meaningful differs (confirmed: same line count, same rendered output, only the `\r` bytes differ). This is a Python text-mode/OS default behavior, not a build bug. If you need a true byte-identical check, normalize line endings on both sides first (e.g. `dos2unix` or `sed 's/\r$//'`) before comparing.

## Where the vendored libraries come from

- `vendor/marked.min.js`, [marked](https://github.com/markedjs/marked) v12.0.2 (MIT).
- `vendor/highlight.min.js`, [highlight.js](https://github.com/highlightjs/highlight.js) v11.9.0 (BSD-3-Clause), core + common languages.
- `vendor/base.css`, the viewer's own styles, written for this project (theme variables + layout + component styles); not a third-party file despite living in `vendor/`. See `vendor/README.md`.

Keep these libraries as they ship, do not minify or transform them beyond their upstream build. To upgrade a library, drop in the new minified file under the same name and rebuild.

## How to test changes

Open the generated `index.html` directly in a real browser (double-click, or drag onto an open window). No server or special tooling is required, the app runs from `file://`. Exercise the toolbar features (theme, TOC, search, tabs, raw toggle, edit mode, Copy MD, Save, Export, copy-code buttons) manually. The fixtures under [`../test/fixtures/`](../test/fixtures/) (two `README.md` files with distinct markers, `projectA` and `projectB`) are handy for confirming that same-named files open as separate tabs rather than colliding.

> **Note on internal verification harness (not part of your workflow):** during the build sessions an automated browser harness was occasionally used to drive the running app headlessly (calling `openTab`, `readFile`, `renderTabs`, `setActive`, dispatching synthetic keyboard events, etc.). That harness runs in a sandboxed context that blocks browser primitives like `localStorage` and `fetch` and intercepts programmatic downloads, so those paths had to be stubbed or verified by a separate human smoke test. This harness is internal test scaffolding only, excluded from this source folder. As a contributor, simply open `index.html` in a real browser, which has none of those restrictions, and verify visually.

## Known unverified edge cases (from the build sessions)

- **Export visual styling** was structurally verified (the exported file is self-contained, embeds highlighted code and rendered content, and closes properly), but the exported HTML was not opened in a second browser session to confirm its visual styling end-to-end. If you change export logic, open an exported file in a browser to confirm it looks right.
- **Cross-session `localStorage` persistence** (restore of open tabs/edits across reload) was confirmed by a human smoke test, not by the sandbox harness (which blocks storage). Re-test by opening same-named files, reloading, and confirming the tabs return.

## Code hygiene expectations

- Never hand-edit the shipped `index.html`; edit `src/template.html`, `src/app.js`, or the vendor files, then rebuild.
- Don't introduce `eval()` or dynamic execution of your own code.
- Keep the vendored libraries untouched.
- Bump `VERSION` when you ship a change; never edit the footer string by hand.
