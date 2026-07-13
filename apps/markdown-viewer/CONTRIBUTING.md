# Contributing

Thanks for your interest in improving the Markdown Viewer. This is a small, self-contained project; there is no build server, no package registry, and (by design of this packaging pass) no remote repository yet. Contributions are made by editing the source and rebuilding the single-file output.

## Where things live

All source is in `src/`. The shipped file `release/index.html` is generated, not edited directly; never hand-edit `index.html`; edit `template.html` instead.

## Where a new feature would slot in

Most new functionality is UI- or parse-adjacent. Here is the usual map:

- **New toolbar control** (button, toggle, input): add the element in `template.html`'s `#toolbar` block, style it in the `<style>` section, and wire its event listener in the app `<script>` (look for the `/* TOOLBAR ACTIONS */` section).
- **New rendering behavior**: the parser is `marked` and the highlighter is `highlight.js`, both inlined. Change parsing/theme by editing `template.html` or swapping the bundled lib files in `src/` (then rebuild).
- **New persisted setting**: extend the `store` object and the `saveState`/`loadState` helpers so it survives reloads via `localStorage`.
- **New export option**: the `exportStandalone` function builds a standalone HTML string; extend it there, keeping the output fully self-contained (inline everything, no external URLs).

After any change, rebuild with `python assemble.py` (from `src/`) and open the result in a browser to verify.

## One known unverified edge case

From the original build verification: the **Export** feature produces a structurally valid, self-contained standalone HTML file (it embeds the highlighted code and rendered content and closes properly), but the exported file was **not** opened in a second browser session to confirm its visual styling end-to-end. When working on or near export, please open an actual exported `.html` in a browser to confirm it renders as expected.

## Code hygiene expectations

- Keep bundled libraries as-is (don't minify or transform `marked.min.js` / `highlight.min.js` beyond how they ship).
- Don't introduce `eval()` or dynamic code execution of your own.
- Prefer readable, unobfuscated app code in `template.html`.

## License

No `LICENSE` file is included yet; it needs to be added before this project is published or redistributed. Do not assume reuse rights until one is in place.
