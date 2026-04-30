// Package config handles loading condor-tui configuration from a YAML file
// and merging it with CLI-flag overrides.
//
// Config file location (first found wins):
//
//	$XDG_CONFIG_HOME/condor-tui/config.yaml
//	~/.config/condor-tui/config.yaml
package config

import (
	"errors"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// StatsConfig controls which system metrics are shown in the header.
type StatsConfig struct {
	CPU     bool `yaml:"cpu"`
	Memory  bool `yaml:"memory"`
	Network bool `yaml:"network"`
}

// EngineConfig describes the Conductor Engine connection.
type EngineConfig struct {
	// URL is the remote engine endpoint (empty = local file-based access).
	URL string `yaml:"url"`
	// Kind is a display hint: "local", "remote", "docker", "kubernetes".
	// Auto-detected when empty.
	Kind string `yaml:"kind"`
}

// ToolsConfig lists external executables used to open/diff files.
type ToolsConfig struct {
	// Editor is the program used to open files (e.g. "nvim", "code").
	Editor string `yaml:"editor"`
	// Diff is the program used to diff task input vs result (e.g. "nvim -d", "vimdiff", "code --diff").
	Diff string `yaml:"diff"`
	// Opener is a generic file/URL opener (e.g. "open" on macOS, "xdg-open" on Linux).
	Opener string `yaml:"opener"`
}

// File mirrors the YAML structure of the config file. All fields are optional.
type File struct {
	Store    string        `yaml:"store"`
	Registry string        `yaml:"registry"`
	Log      string        `yaml:"log"`
	Refresh  time.Duration `yaml:"refresh"`
	Theme    string        `yaml:"theme"`
	Stats    StatsConfig   `yaml:"stats"`
	Engine   EngineConfig  `yaml:"engine"`
	Tools    ToolsConfig   `yaml:"tools"`
}

// Flags holds values provided on the command line. Empty string / zero means
// "not set by the user" and the config-file value should be used instead.
type Flags struct {
	Store    string
	Registry string
	Log      string
	Refresh  time.Duration // zero = not set
	Theme    string
}

// Merged is the final resolved configuration after merging file + flags.
type Merged struct {
	Store    string
	Registry string
	Log      string
	Refresh  time.Duration
	Theme    string
	Stats    StatsConfig
	Engine   EngineConfig
	Tools    ToolsConfig
}

// Load reads the config file from the XDG config directory. If the file does
// not exist, an empty File and nil error are returned — a missing config file
// is not an error condition.
func Load() (File, error) {
	path, err := resolvePath()
	if err != nil {
		return File{}, err
	}

	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return File{}, nil
	}
	if err != nil {
		return File{}, err
	}

	var f File
	if err := yaml.Unmarshal(data, &f); err != nil {
		return File{}, err
	}
	return f, nil
}

// Merge combines the file config with CLI flag overrides. Flags always win
// over the file; the file wins over built-in defaults.
func Merge(f File, flags Flags) Merged {
	m := Merged{
		Store:    f.Store,
		Registry: f.Registry,
		Log:      f.Log,
		Refresh:  f.Refresh,
		Theme:    f.Theme,
		Stats:    f.Stats,
		Engine:   f.Engine,
		Tools:    f.Tools,
	}

	if flags.Store != "" {
		m.Store = flags.Store
	}
	if flags.Registry != "" {
		m.Registry = flags.Registry
	}
	if flags.Log != "" {
		m.Log = flags.Log
	}
	if flags.Refresh != 0 {
		m.Refresh = flags.Refresh
	}
	if flags.Theme != "" {
		m.Theme = flags.Theme
	}

	// Apply built-in defaults for anything still unset.
	if m.Refresh == 0 {
		m.Refresh = 2 * time.Second
	}
	if m.Theme == "" {
		m.Theme = "catppuccin-mocha"
	}
	if m.Tools.Editor == "" {
		m.Tools.Editor = defaultEditor()
	}
	if m.Tools.Opener == "" {
		m.Tools.Opener = defaultOpener()
	}

	return m
}

// defaultEditor returns a sensible editor based on environment variables.
func defaultEditor() string {
	if e := os.Getenv("VISUAL"); e != "" {
		return e
	}
	if e := os.Getenv("EDITOR"); e != "" {
		return e
	}
	return "vi"
}

// defaultOpener returns the platform file opener.
func defaultOpener() string {
	if _, err := os.Stat("/usr/bin/xdg-open"); err == nil {
		return "xdg-open"
	}
	return "open" // macOS default
}

// resolvePath returns the path of the config file based on XDG conventions.
func resolvePath() (string, error) {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "condor-tui", "config.yaml"), nil
}

// Path returns the resolved config file path (for display / diagnostics).
func Path() string {
	p, _ := resolvePath()
	return p
}
