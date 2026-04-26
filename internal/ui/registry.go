package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/DanSega1/condor-tui/internal/client"
)

// registryModel is the capability registry browser view.
type registryModel struct {
	entries    []client.CapabilityEntry
	cursor     int
	width      int
	height     int
	detail     viewport.Model
	showDetail bool
	err        error
}

func newRegistryModel(width, height int) registryModel {
	vp := viewport.New(width, height-headerHeight-footerHeight)
	return registryModel{
		width:  width,
		height: height,
		detail: vp,
	}
}

func (m registryModel) Init() tea.Cmd { return nil }

func (m registryModel) Update(msg tea.Msg) (registryModel, tea.Cmd) {
	switch msg := msg.(type) {
	case capabilitiesLoadedMsg:
		m.entries = msg.entries
		m.err = msg.err
		if m.cursor >= len(m.entries) && len(m.entries) > 0 {
			m.cursor = len(m.entries) - 1
		}
		if m.showDetail {
			m.detail.SetContent(m.renderDetail())
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
			if m.showDetail {
				m.detail.SetContent(m.renderDetail())
			}
		case "down", "j":
			if m.cursor < len(m.entries)-1 {
				m.cursor++
			}
			if m.showDetail {
				m.detail.SetContent(m.renderDetail())
			}
		case "enter", " ":
			m.showDetail = !m.showDetail
			if m.showDetail {
				m.detail.SetContent(m.renderDetail())
			}
		case "esc":
			m.showDetail = false
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.detail.Width = msg.Width - 4
		m.detail.Height = msg.Height - headerHeight - footerHeight - 4
	}

	var cmd tea.Cmd
	if m.showDetail {
		m.detail, cmd = m.detail.Update(msg)
	}
	return m, cmd
}

func (m registryModel) View() string {
	var b strings.Builder

	b.WriteString(stylePanelTitle.Render("Capability Registry"))
	b.WriteString("\n")

	if m.err != nil {
		b.WriteString(styleError.Render("Error: " + m.err.Error()))
		return b.String()
	}

	if len(m.entries) == 0 {
		b.WriteString(styleDimmed.Render("No capabilities registered."))
		b.WriteString("\n")
		b.WriteString(styleHelp.Render("hint: check config/conductor.capabilities.yaml"))
		return b.String()
	}

	colName := 20
	colRisk := 10
	colDesc := m.width - colName - colRisk - 10
	if colDesc < 20 {
		colDesc = 20
	}

	header := fmt.Sprintf("  %-*s  %-*s  %s",
		colName, "NAME",
		colRisk, "RISK",
		"DESCRIPTION",
	)
	b.WriteString(styleDimmed.Render(header))
	b.WriteString("\n")
	b.WriteString(styleDimmed.Render(strings.Repeat("─", m.width-2)))
	b.WriteString("\n")

	maxRows := m.height - headerHeight - footerHeight - 4
	if maxRows < 1 {
		maxRows = 8
	}
	start := 0
	if m.cursor >= maxRows {
		start = m.cursor - maxRows + 1
	}
	end := start + maxRows
	if end > len(m.entries) {
		end = len(m.entries)
	}

	for i := start; i < end; i++ {
		e := m.entries[i]
		name := truncate(e.Name, colName)
		risk := truncate(riskBadge(e.RiskLevel), colRisk)
		desc := truncate(e.Description, colDesc)

		row := fmt.Sprintf("  %-*s  %-*s  %s",
			colName, name,
			colRisk, risk,
			desc,
		)

		if i == m.cursor {
			b.WriteString(styleSelected.Render(row))
		} else {
			b.WriteString(row)
		}
		b.WriteString("\n")
	}

	if m.showDetail && len(m.entries) > 0 {
		b.WriteString("\n")
		b.WriteString(styleBorder.Render(m.detail.View()))
	}

	b.WriteString("\n")
	b.WriteString(styleHelp.Render("↑/↓ navigate  enter toggle detail  r refresh  q quit"))

	return b.String()
}

func (m *registryModel) renderDetail() string {
	if len(m.entries) == 0 || m.cursor >= len(m.entries) {
		return ""
	}
	e := m.entries[m.cursor]
	var b strings.Builder

	kv := func(k, v string) {
		b.WriteString(styleKey.Render(k+": "))
		b.WriteString(styleValue.Render(v))
		b.WriteString("\n")
	}

	kv("Name", e.Name)
	kv("Risk", e.RiskLevel)
	if e.Description != "" {
		kv("Description", e.Description)
	}
	if e.ImportPath != "" {
		kv("Import Path", e.ImportPath)
	}
	if len(e.Tags) > 0 {
		kv("Tags", strings.Join(e.Tags, ", "))
	}

	return b.String()
}

// riskBadge returns a styled risk level string.
func riskBadge(level string) string {
	switch strings.ToLower(level) {
	case "low":
		return styleBadgeCompleted.Render("low")
	case "medium":
		return styleBadgeRunning.Render("medium")
	case "high":
		return styleBadgeFailed.Render("high")
	case "critical":
		return lipglossRender(colorRed, true, "critical")
	default:
		return styleDimmed.Render(level)
	}
}

// lipglossRender is a small helper to avoid importing lipgloss directly.
func lipglossRender(color interface{}, bold bool, text string) string {
	s := styleBadgeFailed
	_ = s
	return styleBadgeFailed.Render(text)
}
