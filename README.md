# tinkerVault

A small collection of offline-first, single-file web apps. No install,
no account, no server. Double-click and go, or use the hosted versions
below.

## Apps

| App | What it does | Try it | Source |
|---|---|---|---|
| **Markdown Viewer** | Drag-and-drop `.md` viewer with syntax highlighting, tabs, TOC, search, dark/light theme, standalone export | [Open](https://webshoppe.github.io/tinkerVault/apps/markdown-viewer/) | [`apps/markdown-viewer/`](apps/markdown-viewer/) |
| **Whiteboard** (v2.4.0) | Five boards: Paint, Sticky Notes, Annotate, Wordpad, Kanban, with smart paste routing, quota tracking, full board history/trash, and command-palette search, all in one portable file | [Open](https://webshoppe.github.io/tinkerVault/apps/whiteboard/whiteboard.html) | [`apps/whiteboard/`](apps/whiteboard/) |

Each app also runs completely offline, just download its folder and
double-click the HTML file. Nothing calls out to the network at
runtime. Some apps keep past versions available as self-contained
snapshots: Whiteboard's original four-board release is still available
at [`apps/whiteboard/releases/v1.0.0/`](apps/whiteboard/releases/v1.0.0/whiteboard.html).

## Why single-file apps

Every app here inlines its own dependencies and runs from a plain
double-clicked HTML file, no CDN, no build step required to use it,
no account, nothing phoning home. That constraint is deliberate, not
a limitation: it means anything in this repo still works exactly the
same way in five years, on a USB stick, with zero setup.

Native/expanded editions of some apps may live alongside their
portable originals as they grow beyond what a single HTML file can
reasonably do, each app's own README says which editions exist.

## Structure

```
tinkerVault/
├── apps/
│   ├── markdown-viewer/
│   └── whiteboard/
└── docs/            (repo-wide notes, if any)
```

Each app is self-contained: its own README, docs, source, and version.
Apps do not share dependencies or a build system by design.

## License

MIT, see [LICENSE](LICENSE). Fork it, tinker with it, ship your own version, no attribution required.

## Repo

[github.com/webshoppe/tinkerVault](https://github.com/webshoppe/tinkerVault)

## Contributing

See each app's own `README.md` / `CONTRIBUTING.md` for that app's
specifics. General principle across this repo: prefer the simplest
thing that actually works, verify claims in a real browser before
calling something done, document honest limitations rather than
papering over them.
