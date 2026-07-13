#!/usr/bin/env python3
"""Assemble a single self-contained index.html from template.html + local libs."""
import re, pathlib

base = pathlib.Path(__file__).parent
tpl = (base / "template.html").read_text(encoding="utf-8")

marked = (base / "marked.min.js").read_text(encoding="utf-8")
hljs = (base / "highlight.min.js").read_text(encoding="utf-8")
light = (base / "github-light.css").read_text(encoding="utf-8")
dark = (base / "github-dark.css").read_text(encoding="utf-8")

# Version stamp (single source of truth: VERSION file at the release root)
version = "0.0.0"
try:
    version = (base.parent / "VERSION").read_text(encoding="utf-8").strip() or version
except FileNotFoundError:
    pass

# Scope dark theme rules under html[data-theme="dark"] so they win by specificity
dark_scoped = re.sub(r'([^{}]+\{)', r'html[data-theme="dark"] \1', dark)

# Sanity: ensure libs don't contain a closing script tag that would break the block
assert "</script" not in marked.lower(), "marked contains </script>"
assert "</script" not in hljs.lower(), "hljs contains </script>"

out = tpl
out = out.replace("/*__MARKED__*/", marked)
out = out.replace("/*__HLJS__*/", hljs)
out = out.replace("/*__HLJS_LIGHT__*/", light)
out = out.replace("/*__HLJS_DARK__*/", dark_scoped)

# Expose hljs css sources for the export feature
out = out.replace("HLJS_LIGHT_SRC", light.replace("`", "\\`"))
out = out.replace("HLJS_DARK_SRC", dark_scoped.replace("`", "\\`"))

# Stamp version (single source of truth: VERSION file)
out = out.replace("--VERSION--", version)

target = base / "index.html"
target.write_text(out, encoding="utf-8")
print("Wrote", target, len(out), "bytes")
