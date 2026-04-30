# condor-tui

A standalone terminal UI for monitoring and operating a running
[Conductor Engine](https://github.com/DanSega1/Conductor-Engine) instance.

**Stack:** Go + [Bubble Tea](https://github.com/charmbracelet/bubbletea)

---

## Features

| Tab | Description |
|-----|-------------|
| **[1] Tasks** | Live task queue and status board. Shows every `TaskRecord` from the task store, sorted by creation time. Select a task and press `enter` to expand a detail pane with full result, audit trail, and retry metadata. |
| **[2] Workflows** | Workflow execution trace. Tasks are grouped by `workflow_id` and shown as ordered steps. Select a workflow and press `enter` for a step-by-step result view. |
| **[3] Registry** | Capability registry browser. Lists built-in capabilities (`echo`, `filesystem`, `http`, `memory`) and any plugin capabilities declared in `conductor.capabilities.yaml`. |
| **[4] Logs** | Log tail with live filtering. Streams new lines as they are appended. Press `/` to open a filter input; matching text is highlighted in-place. |

---

## Installation

```bash
go install github.com/DanSega1/condor-tui@latest
```

Or build from source:

```bash
git clone https://github.com/DanSega1/condor-tui.git
cd condor-tui
go build -o condor-tui .
```

---

## Usage

Run from the same directory as your Conductor Engine project (the default paths
match Conductor's local-storage conventions):

```bash
# Defaults: reads .conductor/tasks.json and config/conductor.capabilities.yaml
condor-tui

# Override paths explicitly
condor-tui \
  --store   /path/to/.conductor/tasks.json \
  --registry /path/to/config/conductor.capabilities.yaml \
  --log     /path/to/conductor.log \
  --refresh 1s \
  --theme   dracula
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--store` | `.conductor/tasks.json` | Path to the Conductor JSON task store |
| `--registry` | `config/conductor.capabilities.yaml` | Path to the capability config YAML |
| `--log` | *(empty)* | Path to a log file to tail (log tab is inactive without this) |
| `--refresh` | `2s` | How often the task store and log file are polled for changes |
| `--theme` | `catppuccin-mocha` | Colour theme (see [Themes](#themes)) |

---

## Config File

Flags and advanced options can be persisted in a YAML config file:

```
~/.config/condor-tui/config.yaml          # default location
$XDG_CONFIG_HOME/condor-tui/config.yaml   # if XDG_CONFIG_HOME is set
```

Full example:

```yaml
store:    /home/dan/projects/myapp/.conductor/tasks.json
registry: /home/dan/projects/myapp/config/conductor.capabilities.yaml
log:      /home/dan/projects/myapp/conductor.log
refresh:  1s
theme:    nord

# System stats shown in the header (all off by default)
stats:
  cpu:     true
  memory:  true
  network: true

# Engine context displayed in header
engine:
  url:  http://localhost:8080
  kind: local   # local | remote | k8s | docker

# External tools launched from the TUI
tools:
  editor: nvim         # opened with 'e' or ':open'
  diff:   vimdiff      # opened with 'd' key (input vs result)
  opener: open         # used for URLs / files (:open config)
```

CLI flags always override the config file. A missing config file is not an error.

---

## Themes

| Name | Description |
|------|-------------|
| `catppuccin-mocha` | Dark — default |
| `catppuccin-latte` | Light |
| `dracula` | Dark purple |
| `nord` | Dark arctic blue |
| `gruvbox` | Dark warm retro |

Set via `--theme dracula` or in the config file, or live with `:theme dracula`.

---

## Keyboard Shortcuts

Press `?` inside the app for a live reference. Quick summary:

| Key | Action |
|-----|--------|
| `1` / `2` / `3` / `4` | Switch to Tasks / Workflows / Registry / Logs tab |
| `Tab` / `→` / `l` | Next tab |
| `Shift+Tab` / `←` / `h` | Previous tab |
| `↑` / `k` | Move selection up |
| `↓` / `j` | Move selection down |
| `g` | Jump to top of list |
| `G` | Jump to bottom of list |
| `Enter` / `Space` | Toggle detail pane (Tasks, Workflows, Registry) |
| `Esc` | Close detail pane |
| `r` | Force refresh data from disk |
| `e` | Open store file in configured editor |
| `d` | Diff selected task input vs result (Tasks tab) |
| `:` | Open command palette |
| `/` or `f` | Open filter input (Logs tab) |
| `c` | Clear accumulated log lines (Logs tab) |
| `?` | Show keyboard & mouse reference |
| `q` / `Ctrl+C` | Quit |

### Command Palette

Press `:` to open the vim/k9s-style command palette at the bottom of the screen.
`Tab` auto-completes command names. `Esc` cancels.

| Command | Description |
|---------|-------------|
| `theme <name>` | Switch colour theme live |
| `store <path>` | Load a different task store file |
| `refresh <dur>` | Change poll interval (e.g. `1s`, `500ms`) |
| `stats on\|off` | Toggle CPU/MEM/NET stats in the header |
| `open [config]` | Open store (or config file) in editor |
| `quit` | Quit condor-tui |

### Mouse

| Action | Effect |
|--------|--------|
| Click tab bar | Switch to that tab |
| Scroll wheel | Move list cursor / scroll log viewport |

---

## Header

The header area (inspired by k9s) shows:

```
  Engine: local (http://localhost:8080)       __             __       _
  Store:  .conductor/tasks.json            _______  ___  ___/ /__  ________/ /___
  Version: 0.1.0                          / __/ _ \/ _ \/ _  / _ \/ __/___/ __/
  CPU: 12%  MEM: 34%  NET: ↑2k ↓8k       \__/\___/_//_/\_,_/\___/_/      \__/
```

Stats panel is hidden by default — enable individual metrics in the config file.

---

## Why a Separate Binary

- **Conductor Engine is a Python library and CLI tool** — the TUI has no reason
  to share the runtime process.
- **Go compiles to a single static binary** with no interpreter dependency,
  making it easy to distribute alongside the Python package.
- **Bubble Tea** is purpose-built for this kind of interactive terminal work.

---

## Development

```bash
# Run tests
go test ./...

# Build
go build -o condor-tui .

# Lint (requires golangci-lint)
golangci-lint run
```

### Project Layout

```
condor-tui/
├── main.go                  # CLI entry point, flag parsing, config load
├── internal/
│   ├── config/
│   │   ├── config.go        # YAML config file loading + flag merge
│   │   └── config_test.go
│   ├── sysinfo/
│   │   └── sysinfo.go       # CPU/MEM/NET stats via gopsutil/v3
│   ├── client/
│   │   ├── types.go         # Go types mirroring Conductor Engine data models
│   │   ├── reader.go        # StoreReader, RegistryReader, LogTailer
│   │   └── reader_test.go
│   └── ui/
│       ├── app.go           # Root Bubble Tea model, tab routing, mouse, commands
│       ├── header.go        # k9s-style ASCII logo + info panel
│       ├── cmdpalette.go    # ':' command palette with tab-completion
│       ├── tools.go         # External tool helpers (editor, diff, opener)
│       ├── theme.go         # Theme struct + 5 built-in themes
│       ├── styles.go        # Lipgloss style vars (theme-driven)
│       ├── tasks.go         # Task queue & status board view
│       ├── workflow.go      # Workflow execution trace view
│       ├── registry.go      # Capability registry browser view
│       ├── logs.go          # Log tail with filtering view
│       └── app_test.go
└── go.mod
```
