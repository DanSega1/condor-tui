package client_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DanSega1/condor-tui/internal/client"
)

// helpers

func writeStore(t *testing.T, dir string, records []client.TaskRecord) string {
	t.Helper()
	payload := make(map[string]client.TaskRecord, len(records))
	for _, r := range records {
		payload[r.TaskID] = r
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	p := filepath.Join(dir, "tasks.json")
	if err := os.WriteFile(p, raw, 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	return p
}

func makeRecord(id, name, capability, status string, wfID *string) client.TaskRecord {
	now := time.Now().UTC()
	return client.TaskRecord{
		TaskID:     id,
		Name:       name,
		Capability: capability,
		Status:     client.TaskStatus(status),
		WorkflowID: wfID,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// StoreReader tests

func TestStoreReader_ReadAll_empty(t *testing.T) {
	r := client.NewStoreReader("/tmp/does-not-exist-abc123/tasks.json")
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("unexpected error on missing file: %v", err)
	}
	if len(records) != 0 {
		t.Fatalf("expected 0 records, got %d", len(records))
	}
}

func TestStoreReader_ReadAll_records(t *testing.T) {
	dir := t.TempDir()
	wfID := "wf-1"
	records := []client.TaskRecord{
		makeRecord("b", "step-2", "echo", "completed", &wfID),
		makeRecord("a", "step-1", "echo", "running", &wfID),
	}
	// Records are written in reverse order; ReadAll must sort by created_at.
	// Force distinct timestamps so sort is deterministic.
	records[0].CreatedAt = time.Now().Add(-1 * time.Second)
	records[1].CreatedAt = time.Now().Add(-2 * time.Second)

	path := writeStore(t, dir, records)
	r := client.NewStoreReader(path)

	got, err := r.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 records, got %d", len(got))
	}
	// Sorted ascending by created_at: records[1] (step-1) before records[0] (step-2).
	if got[0].Name != "step-1" {
		t.Errorf("expected first record to be step-1, got %q", got[0].Name)
	}
	if got[1].Name != "step-2" {
		t.Errorf("expected second record to be step-2, got %q", got[1].Name)
	}
}

func TestStoreReader_Exists(t *testing.T) {
	dir := t.TempDir()
	path := writeStore(t, dir, nil)

	r := client.NewStoreReader(path)
	if !r.Exists() {
		t.Error("Exists returned false for a file that exists")
	}

	r2 := client.NewStoreReader(filepath.Join(dir, "nope.json"))
	if r2.Exists() {
		t.Error("Exists returned true for a file that does not exist")
	}
}

func TestStoreReader_ModTime(t *testing.T) {
	dir := t.TempDir()
	path := writeStore(t, dir, nil)

	r := client.NewStoreReader(path)
	mt, err := r.ModTime()
	if err != nil {
		t.Fatalf("ModTime: %v", err)
	}
	if mt.IsZero() {
		t.Error("ModTime returned zero time")
	}
}

// RegistryReader tests

func TestRegistryReader_missing_returns_builtins(t *testing.T) {
	r := client.NewRegistryReader("/tmp/no-such-conductor-config/caps.yaml")
	entries, err := r.Read()
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(entries) != len(client.BuiltinCapabilities) {
		t.Errorf("expected %d builtin entries, got %d",
			len(client.BuiltinCapabilities), len(entries))
	}
}

func TestRegistryReader_include_builtins_true(t *testing.T) {
	dir := t.TempDir()
	yaml := "include_builtins: true\ncapabilities: []\n"
	p := filepath.Join(dir, "caps.yaml")
	if err := os.WriteFile(p, []byte(yaml), 0o600); err != nil {
		t.Fatal(err)
	}

	r := client.NewRegistryReader(p)
	entries, err := r.Read()
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(entries) < len(client.BuiltinCapabilities) {
		t.Errorf("expected at least %d entries, got %d",
			len(client.BuiltinCapabilities), len(entries))
	}
}

func TestRegistryReader_plugin_entry(t *testing.T) {
	dir := t.TempDir()
	yaml := `include_builtins: false
capabilities:
  - import_path: "mypackage.caps.custom:CustomCapability"
`
	p := filepath.Join(dir, "caps.yaml")
	if err := os.WriteFile(p, []byte(yaml), 0o600); err != nil {
		t.Fatal(err)
	}

	r := client.NewRegistryReader(p)
	entries, err := r.Read()
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Name != "CustomCapability" {
		t.Errorf("expected name CustomCapability, got %q", entries[0].Name)
	}
	if entries[0].ImportPath != "mypackage.caps.custom:CustomCapability" {
		t.Errorf("unexpected import path: %q", entries[0].ImportPath)
	}
}

// LogTailer tests

func TestLogTailer_missing_file(t *testing.T) {
	tl := client.NewLogTailer("/tmp/no-such-log-abc.log")
	lines, err := tl.Poll(100)
	if err != nil {
		t.Fatalf("Poll on missing file should not error: %v", err)
	}
	if len(lines) != 0 {
		t.Errorf("expected 0 lines, got %d", len(lines))
	}
}

func TestLogTailer_reads_lines(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "test.log")
	content := "line one\nline two\nline three\n"
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	tl := client.NewLogTailer(p)
	lines, err := tl.Poll(100)
	if err != nil {
		t.Fatalf("Poll: %v", err)
	}
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	if lines[0].Text != "line one" {
		t.Errorf("unexpected first line: %q", lines[0].Text)
	}
}

func TestLogTailer_tails_incrementally(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "test.log")
	if err := os.WriteFile(p, []byte("first\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	tl := client.NewLogTailer(p)
	lines, _ := tl.Poll(100)
	if len(lines) != 1 {
		t.Fatalf("expected 1 line on first poll, got %d", len(lines))
	}

	// Append more lines.
	f, err := os.OpenFile(p, os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString("second\nthird\n")
	f.Close()

	lines2, _ := tl.Poll(100)
	if len(lines2) != 2 {
		t.Fatalf("expected 2 new lines on second poll, got %d", len(lines2))
	}
	if lines2[0].Text != "second" {
		t.Errorf("unexpected line: %q", lines2[0].Text)
	}
}

func TestLogTailer_max_limit(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "test.log")

	var sb []byte
	for i := 0; i < 20; i++ {
		sb = append(sb, []byte("line\n")...)
	}
	if err := os.WriteFile(p, sb, 0o600); err != nil {
		t.Fatal(err)
	}

	tl := client.NewLogTailer(p)
	lines, _ := tl.Poll(5)
	if len(lines) != 5 {
		t.Errorf("expected 5 lines with max=5, got %d", len(lines))
	}
}
