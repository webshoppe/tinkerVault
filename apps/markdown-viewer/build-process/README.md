# Build Process Artifacts

These files aren't part of the app and aren't needed to run or rebuild it (see [`../src/DEV_GUIDE.md`](../src/DEV_GUIDE.md) for the real rebuild workflow). They're kept here because they're a genuine, unedited record of how an agent worked around its own tooling limitations mid-build, which is the kind of thing that's often thrown away and, on reflection, is worth keeping.

- **`harness.html`**, a full copy of the built app with a test fixture (`test.md`) injected directly into the page via a `<script type="text/markdown">` tag, instead of `fetch()`. Built because the agent's own sandboxed browser tool blocked `fetch()` and couldn't drive a real file picker, so this was how it exercised real rendering during verification.
- **`build_harness.py`**, the small script that generates `harness.html` from `index.html` + `test.md`. Notably has to inject before the *last* `</body>` specifically, since the app's own export feature embeds a template literal containing a `</body>` string, and a naive first-match replace would have corrupted the harness.
- **`app_extracted.js`**, the app's JavaScript, pulled out of `index.html` in isolation, for easier review outside the full page.
- **`app_harness.js`**, the same extracted JS with the fixture-loading snippet appended directly, an earlier iteration of the same workaround before `harness.html`'s full-page approach was settled on.
- **`_verify_export.py`**, a tiny self-containment check: confirms no unexpected external URLs exist in `index.html` (filtering out the deliberate example links inside the test fixture text itself).

None of this reflects a limitation of the app. All of it reflects a limitation of the *agent's own browser tool* during the build, and the workarounds it found for that. Kept as-is, unedited, for anyone curious how that kind of constraint actually gets worked around in practice.
