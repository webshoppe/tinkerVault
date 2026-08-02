# Developer guide; Dossier 1.0.3

For people who want to **read or change** the codebase.  
End-user instructions: [USER_GUIDE.md](./USER_GUIDE.md). Overview: [README.md](../README.md).

---

## Big picture

```
┌────────────────────────────────────────────────────────────┐
│  main.go (windows)  go-webview2 Win32 window + WebView2    │
│  embeds ui/index.html + flag-data.inc.js                   │
│  binds JS ↔ Go API (internal/app)                          │
└───────────────────────────┬────────────────────────────────┘
                            │
        ┌───────────────────┼───────────────────┐
        ▼                   ▼                   ▼
 internal/app          internal/dossier      internal/dialog
 (API, agent HTTP,     (SQLite FTS5 store,   (folder/file
  settings, version)    import, boards)       pickers, shell)
```

- **One process, one window.** Switching dossiers reopens the store; it does not spawn extra processes.
- **UI** is a single embedded HTML/CSS/JS file (plus offline Twemoji flag SVGs).
- **Data** is mostly on disk in the dossier folder; app config is in `%APPDATA%\Dossier`.

---

## Module layout

| Path | Role |
|------|------|
| `main.go` | Windows entry: WebView2, binds, `DOSSIER_AUTO_OPEN`, logging |
| `internal/app/` | JS-facing API, agent client, settings, `Version` |
| `internal/dossier/` | Store, import, PDF/Office extract, notes, stickies, boards, kanban, decisions, FTS |
| `internal/dialog/` | Windows open/save/shell helpers |
| `ui/index.html` | Entire UI |
| `ui/flag-data.inc.js` | Injected Twemoji SVGs (`/*__FLAG_DATA__*/`) |
| `winres/` | Icon PNGs + `winres.json` → `rsrc_windows_amd64.syso` |
| `third_party/go-webview2` | Local patch of jchv/go-webview2 (context menus vs DevTools) |
| `cmd/smoke/` | Headless Windows store smoke test |
| `build-process/` | Re-runnable verify scripts |
| `releases/<version>/` | Ship package |

---

## Build & resources

```bash
export PATH="$HOME/.local/go/bin:$HOME/go/bin:$PATH"
make test
make build      # go-winres + GOOS=windows go build
make release    # copies exe, icon, docs → releases/$(cat VERSION)/
```

**Icon + PE version:** [go-winres](https://github.com/tc-hib/go-winres) reads `winres/winres.json` and writes `rsrc_windows_amd64.syso`, which `go build` links automatically. Why go-winres: pure Go, no CGO/windres, works from WSL cross-compile, handles icon + `RT_VERSION` + manifest in one step.

**App version constant:** `internal/app/version.go` (`Version = "1.0.3"`). Keep in sync with `VERSION`, `winres/winres.json`, and UI footer default.

**go-webview2 replace:** `go.mod` replaces upstream with `./third_party/go-webview2` so default **context menus stay enabled** while **DevTools stay off** unless `DOSSIER_DEBUG=1` (upstream had both gated on `Debug`).

---

## Feature organization by tier (historical)

| Tier | Highlights |
|------|------------|
| **1** | Native window, dossier folder, SQLite FTS5, import, notes, stickies |
| **2** | DnD import, paint/kanban/annotate, multi-dossier, decisions, doc detail |
| **3 / 3.1** | Twemoji flags, polish, sticky CSS, intros, settings, icon asset |
| **4 / 4.1 / 4.2** | Layout 1320, optional agent, docx/xlsx, host first-save, FTS OR+bm25 cap, `~$` skip, toast |
| **5** | Ask clear conversation; Documents sort |
| **Packaging (1.0.0)** | Host-gate fix, context menus, PE icon/version, Notes Copy, docs, release folder |
| **Packaging gap-fix (1.0.1)** | Icon alpha (RGBA) fix, per-dossier icon variant skipped (concrete reason), Open externally per-type override, Notes auto-save safety net |
| **Packaging gap-fix (1.0.2)** | Icon regenerated fresh for size/legibility (no alpha regression) |
| **Packaging gap-fix (1.0.3)** | Icon paper/sticky art direction reworked (rounded corners, diagonal placement); folder body unchanged |

Do not re-litigate shipped FTS ranking, toast arrows, Documents sort, the host/port Settings gate, the context-menu fix, or the icon alpha pipeline, unless fixing a regression.

---

## Settings / Ask panel gate

`AgentConfigured` requires **non-empty host** and **port &gt; 0**.

- **Clear host + save** → panel hidden (toast: “Ask panel hidden”), even if port remains.
- Empty host is **not** auto-filled back to `127.0.0.1` (that was the packaging-tier bug).
- UI pre-fills `127.0.0.1` only when the agent was **never** configured (first-time convenience).

Relevant code: `internal/app/agent.go` (`AgentConfigured`), `SaveAppSettings` in `api.go`, `renderSettings` / save handler in `ui/index.html`.

---

## Right-click / DevTools

Upstream go-webview2 set both:

- `AreDefaultContextMenusEnabled = Debug`
- `AreDevToolsEnabled = Debug`

Production (`Debug=false`) therefore killed Cut/Copy/Paste menus. Local patch enables **context menus always**, DevTools **only when** `DOSSIER_DEBUG=1`.

---

## Agent context selection (do not casually change)

Documented from the Tier 4.2 gap-fix pass (see PROJECT_SUMMARY.md's Tier history):

- `SearchLoose`: stopwords stripped, remaining terms OR-matched  
- Ordered by **bm25**, limited to **top 4** hits  
- Per-hit body cap + total context rune cap in `agent.go`

---

## Known limitations

| Topic | Limit |
|-------|--------|
| PDF | Image-only / scanned pages often have no extractable text |
| Office | Best-effort ZIP/XML text; no formulas/charts fidelity; encrypted OOXML unsupported |
| Agent | No embedded model; HTTP client only; live agent optional |
| Per-dossier taskbar icon | **Still skipped**; one process/one window; no separate taskbar buttons per dossier (see [releases/1.0.3/STATUS.md](../releases/1.0.3/STATUS.md)) |
| Open externally prefs | `AppSettings.OpenWith` map `".ext" → exe path` in `%APPDATA%\Dossier\config.json`; `OpenDocumentExternally(id, forceDefault)` |
| Notes safety net | UI-only snapshot at `openNote`; **Revert edits** restores that snapshot (not multi-version history) |
| Icon alpha | `assets/dossier-icon.png` and `winres/icon_*.png` must be **RGBA** (color type 6); packaging briefly shipped RGB-only; fixed in 1.0.1 |
| Cross-compile | Build from Linux/WSL with `CGO_ENABLED=0`; needs Go + go-winres |
| `ui/index.html` edits | Single ~3k+ line file; a bulk find-replace once truncated it mid-pass (lost the annotate/kanban sections), the actual cause of the Tier 3 blank-window regression (see PROJECT_SUMMARY.md's Tier 3 / 3.1 history). Use surgical, scoped patches, never a bulk rewrite of the whole file |

---

## Verify

```bash
# Unit tests (Linux OK)
go test ./internal/... -count=1

# Windows smoke (on Windows host / via wine-less path to .exe)
build/smoke.exe

# Packaging verify scripts
bash build-process/verify-build.sh
powershell.exe -File build-process/verify-windows.ps1
```

---

## Debug

| Env | Effect |
|-----|--------|
| `DOSSIER_DEBUG=1` | WebView2 DevTools enabled |
| `DOSSIER_AUTO_OPEN=1` | Open last dossier before UI paints |

Logs: `%LOCALAPPDATA%\Dossier\dossier.log` (GUI binary has no console).
