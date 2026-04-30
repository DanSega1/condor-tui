package client_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DanSega1/condor-tui/internal/client"
)

// Edge cases for StoreReader.

func TestStoreReader_MalformedJSON(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "tasks.json")
	if err := os.WriteFile(p, []byte("{invalid json"), 0o600); err != nil {
		t.Fatal(err)
	}

	r := client.NewStoreReader(p)
	_, err := r.ReadAll()
	if err == nil {
		t.Error("expected error for malformed JSON, got nil")
	}
}

func TestStoreReader_EmptyJSON(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "tasks.json")
	if err := os.WriteFile(p, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}

	r := client.NewStoreReader(p)
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("empty object should parse fine: %v", err)
	}
	if len(records) != 0 {
		t.Errorf("expected 0 records from empty JSON object, got %d", len(records))
	}
}

func TestStoreReader_PartiallyCorruptRecords(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "tasks.json")
	// One valid, one invalid record.
	json := `{
		"t1": {
			"task_id": "t1",
			"name": "valid",
			"capability": "echo",
			"status": "completed",
			"created_at": "2024-01-01T12:00:00Z",
			"updated_at": "2024-01-01T12:00:00Z"
		},
		"t2": {
			"task_id": "t2",
			"created_at": "not a timestamp"
		}
	}`
	if err := os.WriteFile(p, []byte(json), 0o600); err != nil {
		t.Fatal(err)
	}

	r := client.NewStoreReader(p)
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll should skip invalid records: %v", err)
	}
	// Only the valid record should be returned.
	if len(records) != 1 {
		t.Errorf("expected 1 valid record, got %d", len(records))
	}
	if len(records) > 0 && records[0].TaskID != "t1" {
		t.Errorf("expected first record to be t1, got %s", records[0].TaskID)
	}
}

func TestStoreReader_ModTime_missing_file(t *testing.T) {
	r := client.NewStoreReader("/no/such/file/tasks.json")
	_, err := r.ModTime()
	if err == nil {
		t.Error("ModTime should return error for missing file")
	}
}

// Edge cases for RegistryReader.

func TestRegistryReader_MalformedYAML(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "caps.yaml")
	yaml := "include_builtins: [invalid yaml structure"
	if err := os.WriteFile(p, []byte(yaml), 0o600); err != nil {
		t.Fatal(err)
	}

	r := client.NewRegistryReader(p)
	_, err := r.Read()
	if err == nil {
		t.Error("expected error for malformed YAML, got nil")
	}
}

func TestRegistryReader_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "caps.yaml")
	if err := os.WriteFile(p, []byte(""), 0o600); err != nil {
		t.Fatal(err)
	}

	r := client.NewRegistryReader(p)
	_, err := r.Read()
	// Empty file causes YAML parsing error (EOF).
	if err == nil {
		t.Error("expected error for empty YAML file (EOF), got nil")
	}
}

func TestRegistryReader_MissingImportPath(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "caps.yaml")
	// Capability entry without import_path.
	yaml := `include_builtins: false
capabilities:
  - config:
      some_key: value
`
	if err := os.WriteFile(p, []byte(yaml), 0o600); err != nil {
		t.Fatal(err)
	}

	r := client.NewRegistryReader(p)
	entries, err := r.Read()
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	// Entry should be created with empty name and import_path.
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Name != "" {
		t.Errorf("expected empty name for missing import_path, got %q", entries[0].Name)
	}
	if entries[0].ImportPath != "" {
		t.Errorf("expected empty import_path, got %q", entries[0].ImportPath)
	}
}

func TestRegistryReader_ImportPathWithoutColon(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "caps.yaml")
	yaml := `include_builtins: false
capabilities:
  - import_path: "just.a.module.path"
`
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
	// Without colon, the entire string should be used as name.
	if entries[0].Name != "just.a.module.path" {
		t.Errorf("expected name=import_path when no colon, got %q", entries[0].Name)
	}
}

// Edge cases for LogTailer.

func TestLogTailer_FileRotation(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "test.log")
	if err := os.WriteFile(p, []byte("line 1\nline 2\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	tl := client.NewLogTailer(p)
	lines1, _ := tl.Poll(100)
	if len(lines1) != 2 {
		t.Fatalf("expected 2 lines on first poll, got %d", len(lines1))
	}

	// Simulate log rotation: replace file with smaller content.
	if err := os.WriteFile(p, []byte("new line 1\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	lines2, _ := tl.Poll(100)
	// LogTailer should detect rotation and restart from beginning.
	if len(lines2) != 1 {
		t.Fatalf("expected 1 line after rotation, got %d", len(lines2))
	}
	if lines2[0].Text != "new line 1" {
		t.Errorf("expected 'new line 1', got %q", lines2[0].Text)
	}
}

func TestLogTailer_NoTrailingNewline(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "test.log")
	// Write content without trailing newline.
	if err := os.WriteFile(p, []byte("incomplete"), 0o600); err != nil {
		t.Fatal(err)
	}

	tl := client.NewLogTailer(p)
	lines, _ := tl.Poll(100)
	if len(lines) != 1 {
		t.Fatalf("expected 1 line for content without newline, got %d", len(lines))
	}
	if lines[0].Text != "incomplete" {
		t.Errorf("expected 'incomplete', got %q", lines[0].Text)
	}
}

func TestLogTailer_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "test.log")
	if err := os.WriteFile(p, []byte(""), 0o600); err != nil {
		t.Fatal(err)
	}

	tl := client.NewLogTailer(p)
	lines, _ := tl.Poll(100)
	if len(lines) != 0 {
		t.Errorf("expected 0 lines from empty file, got %d", len(lines))
	}
}

func TestLogTailer_OnlyNewlines(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "test.log")
	// Only newlines, no actual text.
	if err := os.WriteFile(p, []byte("\n\n\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	tl := client.NewLogTailer(p)
	lines, _ := tl.Poll(100)
	// Empty lines are filtered out in Poll implementation.
	if len(lines) != 0 {
		t.Errorf("expected 0 lines (empty lines filtered), got %d", len(lines))
	}
}

// Execution controls parsing tests.

func TestRegistryReader_ExecutionControls_SingleCapability(t *testing.T) {
	dir := t.TempDir()
	yaml := `include_builtins: false
capabilities:
  - import_path: "mypackage.caps.custom:CustomCapability"
execution_controls:
  CustomCapability:
    timeout_seconds: 30.5
    min_interval_seconds: 2.0
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
	if entries[0].ExecutionControls == nil {
		t.Fatal("expected ExecutionControls to be set, got nil")
	}
	if *entries[0].ExecutionControls.TimeoutSeconds != 30.5 {
		t.Errorf("expected TimeoutSeconds=30.5, got %f", *entries[0].ExecutionControls.TimeoutSeconds)
	}
	if *entries[0].ExecutionControls.MinIntervalSeconds != 2.0 {
		t.Errorf("expected MinIntervalSeconds=2.0, got %f", *entries[0].ExecutionControls.MinIntervalSeconds)
	}
}

func TestRegistryReader_ExecutionControls_ForBuiltins(t *testing.T) {
	dir := t.TempDir()
	yaml := `include_builtins: true
capabilities: []
execution_controls:
  echo:
    timeout_seconds: 5.0
  http:
    timeout_seconds: 60.0
    min_interval_seconds: 1.0
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

	// Find echo and http in the list.
	var echoEntry, httpEntry *client.CapabilityEntry
	for i := range entries {
		if entries[i].Name == "echo" {
			echoEntry = &entries[i]
		}
		if entries[i].Name == "http" {
			httpEntry = &entries[i]
		}
	}
	if echoEntry == nil {
		t.Fatal("expected to find echo builtin")
	}
	if httpEntry == nil {
		t.Fatal("expected to find http builtin")
	}

	// Verify echo has execution controls.
	if echoEntry.ExecutionControls == nil {
		t.Fatal("expected echo ExecutionControls to be set")
	}
	if *echoEntry.ExecutionControls.TimeoutSeconds != 5.0 {
		t.Errorf("expected echo TimeoutSeconds=5.0, got %f", *echoEntry.ExecutionControls.TimeoutSeconds)
	}
	if echoEntry.ExecutionControls.MinIntervalSeconds != nil {
		t.Errorf("expected echo MinIntervalSeconds to be nil, got %f", *echoEntry.ExecutionControls.MinIntervalSeconds)
	}

	// Verify http has both controls.
	if httpEntry.ExecutionControls == nil {
		t.Fatal("expected http ExecutionControls to be set")
	}
	if *httpEntry.ExecutionControls.TimeoutSeconds != 60.0 {
		t.Errorf("expected http TimeoutSeconds=60.0, got %f", *httpEntry.ExecutionControls.TimeoutSeconds)
	}
	if *httpEntry.ExecutionControls.MinIntervalSeconds != 1.0 {
		t.Errorf("expected http MinIntervalSeconds=1.0, got %f", *httpEntry.ExecutionControls.MinIntervalSeconds)
	}
}

func TestRegistryReader_ExecutionControls_PartiallySet(t *testing.T) {
	dir := t.TempDir()
	yaml := `include_builtins: false
capabilities:
  - import_path: "pkg:Cap1"
  - import_path: "pkg:Cap2"
execution_controls:
  Cap1:
    timeout_seconds: 10.0
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
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	// Cap1 should have controls, Cap2 should not.
	cap1 := entries[0]
	cap2 := entries[1]
	if cap1.Name != "Cap1" {
		t.Errorf("expected first entry name=Cap1, got %q", cap1.Name)
	}
	if cap1.ExecutionControls == nil {
		t.Error("expected Cap1 ExecutionControls to be set")
	}
	if cap2.Name != "Cap2" {
		t.Errorf("expected second entry name=Cap2, got %q", cap2.Name)
	}
	if cap2.ExecutionControls != nil {
		t.Error("expected Cap2 ExecutionControls to be nil")
	}
}

func TestRegistryReader_ExecutionControls_EmptyBlock(t *testing.T) {
	dir := t.TempDir()
	yaml := `include_builtins: true
capabilities: []
execution_controls: {}
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

	// All builtins should have nil ExecutionControls.
	for _, e := range entries {
		if e.ExecutionControls != nil {
			t.Errorf("expected builtin %q ExecutionControls to be nil, got %v", e.Name, e.ExecutionControls)
		}
	}
}
