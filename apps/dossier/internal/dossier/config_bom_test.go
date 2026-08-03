package dossier

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_stripsUTF8BOM(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("APPDATA", tmp)
	dir := filepath.Join(tmp, AppDataRelPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	// JSON string with escaped backslashes → unmarshals to C:\Work\D
	body := append([]byte{0xEF, 0xBB, 0xBF}, []byte(`{"lastDossierPath":"C:\\Work\\D","settings":{"autoOpenLast":true}}`)...)
	if err := os.WriteFile(filepath.Join(dir, ConfigFileName), body, 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	want := `C:\Work\D`
	if c.LastDossierPath != want {
		t.Fatalf("path=%q want %q", c.LastDossierPath, want)
	}
	if !c.Settings.AutoOpenLastOrDefault() {
		t.Fatal("autoOpenLast not true")
	}
}
