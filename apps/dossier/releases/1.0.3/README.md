# Dossier

<img src="dossier-icon.png" width="48" height="48" alt="Dossier icon" />

**Version 1.0.3**; a portable, offline **dossier workspace** for Windows.

One native `.exe` (not a browser tab). Pick a folder; that folder becomes a self-contained workspace with real documents, notes, stickies, boards, search, and an optional local agent panel. Backup = copy the folder.

---

## What it is

Dossier is a multi-surface productivity app built around a **dossier folder**:

| Surface | Purpose |
|---------|---------|
| **Documents** | Import files (md/txt/pdf/docx/xlsx + attachments); sort, preview, open externally |
| **Notes** | Markdown notes as real files under `notes/` |
| **Sticky Notes** | Freeform board with colors, sizes, emoji/flags |
| **Paint** | Freehand canvas |
| **Annotate** | Image markup (highlight, arrow, label, redact, crop) |
| **Kanban** | Columns, WIP limits, drag cards |
| **Decisions** | Chronological spine + optional frozen document versions |
| **Search** | Full-text (SQLite FTS5) across the dossier |
| **Ask dossier** | Optional: send FTS-selected context to *your* local HTTP agent |
| **Settings** | App-wide prefs in `%APPDATA%\Dossier` |

Requires the **Microsoft Edge WebView2 Runtime** (preinstalled on modern Windows 10/11).

---

## How to run

1. Get `Dossier.exe` from `releases/1.0.3/` (or build it; see below).
2. Double-click it on Windows.
3. **Create / open folder…** or **Open last** / pick a recent workspace.
4. Work. Everything for that workspace lives inside the folder you chose.

```powershell
# Example
C:\path\to\releases\1.0.3\Dossier.exe
```

Optional automation (debug / verify only): set `DOSSIER_AUTO_OPEN=1` to open the last dossier on launch.

---

## Where data lives

### Per dossier (portable)

```
MyDossier/
  .dossier                 # marker
  dossier.db               # SQLite + FTS5 index & board state
  documents/               # imported files (real files on disk)
  notes/                   # typed notes as .md files
  boards/                  # paint / annotate / kanban assets
  decision-versions/       # frozen document snapshots for decisions
```

Backup or move a project by copying that whole folder.

### App-wide (not inside a dossier)

| Path | Contents |
|------|----------|
| `%APPDATA%\Dossier\config.json` | Last path, recents, Settings (agent host/port, intros, …) |
| `%LOCALAPPDATA%\Dossier\dossier.log` | App log |
| `%LOCALAPPDATA%\Dossier\webview\` | WebView2 user data |

---

## Build (from WSL2 Ubuntu)

```bash
export PATH="$HOME/.local/go/bin:$HOME/go/bin:$PATH"
cd /path/to/dossier-app
go test ./internal/... -count=1
make build          # → build/Dossier.exe  (icon + version embedded)
make release        # → releases/1.0.3/
```

Or without Make:

```bash
go-winres make --in winres/winres.json --out rsrc --arch amd64
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-H windowsgui -s -w" -o build/Dossier.exe .
```

Console binary (logs to console): `make build-console`.

---

## Version

In-app footer and Windows **File → Properties → Details** both report **1.0.3**.  
See [STATUS.md](./STATUS.md) in this folder for this version's verification.

---

## Docs

| Doc | Audience |
|-----|----------|
| [USER_GUIDE.md](./docs/USER_GUIDE.md) | End users; one section per surface |
| [DEV_GUIDE.md](./docs/DEV_GUIDE.md) | Developers; architecture, tiers, limits |
| [PROJECT_SUMMARY.md](./PROJECT_SUMMARY.md) | Tier history + room for a cold-read note |
| [build-process/](../../build-process/) | Re-runnable verify scripts (lives at the app root, not duplicated per release) |
| [STATUS.md](./STATUS.md) | This version's packaging verification |

---

## License / runtime

- App code: MIT, see the repo's [LICENSE](../../../../LICENSE) file.
- Flag images: Twemoji (CC-BY 4.0 Twitter/Twemoji), embedded offline.
- WebView2 Runtime: Microsoft (system component). Also vendors a locally-patched go-webview2 (MIT) and Microsoft's WebView2Loader SDK (BSD-style redistribution license); see [third_party/go-webview2/](../../third_party/go-webview2/) in the main app source for both LICENSE files.
