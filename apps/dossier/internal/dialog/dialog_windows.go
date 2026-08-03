//go:build windows

package dialog

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

// PickFolder opens a native Windows folder browser dialog.
func PickFolder(title string) (string, error) {
	if title == "" {
		title = "Select folder"
	}
	// Escape single quotes for PowerShell string
	t := strings.ReplaceAll(title, "'", "''")
	script := fmt.Sprintf(`
Add-Type -AssemblyName System.Windows.Forms
$d = New-Object System.Windows.Forms.FolderBrowserDialog
$d.Description = '%s'
$d.ShowNewFolderButton = $true
$r = $d.ShowDialog()
if ($r -eq [System.Windows.Forms.DialogResult]::OK) { Write-Output $d.SelectedPath }
`, t)
	return runPS(script)
}

// PickFiles opens a multi-select file dialog. Returns selected paths.
func PickFiles(title string) ([]string, error) {
	if title == "" {
		title = "Select files"
	}
	t := strings.ReplaceAll(title, "'", "''")
	script := fmt.Sprintf(`
Add-Type -AssemblyName System.Windows.Forms
$d = New-Object System.Windows.Forms.OpenFileDialog
$d.Title = '%s'
$d.Multiselect = $true
$d.Filter = 'Documents|*.md;*.txt;*.pdf;*.docx;*.xlsx;*.odt;*.ods;*.markdown;*.text;*.log;*.csv;*.json|Office|*.docx;*.xlsx;*.odt;*.ods|All files|*.*'
$r = $d.ShowDialog()
if ($r -eq [System.Windows.Forms.DialogResult]::OK) {
  foreach ($f in $d.FileNames) { Write-Output $f }
}
`, t)
	out, err := runPS(script)
	if err != nil {
		return nil, err
	}
	if out == "" {
		return []string{}, nil
	}
	lines := strings.Split(out, "\n")
	var paths []string
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l != "" {
			paths = append(paths, l)
		}
	}
	return paths, nil
}

// OpenExternally opens a file with the default Windows application.
func OpenExternally(path string) error {
	path = filepath.Clean(path)
	if _, err := os.Stat(path); err != nil {
		return err
	}
	// Use cmd start so associated app launches
	cmd := exec.Command("cmd", "/c", "start", "", path)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd.Start()
}

// OpenURL opens an http(s) or other URL in the default browser/handler.
// Does not Stat the path (unlike OpenExternally).
func OpenURL(raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fmt.Errorf("empty url")
	}
	// Basic safety: only allow common external schemes
	lower := strings.ToLower(raw)
	if !strings.HasPrefix(lower, "http://") && !strings.HasPrefix(lower, "https://") &&
		!strings.HasPrefix(lower, "mailto:") {
		return fmt.Errorf("unsupported url scheme")
	}
	cmd := exec.Command("cmd", "/c", "start", "", raw)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd.Start()
}

// OpenWithApp opens path using a specific executable (Dossier preferred opener).
// appExe must be an absolute path to an .exe (or other launcher Windows accepts).
func OpenWithApp(path, appExe string) error {
	path = filepath.Clean(path)
	appExe = filepath.Clean(strings.TrimSpace(appExe))
	if appExe == "" {
		return OpenExternally(path)
	}
	if _, err := os.Stat(path); err != nil {
		return err
	}
	if _, err := os.Stat(appExe); err != nil {
		return fmt.Errorf("preferred app not found: %s", appExe)
	}
	cmd := exec.Command(appExe, path)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd.Start()
}

// PickExecutable opens a file dialog filtered to applications (.exe).
func PickExecutable(title string) (string, error) {
	if title == "" {
		title = "Choose application"
	}
	t := strings.ReplaceAll(title, "'", "''")
	script := fmt.Sprintf(`
Add-Type -AssemblyName System.Windows.Forms
$d = New-Object System.Windows.Forms.OpenFileDialog
$d.Title = '%s'
$d.Multiselect = $false
$d.Filter = 'Programs|*.exe;*.bat;*.cmd;*.com|All files|*.*'
$d.InitialDirectory = $env:ProgramFiles
$r = $d.ShowDialog()
if ($r -eq [System.Windows.Forms.DialogResult]::OK) { Write-Output $d.FileName }
`, t)
	return runPS(script)
}

// RevealInExplorer opens Explorer with the file selected.
func RevealInExplorer(path string) error {
	path = filepath.Clean(path)
	cmd := exec.Command("explorer", "/select,", path)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd.Start()
}

func runPS(script string) (string, error) {
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-STA", "-Command", script)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("dialog failed: %s", strings.TrimSpace(string(ee.Stderr)))
		}
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
