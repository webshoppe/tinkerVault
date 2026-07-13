#!/usr/bin/env python3
# Pull the captured export string from the live browser session via a separate fetch is not possible;
# instead this just checks the export template logic by confirming no external URLs in index.html.
import pathlib, re
s = pathlib.Path(__file__).parent.read_text() if False else open(__file__.parent / "index.html", encoding="utf-8").read()
ext = re.findall(r'(https?:)?//[^\s"\'<>]+', s)
ext = [e for e in ext if 'example.com' not in e and 'nousresearch' not in e]
# Allow only the anchor/href example refs inside markdown text, not script/link deps
print("external URLs in index.html:", ext[:10], "count", len(ext))
