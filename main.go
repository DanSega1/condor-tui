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
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/DanSega1/condor-tui/internal/ui"
)

func main() {
	var (
		storePath    = flag.String("store", "", "path to .conductor/tasks.json (default: .conductor/tasks.json)")
		registryPath = flag.String("registry", "", "path to conductor.capabilities.yaml (default: config/conductor.capabilities.yaml)")
		logPath      = flag.String("log", "", "path to a log file to tail (optional)")
		refresh      = flag.Duration("refresh", 2*time.Second, "data refresh interval")
	)
	flag.Parse()

	cfg := ui.AppConfig{
		StorePath:    *storePath,
		RegistryPath: *registryPath,
		LogPath:      *logPath,
		RefreshRate:  *refresh,
	}

	app := ui.New(cfg)
	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
