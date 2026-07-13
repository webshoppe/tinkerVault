#!/usr/bin/env python3
import pathlib
base = pathlib.Path(__file__).parent
idx = (base / "index.html").read_text(encoding="utf-8")
md = (base / "test.md").read_text(encoding="utf-8")
loader = '<script type="text/markdown" id="fixture">' + md + '</script>\n'
loader += '<script>\nwindow.addEventListener("load",function(){var t=document.getElementById("fixture").textContent;window.openTab("test.md",t);});\n</script>\n'
# Replace ONLY the final </body> (the export template literal also contains a </body>).
last = idx.rfind('</body>')
out = idx[:last] + loader + idx[last:]
(base / "harness.html").write_text(out, encoding="utf-8")
import re
print("harness written", len(out))
print("fixture tags:", len(re.findall(r'<script type="text/markdown" id="fixture">', out)))
print("injector:", out.count('window.addEventListener("load"'))
