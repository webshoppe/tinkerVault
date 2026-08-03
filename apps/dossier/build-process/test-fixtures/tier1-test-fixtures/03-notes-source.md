# Dossier Tier 1 — Markdown Import Test

This file exists to verify Dossier's markdown import path, since only `.txt` and `.pdf` had been manually tested so far, not `.md`.

## Search keyword

FTS5 search keyword for this file: `MARIGOLD7731`

## Structure check

This document intentionally includes the kinds of elements a real dossier note might contain:

- A bullet list item
- Another bullet list item
- A third one, for good measure

1. First numbered step
2. Second numbered step

```
a small fenced code block
to confirm formatting survives import
```

## Expected result

This should appear in the Documents list as a markdown file, be fully searchable via FTS5 (try searching `MARIGOLD7731`), and open/render correctly without needing "open externally."
