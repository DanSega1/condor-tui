package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/DanSega1/condor-tui/internal/client"
)

// tasksModel is the live task queue / status board view.
type tasksModel struct {
	records    []client.TaskRecord
	cursor     int
	width      int
	height     int
	detail     viewport.Model
	showDetail bool
	err        error
}

func newTasksModel(width, height int) tasksModel {
	vp := viewport.New(width, height-headerHeight-footerHeight)
	return tasksModel{
		width:  width,
		height: height,
		detail: vp,
	}
}

const headerHeight = 4
const footerHeight = 2

func (m tasksModel) Init() tea.Cmd { return nil }

func (m tasksModel) Update(msg tea.Msg) (tasksModel, tea.Cmd) {
	switch msg := msg.(type) {
	case recordsLoadedMsg:
		m.records = msg.records
		m.err = msg.err
		if m.cursor >= len(m.records) && len(m.records) > 0 {
			m.cursor = len(m.records) - 1
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
			if m.cursor < len(m.records)-1 {
				m.cursor++
			}
			if m.showDetail {
				m.detail.SetContent(m.renderDetail())
			}
		case "g":
			m.cursor = 0
			if m.showDetail {
				m.detail.SetContent(m.renderDetail())
			}
		case "G":
			if len(m.records) > 0 {
				m.cursor = len(m.records) - 1
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
	case tea.MouseMsg:
		switch msg.Button {
		case tea.MouseButtonWheelUp:
			if m.cursor > 0 {
				m.cursor--
			}
		case tea.MouseButtonWheelDown:
			if m.cursor < len(m.records)-1 {
				m.cursor++
			}
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

func (m tasksModel) View() string {
	var b strings.Builder

	// Header.
	b.WriteString(stylePanelTitle.Render("Task Queue & Status Board"))
	b.WriteString("\n")

	if m.err != nil {
		b.WriteString(styleError.Render("Error: " + m.err.Error()))
		b.WriteString("\n")
		b.WriteString(styleHelp.Render("hint: run `cond run` to create tasks"))
		return b.String()
	}

	if len(m.records) == 0 {
		b.WriteString(styleDimmed.Render("No tasks found."))
		b.WriteString("\n")
		b.WriteString(styleHelp.Render("hint: run `cond run <task.yaml>` to create tasks"))
		b.WriteString("\n")
		b.WriteString(styleHelp.Render("or check if the store path is correct: --store <path>"))
		return b.String()
	}

	// Column widths.
	colStatus := 22
	colCap := 16
	colName := m.width - colStatus - colCap - 28
	if colName < 12 {
		colName = 12
	}

	// Table header.
	header := fmt.Sprintf("  %-*s  %-*s  %-*s  %s",
		colName, "NAME",
		colCap, "CAPABILITY",
		colStatus, "STATUS",
		"UPDATED",
	)
	b.WriteString(styleDimmed.Render(header))
	b.WriteString("\n")
	b.WriteString(styleDimmed.Render(strings.Repeat("─", m.width-2)))
	b.WriteString("\n")

	// Rows.
	maxRows := m.height - headerHeight - footerHeight - 4
	if maxRows < 1 {
		maxRows = 10
	}

	start := 0
	if m.cursor >= maxRows {
		start = m.cursor - maxRows + 1
	}
	end := start + maxRows
	if end > len(m.records) {
		end = len(m.records)
	}

	for i := start; i < end; i++ {
		r := m.records[i]
		displayName := r.Name
		if displayName == "" {
			displayName = r.TaskID
		}
		name := truncate(displayName, colName)
		cap_ := truncate(r.Capability, colCap)
		updated := r.UpdatedAt.Format("15:04:05")

		row := fmt.Sprintf("  %-*s  %-*s  %-*s  %s",
			colName, name,
			colCap, cap_,
			colStatus, statusBadge(string(r.Status)),
			updated,
		)

		if i == m.cursor {
			b.WriteString(styleSelected.Render(row))
		} else {
			b.WriteString(row)
		}
		b.WriteString("\n")
	}

	// Scroll indicator.
	if len(m.records) > maxRows {
		b.WriteString(styleDimmed.Render(fmt.Sprintf(
			"  … %d of %d tasks (↑↓ to scroll)",
			m.cursor+1, len(m.records),
		)))
		b.WriteString("\n")
	}

	// Detail pane.
	if m.showDetail && len(m.records) > 0 {
		b.WriteString("\n")
		b.WriteString(styleBorder.Render(m.detail.View()))
	}

	// Footer.
	b.WriteString("\n")
	b.WriteString(styleHelp.Render("↑/↓ navigate  enter toggle detail  r refresh  q quit"))

	return b.String()
}

// renderDetail builds the detail view for the currently selected task.
func (m *tasksModel) renderDetail() string {
	if len(m.records) == 0 || m.cursor >= len(m.records) {
		return ""
	}
	r := m.records[m.cursor]
	var b strings.Builder

	kv := func(k, v string) {
		b.WriteString(styleKey.Render(k + ": "))
		b.WriteString(styleValue.Render(v))
		b.WriteString("\n")
	}

	kv("ID", r.TaskID)
	kv("Name", r.Name)
	kv("Capability", r.Capability)
	kv("Status", string(r.Status))

	// Highlight retry information prominently
	retryInfo := fmt.Sprintf("%d / %d", r.Attempt, r.MaxRetries+1)
	if r.Attempt > 1 {
		retryInfo += styleError.Render(" (retried)")
	}
	kv("Attempt", retryInfo)

	// Show retry eligibility for failed tasks
	if r.Status == client.StatusFailed && r.Attempt <= r.MaxRetries {
		b.WriteString(styleKey.Render("Retry Status: "))
		b.WriteString(styleBadgeRunning.Render("eligible for retry"))
		b.WriteString("\n")
	}

	if r.WorkflowID != nil {
		kv("Workflow", *r.WorkflowID)
	}
	kv("Created", r.CreatedAt.Format("2006-01-02 15:04:05 UTC"))
	kv("Updated", r.UpdatedAt.Format("2006-01-02 15:04:05 UTC"))

	if r.Result != nil {
		b.WriteString("\n")
		b.WriteString(styleKey.Render("Result:") + "\n")
		if r.Result.Error != nil {
			b.WriteString(styleError.Render("  error: "+*r.Result.Error) + "\n")
		}
		if r.Result.Output != nil {
			b.WriteString(styleDetail.Render(fmt.Sprintf("  output: %v", r.Result.Output)) + "\n")
		}
		if r.Result.StartedAt != nil {
			b.WriteString(styleDetail.Render("  started:   "+r.Result.StartedAt.Format("15:04:05")) + "\n")
		}
		if r.Result.CompletedAt != nil {
			b.WriteString(styleDetail.Render("  completed: "+r.Result.CompletedAt.Format("15:04:05")) + "\n")
		}
	}

	if len(r.AuditTrail) > 0 {
		b.WriteString("\n")
		b.WriteString(styleKey.Render("Audit Trail:") + "\n")
		for _, e := range r.AuditTrail {
			line := fmt.Sprintf("  %s  %s  %s",
				e.Timestamp.Format("15:04:05"),
				styleKey.Render(e.Actor),
				e.Action,
			)
			if e.FromStatus != nil && e.ToStatus != nil {
				line += fmt.Sprintf(" (%s → %s)", *e.FromStatus, *e.ToStatus)
			}
			b.WriteString(styleDetail.Render(line) + "\n")

			// Show audit metadata if present
			if len(e.Metadata) > 0 {
				for k, v := range e.Metadata {
					b.WriteString(styleDetail.Render(fmt.Sprintf("    %s: %v", k, v)) + "\n")
				}
			}
		}
	}

	return b.String()
}

// truncate shortens s to n runes, adding "…" if needed.
func truncate(s string, n int) string {
	if n <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n-1]) + "…"
}

// padRight pads s to width w.
func padRight(s string, w int) string {
	n := w - len([]rune(s))
	if n <= 0 {
		return s
	}
	return s + strings.Repeat(" ", n)
}
