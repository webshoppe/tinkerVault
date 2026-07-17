#!/usr/bin/env python3
"""Assemble a single self-contained index.html from template.html + src files.

Run:  python src/assemble.py
Outputs: index.html (in the project root, sibling of src/)

The shipped index.html is one file with no build step so it opens straight from
file://. This script exists only so future edits can happen on small, readable
source files (src/template.html, src/app.js, src/vendor/*) instead of a ~190 KB
single file. It performs a lossless round-trip of the current build.
"""
import pathlib

ROOT = pathlib.Path(__file__).resolve().parent.parent
SRC = ROOT / "src"
VENDOR = SRC / "vendor"

template = (SRC / "template.html").read_text(encoding="utf-8")
style = (VENDOR / "base.css").read_text(encoding="utf-8")
marked = (VENDOR / "marked.min.js").read_text(encoding="utf-8")
hljs = (VENDOR / "highlight.min.js").read_text(encoding="utf-8")
app = (SRC / "app.js").read_text(encoding="utf-8")

# Sanity: vendored libs must not contain a closing script tag that would
# prematurely terminate the <script> block in the assembled file.
assert "</script" not in marked.lower(), "marked.min.js contains </script>"
assert "</script" not in hljs.lower(), "highlight.min.js contains </script>"

out = template
out = out.replace("/*__STYLE__*/", style)
out = out.replace("/*__MARKED__*/", marked)
out = out.replace("/*__HLJS__*/", hljs)
out = out.replace("/*__APP__*/", app)

# Version stamp (single source of truth: VERSION file at the project root).
version = "0.0.0"
try:
    version = (ROOT / "VERSION").read_text(encoding="utf-8").strip() or version
except FileNotFoundError:
    pass
out = out.replace("--VERSION--", version)

# Guard: every token must have been consumed.
for tok in ["/*__STYLE__*/", "/*__MARKED__*/", "/*__HLJS__*/", "/*__APP__*/"]:
    assert tok not in out, f"unreplaced token: {tok}"

target = ROOT / "index.html"
target.write_text(out, encoding="utf-8")
print("Wrote", target, len(out), "bytes")
