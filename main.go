// condor-tui is a standalone terminal UI for monitoring and operating a
// running Conductor Engine instance.
//
// Usage:
//
//	condor-tui [flags]
//
// Flags:
//
//	--store      path to the Conductor task store JSON file
//	             (default: .conductor/tasks.json)
//	--registry   path to the Conductor capabilities YAML file
//	             (default: config/conductor.capabilities.yaml)
//	--log        path to a log file to tail (optional)
//	--refresh    data refresh interval, e.g. 2s, 500ms (default: 2s)
//	--theme      UI colour theme (default: catppuccin-mocha)
//	             available: catppuccin-mocha, catppuccin-latte, dracula, nord, gruvbox
package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/DanSega1/condor-tui/internal/config"
	"github.com/DanSega1/condor-tui/internal/ui"
)

func main() {
	var (
		storePath    = flag.String("store", "", "path to .conductor/tasks.json")
		registryPath = flag.String("registry", "", "path to conductor.capabilities.yaml")
		logPath      = flag.String("log", "", "path to a log file to tail (optional)")
		refresh      = flag.Duration("refresh", 0, "data refresh interval (e.g. 2s, 500ms)")
		theme        = flag.String("theme", "", "colour theme: catppuccin-mocha, catppuccin-latte, dracula, nord, gruvbox")
	)
	flag.Parse()

	fileCfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not read config file: %v\n", err)
	}

	merged := config.Merge(fileCfg, config.Flags{
		Store:    *storePath,
		Registry: *registryPath,
		Log:      *logPath,
		Refresh:  *refresh,
		Theme:    *theme,
	})

	cfg := ui.AppConfig{
		StorePath:    merged.Store,
		RegistryPath: merged.Registry,
		LogPath:      merged.Log,
		RefreshRate:  merged.Refresh,
		Theme:        merged.Theme,
		Stats:        merged.Stats,
		Engine:       merged.Engine,
		Tools:        merged.Tools,
	}

	app := ui.New(cfg)
	p := tea.NewProgram(app, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
