# apps/

Every app in the tinkerVault collection lives here, each self-contained in its own subfolder: own README, docs, source, and version. Apps do not share dependencies or a build system by design.

- [`markdown-viewer/`](markdown-viewer/), Markdown Viewer: drag-and-drop `.md` viewer with tabs, syntax highlighting, TOC, search, dark/light theme, edit mode, and standalone export.
- [`whiteboard/`](whiteboard/), Whiteboard: five boards (Paint, Sticky Notes, Annotate, Wordpad, Kanban) with smart paste routing, quota tracking, board history, and command-palette search.
- [`hermes-console/`](hermes-console), Hermes Console: A single-file Progressive Web App for resolving Hermes approval-gated tool calls and browsing sessions from your phone or any other device.
- [`dossier/`](dossier/), Dossier: portable, offline dossier workspace for Windows (compiled `.exe`, not a browser app), one folder per workspace holding real documents, notes, sticky notes, paint, kanban, annotate, and a decision timeline, plus full-text search and an optional local agent panel.

See the [root README](../README.md) for the collection overview, hosted "Try it" links, and the full version history across both apps.
