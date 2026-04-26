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
  --refresh 1s
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--store` | `.conductor/tasks.json` | Path to the Conductor JSON task store |
| `--registry` | `config/conductor.capabilities.yaml` | Path to the capability config YAML |
| `--log` | *(empty)* | Path to a log file to tail (log tab is inactive without this) |
| `--refresh` | `2s` | How often the task store and log file are polled for changes |

---

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `1` / `2` / `3` / `4` | Switch to Tasks / Workflows / Registry / Logs tab |
| `Tab` / `Shift+Tab` | Cycle through tabs |
| `↑` / `k` | Move selection up |
| `↓` / `j` | Move selection down |
| `Enter` / `Space` | Toggle detail pane (Tasks, Workflows, Registry) |
| `Esc` | Close detail pane |
| `r` | Force refresh data from disk |
| `/` or `f` | Open filter input (Logs tab) |
| `c` | Clear accumulated log lines (Logs tab) |
| `q` / `Ctrl+C` | Quit |

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
├── main.go                  # CLI entry point, flag parsing
├── internal/
│   ├── client/
│   │   ├── types.go         # Go types mirroring Conductor Engine data models
│   │   ├── reader.go        # StoreReader, RegistryReader, LogTailer
│   │   └── reader_test.go
│   └── ui/
│       ├── app.go           # Root Bubble Tea model, tab routing, commands
│       ├── styles.go        # Lipgloss colour palette and style definitions
│       ├── tasks.go         # Task queue & status board view
│       ├── workflow.go      # Workflow execution trace view
│       ├── registry.go      # Capability registry browser view
│       ├── logs.go          # Log tail with filtering view
│       └── app_test.go
└── go.mod
```

