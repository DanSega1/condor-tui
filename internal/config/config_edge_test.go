package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DanSega1/condor-tui/internal/config"
)

// Edge cases for config loading.

func TestLoad_MalformedYAML(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfgDir := filepath.Join(dir, "condor-tui")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	yaml := "theme: [invalid yaml structure"
	if err := os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(yaml), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := config.Load()
	if err == nil {
		t.Error("expected error for malformed YAML, got nil")
	}
}

func TestLoad_EmptyYAML(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfgDir := filepath.Join(dir, "condor-tui")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}

	f, err := config.Load()
	if err != nil {
		t.Fatalf("empty YAML should be valid: %v", err)
	}
	if f.Theme != "" || f.Refresh != 0 {
		t.Errorf("expected zero-valued File for empty YAML, got %+v", f)
	}
}

func TestLoad_PartialConfig(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfgDir := filepath.Join(dir, "condor-tui")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	yaml := `theme: dracula
# other fields omitted
`
	if err := os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(yaml), 0o644); err != nil {
		t.Fatal(err)
	}

	f, err := config.Load()
	if err != nil {
		t.Fatalf("partial config should be valid: %v", err)
	}
	if f.Theme != "dracula" {
		t.Errorf("theme: got %q want dracula", f.Theme)
	}
	if f.Store != "" {
		t.Errorf("store should be empty when not specified, got %q", f.Store)
	}
}

func TestLoad_InvalidDuration(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfgDir := filepath.Join(dir, "condor-tui")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	yaml := `refresh: not-a-duration
`
	if err := os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(yaml), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := config.Load()
	if err == nil {
		t.Error("expected error for invalid duration string, got nil")
	}
}

func TestLoad_NestedStructs(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfgDir := filepath.Join(dir, "condor-tui")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	yaml := `
stats:
  cpu: true
  memory: false
  network: true
engine:
  url: http://localhost:8080
  kind: docker
tools:
  editor: nvim
  diff: vimdiff
  opener: xdg-open
`
	if err := os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(yaml), 0o644); err != nil {
		t.Fatal(err)
	}

	f, err := config.Load()
	if err != nil {
		t.Fatalf("nested structs should parse: %v", err)
	}
	if !f.Stats.CPU {
		t.Error("stats.cpu should be true")
	}
	if f.Stats.Memory {
		t.Error("stats.memory should be false")
	}
	if f.Engine.URL != "http://localhost:8080" {
		t.Errorf("engine.url: got %q want http://localhost:8080", f.Engine.URL)
	}
	if f.Tools.Editor != "nvim" {
		t.Errorf("tools.editor: got %q want nvim", f.Tools.Editor)
	}
}

// Edge cases for Merge.

func TestMerge_AllFlagsProvided(t *testing.T) {
	f := config.File{
		Store:    "/file/store",
		Registry: "/file/registry",
		Log:      "/file/log",
		Refresh:  3 * time.Second,
		Theme:    "nord",
	}
	flags := config.Flags{
		Store:    "/flag/store",
		Registry: "/flag/registry",
		Log:      "/flag/log",
		Refresh:  5 * time.Second,
		Theme:    "dracula",
	}

	m := config.Merge(f, flags)

	// All flags should override file values.
	if m.Store != "/flag/store" {
		t.Errorf("store: got %q want /flag/store", m.Store)
	}
	if m.Registry != "/flag/registry" {
		t.Errorf("registry: got %q want /flag/registry", m.Registry)
	}
	if m.Log != "/flag/log" {
		t.Errorf("log: got %q want /flag/log", m.Log)
	}
	if m.Refresh != 5*time.Second {
		t.Errorf("refresh: got %v want 5s", m.Refresh)
	}
	if m.Theme != "dracula" {
		t.Errorf("theme: got %q want dracula", m.Theme)
	}
}

func TestMerge_NoFlagsProvided(t *testing.T) {
	f := config.File{
		Store:    "/file/store",
		Registry: "/file/registry",
		Log:      "/file/log",
		Refresh:  3 * time.Second,
		Theme:    "nord",
	}

	m := config.Merge(f, config.Flags{})

	// All file values should be preserved.
	if m.Store != "/file/store" {
		t.Errorf("store: got %q want /file/store", m.Store)
	}
	if m.Registry != "/file/registry" {
		t.Errorf("registry: got %q want /file/registry", m.Registry)
	}
	if m.Log != "/file/log" {
		t.Errorf("log: got %q want /file/log", m.Log)
	}
	if m.Refresh != 3*time.Second {
		t.Errorf("refresh: got %v want 3s", m.Refresh)
	}
	if m.Theme != "nord" {
		t.Errorf("theme: got %q want nord", m.Theme)
	}
}

func TestMerge_DefaultsApplied_EmptyFileAndFlags(t *testing.T) {
	m := config.Merge(config.File{}, config.Flags{})

	if m.Refresh != 2*time.Second {
		t.Errorf("default refresh: got %v want 2s", m.Refresh)
	}
	if m.Theme != "catppuccin-mocha" {
		t.Errorf("default theme: got %q want catppuccin-mocha", m.Theme)
	}
	// Editor should fall back to VISUAL, EDITOR, or "vi".
	if m.Tools.Editor == "" {
		t.Error("default editor should not be empty")
	}
	// Opener should have a default value.
	if m.Tools.Opener == "" {
		t.Error("default opener should not be empty")
	}
}

func TestMerge_NestedStructsMerged(t *testing.T) {
	f := config.File{}
	f.Stats.CPU = true
	f.Stats.Memory = false
	f.Engine.URL = "http://localhost:9000"
	f.Tools.Editor = "code"

	m := config.Merge(f, config.Flags{})

	if !m.Stats.CPU {
		t.Error("stats.cpu should be preserved from file")
	}
	if m.Stats.Memory {
		t.Error("stats.memory should be false from file")
	}
	if m.Engine.URL != "http://localhost:9000" {
		t.Errorf("engine.url: got %q want http://localhost:9000", m.Engine.URL)
	}
	if m.Tools.Editor != "code" {
		t.Errorf("tools.editor: got %q want code", m.Tools.Editor)
	}
}

func TestMerge_ZeroRefreshFlag(t *testing.T) {
	// Zero refresh in flags means "not set", so file value should be used.
	f := config.File{Refresh: 7 * time.Second}
	flags := config.Flags{Refresh: 0}

	m := config.Merge(f, flags)

	if m.Refresh != 7*time.Second {
		t.Errorf("refresh: got %v want 7s (flag=0 means unset)", m.Refresh)
	}
}

func TestMerge_EmptyStringFlags(t *testing.T) {
	// Empty string in flags means "not set", so file value should be used.
	f := config.File{
		Store:    "/file/store",
		Registry: "/file/registry",
		Theme:    "nord",
	}
	flags := config.Flags{
		Store:    "",
		Registry: "",
		Theme:    "",
	}

	m := config.Merge(f, flags)

	if m.Store != "/file/store" {
		t.Errorf("store: got %q want /file/store (empty flag means unset)", m.Store)
	}
	if m.Registry != "/file/registry" {
		t.Errorf("registry: got %q want /file/registry", m.Registry)
	}
	if m.Theme != "nord" {
		t.Errorf("theme: got %q want nord", m.Theme)
	}
}
