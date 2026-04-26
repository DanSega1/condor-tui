package ui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/DanSega1/condor-tui/internal/client"
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
}

// App is the root Bubble Tea model.
type App struct {
	cfg      AppConfig
	store    *client.StoreReader
	registry *client.RegistryReader
	tailer   *client.LogTailer

	activeTab tab
	width     int
	height    int

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
	return &App{
		cfg:      cfg,
		store:    client.NewStoreReader(cfg.StorePath),
		registry: client.NewRegistryReader(cfg.RegistryPath),
		tailer:   client.NewLogTailer(cfg.LogPath),
	}
}

// Init starts the tick and fires the initial data loads.
func (a *App) Init() tea.Cmd {
	return tea.Batch(
		tickEvery(a.cfg.RefreshRate),
		a.loadRecords(),
		a.loadCapabilities(),
		a.loadLogLines(),
	)
}

// Update is the central Bubble Tea update handler.
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.tasks = newTasksModel(msg.Width, msg.Height)
		a.workflow = newWorkflowModel(msg.Width, msg.Height)
		a.reg = newRegistryModel(msg.Width, msg.Height)
		a.logs = newLogsModel(msg.Width, msg.Height)
		// Re-broadcast to sub-models.
		a.dispatchToAll(msg)
		return a, nil

	case tea.KeyMsg:
		// Global shortcuts always take precedence.
		switch msg.String() {
		case "q", "ctrl+c":
			return a, tea.Quit
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
		case "tab":
			a.activeTab = (a.activeTab + 1) % tabCount
			return a, nil
		case "shift+tab":
			a.activeTab = (a.activeTab - 1 + tabCount) % tabCount
			return a, nil
		case "r":
			return a, tea.Batch(
				a.loadRecords(),
				a.loadCapabilities(),
				a.loadLogLines(),
			)
		}

	case tickMsg:
		return a, tea.Batch(
			tickEvery(a.cfg.RefreshRate),
			a.loadRecords(),
			a.loadLogLines(),
		)

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

// View renders the full TUI screen.
func (a *App) View() string {
	if a.width == 0 {
		return "Loading…"
	}

	tabBar := a.renderTabBar()
	separator := lipgloss.NewStyle().
		Foreground(colorSurface).
		Render(fmt.Sprintf("  %s\n", storeStatusLine(a.store)))

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

	return tabBar + "\n" + separator + panel
}

// renderTabBar draws the top tab bar.
func (a *App) renderTabBar() string {
	tabs := make([]string, tabCount)
	for i := 0; i < tabCount; i++ {
		if tab(i) == a.activeTab {
			tabs[i] = styleActiveTab.Render(tabNames[i])
		} else {
			tabs[i] = styleInactiveTab.Render(tabNames[i])
		}
	}
	title := lipgloss.NewStyle().
		Foreground(colorTeal).
		Bold(true).
		PaddingRight(2).
		Render("condor-tui")

	bar := lipgloss.JoinHorizontal(lipgloss.Top, append([]string{title}, tabs...)...)
	return styleTabBar.Width(a.width).Render(bar)
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
