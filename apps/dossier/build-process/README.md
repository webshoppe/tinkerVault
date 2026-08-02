# build-process; Dossier verify scripts

Reusable checks for the packaging and regression loop. These are **not** the one-off `build/capture-*.ps1` screenshot helpers.

| Script | Host | What it does |
|--------|------|----------------|
| `verify-build.sh` | WSL/Linux | unit tests, resources, cross-compile, PE icon/version string presence |
| `verify-windows.ps1` | Windows | smoke.exe, launch GUI with auto-open, report version resource + process |

## Typical flow

```bash
# From WSL
bash build-process/verify-build.sh

# Then on Windows (or powershell.exe from WSL)
powershell.exe -NoProfile -ExecutionPolicy Bypass -File build-process/verify-windows.ps1
```

Exit code `0` means the automated checks passed. Hand-check still required for right-click menus and visual taskbar icon (see [releases/1.0.3/STATUS.md](../releases/1.0.3/STATUS.md)).
