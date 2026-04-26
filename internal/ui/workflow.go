package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/DanSega1/condor-tui/internal/client"
)

// workflowModel groups tasks by workflow ID and renders a step-by-step trace.
type workflowModel struct {
	records    []client.TaskRecord
	workflows  []workflowGroup // ordered list of workflow groups
	cursor     int            // index into workflows
	width      int
	height     int
	detail     viewport.Model
	showDetail bool
	err        error
}

// workflowGroup collects all task records for a single workflow execution.
type workflowGroup struct {
	id     string // workflow_id or "(standalone)" for tasks without one
	tasks  []client.TaskRecord
	status string // derived overall status
}

func newWorkflowModel(width, height int) workflowModel {
	vp := viewport.New(width, height-headerHeight-footerHeight)
	return workflowModel{
		width:  width,
		height: height,
		detail: vp,
	}
}

func (m workflowModel) Init() tea.Cmd { return nil }

func (m workflowModel) Update(msg tea.Msg) (workflowModel, tea.Cmd) {
	switch msg := msg.(type) {
	case recordsLoadedMsg:
		m.records = msg.records
		m.err = msg.err
		m.workflows = buildWorkflowGroups(m.records)
		if m.cursor >= len(m.workflows) && len(m.workflows) > 0 {
			m.cursor = len(m.workflows) - 1
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
			if m.cursor < len(m.workflows)-1 {
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

func (m workflowModel) View() string {
	var b strings.Builder

	b.WriteString(stylePanelTitle.Render("Workflow Execution Trace"))
	b.WriteString("\n")

	if m.err != nil {
		b.WriteString(styleError.Render("Error: " + m.err.Error()))
		return b.String()
	}

	if len(m.workflows) == 0 {
		b.WriteString(styleDimmed.Render("No workflows found."))
		b.WriteString("\n")
		b.WriteString(styleHelp.Render("hint: run `cond workflow run <workflow.yaml>` to create workflows"))
		return b.String()
	}

	// Column widths.
	colID := 40
	colStatus := 22
	colSteps := 7
	colName := m.width - colID - colStatus - colSteps - 10
	if colName < 10 {
		colName = 10
	}

	header := fmt.Sprintf("  %-*s  %-*s  %-*s  %s",
		colID, "WORKFLOW ID",
		colStatus, "STATUS",
		colSteps, "STEPS",
		"GOAL",
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
	if end > len(m.workflows) {
		end = len(m.workflows)
	}

	for i := start; i < end; i++ {
		wf := m.workflows[i]
		id := truncate(wf.id, colID)
		goal := truncate(goalFromTasks(wf.tasks), colName)
		steps := fmt.Sprintf("%d", len(wf.tasks))

		row := fmt.Sprintf("  %-*s  %-*s  %-*s  %s",
			colID, id,
			colStatus, statusBadge(wf.status),
			colSteps, steps,
			goal,
		)

		if i == m.cursor {
			b.WriteString(styleSelected.Render(row))
		} else {
			b.WriteString(row)
		}
		b.WriteString("\n")
	}

	// Detail pane.
	if m.showDetail && len(m.workflows) > 0 {
		b.WriteString("\n")
		b.WriteString(styleBorder.Render(m.detail.View()))
	}

	b.WriteString("\n")
	b.WriteString(styleHelp.Render("↑/↓ navigate  enter toggle steps  r refresh  q quit"))

	return b.String()
}

// renderDetail renders a step-by-step trace of the selected workflow.
func (m *workflowModel) renderDetail() string {
	if len(m.workflows) == 0 || m.cursor >= len(m.workflows) {
		return ""
	}
	wf := m.workflows[m.cursor]
	var b strings.Builder

	b.WriteString(styleKey.Render("Workflow: "))
	b.WriteString(styleValue.Render(wf.id))
	b.WriteString("\n")
	b.WriteString(styleKey.Render("Overall: "))
	b.WriteString(statusBadge(wf.status))
	b.WriteString("\n\n")

	for i, t := range wf.tasks {
		step := fmt.Sprintf("Step %d: %s", i+1, t.Name)
		b.WriteString(styleKey.Render(step))
		b.WriteString("\n")
		b.WriteString(styleDetail.Render(fmt.Sprintf("  capability : %s", t.Capability)))
		b.WriteString("\n")
		b.WriteString(styleDetail.Render(fmt.Sprintf("  status     : %s", statusBadge(string(t.Status)))))
		b.WriteString("\n")
		b.WriteString(styleDetail.Render(fmt.Sprintf("  updated    : %s", t.UpdatedAt.Format("15:04:05"))))
		b.WriteString("\n")

		if t.Result != nil {
			if t.Result.Error != nil {
				b.WriteString(styleError.Render("  error: "+*t.Result.Error))
				b.WriteString("\n")
			} else if t.Result.Output != nil {
				b.WriteString(styleDetail.Render(fmt.Sprintf("  output: %v", t.Result.Output)))
				b.WriteString("\n")
			}
		}
		b.WriteString("\n")
	}

	return b.String()
}

// buildWorkflowGroups organises records into workflow groups.
// Records with no workflow_id are bucketed as standalone individual tasks
// keyed by their task_id.
func buildWorkflowGroups(records []client.TaskRecord) []workflowGroup {
	order := []string{}
	groups := map[string]*workflowGroup{}

	for _, r := range records {
		key := ""
		if r.WorkflowID != nil && *r.WorkflowID != "" {
			key = *r.WorkflowID
		} else {
			key = "(standalone) " + r.TaskID
		}

		g, ok := groups[key]
		if !ok {
			g = &workflowGroup{id: key}
			groups[key] = g
			order = append(order, key)
		}
		g.tasks = append(g.tasks, r)
	}

	result := make([]workflowGroup, 0, len(order))
	for _, key := range order {
		g := groups[key]
		g.status = deriveWorkflowStatus(g.tasks)
		result = append(result, *g)
	}
	return result
}

// deriveWorkflowStatus collapses a set of task statuses into one summary.
func deriveWorkflowStatus(tasks []client.TaskRecord) string {
	if len(tasks) == 0 {
		return "pending"
	}
	counts := map[string]int{}
	for _, t := range tasks {
		counts[string(t.Status)]++
	}
	switch {
	case counts["running"] > 0:
		return "running"
	case counts["failed"] > 0, counts["policy_denied"] > 0:
		return "failed"
	case counts["pending"] > 0, counts["awaiting_approval"] > 0:
		return "pending"
	case counts["completed"] == len(tasks):
		return "completed"
	default:
		return string(tasks[len(tasks)-1].Status)
	}
}

// goalFromTasks attempts to reconstruct a human-readable goal string.
func goalFromTasks(tasks []client.TaskRecord) string {
	if len(tasks) == 0 {
		return ""
	}
	// The goal is stored in task.metadata["goal"] by the orchestrator.
	for _, t := range tasks {
		if g, ok := t.Metadata["goal"].(string); ok && g != "" {
			return g
		}
	}
	// Fallback: list step names.
	names := make([]string, 0, len(tasks))
	for _, t := range tasks {
		names = append(names, t.Name)
	}
	return strings.Join(names, " → ")
}
