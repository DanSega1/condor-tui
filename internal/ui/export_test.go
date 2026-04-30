package ui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/DanSega1/condor-tui/internal/client"
)

// RecordsLoadedMsgForTest returns a recordsLoadedMsg for tests in package ui_test.
func RecordsLoadedMsgForTest(records []client.TaskRecord, err error) tea.Msg {
	return recordsLoadedMsg{records: records, err: err}
}

// CapabilitiesLoadedMsgForTest returns a capabilitiesLoadedMsg for tests in package ui_test.
func CapabilitiesLoadedMsgForTest(entries []client.CapabilityEntry, err error) tea.Msg {
	return capabilitiesLoadedMsg{entries: entries, err: err}
}

// LogLinesMsgForTest returns a logLinesMsg for tests in package ui_test.
func LogLinesMsgForTest(lines []client.LogLine, err error) tea.Msg {
	return logLinesMsg{lines: lines, err: err}
}

// RenderStatusBar returns a one-line summary of task counts by status.
// Exported for testing.
func RenderStatusBar(records []client.TaskRecord) string {
	return renderStatusBar(records)
}
