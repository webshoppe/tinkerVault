//go:build !windows

// This file allows non-Windows hosts to compile package-level tests and
// headless verification of the dossier package without WebView2.

package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "Dossier is a Windows desktop app.")
	fmt.Fprintln(os.Stderr, "Build with: GOOS=windows GOARCH=amd64 go build -ldflags=\"-H windowsgui\" -o build/Dossier.exe .")
	fmt.Fprintln(os.Stderr, "Or run: make build")
	os.Exit(1)
}
