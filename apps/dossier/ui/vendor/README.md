# Vendored libraries (do not edit by hand)

These are the bundled third-party assets used by Notes' Markdown workbench (Edit / Split / Preview):

- `marked.min.js`; Markdown parser (marked v12.0.2, MIT).
- `highlight.min.js`; syntax highlighter (highlight.js v11.9.0, BSD-3-Clause).

Both are embedded directly by `ui/index.html` (a single hand-maintained file; see the standing edit-discipline note in DEV_GUIDE.md, surgical patches only, never a bulk rewrite). To update a library version, replace the file here and update the corresponding `<script>` reference in `ui/index.html` by hand; there is no separate build/assemble step.
