package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DanSega1/condor-tui/internal/config"
)

func TestLoad_missing_file(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	f, err := config.Load()
	if err != nil {
		t.Fatalf("expected nil error for missing config, got %v", err)
	}
	if f.Theme != "" || f.Store != "" {
		t.Errorf("expected empty File for missing config, got %+v", f)
	}
}

func TestLoad_parses_yaml(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfgDir := filepath.Join(dir, "condor-tui")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	yaml := `
store: /tmp/tasks.json
registry: /tmp/caps.yaml
log: /tmp/conductor.log
refresh: 5s
theme: dracula
`
	if err := os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(yaml), 0o644); err != nil {
		t.Fatal(err)
	}

	f, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.Store != "/tmp/tasks.json" {
		t.Errorf("store: got %q want /tmp/tasks.json", f.Store)
	}
	if f.Theme != "dracula" {
		t.Errorf("theme: got %q want dracula", f.Theme)
	}
	if f.Refresh != 5*time.Second {
		t.Errorf("refresh: got %v want 5s", f.Refresh)
	}
}

func TestMerge_flags_override_file(t *testing.T) {
	f := config.File{Store: "/file/store", Theme: "nord", Refresh: 3 * time.Second}
	flags := config.Flags{Store: "/flag/store", Theme: "dracula"}

	m := config.Merge(f, flags)

	if m.Store != "/flag/store" {
		t.Errorf("store: got %q want /flag/store", m.Store)
	}
	if m.Theme != "dracula" {
		t.Errorf("theme: got %q want dracula", m.Theme)
	}
	// Refresh not overridden by flags → should come from file.
	if m.Refresh != 3*time.Second {
		t.Errorf("refresh: got %v want 3s", m.Refresh)
	}
}

func TestMerge_defaults_applied(t *testing.T) {
	m := config.Merge(config.File{}, config.Flags{})
	if m.Refresh != 2*time.Second {
		t.Errorf("default refresh: got %v want 2s", m.Refresh)
	}
	if m.Theme != "catppuccin-mocha" {
		t.Errorf("default theme: got %q want catppuccin-mocha", m.Theme)
	}
}
