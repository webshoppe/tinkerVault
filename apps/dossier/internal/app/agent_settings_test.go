package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/webshoppe/dossier/internal/dossier"
)

// Host is the enable/hide gate: empty host disables the Ask panel even if port remains set.
// Port-only no longer auto-fills host (that made “clear host to hide” impossible).
func TestSaveAppSettings_hostIsGate(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("APPDATA", tmp)
	api := NewAPI()

	host := "127.0.0.1"
	port := 8765
	st, err := api.SaveAppSettings(SettingsPatch{AgentHost: &host, AgentPort: &port})
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := dossier.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Settings.AgentHost != "127.0.0.1" {
		t.Fatalf("host=%q", cfg.Settings.AgentHost)
	}
	if cfg.Settings.AgentPort != 8765 {
		t.Fatalf("port=%d", cfg.Settings.AgentPort)
	}
	if on, _ := st["agentConfigured"].(bool); !on {
		t.Fatal("agentConfigured false after save with host+port")
	}
	if !AgentConfigured(cfg) {
		t.Fatal("AgentConfigured false")
	}

	// Clear host alone (port still set) → panel hidden
	empty := ""
	st2, err := api.SaveAppSettings(SettingsPatch{AgentHost: &empty, AgentPort: &port})
	if err != nil {
		t.Fatal(err)
	}
	cfg2, _ := dossier.LoadConfig()
	if cfg2.Settings.AgentHost != "" {
		t.Fatalf("expected empty host after clear, got %q", cfg2.Settings.AgentHost)
	}
	if cfg2.Settings.AgentPort != 8765 {
		t.Fatalf("port should remain %d, got %d", 8765, cfg2.Settings.AgentPort)
	}
	if on, _ := st2["agentConfigured"].(bool); on {
		t.Fatal("agentConfigured true after clearing host")
	}
	if AgentConfigured(cfg2) {
		t.Fatal("AgentConfigured true after clearing host")
	}

	// Clear port too (fully off)
	zero := 0
	_, _ = api.SaveAppSettings(SettingsPatch{AgentHost: &empty, AgentPort: &zero})
	cfg3, _ := dossier.LoadConfig()
	if AgentConfigured(cfg3) {
		t.Fatal("expected disabled when port 0 and host empty")
	}
	_ = filepath.Join(tmp, "x")
	_ = os.Getenv("APPDATA")
}

func TestSaveAppSettings_portOnlyDoesNotAutofillHost(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("APPDATA", tmp)
	api := NewAPI()
	host := ""
	port := 8765
	st, err := api.SaveAppSettings(SettingsPatch{AgentHost: &host, AgentPort: &port})
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := dossier.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Settings.AgentHost != "" {
		t.Fatalf("host should stay empty without auto-fill, got %q", cfg.Settings.AgentHost)
	}
	if on, _ := st["agentConfigured"].(bool); on {
		t.Fatal("agentConfigured should be false when host empty")
	}
}

func TestSaveAppSettings_autoOpenLastDefaultOff(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("APPDATA", tmp)
	api := NewAPI()
	st := api.GetAppSettings()
	if on, _ := st["autoOpenLast"].(bool); on {
		t.Fatal("autoOpenLast should default false")
	}
	on := true
	st2, err := api.SaveAppSettings(SettingsPatch{AutoOpenLast: &on})
	if err != nil {
		t.Fatal(err)
	}
	if v, _ := st2["autoOpenLast"].(bool); !v {
		t.Fatal("autoOpenLast should be true after save")
	}
	cfg, err := dossier.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Settings.AutoOpenLastOrDefault() {
		t.Fatal("config AutoOpenLastOrDefault false")
	}
	off := false
	st3, err := api.SaveAppSettings(SettingsPatch{AutoOpenLast: &off})
	if err != nil {
		t.Fatal(err)
	}
	if v, _ := st3["autoOpenLast"].(bool); v {
		t.Fatal("autoOpenLast should be false after clear")
	}
}
