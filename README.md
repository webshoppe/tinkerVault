# tinkerVault

A small collection of offline-first, single-file web apps. No install,
no account, no server. Double-click and go, or use the hosted versions
below.

## Apps

| App | What it does | Try it | Source |
|---|---|---|---|
| **Markdown Viewer** (v1.1.0) | Drag-and-drop `.md` viewer with tabs, syntax highlighting, TOC, search, dark/light theme, edit mode, and standalone export | [Open](https://webshoppe.github.io/tinkerVault/apps/markdown-viewer/) | [`apps/markdown-viewer/`](apps/markdown-viewer/) |
| **Whiteboard** (v2.4.0) | Five boards: Paint, Sticky Notes, Annotate, Wordpad, Kanban, with smart paste routing, quota tracking, full board history/trash, and command-palette search, all in one portable file | [Open](https://webshoppe.github.io/tinkerVault/apps/whiteboard/whiteboard.html) | [`apps/whiteboard/`](apps/whiteboard/) |
| **Hermes Console** (v1.0.0) | Resolve Hermes approval-gated tool calls and browse sessions from any device, a phone-friendly PWA that talks straight to a running Hermes API server (approve/deny tool calls, chat, session browse + fork) | [Open](https://webshoppe.github.io/tinkerVault/apps/hermes-console/) | [`apps/hermes-console/`](apps/hermes-console/) |

Each **Try it** link above always points to that app's current version,
hosted live via GitHub Pages. GitHub Pages only ever serves the current
version, it doesn't host old releases live, so looking at an older
version means downloading its `releases/vX.X.X/` folder below and
opening its HTML file directly; it runs exactly the same way, just
without the hosted link. The landing page itself lives at
[webshoppe.github.io/tinkerVault](https://webshoppe.github.io/tinkerVault/).

Each app also runs completely offline, just download its folder and
double-click the HTML file. Nothing calls out to the network at
runtime. Old versions never disappear, every app keeps its full
version history as self-contained snapshots inside its own
`releases/` folder:

- Markdown Viewer: [v1.1.0](apps/markdown-viewer/) current, [v1.0.0](apps/markdown-viewer/releases/v1.0.0/index.html) archived
- Whiteboard: [v2.4.0](apps/whiteboard/whiteboard.html) current, [v1.0.0](apps/whiteboard/releases/v1.0.0/whiteboard.html) archived
- Hermes Console: [v1.0.0](apps/hermes-console/) current, no earlier version yet

## Why single-file apps

Every app here inlines its own dependencies and runs from a plain
double-clicked HTML file, no CDN, no build step required to use it,
no account, nothing phoning home. That constraint is deliberate, not
a limitation: it means anything in this repo still works exactly the
same way in five years, on a USB stick, with zero setup.

Native/expanded editions of some apps may live alongside their
portable originals as they grow beyond what a single HTML file can
reasonably do, each app's own README says which editions exist.

Curious about the hardware and process behind this, and what
"verified" actually means in this repo? See
[`docs/PHILOSOPHY.md`](docs/PHILOSOPHY.md).

## Structure

```
tinkerVault/
├── apps/
│   ├── hermes-console/  (currently v1.0.0)
│   │   └── releases/
│   │       └── v1.0.0/  (duplicate of current, kept as a snapshot)
│   ├── markdown-viewer/  (currently v1.1.0)
│   │   └── releases/
│   │       ├── v1.0.0/  (archived)
│   │       └── v1.1.0/  (duplicate of current, kept as a snapshot)
│   └── whiteboard/  (currently v2.4.0)
│       └── releases/
│           ├── v1.0.0/  (archived)
│           └── v2.4.0/  (duplicate of current, kept as a snapshot)
├── docs/
│   └── PHILOSOPHY.md
├── DESKTOP-LAUNCHER.md
├── LICENSE
├── README.md
└── index.html
```

Each app is self-contained: its own README, docs, source, and version.
Apps do not share dependencies or a build system by design.

## Changelog

Repo-level changes (structure, conventions, the landing page, cross-app housekeeping) are tracked in [`CHANGELOG.md`](CHANGELOG.md). Each app keeps its own feature-level changelog too, see that app's own README.

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
