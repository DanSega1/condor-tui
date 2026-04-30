package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/DanSega1/condor-tui/internal/client"
	"github.com/DanSega1/condor-tui/internal/config"
	"github.com/DanSega1/condor-tui/internal/sysinfo"
)

// tab represents the currently active panel.
type tab int

const (
	tabTasks tab = iota
	tabWorkflow
	tabRegistry
	tabLogs

	tabCount = 4
)

var tabNames = [tabCount]string{
	"[1] Tasks",
	"[2] Workflows",
	"[3] Registry",
	"[4] Logs",
}

// recordsLoadedMsg is sent when the task store is refreshed.
type recordsLoadedMsg struct {
	records []client.TaskRecord
	err     error
}

// capabilitiesLoadedMsg is sent when the capability registry is refreshed.
type capabilitiesLoadedMsg struct {
	entries []client.CapabilityEntry
	err     error
}

// logLinesMsg carries new log lines from the tailer.
type logLinesMsg struct {
	lines []client.LogLine
	err   error
}

// tickMsg is the periodic refresh tick.
type tickMsg time.Time

// AppConfig holds runtime configuration provided via CLI flags.
type AppConfig struct {
	StorePath    string
	RegistryPath string
	LogPath      string
	RefreshRate  time.Duration
	Theme        string
	Stats        config.StatsConfig
	Engine       config.EngineConfig
	Tools        config.ToolsConfig
}

// App is the root Bubble Tea model.
type App struct {
	cfg      AppConfig
	store    *client.StoreReader
	registry *client.RegistryReader
	tailer   *client.LogTailer

	activeTab  tab
	width      int
	height     int
	showHelp   bool
	tabOffsets []int // x-offset of each tab in the tab bar (for mouse clicks)

	sysStats   sysinfo.Stats
	palette    cmdPalette
	toolStatus string // brief message after a tool returns

	tasks    tasksModel
	workflow workflowModel
	reg      registryModel
	logs     logsModel
}

// New constructs a fully initialised App.
func New(cfg AppConfig) *App {
	if cfg.RefreshRate == 0 {
		cfg.RefreshRate = 2 * time.Second
	}
	applyTheme(ThemeByName(cfg.Theme))
	return &App{
		cfg:      cfg,
		store:    client.NewStoreReader(cfg.StorePath),
		registry: client.NewRegistryReader(cfg.RegistryPath),
		tailer:   client.NewLogTailer(cfg.LogPath),
		palette:  newCmdPalette(80),
	}
}

// Init starts the tick and fires the initial data loads.
func (a *App) Init() tea.Cmd {
	cmds := []tea.Cmd{
		tickEvery(a.cfg.RefreshRate),
		a.loadRecords(),
		a.loadCapabilities(),
		a.loadLogLines(),
	}
	if a.cfg.Stats.CPU || a.cfg.Stats.Memory || a.cfg.Stats.Network {
		cmds = append(cmds, func() tea.Msg { return sysInfoMsg{sysinfo.Collect(a.cfg.Stats.CPU, a.cfg.Stats.Memory, a.cfg.Stats.Network)} })
	}
	return tea.Batch(cmds...)
}

// Update is the central Bubble Tea update handler.
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.palette.width = msg.Width
		a.palette.input.Width = msg.Width - 4
		a.tasks = newTasksModel(msg.Width, a.contentHeight())
		a.workflow = newWorkflowModel(msg.Width, a.contentHeight())
		a.reg = newRegistryModel(msg.Width, a.contentHeight())
		a.logs = newLogsModel(msg.Width, a.contentHeight())
		a.dispatchToAll(msg)
		return a, nil

	case sysInfoMsg:
		a.sysStats = msg.stats
		return a, nil

	case toolDoneMsg:
		a.toolStatus = msg.msg
		return a, nil

	case toolErrMsg:
		a.toolStatus = styleError.Render("tool error: " + msg.err.Error())
		return a, nil

	case tea.KeyMsg:
		// Command palette intercepts all keys when active.
		if a.palette.active {
			var cmd tea.Cmd
			a.palette, cmd = a.palette.Update(msg)
			if a.palette.lastCmd != "" {
				return a, a.applyPaletteCmd(a.palette.lastCmd)
			}
			return a, cmd
		}

		// Help overlay.
		if a.showHelp {
			a.showHelp = false
			return a, nil
		}

		// Global shortcuts.
		switch msg.String() {
		case "q", "ctrl+c":
			return a, tea.Quit
		case ":":
			a.palette.active = true
			a.palette.ClearResult()
			a.toolStatus = ""
			return a, a.palette.input.Focus()
		case "?":
			a.showHelp = true
			return a, nil
		case "1":
			a.activeTab = tabTasks
			return a, nil
		case "2":
			a.activeTab = tabWorkflow
			return a, nil
		case "3":
			a.activeTab = tabRegistry
			return a, nil
		case "4":
			a.activeTab = tabLogs
			return a, nil
		case "tab", "right", "l":
			a.activeTab = (a.activeTab + 1) % tabCount
			return a, nil
		case "shift+tab", "left", "h":
			a.activeTab = (a.activeTab - 1 + tabCount) % tabCount
			return a, nil
		case "r":
			return a, tea.Batch(
				a.loadRecords(),
				a.loadCapabilities(),
				a.loadLogLines(),
			)
		case "e":
			return a, openInEditor(a.cfg.Tools.Editor, a.store.Path)
		case "d":
			if a.activeTab == tabTasks && len(a.tasks.records) > 0 {
				rec := a.tasks.records[a.tasks.cursor]
				var resMap map[string]any
				if rec.Result != nil && rec.Result.Output != nil {
					if m, ok := rec.Result.Output.(map[string]any); ok {
						resMap = m
					}
				}
				return a, diffTaskJSON(a.cfg.Tools.Diff, rec.Input, resMap)
			}
		}

	case tea.MouseMsg:
		return a.handleMouse(msg)

	case tickMsg:
		cmds := []tea.Cmd{
			tickEvery(a.cfg.RefreshRate),
			a.loadRecords(),
			a.loadLogLines(),
		}
		if a.cfg.Stats.CPU || a.cfg.Stats.Memory || a.cfg.Stats.Network {
			cmds = append(cmds, func() tea.Msg {
				return sysInfoMsg{sysinfo.Collect(a.cfg.Stats.CPU, a.cfg.Stats.Memory, a.cfg.Stats.Network)}
			})
		}
		return a, tea.Batch(cmds...)

	case recordsLoadedMsg:
		a.tasks, _ = a.tasks.Update(msg)
		a.workflow, _ = a.workflow.Update(msg)
		return a, nil

	case capabilitiesLoadedMsg:
		a.reg, _ = a.reg.Update(msg)
		return a, nil

	case logLinesMsg:
		a.logs, _ = a.logs.Update(msg)
		return a, nil
	}

	// Route other messages to the active tab.
	var cmd tea.Cmd
	switch a.activeTab {
	case tabTasks:
		a.tasks, cmd = a.tasks.Update(msg)
	case tabWorkflow:
		a.workflow, cmd = a.workflow.Update(msg)
	case tabRegistry:
		a.reg, cmd = a.reg.Update(msg)
	case tabLogs:
		a.logs, cmd = a.logs.Update(msg)
	}
	return a, cmd
}

// applyPaletteCmd interprets a confirmed palette command and returns a tea.Cmd.
func (a *App) applyPaletteCmd(raw string) tea.Cmd {
	defer func() { a.palette.lastCmd = "" }()

	parts := strings.Fields(raw)
	if len(parts) == 0 {
		return nil
	}

	switch parts[0] {
	case "quit":
		return tea.Quit
	case "theme":
		if len(parts) >= 2 {
			applyTheme(ThemeByName(parts[1]))
			a.cfg.Theme = parts[1]
		}
	case "store":
		if len(parts) >= 2 {
			a.cfg.StorePath = parts[1]
			a.store = client.NewStoreReader(parts[1])
			return a.loadRecords()
		}
	case "refresh":
		if len(parts) >= 2 {
			if d, err := time.ParseDuration(parts[1]); err == nil {
				a.cfg.RefreshRate = d
			}
		}
	case "stats":
		if len(parts) >= 2 {
			on := parts[1] == "on"
			a.cfg.Stats.CPU = on
			a.cfg.Stats.Memory = on
			a.cfg.Stats.Network = on
		}
	case "open":
		target := a.store.Path
		if len(parts) >= 2 && parts[1] == "config" {
			target = config.Path()
		}
		return openInEditor(a.cfg.Tools.Editor, target)
	case "help":
		a.showHelp = true
	}
	return nil
}

// contentHeight returns the usable panel height accounting for the header.
func (a *App) contentHeight() int {
	headerRows := len(logoLines) + 2 // logo lines + divider + tab bar
	h := a.height - headerRows
	if h < 10 {
		h = 10
	}
	return h
}

// handleMouse processes mouse events: tab bar clicks and scroll.
func (a *App) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	switch msg.Action { //nolint:exhaustive
	case tea.MouseActionPress:
		if msg.Button == tea.MouseButtonLeft && msg.Y == 0 {
			if t := a.tabAtX(msg.X); t >= 0 {
				a.activeTab = tab(t)
				return a, nil
			}
		}
	case tea.MouseActionRelease:
		// nothing
	}

	// Scroll wheel — forward to active tab.
	var cmd tea.Cmd
	switch a.activeTab {
	case tabTasks:
		a.tasks, cmd = a.tasks.Update(msg)
	case tabWorkflow:
		a.workflow, cmd = a.workflow.Update(msg)
	case tabRegistry:
		a.reg, cmd = a.reg.Update(msg)
	case tabLogs:
		a.logs, cmd = a.logs.Update(msg)
	}
	return a, cmd
}

// tabAtX returns the tab index for a given x position in the tab bar, or -1.
func (a *App) tabAtX(x int) int {
	for i, off := range a.tabOffsets {
		end := off
		if i+1 < len(a.tabOffsets) {
			end = a.tabOffsets[i+1]
		} else {
			end = a.width
		}
		if x >= off && x < end {
			return i
		}
	}
	return -1
}

// View renders the full TUI screen.
func (a *App) View() string {
	if a.width == 0 {
		return "Loading…"
	}

	if a.showHelp {
		return a.renderHelp()
	}

	header := renderHeader(a.width, a.cfg, a.sysStats)
	tabBar := a.renderTabBar()

	var panel string
	switch a.activeTab {
	case tabTasks:
		panel = a.tasks.View()
	case tabWorkflow:
		panel = a.workflow.View()
	case tabRegistry:
		panel = a.reg.View()
	case tabLogs:
		panel = a.logs.View()
	}

	bottom := a.palette.View()
	if bottom == "" && a.toolStatus != "" {
		bg := lipgloss.NewStyle().Background(colorSurface).Width(a.width)
		bottom = bg.Render("  " + a.toolStatus)
	}

	out := header + tabBar + "\n" + panel
	if bottom != "" {
		out += "\n" + bottom
	}
	return out
}

// renderTabBar draws the top tab bar and records each tab's x-offset for mouse hit-testing.
func (a *App) renderTabBar() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(colorTeal).
		Bold(true).
		PaddingRight(2)
	title := titleStyle.Render("condor-tui")
	titleWidth := lipgloss.Width(title)

	a.tabOffsets = make([]int, tabCount)
	x := titleWidth
	tabs := make([]string, tabCount)
	for i := 0; i < tabCount; i++ {
		a.tabOffsets[i] = x
		var rendered string
		if tab(i) == a.activeTab {
			rendered = styleActiveTab.Render(tabNames[i])
		} else {
			rendered = styleInactiveTab.Render(tabNames[i])
		}
		tabs[i] = rendered
		x += lipgloss.Width(rendered)
	}

	hint := lipgloss.NewStyle().
		Foreground(colorMuted).
		Background(colorSurface).
		PaddingLeft(2).
		Render("? help")

	tabsRow := lipgloss.JoinHorizontal(lipgloss.Top, append([]string{title}, tabs...)...)
	// Fill the gap between tabs and hint with the surface colour, then pin hint to the right.
	gap := a.width - lipgloss.Width(tabsRow) - lipgloss.Width(hint)
	if gap < 0 {
		gap = 0
	}
	filler := lipgloss.NewStyle().Background(colorSurface).Render(strings.Repeat(" ", gap))

	return styleTabBar.Render(tabsRow + filler + hint)
}

// renderHelp shows a keyboard shortcuts and mouse reference overlay.
func (a *App) renderHelp() string {
	box := styleBorder.
		Width(a.width - 4).
		BorderForeground(colorBlue)

	header := stylePanelTitle.Render("Keyboard & Mouse Reference")
	hint := styleHelp.Render("press any key to close")

	type row struct{ key, desc string }
	sections := []struct {
		title string
		rows  []row
	}{
		{"Navigation", []row{
			{"1 / 2 / 3 / 4", "Switch to Tasks / Workflows / Registry / Logs"},
			{"Tab / → / l", "Next tab"},
			{"Shift+Tab / ← / h", "Previous tab"},
			{"↑ / k", "Move selection up"},
			{"↓ / j", "Move selection down"},
			{"g", "Jump to top of list"},
			{"G", "Jump to bottom of list"},
		}},
		{"Actions", []row{
			{"Enter / Space", "Toggle detail pane"},
			{"Esc", "Close detail pane"},
			{"r", "Force refresh from disk"},
			{"e", "Open store file in editor"},
			{"d", "Diff selected task input vs result"},
			{":", "Open command palette"},
			{"/ or f", "Open filter input (Logs tab)"},
			{"c", "Clear log lines (Logs tab)"},
			{"?", "Show this help"},
			{"q / Ctrl+C", "Quit"},
		}},
		{"Command palette (:)", []row{
			{"theme <name>", "Switch theme (catppuccin-mocha, dracula, nord…)"},
			{"store <path>", "Load a different task store file"},
			{"refresh <dur>", "Change poll interval (e.g. 1s, 500ms)"},
			{"stats on|off", "Toggle CPU/MEM/NET stats in header"},
			{"open [config]", "Open store (or config file) in editor"},
			{"quit", "Quit condor-tui"},
		}},
		{"Mouse", []row{
			{"Click tab bar", "Switch to that tab"},
			{"Scroll wheel", "Move list cursor up / down"},
		}},
	}

	var sb strings.Builder
	sb.WriteString(header + "\n\n")
	for _, sec := range sections {
		sb.WriteString(styleKey.Render(sec.title) + "\n")
		for _, r := range sec.rows {
			sb.WriteString(fmt.Sprintf("  %-30s %s\n",
				styleMauve.Render(r.key),
				styleDetail.Render(r.desc),
			))
		}
		sb.WriteString("\n")
	}
	sb.WriteString(hint)

	return box.Render(sb.String())
}

// storeStatusLine returns a short string summarising the store state.
func storeStatusLine(r *client.StoreReader) string {
	if !r.Exists() {
		return styleError.Render("store not found: " + r.Path)
	}
	if mt, err := r.ModTime(); err == nil {
		return styleDimmed.Render("store: " + r.Path + "  last modified: " + mt.Format("15:04:05"))
	}
	return styleDimmed.Render("store: " + r.Path)
}

// dispatchToAll sends a message to every sub-model (used for WindowSizeMsg).
func (a *App) dispatchToAll(msg tea.Msg) {
	a.tasks, _ = a.tasks.Update(msg)
	a.workflow, _ = a.workflow.Update(msg)
	a.reg, _ = a.reg.Update(msg)
	a.logs, _ = a.logs.Update(msg)
}

// ─── commands ────────────────────────────────────────────────────────────────

func tickEvery(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (a *App) loadRecords() tea.Cmd {
	return func() tea.Msg {
		records, err := a.store.ReadAll()
		return recordsLoadedMsg{records: records, err: err}
	}
}

func (a *App) loadCapabilities() tea.Cmd {
	return func() tea.Msg {
		entries, err := a.registry.Read()
		return capabilitiesLoadedMsg{entries: entries, err: err}
	}
}

func (a *App) loadLogLines() tea.Cmd {
	if a.cfg.LogPath == "" {
		return nil
	}
	return func() tea.Msg {
		lines, err := a.tailer.Poll(200)
		return logLinesMsg{lines: lines, err: err}
	}
}
