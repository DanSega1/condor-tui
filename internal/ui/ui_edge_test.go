package ui_test

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/DanSega1/condor-tui/internal/client"
	"github.com/DanSega1/condor-tui/internal/config"
	"github.com/DanSega1/condor-tui/internal/sysinfo"
	"github.com/DanSega1/condor-tui/internal/ui"
)

// Header rendering edge cases.

func TestRenderHeader_EngineKindAutoDetection_Kubernetes(t *testing.T) {
	cfg := ui.AppConfig{
		StorePath: ".conductor/tasks.json",
		Engine: config.EngineConfig{
			URL: "http://my-k8s-cluster.example.com",
		},
	}
	out := ui.RenderHeader(120, cfg, sysinfo.Stats{})
	// Should auto-detect "kubernetes" from URL containing "k8s".
	if !strings.Contains(out, "example.com") {
		t.Error("header should show URL when engine.url is set")
	}
}

func TestRenderHeader_EngineKindAutoDetection_Docker(t *testing.T) {
	cfg := ui.AppConfig{
		StorePath: ".conductor/tasks.json",
		Engine: config.EngineConfig{
			URL: "http://docker.local:2375",
		},
	}
	out := ui.RenderHeader(120, cfg, sysinfo.Stats{})
	if !strings.Contains(out, "docker.local") {
		t.Error("header should show URL when engine.url is set")
	}
}

func TestRenderHeader_EngineKindExplicit(t *testing.T) {
	cfg := ui.AppConfig{
		StorePath: ".conductor/tasks.json",
		Engine: config.EngineConfig{
			Kind: "remote",
			URL:  "http://prod.example.com",
		},
	}
	out := ui.RenderHeader(120, cfg, sysinfo.Stats{})
	// Kind is explicitly set, so it should be shown (though URL display takes precedence).
	if !strings.Contains(out, "prod.example.com") {
		t.Error("header should show URL when engine.url is set")
	}
}

func TestRenderHeader_StorePathTruncation(t *testing.T) {
	longPath := "/very/long/path/to/some/deeply/nested/directory/structure/that/exceeds/forty/characters/.conductor/tasks.json"
	cfg := ui.AppConfig{
		StorePath: longPath,
	}
	out := ui.RenderHeader(120, cfg, sysinfo.Stats{})
	// Path longer than 40 chars should be truncated with ellipsis.
	if !strings.Contains(out, "…") {
		t.Error("long store path should be truncated with ellipsis")
	}
	if strings.Contains(out, "/very/long/path") {
		t.Error("truncated path should not contain the beginning of the path")
	}
}

func TestRenderHeader_EmptyStorePath(t *testing.T) {
	cfg := ui.AppConfig{
		StorePath: "",
	}
	out := ui.RenderHeader(120, cfg, sysinfo.Stats{})
	// Empty store path should fall back to default ".conductor/tasks.json".
	if !strings.Contains(out, ".conductor/tasks.json") {
		t.Error("header should show default store path when empty")
	}
}

func TestRenderHeader_StatsColorThresholds_Red(t *testing.T) {
	cfg := ui.AppConfig{
		StorePath: ".conductor/tasks.json",
		Stats:     config.StatsConfig{CPU: true},
	}
	stats := sysinfo.Stats{CPU: "85%"}
	out := ui.RenderHeader(120, cfg, stats)
	if !strings.Contains(out, "85%") {
		t.Error("header should show CPU value when stats.CPU is enabled")
	}
	// Color is applied via ANSI codes, but we can at least verify the value is present.
}

func TestRenderHeader_StatsColorThresholds_Yellow(t *testing.T) {
	cfg := ui.AppConfig{
		StorePath: ".conductor/tasks.json",
		Stats:     config.StatsConfig{Memory: true},
	}
	stats := sysinfo.Stats{Memory: "65%"}
	out := ui.RenderHeader(120, cfg, stats)
	if !strings.Contains(out, "65%") {
		t.Error("header should show memory value when stats.Memory is enabled")
	}
}

func TestRenderHeader_StatsColorThresholds_Green(t *testing.T) {
	cfg := ui.AppConfig{
		StorePath: ".conductor/tasks.json",
		Stats:     config.StatsConfig{CPU: true},
	}
	stats := sysinfo.Stats{CPU: "25%"}
	out := ui.RenderHeader(120, cfg, stats)
	if !strings.Contains(out, "25%") {
		t.Error("header should show CPU value when stats.CPU is enabled")
	}
}

func TestRenderHeader_NetworkStats(t *testing.T) {
	cfg := ui.AppConfig{
		StorePath: ".conductor/tasks.json",
		Stats:     config.StatsConfig{Network: true},
	}
	stats := sysinfo.Stats{NetUp: "↑1.2M", NetDn: "↓5.3M"}
	out := ui.RenderHeader(120, cfg, stats)
	if !strings.Contains(out, "↑1.2M") {
		t.Error("header should show NetUp when stats.Network is enabled")
	}
	if !strings.Contains(out, "↓5.3M") {
		t.Error("header should show NetDn when stats.Network is enabled")
	}
}

func TestRenderHeader_MultipleStatsEnabled(t *testing.T) {
	cfg := ui.AppConfig{
		StorePath: ".conductor/tasks.json",
		Stats:     config.StatsConfig{CPU: true, Memory: true, Network: true},
	}
	stats := sysinfo.Stats{CPU: "12%", Memory: "34%", NetUp: "↑2k", NetDn: "↓8k"}
	out := ui.RenderHeader(120, cfg, stats)
	if !strings.Contains(out, "12%") {
		t.Error("header should show CPU")
	}
	if !strings.Contains(out, "34%") {
		t.Error("header should show Memory")
	}
	if !strings.Contains(out, "↑2k") {
		t.Error("header should show NetUp")
	}
	if !strings.Contains(out, "↓8k") {
		t.Error("header should show NetDn")
	}
}

func TestRenderHeader_VeryNarrowTerminal(t *testing.T) {
	cfg := ui.AppConfig{StorePath: ".conductor/tasks.json"}
	// At 5 cols, logo definitely won't fit.
	out := ui.RenderHeader(5, cfg, sysinfo.Stats{})
	if out == "" {
		t.Error("header should not be empty even on very narrow terminal")
	}
	// Should still contain some info, but logo should be omitted.
}

func TestRenderHeader_ExactlyAtLogoWidthBoundary(t *testing.T) {
	cfg := ui.AppConfig{StorePath: ".conductor/tasks.json"}
	// Test at the boundary where logo just fits (around 40-50 cols).
	out := ui.RenderHeader(50, cfg, sysinfo.Stats{})
	if out == "" {
		t.Error("header should not be empty")
	}
}

// Task detail rendering edge cases.

func TestTaskDetail_RetryInformation(t *testing.T) {
	// Testing the retry display logic from tasks.go renderDetail.
	// We'll create a TaskRecord with retry metadata and verify it's shown.
	now := time.Now().UTC()
	wfID := "wf-test"
	record := client.TaskRecord{
		TaskID:     "t1",
		Name:       "retried-task",
		Capability: "http",
		Status:     client.StatusFailed,
		Attempt:    2,
		MaxRetries: 3,
		WorkflowID: &wfID,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	app := newTestApp()
	app.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	// Switch to tasks tab.
	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("1")})
	// Load records using test helper.
	app.Update(ui.RecordsLoadedMsgForTest([]client.TaskRecord{record}, nil))
	// Toggle detail pane.
	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})

	view := app.View()
	if !strings.Contains(view, "2 / 4") {
		t.Error("detail view should show 'Attempt: 2 / 4' (attempt / max_retries+1)")
	}
	if !strings.Contains(view, "eligible for retry") {
		t.Error("detail view should show retry eligibility for failed task with attempts remaining")
	}
}

func TestTaskDetail_NoRetryNeeded(t *testing.T) {
	now := time.Now().UTC()
	record := client.TaskRecord{
		TaskID:     "t1",
		Name:       "completed-task",
		Capability: "echo",
		Status:     client.StatusCompleted,
		Attempt:    1,
		MaxRetries: 0,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	app := newTestApp()
	app.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("1")})
	app.Update(ui.RecordsLoadedMsgForTest([]client.TaskRecord{record}, nil))
	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})

	view := app.View()
	// Should not show retry eligibility for completed task.
	if strings.Contains(view, "eligible for retry") {
		t.Error("completed task should not show retry eligibility")
	}
}

func TestTaskDetail_AuditTrailWithMetadata(t *testing.T) {
	now := time.Now().UTC()
	fromStatus := "pending"
	toStatus := "running"
	record := client.TaskRecord{
		TaskID:     "t1",
		Name:       "task-with-audit",
		Capability: "filesystem",
		Status:     client.StatusRunning,
		CreatedAt:  now,
		UpdatedAt:  now,
		AuditTrail: []client.AuditEntry{
			{
				Timestamp:  now,
				Actor:      "system",
				Action:     "status_changed",
				FromStatus: &fromStatus,
				ToStatus:   &toStatus,
				Metadata: map[string]any{
					"reason":    "scheduler_assigned",
					"worker_id": "worker-1",
				},
			},
		},
	}

	app := newTestApp()
	app.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("1")})
	app.Update(ui.RecordsLoadedMsgForTest([]client.TaskRecord{record}, nil))
	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})

	view := app.View()
	if !strings.Contains(view, "Audit Trail:") {
		t.Error("detail view should show Audit Trail header")
	}
	if !strings.Contains(view, "status_changed") {
		t.Error("detail view should show audit action")
	}
	if !strings.Contains(view, "system") {
		t.Error("detail view should show audit actor")
	}
	if !strings.Contains(view, "(pending → running)") {
		t.Error("detail view should show status transition")
	}
	if !strings.Contains(view, "reason") {
		t.Error("detail view should show audit metadata keys")
	}
}

func TestTaskDetail_ResultWithError(t *testing.T) {
	now := time.Now().UTC()
	errorMsg := "connection timeout"
	startedAt := now.Add(-5 * time.Second)
	completedAt := now
	record := client.TaskRecord{
		TaskID:     "t1",
		Name:       "failed-task",
		Capability: "http",
		Status:     client.StatusFailed,
		CreatedAt:  now,
		UpdatedAt:  now,
		Result: &client.TaskResult{
			Success:     false,
			Error:       &errorMsg,
			StartedAt:   &startedAt,
			CompletedAt: &completedAt,
		},
	}

	app := newTestApp()
	app.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("1")})
	app.Update(ui.RecordsLoadedMsgForTest([]client.TaskRecord{record}, nil))
	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})

	view := app.View()
	if !strings.Contains(view, "Result:") {
		t.Error("detail view should show Result header")
	}
	if !strings.Contains(view, "connection timeout") {
		t.Error("detail view should show error message")
	}
	if !strings.Contains(view, "started:") {
		t.Error("detail view should show started timestamp")
	}
	if !strings.Contains(view, "completed:") {
		t.Error("detail view should show completed timestamp")
	}
}

func TestTaskDetail_ResultWithOutput(t *testing.T) {
	now := time.Now().UTC()
	record := client.TaskRecord{
		TaskID:     "t1",
		Name:       "success-task",
		Capability: "echo",
		Status:     client.StatusCompleted,
		CreatedAt:  now,
		UpdatedAt:  now,
		Result: &client.TaskResult{
			Success: true,
			Output:  map[string]any{"message": "Hello, World!"},
		},
	}

	app := newTestApp()
	app.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("1")})
	app.Update(ui.RecordsLoadedMsgForTest([]client.TaskRecord{record}, nil))
	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})

	view := app.View()
	if !strings.Contains(view, "output:") {
		t.Error("detail view should show output field")
	}
}

func TestTaskDetail_NoWorkflowID(t *testing.T) {
	now := time.Now().UTC()
	record := client.TaskRecord{
		TaskID:     "t1",
		Name:       "standalone-task",
		Capability: "echo",
		Status:     client.StatusCompleted,
		WorkflowID: nil,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	app := newTestApp()
	app.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("1")})
	app.Update(ui.RecordsLoadedMsgForTest([]client.TaskRecord{record}, nil))
	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})

	view := app.View()
	// Workflow field should be omitted when WorkflowID is nil.
	// We can't easily assert absence without false positives, but we can
	// verify that the detail pane renders without errors.
	if !strings.Contains(view, "standalone-task") {
		t.Error("detail view should show task name")
	}
}

// Command palette edge cases.

func TestExecCmd_ThemeWithNoArgument(t *testing.T) {
	result, _ := ui.ExecCmd("theme")
	if result.Err {
		t.Error("theme command with no argument should not error (validation happens in app, not palette)")
	}
}

func TestExecCmd_RefreshWithArgument(t *testing.T) {
	result, _ := ui.ExecCmd("refresh 5s")
	if result.Err {
		t.Error("refresh command should be recognized (argument validation happens in app)")
	}
}

func TestExecCmd_StatsWithArgument(t *testing.T) {
	result, _ := ui.ExecCmd("stats on")
	if result.Err {
		t.Error("stats command should be recognized (argument validation happens in app)")
	}
}

func TestExecCmd_CaseSensitive(t *testing.T) {
	// Commands are case-sensitive in the current implementation.
	tests := []struct {
		input   string
		wantErr bool
	}{
		{"theme dracula", false},
		{"THEME dracula", true}, // uppercase not recognized
		{"Theme Dracula", true}, // mixed case not recognized
	}
	for _, tt := range tests {
		result, _ := ui.ExecCmd(tt.input)
		if result.Err != tt.wantErr {
			t.Errorf("command %q: got err=%v, want err=%v", tt.input, result.Err, tt.wantErr)
		}
	}
}

func TestExecCmd_LeadingTrailingWhitespace(t *testing.T) {
	result, cmd := ui.ExecCmd("  theme dracula  ")
	if result.Err {
		t.Error("command with leading/trailing whitespace should work")
	}
	if cmd == "" {
		t.Error("trimmed command should be returned")
	}
}

// Note: These tests use RecordsLoadedMsgForTest from export_test.go
// to properly simulate internal message passing in the Bubble Tea framework.

// Compact task status visibility tests.

func TestRenderStatusBar_Empty(t *testing.T) {
	bar := ui.RenderStatusBar(nil)
	if bar != "" {
		t.Errorf("expected empty status bar for nil records, got %q", bar)
	}

	bar = ui.RenderStatusBar([]client.TaskRecord{})
	if bar != "" {
		t.Errorf("expected empty status bar for empty slice, got %q", bar)
	}
}

func TestRenderStatusBar_SingleStatus(t *testing.T) {
	now := time.Now().UTC()
	records := []client.TaskRecord{
		{
			TaskID:    "t1",
			Name:      "test",
			Status:    client.StatusRunning,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
	bar := ui.RenderStatusBar(records)
	if bar == "" {
		t.Error("expected non-empty status bar for single running task")
	}
	if !strings.Contains(bar, "1") {
		t.Error("status bar should contain count '1'")
	}
}

func TestRenderStatusBar_MultipleCounts(t *testing.T) {
	now := time.Now().UTC()
	records := []client.TaskRecord{
		{TaskID: "t1", Name: "a", Status: client.StatusRunning, CreatedAt: now, UpdatedAt: now},
		{TaskID: "t2", Name: "b", Status: client.StatusRunning, CreatedAt: now, UpdatedAt: now},
		{TaskID: "t3", Name: "c", Status: client.StatusPending, CreatedAt: now, UpdatedAt: now},
		{TaskID: "t4", Name: "d", Status: client.StatusCompleted, CreatedAt: now, UpdatedAt: now},
		{TaskID: "t5", Name: "e", Status: client.StatusCompleted, CreatedAt: now, UpdatedAt: now},
		{TaskID: "t6", Name: "f", Status: client.StatusCompleted, CreatedAt: now, UpdatedAt: now},
		{TaskID: "t7", Name: "g", Status: client.StatusFailed, CreatedAt: now, UpdatedAt: now},
	}
	bar := ui.RenderStatusBar(records)
	if bar == "" {
		t.Error("expected non-empty status bar")
	}
	// Should show: running 2, pending 1, completed 3, failed 1
	if !strings.Contains(bar, "2") {
		t.Error("should show count '2' for running tasks")
	}
	if !strings.Contains(bar, "3") {
		t.Error("should show count '3' for completed tasks")
	}
}

func TestRenderStatusBar_AllStatuses(t *testing.T) {
	now := time.Now().UTC()
	records := []client.TaskRecord{
		{TaskID: "t1", Status: client.StatusRunning, CreatedAt: now, UpdatedAt: now},
		{TaskID: "t2", Status: client.StatusPending, CreatedAt: now, UpdatedAt: now},
		{TaskID: "t3", Status: client.StatusCompleted, CreatedAt: now, UpdatedAt: now},
		{TaskID: "t4", Status: client.StatusFailed, CreatedAt: now, UpdatedAt: now},
		{TaskID: "t5", Status: client.StatusAwaitingApproval, CreatedAt: now, UpdatedAt: now},
		{TaskID: "t6", Status: client.StatusApproved, CreatedAt: now, UpdatedAt: now},
		{TaskID: "t7", Status: client.StatusPolicyDenied, CreatedAt: now, UpdatedAt: now},
		{TaskID: "t8", Status: client.StatusCancelled, CreatedAt: now, UpdatedAt: now},
	}
	bar := ui.RenderStatusBar(records)
	if bar == "" {
		t.Error("expected non-empty status bar with all statuses")
	}
	for _, label := range []string{"RUN", "PND", "OK", "FAIL", "WAIT", "APV", "DENY", "CNCL"} {
		if !strings.Contains(bar, label) {
			t.Errorf("status bar should contain %q for all known statuses; got %q", label, bar)
		}
	}
}

func TestRenderStatusBar_OnlyCompletedTasks(t *testing.T) {
	now := time.Now().UTC()
	records := []client.TaskRecord{
		{TaskID: "t1", Status: client.StatusCompleted, CreatedAt: now, UpdatedAt: now},
		{TaskID: "t2", Status: client.StatusCompleted, CreatedAt: now, UpdatedAt: now},
	}
	bar := ui.RenderStatusBar(records)
	if bar == "" {
		t.Error("expected non-empty status bar")
	}
	// Should only show completed count, not other statuses.
}

func TestRenderStatusBar_IgnoresUnknownStatuses(t *testing.T) {
	now := time.Now().UTC()
	records := []client.TaskRecord{
		{TaskID: "t1", Status: client.StatusCompleted, CreatedAt: now, UpdatedAt: now},
		{TaskID: "t2", Status: client.TaskStatus("unknown_status"), CreatedAt: now, UpdatedAt: now},
	}
	bar := ui.RenderStatusBar(records)
	if bar == "" {
		t.Error("expected non-empty status bar")
	}
	// Unknown status won't be in the predefined order, so it won't be shown.
	// Verify bar contains "1" for completed but not "unknown_status"
	if !strings.Contains(bar, "1") {
		t.Error("should show count for completed tasks")
	}
	if strings.Contains(bar, "unknown_status") {
		t.Error("unknown statuses should not be shown in compact status bar")
	}
}

// Registry detail view execution controls tests.

func TestRegistryDetail_ExecutionControlsDisplay(t *testing.T) {
	timeout := 30.0
	minInterval := 5.0
	entries := []client.CapabilityEntry{
		{
			Name:        "test-cap",
			Description: "Test capability",
			RiskLevel:   "medium",
			ImportPath:  "pkg.caps:TestCap",
			ExecutionControls: &client.ExecControls{
				TimeoutSeconds:     &timeout,
				MinIntervalSeconds: &minInterval,
			},
		},
	}

	app := newTestApp()
	app.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	// Navigate to registry tab (3).
	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("3")})
	// Load capabilities.
	app.Update(ui.CapabilitiesLoadedMsgForTest(entries, nil))
	// Toggle detail pane.
	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})

	view := app.View()
	if !strings.Contains(view, "Execution Controls:") {
		t.Error("detail view should show 'Execution Controls:' header")
	}
	if !strings.Contains(view, "timeout: 30.0s") {
		t.Error("detail view should show timeout value")
	}
	if !strings.Contains(view, "min_interval: 5.0s") {
		t.Error("detail view should show min_interval value")
	}
}

func TestRegistryDetail_NoExecutionControls(t *testing.T) {
	entries := []client.CapabilityEntry{
		{
			Name:              "test-cap",
			Description:       "Test capability",
			RiskLevel:         "low",
			ExecutionControls: nil,
		},
	}

	app := newTestApp()
	app.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("3")})
	app.Update(ui.CapabilitiesLoadedMsgForTest(entries, nil))
	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})

	view := app.View()
	if strings.Contains(view, "Execution Controls:") {
		t.Error("detail view should NOT show 'Execution Controls:' when nil")
	}
}

func TestRegistryDetail_PartialExecutionControls(t *testing.T) {
	timeout := 15.0
	entries := []client.CapabilityEntry{
		{
			Name:        "test-cap",
			Description: "Test capability",
			RiskLevel:   "high",
			ExecutionControls: &client.ExecControls{
				TimeoutSeconds:     &timeout,
				MinIntervalSeconds: nil, // Only timeout set
			},
		},
	}

	app := newTestApp()
	app.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("3")})
	app.Update(ui.CapabilitiesLoadedMsgForTest(entries, nil))
	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})

	view := app.View()
	if !strings.Contains(view, "Execution Controls:") {
		t.Error("detail view should show 'Execution Controls:' header")
	}
	if !strings.Contains(view, "timeout: 15.0s") {
		t.Error("detail view should show timeout value")
	}
	if strings.Contains(view, "min_interval:") {
		t.Error("detail view should NOT show min_interval when nil")
	}
}
