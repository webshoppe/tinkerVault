#!/usr/bin/env bash
# Dossier packaging verify; run from WSL/Linux project root.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
export PATH="${HOME}/.local/go/bin:${HOME}/go/bin:${PATH}"

VERSION="$(cat VERSION 2>/dev/null || echo unknown)"
echo "=== Dossier verify-build ==="
echo "version file: $VERSION"
echo "root: $ROOT"

echo "--- go test ---"
go test ./internal/... -count=1 -timeout 90s

echo "--- resources (go-winres) ---"
if [[ ! -f rsrc_windows_amd64.syso ]]; then
  go-winres make --in winres/winres.json --out rsrc --arch amd64
fi
test -f rsrc_windows_amd64.syso
file rsrc_windows_amd64.syso | grep -q 'COFF\|syso\|object' || file rsrc_windows_amd64.syso

echo "--- build GUI exe ---"
mkdir -p build
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-H windowsgui -s -w" -o build/Dossier.exe .
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o build/smoke.exe ./cmd/smoke
ls -lh build/Dossier.exe build/smoke.exe
file build/Dossier.exe | grep -q 'PE32+'

echo "--- PE resource strings ---"
# Version is often UTF-16 in RT_VERSION; also check ASCII app constant.
if command -v strings >/dev/null; then
  # Avoid pipefail+head SIGPIPE under set -e
  set +o pipefail
  strings build/Dossier.exe | grep -F "$VERSION" | head -5 || true
  strings -el build/Dossier.exe 2>/dev/null | grep -E "FileVersion|ProductVersion|${VERSION//./\\.}" | head -10 || true
  set -o pipefail
  if ! strings build/Dossier.exe | grep -Fq "$VERSION" \
    && ! strings -el build/Dossier.exe 2>/dev/null | grep -Fq "$VERSION"; then
    echo "FAIL: $VERSION not found in exe strings (ascii or utf-16)"
    exit 1
  fi
fi

echo "--- source gates ---"
grep -q "const Version = \"$VERSION\"" internal/app/version.go
grep -q 'PutAreDefaultContextMenusEnabled(true)' third_party/go-webview2/webview.go
grep -q 'PutAreDevToolsEnabled(options.Debug)' third_party/go-webview2/webview.go
grep -q 'note-copy' ui/index.html
grep -q 'Clear the host field entirely' ui/index.html
# Ensure we do NOT auto-fill host on empty+port in SaveAppSettings
if grep -n 'AgentPort > 0 && strings.TrimSpace(cfg.Settings.AgentHost) == ""' internal/app/api.go | grep -v '//' ; then
  echo "FAIL: host auto-fill on empty still present in api.go"
  exit 1
fi

echo "VERIFY_BUILD_OK version=$VERSION"
