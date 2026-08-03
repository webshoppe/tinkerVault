# Dossier packaging verify - Windows host (or powershell.exe from WSL).
# Checks: smoke, PE version resource, process launch with auto-open.
$ErrorActionPreference = "Stop"
$Root = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$Exe = Join-Path $Root "build\Dossier.exe"
$Smoke = Join-Path $Root "build\smoke.exe"
$VersionFile = Join-Path $Root "VERSION"
$ExpectVer = (Get-Content $VersionFile -Raw).Trim()

Write-Host "=== Dossier verify-windows ==="
Write-Host "root=$Root expect=$ExpectVer"

if (-not (Test-Path $Exe)) { throw "missing $Exe - run build-process/verify-build.sh first" }
if (-not (Test-Path $Smoke)) { throw "missing $Smoke" }

Write-Host "--- smoke ---"
& $Smoke
if ($LASTEXITCODE -ne 0) { throw "smoke failed: $LASTEXITCODE" }

Write-Host "--- PE version resource ---"
$vi = [System.Diagnostics.FileVersionInfo]::GetVersionInfo($Exe)
Write-Host ("FileVersion={0} ProductVersion={1} ProductName={2} FileDescription={3}" -f `
  $vi.FileVersion, $vi.ProductVersion, $vi.ProductName, $vi.FileDescription)
if ([string]::IsNullOrWhiteSpace($vi.FileVersion) -and [string]::IsNullOrWhiteSpace($vi.ProductVersion)) {
  throw "PE version resource blank - go-winres syso missing or not linked"
}
$pv = "$($vi.ProductVersion)"
$fv = "$($vi.FileVersion)"
# Accept exact product string or PE fixed prefix (e.g. 2.0.0 / 2.0.0.23)
if (-not ($pv -eq $ExpectVer -or $fv -eq $ExpectVer -or $pv.StartsWith($ExpectVer) -or $fv.StartsWith($ExpectVer))) {
  throw "version mismatch: PE=$pv/$fv expect=$ExpectVer"
}

Write-Host "--- launch GUI (DOSSIER_AUTO_OPEN) ---"
Get-Process -Name Dossier -ErrorAction SilentlyContinue | Stop-Process -Force -ErrorAction SilentlyContinue
Start-Sleep -Seconds 1
# cmd.exe so env is set for the child (Start-Process -Environment needs PS 7+)
$cmdLine = 'set DOSSIER_AUTO_OPEN=1& start "" "' + $Exe + '"'
Start-Process -FilePath "cmd.exe" -ArgumentList @("/c", $cmdLine) | Out-Null
$deadline = (Get-Date).AddSeconds(18)
$proc = $null
while ((Get-Date) -lt $deadline) {
  Start-Sleep -Milliseconds 400
  $proc = Get-Process -Name Dossier -ErrorAction SilentlyContinue | Select-Object -First 1
  if ($proc -and $proc.MainWindowHandle -ne 0) { break }
}
if (-not $proc) { throw "Dossier.exe did not start" }
Write-Host ("PROCESS_OK pid={0} title={1}" -f $proc.Id, $proc.MainWindowTitle)
Start-Sleep -Seconds 3

$shotDir = Join-Path $Root "build"
$code = @'
using System;
using System.Runtime.InteropServices;
using System.Drawing;
using System.Drawing.Imaging;
public class PackShot {
  [DllImport("user32.dll")] public static extern bool GetWindowRect(IntPtr hWnd, out RECT lpRect);
  [DllImport("user32.dll")] public static extern bool PrintWindow(IntPtr hWnd, IntPtr hdcBlt, int nFlags);
  [DllImport("user32.dll")] public static extern bool SetForegroundWindow(IntPtr hWnd);
  [StructLayout(LayoutKind.Sequential)] public struct RECT { public int Left, Top, Right, Bottom; }
  public static void Capture(IntPtr hwnd, string path) {
    SetForegroundWindow(hwnd);
    System.Threading.Thread.Sleep(200);
    RECT r; GetWindowRect(hwnd, out r);
    int w = r.Right - r.Left, h = r.Bottom - r.Top;
    using (var bmp = new Bitmap(w, h, PixelFormat.Format32bppArgb)) {
      using (var g = Graphics.FromImage(bmp)) {
        IntPtr hdc = g.GetHdc();
        PrintWindow(hwnd, hdc, 2);
        g.ReleaseHdc(hdc);
      }
      bmp.Save(path, ImageFormat.Png);
    }
  }
}
'@
Add-Type -TypeDefinition $code -ReferencedAssemblies System.Drawing
$shot = Join-Path $shotDir "packaging-shell-shot.png"
[PackShot]::Capture($proc.MainWindowHandle, $shot)
Write-Host "shot $shot size=$((Get-Item $shot).Length)"

# Icon resource: ExtractAssociatedIcon
try {
  $icon = [System.Drawing.Icon]::ExtractAssociatedIcon($Exe)
  if ($icon -eq $null) { throw "ExtractAssociatedIcon returned null" }
  $iconPath = Join-Path $shotDir "packaging-exe-icon.png"
  $icon.ToBitmap().Save($iconPath, [System.Drawing.Imaging.ImageFormat]::Png)
  Write-Host "ICON_OK extracted $iconPath size=$((Get-Item $iconPath).Length)"
} catch {
  Write-Host "ICON_WARN $($_.Exception.Message)"
}

Write-Host "VERIFY_WINDOWS_OK version=$ExpectVer"
