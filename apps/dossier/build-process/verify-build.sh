#!/usr/bin/env bash
# Dossier packaging verify; run from WSL/Linux project root.
# Version is always read from the VERSION file (never a hardcoded product string).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
export PATH="${HOME}/.local/go/bin:${HOME}/go/bin:${PATH}"

VERSION="$(cat VERSION 2>/dev/null | tr -d ' \t\r\n' || echo unknown)"
if [[ -z "$VERSION" || "$VERSION" == "unknown" ]]; then
  echo "FAIL: VERSION file missing or empty"
  exit 1
fi

echo "=== Dossier verify-build ==="
echo "version file: $VERSION"
echo "root: $ROOT"

echo "--- go test ---"
go test ./internal/... -count=1 -timeout 90s

echo "--- resources (go-winres) for GUI OriginalFilename ---"
python3 -c "import json; p='winres/winres.json'; d=json.load(open(p)); d['RT_VERSION']['#1']['0000']['info']['0409']['OriginalFilename']='Dossier.exe'; json.dump(d, open(p,'w'), indent=2); open(p,'a').write('\n')"
go-winres make --in winres/winres.json --out rsrc --arch amd64
test -f rsrc_windows_amd64.syso
file rsrc_windows_amd64.syso | grep -q 'COFF\|syso\|object' || file rsrc_windows_amd64.syso

echo "--- build GUI + console + smoke ---"
mkdir -p build
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-H windowsgui -s -w" -o build/Dossier.exe .
python3 -c "import json; p='winres/winres.json'; d=json.load(open(p)); d['RT_VERSION']['#1']['0000']['info']['0409']['OriginalFilename']='Dossier-console.exe'; json.dump(d, open(p,'w'), indent=2); open(p,'a').write('\n')"
go-winres make --in winres/winres.json --out rsrc --arch amd64
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o build/Dossier-console.exe .
python3 -c "import json; p='winres/winres.json'; d=json.load(open(p)); d['RT_VERSION']['#1']['0000']['info']['0409']['OriginalFilename']='Dossier.exe'; json.dump(d, open(p,'w'), indent=2); open(p,'a').write('\n')"
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o build/smoke.exe ./cmd/smoke
ls -lh build/Dossier.exe build/Dossier-console.exe build/smoke.exe
file build/Dossier.exe | grep -q 'PE32+'
file build/Dossier-console.exe | grep -q 'PE32+'

echo "--- PE resource strings (both exes) ---"
# Version is often UTF-16 in RT_VERSION; also check ASCII app constant.
check_version_in_exe() {
  local exe="$1"
  if ! command -v strings >/dev/null; then
    echo "WARN: strings not installed; skip string grep for $exe"
    return 0
  fi
  set +o pipefail
  strings "$exe" | grep -F "$VERSION" | head -5 || true
  strings -el "$exe" 2>/dev/null | grep -E "FileVersion|ProductVersion|${VERSION//./\\.}" | head -10 || true
  set -o pipefail
  if ! strings "$exe" | grep -Fq "$VERSION" \
    && ! strings -el "$exe" 2>/dev/null | grep -Fq "$VERSION"; then
    echo "FAIL: $VERSION not found in $exe strings (ascii or utf-16)"
    exit 1
  fi
}
check_version_in_exe build/Dossier.exe
check_version_in_exe build/Dossier-console.exe

echo "--- source gates ---"
grep -q "const Version = \"$VERSION\"" internal/app/version.go
grep -q 'PutAreDefaultContextMenusEnabled(true)' third_party/go-webview2/webview.go
grep -q 'PutAreDevToolsEnabled(options.Debug)' third_party/go-webview2/webview.go
grep -q 'note-copy' ui/index.html
grep -q 'Clear the host field entirely' ui/index.html
# Escape KEY_UP path must remain (sticky picker a11y)
grep -q 'COREWEBVIEW2_KEY_EVENT_KIND_KEY_UP' third_party/go-webview2/pkg/edge/chromium.go
# Ensure we do NOT auto-fill host on empty+port in SaveAppSettings
if grep -n 'AgentPort > 0 && strings.TrimSpace(cfg.Settings.AgentHost) == ""' internal/app/api.go | grep -v '//' ; then
  echo "FAIL: host auto-fill on empty still present in api.go"
  exit 1
fi

echo "VERIFY_BUILD_OK version=$VERSION"
