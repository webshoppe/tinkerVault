# Developer Guide

This document is for anyone who wants to read, modify, or rebuild the Markdown Viewer from its source.

## What this folder contains

```
src/
├── template.html        # The full app: HTML structure, CSS, and the app JavaScript.
│                         #   Library code is injected at build time via placeholders.
├── assemble.py          # Build script. Reads template.html + the four lib files below
│                         #   and writes a single self-contained index.html.
├── marked.min.js        # Markdown parser (marked, v12.0.2).
├── highlight.min.js     # Syntax highlighter (highlight.js, v11.9.0), core + common langs.
├── github-light.css     # highlight.js "github" theme (light).
├── github-dark.css      # highlight.js "github-dark" theme (dark).
└── DEV_GUIDE.md         # This file.
```

The `release/index.html` you ship is the *output* of `assemble.py`. The source of truth is `template.html` plus the lib files.

## Architecture in one paragraph

The viewer is a single static page. `marked` parses Markdown into HTML. `highlight.js` then colorizes any `<code>` blocks. The app's own JavaScript (in `template.html`) handles the UI: file loading (picker + drag/drop via `FileReader`), tab management, theme toggle, table-of-contents generation from heading IDs, in-document search/highlight, raw/rendered toggle, clipboard copy, and export to a standalone HTML file. All state the user would want remembered (open files, theme, TOC state) is persisted to `localStorage`. Crucially, **everything is inlined**, no CDN, no network. The page works from `file://` with no server.

## Placeholders

`template.html` contains fixed tokens that `assemble.py` replaces with the bundled libraries:

| Placeholder            | Replaced with                              |
|------------------------|--------------------------------------------|
| `/*__MARKED__*/`       | contents of `marked.min.js`                |
| `/*__HLJS__*/`         | contents of `highlight.min.js`             |
| `/*__HLJS_LIGHT__*/`   | contents of `github-light.css`             |
| `/*__HLJS_DARK__*/`    | `github-dark.css`, scoped under `html[data-theme="dark"]` |
| `HLJS_LIGHT_SRC`       | light CSS, exposed to the export feature   |
| `HLJS_DARK_SRC`        | dark CSS (scoped), exposed to the export feature |

The dark theme CSS is scoped with a `html[data-theme="dark"] ` prefix so it only applies in dark mode and wins by specificity over the light rules.

## How to rebuild

From inside this `src/` folder:

```bash
python assemble.py
```

This reads `template.html` and the four library files from the *same folder*, then writes `index.html` next to them. To produce the distributable, copy that generated `index.html` into the release root (or run the build directly in the location where you want the output).

Requirements: Python 3 only. No pip packages, no network. The script asserts that the library files do not contain a literal `</script>` (which would prematurely close the inlined `<script>` block) before writing.

### Typical edit-and-rebuild loop

1. Edit `template.html` (CSS, HTML structure, or app JavaScript).
2. If you swap a library, replace the corresponding `.js`/`.css` file and keep its filename, or update `assemble.py` to match.
3. Run `python assemble.py`.
4. Open the generated `index.html` in a browser to verify.

## How to test changes

Open the generated `index.html` directly in a real browser (double-click it, or drag it onto a browser window). No server or special tooling is required; the app loads and runs from `file://`. Use `test/test.md` (the fixture with a table, fenced code block, task list, and nested headings) as a quick sanity check, and exercise the toolbar features (theme, TOC, search, tabs, raw toggle, export, copy) manually.

> Note on test scaffolding: during the original build session, a `harness.html` / `build_harness.py` pair (plus a couple of related verification scripts) was created as *internal* test scaffolding to work around the automated browser tool's sandbox restrictions (it blocked `fetch` and clipboard readback). Those files are **not** part of the normal developer workflow and are intentionally kept out of this `src/` folder, they live in [`../build-process/`](../build-process/) instead, preserved as a record of the workaround rather than something you need to rebuild the app. A real developer opens `index.html` in a real browser, which has none of those sandbox restrictions, and never needs a harness.

## Known unverified edge case

Per the build session's own report: the **export** feature was structurally verified, the generated standalone file is self-contained, embeds the highlighted code and rendered content, and closes properly. However, the exported `.html` file was *not* opened in a second browser session to confirm its visual styling end-to-end. If you change the export logic, please open an exported file in a browser to confirm it looks right.
