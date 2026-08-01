package app

import (
	"path/filepath"
	"testing"

	"github.com/webshoppe/dossier/internal/dossier"
)

func TestSetOpenWithPreference(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("APPDATA", tmp)
	api := NewAPI()
	st, err := api.SetOpenWithPreference("report.PDF", `C:\Apps\Viewer.exe`)
	if err != nil {
		t.Fatal(err)
	}
	ow, _ := st["openWith"].(map[string]interface{})
	// JSON may decode as map[string]string depending on path
	cfg, _ := dossier.LoadConfig()
	if cfg.Settings.OpenWith[".pdf"] != `C:\Apps\Viewer.exe` {
		t.Fatalf("openWith=%v", cfg.Settings.OpenWith)
	}
	if api.preferredAppFor("x.Pdf") != `C:\Apps\Viewer.exe` {
		t.Fatal("preferredAppFor")
	}
	_, err = api.SetOpenWithPreference(".pdf", "")
	if err != nil {
		t.Fatal(err)
	}
	cfg2, _ := dossier.LoadConfig()
	if len(cfg2.Settings.OpenWith) != 0 {
		t.Fatalf("expected cleared, got %v", cfg2.Settings.OpenWith)
	}
	_ = filepath.Base(tmp)
	_ = ow
}
