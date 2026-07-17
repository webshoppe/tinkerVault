# Vendored libraries (do not edit by hand)

These are the bundled third-party assets that ship inside `index.html`:

- `marked.min.js` — Markdown parser (marked v12.0.2, MIT).
- `highlight.min.js` — syntax highlighter (highlight.js v11.9.0, BSD-3-Clause).
- `base.css` — the viewer's own `<style>` block (theme variables + layout + component styles).

`assemble.py` (in `src/`) inlines these into `template.html` to produce the
single self-contained `index.html` in the project root. Swap a file here and
re-run `python src/assemble.py` to rebuild.
