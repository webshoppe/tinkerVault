# Developer guide — Dossier 2.0.1

For people who want to **read or change** the codebase.  
End-user instructions: [USER_GUIDE.md](./USER_GUIDE.md). Overview: [README.md](../README.md).

---

## Big picture

```
┌────────────────────────────────────────────────────────────┐
│  main.go (windows)  go-webview2 Win32 window + WebView2    │
│  embeds ui/index.html + flag-data.inc.js + vendor JS       │
│  binds JS ↔ Go API (internal/app)                          │
└───────────────────────────┬────────────────────────────────┘
                            │
        ┌───────────────────┼───────────────────┐
        ▼                   ▼                   ▼
 internal/app          internal/dossier      internal/dialog
 (API, agent HTTP,     (SQLite FTS5 store,   (folder/file
  settings, version)    import, boards,       pickers, shell)
                        collections helper)
```

- **One process, one window.** Switching dossiers reopens the store; no extra processes or taskbar icons per dossier.
- **UI** is a single embedded HTML/CSS/JS file (plus offline Twemoji flag SVGs and vendored Markdown libs).
- **Workspace data** lives in the dossier folder; **app config** is in `%APPDATA%\Dossier`.

---

## Module layout

| Path | Role |
|------|------|
| `main.go` | Windows entry: WebView2, binds, `DOSSIER_AUTO_OPEN`, logging |
| `internal/app/` | JS-facing API, agent client, settings, `Version` |
| `internal/dossier/` | Store, import, PDF/Office/ODF extract, notes, stickies, boards, kanban, decisions, bookmarks, collections helpers, FTS |
| `internal/dialog/` | Windows open/save/shell helpers |
| `ui/index.html` | Entire UI (~5k lines) |
| `ui/flag-data.inc.js` | Injected Twemoji SVGs (`/*__FLAG_DATA__*/`) |
| `ui/vendor/marked.min.js` | Markdown (`/*__MARKED__*/`) |
| `ui/vendor/highlight.min.js` | Code highlighting (`/*__HIGHLIGHT__*/`) |
| `winres/` | Icon PNGs + `winres.json` → `rsrc_windows_amd64.syso` |
| `third_party/go-webview2` | Local patch of jchv/go-webview2 |
| `cmd/smoke/` | Headless store smoke (Windows) |
| `build-process/` | Re-runnable verify scripts |
| `releases/<version>/` | Ship package |
| `Makefile` | `test`, `build` (both exes), `release` |

---

## Build walkthrough (current toolchain)

Works from **WSL2 Ubuntu** (or other Linux) with Go 1.22+ and go-winres on `PATH`.

```bash
# Tooling (once)
# Go: https://go.dev/dl/  → often $HOME/.local/go/bin/go
# go-winres: go install github.com/tc-hib/go-winres@latest  → $HOME/go/bin

export PATH="$HOME/.local/go/bin:$HOME/go/bin:$PATH"
cd /path/to/dossier-app

# Tests (Linux)
go test ./internal/... -count=1 -timeout 90s

# Production GUI + console twin (always rebuild both)
make build
# → build/Dossier.exe          (windowsgui, OriginalFilename=Dossier.exe)
# → build/Dossier-console.exe  (console, OriginalFilename=Dossier-console.exe)

# Ship folder
make release
# → releases/$(cat VERSION)/  with both exes (see Makefile), icon, docs, sample-files

# Packaging gate (reads VERSION dynamically; rebuilds GUI + smoke)
bash build-process/verify-build.sh
```

**Without Make (GUI only):**

```bash
go-winres make --in winres/winres.json --out rsrc --arch amd64
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-H windowsgui -s -w" -o build/Dossier.exe .
```

Prefer `make build` so both binaries and PE **OriginalFilename** stay correct (winres is regenerated between the two links).

### Version fields (must stay in agreement)

| Location | 2.0.1 ship |
|----------|------------|
| `VERSION` | `2.0.1` |
| `internal/app/version.go` `Version` | `2.0.1` |
| `winres/winres.json` FileVersion / ProductVersion strings | `2.0.1` |
| PE fixed quad | `2.0.1.24` |
| UI footer / brand / launcher defaults | `2.0.1` |

### go-webview2 local patches (`replace` in go.mod)

| Patch | Why |
|-------|-----|
| Context menus always on; DevTools only if `DOSSIER_DEBUG=1` | Upstream gated both on `Debug` |
| `AreBrowserAcceleratorKeysEnabled(false)` | Desktop app; less browser chrome key theft |
| `AcceleratorKeyCallback` Eval-closes open sticky popovers on Escape | Host backup for picker dismiss |
| Escape on **KEY_UP** (vk 0x1B) as well as KEY_DOWN | WebView2 delivers Escape as KEY_UP on this host |
| `COREWEBVIEW2_PHYSICAL_KEY_STATUS` BOOL as `int32` | Go `bool` misaligned WasKeyDown; callback never ran |

Do not “simplify” the Escape path without re-hand-testing sticky pickers with NVDA.

### Icon design history

Factual trail from packaging STATUS / PROJECT_SUMMARY:

- Tier 3: `dossier-icon.png` generated (folder + papers/sticky motif).
- **1.0.0:** PE embed; alpha broken (flattened JPEG-style under `.png`).
- **1.0.1:** RGBA mask fix; washed/small look.
- **1.0.2:** Fresh regenerate; folder body good; paper stack still stiff/jagged.
- **1.0.3 (final):** Paper restyle only (rounded, diagonal, softer tuck); flattened jpg as **style reference** only. Current `winres/icon_*.png` / `assets/dossier-icon.png`.

Owner’s more detailed concept was set aside (too busy at small sizes). **Per-dossier taskbar badges rejected:** one process, one window; icon set once from PE at class creation.

---

## Data model notes (v2)

| Topic | Detail |
|-------|--------|
| Schema | `SchemaVersion = "3"` (in store migrate) |
| Sticky due | `stickies.due_at` nullable; UI + Agenda |
| Sticky↔Kanban | Non-destructive link: card holds sticky id; text/due sync; unlink keeps both |
| Agenda | Aggregates sticky due, kanban card due, decision dates (`ListCalendarItems`) |
| Collections | In `%APPDATA%\Dossier\config.json` (ids, names, member paths); not per-dossier |
| Bookmarks | Per-dossier `bookmarks.json` (portable with the folder) |
| Notes | Files under `notes/`; UI snapshot for **Revert edits** only |

---

## Settings / Ask panel gate

`AgentConfigured` requires **non-empty host** and **port &gt; 0**.

- Clear host + save → panel hidden (even if port remains).
- Empty host is **not** auto-filled back to `127.0.0.1`.
- UI pre-fills `127.0.0.1` only when the agent was **never** configured.

---

## Agent context selection (do not casually change)

- `SearchLoose`: stopwords stripped, remaining terms OR-matched  
- Ordered by **bm25**, limited to **top 4** hits  
- Per-hit body cap + total context rune cap in `agent.go`

---

## Feature organization by tier (abbreviated)

See [PROJECT_SUMMARY.md](./PROJECT_SUMMARY.md) for the full narrative. High level:

| Era | Highlights |
|-----|------------|
| v1 1.x | Core surfaces, packaging, icon 1.0.0–1.0.3 |
| v2 T1–T2 | Auto-resume, ODF, notes split-pane, display names, color presets |
| v2 T3–T4 | Due dates, sticky↔kanban link, Agenda, Collections, bookmarks |
| v2 T5+ | Sort, a11y landmarks/keyboard/scroll/toasts, Escape KEY_UP fix |
| v2 T6/T7 | Polish, version in UI, route focus, kanban arrows, OriginalFilename |
| Packaging 2.0.0 | Docs + release folder + verify script |
| **Gap-fix 2.0.1** | Notes/Decisions two-pane width fix (hard-pinned flex basis, see Layout patterns below) |

---

## Layout patterns: list+detail panes (added 2.0.1)

`ui/index.html` has two views (`#notes-view`, `#decisions-view`) that put a list/spine pane and a detail pane as **direct flex children of the view root** with `flex-direction: row !important`. Documents avoids this by nesting its row (`.docs-layout`) under a column view; that structural difference is why only Notes/Decisions were affected by the 2.0.1 bug.

**The lesson:** a soft flex basis (`flex: 0 1 <n>px` plus a `width: min(..., vw)` cap) is not equivalent to a hard pin (`flex: 0 0 <n>px`). Under min-content pressure (long unbroken titles, no wrap points), the soft-basis pane can balloon past its intended width and starve the sibling `flex: 1` pane, even with a `max-width` set. This is exactly what happened here: the v1-era containment fix used soft flex to solve horizontal-scroll/no-stacking, which fixed that problem but never locked the actual width ratio, and the squish symptom resurfaced at wide viewports.

**Going forward, any new list+detail split should use:**
- List/rail pane: `width: <n>px; flex: 0 0 <n>px; min-width: 0`
- Detail/editor pane: `flex: 1 1 0%; min-width: 0`
- Any banner/intro that could become a stray flex sibling: `flex: 0 0 100%` (forces its own row)

See `releases/2.0.1/STATUS.md` for the full diagnosis (headless Chrome measurements at 1920/1600/1280/1100) and fix.

---

## Known limitations

| Topic | Limit |
|-------|--------|
| PDF | Image-only / scanned pages often have no extractable text |
| Office / ODF | Best-effort text; no fidelity for layouts/formulas; encrypted OOXML unsupported |
| Agent | No embedded model; HTTP client only |
| Per-dossier taskbar icon | Rejected (one window) |
| Notes Revert | One open-session snapshot, not history |
| Collections | App config only; copying a folder does not copy membership |
| `ui/index.html` | Surgical edits only; bulk replace once truncated the file (Tier 3 blank window) |
| WebView2 automation | Coordinate/SendKeys probes often hang; unit tests + hand-check are the gate |
| Cross-compile | `CGO_ENABLED=0` from Linux/WSL; needs Go + go-winres |
| Monorepo upload history (resolved) | During the 1.0.3 tinkerVault upload, source briefly drifted ahead of docs (early v2 Tier 1 code merged in before a version bump), and a follow-up `winres.json` fix was needed for the `RT_VERSION` block specifically (not just the cosmetic `RT_MANIFEST` field). Both closed and verified on real hardware before this v2.0.0 upload began; see PROJECT_SUMMARY.md's incident section for the full writeup |

---

## Verify

```bash
go test ./internal/... -count=1
bash build-process/verify-build.sh   # uses VERSION from file, not a hardcoded string
# On Windows (optional):
powershell.exe -File build-process/verify-windows.ps1
```

---

## Debug

| Env | Effect |
|-----|--------|
| `DOSSIER_DEBUG=1` | WebView2 DevTools |
| `DOSSIER_AUTO_OPEN=1` | Force-open last dossier (automation; independent of Settings) |

Logs: `%LOCALAPPDATA%\Dossier\dossier.log` (GUI binary has no console; use `Dossier-console.exe` for stderr).
