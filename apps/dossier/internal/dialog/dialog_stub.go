//go:build !windows

package dialog

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// Stub implementations for non-Windows builds (compile-check only).
// The product target is Windows; these exist so `go test` can run on Linux CI.

func PickFolder(title string) (string, error) {
	return "", fmt.Errorf("folder picker is only available on Windows (got %s)", runtime.GOOS)
}

func PickFiles(title string) ([]string, error) {
	return nil, fmt.Errorf("file picker is only available on Windows (got %s)", runtime.GOOS)
}

func OpenExternally(path string) error {
	if runtime.GOOS == "linux" {
		// Best-effort for local dev tooling
		if _, err := os.Stat(path); err != nil {
			return err
		}
		return exec.Command("xdg-open", path).Start()
	}
	return fmt.Errorf("open externally only fully supported on Windows")
}

func OpenURL(raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fmt.Errorf("empty url")
	}
	if runtime.GOOS == "linux" {
		return exec.Command("xdg-open", raw).Start()
	}
	return fmt.Errorf("open url only fully supported on Windows")
}

func OpenWithApp(path, appExe string) error {
	if strings.TrimSpace(appExe) == "" {
		return OpenExternally(path)
	}
	return fmt.Errorf("open with app only fully supported on Windows")
}

func PickExecutable(title string) (string, error) {
	return "", fmt.Errorf("pick executable only available on Windows (got %s)", runtime.GOOS)
}

func RevealInExplorer(path string) error {
	return fmt.Errorf("reveal in explorer only available on Windows")
}
