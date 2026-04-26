package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/DanSega1/condor-tui/internal/client"
)

// logsModel is the log tail + filter view.
type logsModel struct {
	lines     []client.LogLine // all accumulated lines
	filtered  []client.LogLine // lines matching the current filter
	filter    textinput.Model
	viewport  viewport.Model
	filtering bool // whether the filter input is focused
	width     int
	height    int
	err       error
}

func newLogsModel(width, height int) logsModel {
	ti := textinput.New()
	ti.Placeholder = "filter…"
	ti.CharLimit = 120
	ti.Width = width - 10

	vp := viewport.New(width, height-headerHeight-footerHeight-3)

	return logsModel{
		filter:  ti,
		viewport: vp,
		width:   width,
		height:  height,
	}
}

func (m logsModel) Init() tea.Cmd { return nil }

func (m logsModel) Update(msg tea.Msg) (logsModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case logLinesMsg:
		m.err = msg.err
		if msg.err == nil {
			m.lines = append(m.lines, msg.lines...)
			m.applyFilter()
		}

	case tea.KeyMsg:
		if m.filtering {
			switch msg.String() {
			case "esc", "enter":
				m.filtering = false
				m.filter.Blur()
				m.applyFilter()
			default:
				var cmd tea.Cmd
				m.filter, cmd = m.filter.Update(msg)
				cmds = append(cmds, cmd)
				m.applyFilter()
			}
		} else {
			switch msg.String() {
			case "/", "f":
				m.filtering = true
				m.filter.Focus()
			case "c":
				m.lines = nil
				m.filtered = nil
				m.filter.SetValue("")
				m.viewport.SetContent("")
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.filter.Width = msg.Width - 10
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - headerHeight - footerHeight - 3
	}

	var vpCmd tea.Cmd
	m.viewport, vpCmd = m.viewport.Update(msg)
	cmds = append(cmds, vpCmd)

	return m, tea.Batch(cmds...)
}

func (m logsModel) View() string {
	var b strings.Builder

	b.WriteString(stylePanelTitle.Render("Log Tail"))
	b.WriteString("\n")

	// Filter bar.
	filterLabel := styleKey.Render("filter: ")
	if m.filtering {
		b.WriteString(filterLabel + m.filter.View())
	} else {
		fv := m.filter.Value()
		if fv != "" {
			b.WriteString(filterLabel + styleValue.Render(fv))
		} else {
			b.WriteString(filterLabel + styleDimmed.Render("(none — press / to filter)"))
		}
	}
	b.WriteString("\n")

	if m.err != nil {
		b.WriteString(styleError.Render("Error reading log: " + m.err.Error()))
		b.WriteString("\n")
		b.WriteString(styleHelp.Render("hint: specify a log file with --log <path>"))
		return b.String()
	}

	b.WriteString(m.viewport.View())
	b.WriteString("\n")
	b.WriteString(styleHelp.Render("/ filter  c clear  ↑/↓ scroll  r refresh  q quit"))

	return b.String()
}

// applyFilter rebuilds the filtered list and refreshes the viewport.
func (m *logsModel) applyFilter() {
	q := strings.ToLower(m.filter.Value())
	if q == "" {
		m.filtered = m.lines
	} else {
		m.filtered = m.filtered[:0:0]
		for _, l := range m.lines {
			if strings.Contains(strings.ToLower(l.Text), q) {
				m.filtered = append(m.filtered, l)
			}
		}
	}

	// Render into the viewport.
	var sb strings.Builder
	maxLines := 2000
	start := 0
	if len(m.filtered) > maxLines {
		start = len(m.filtered) - maxLines
	}
	for _, l := range m.filtered[start:] {
		sb.WriteString(highlightFilter(l.Text, m.filter.Value()))
		sb.WriteString("\n")
	}
	m.viewport.SetContent(sb.String())
	m.viewport.GotoBottom()
}

// highlightFilter wraps occurrences of q in l with the highlight style.
func highlightFilter(line, q string) string {
	if q == "" {
		return styleDetail.Render(line)
	}
	lLower := strings.ToLower(line)
	qLower := strings.ToLower(q)
	var b strings.Builder
	cursor := 0
	for {
		i := strings.Index(lLower[cursor:], qLower)
		if i < 0 {
			b.WriteString(styleDetail.Render(line[cursor:]))
			break
		}
		b.WriteString(styleDetail.Render(line[cursor : cursor+i]))
		b.WriteString(styleSelected.Render(line[cursor+i : cursor+i+len(q)]))
		cursor += i + len(q)
	}
	return b.String()
}
