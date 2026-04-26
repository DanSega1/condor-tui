// Package client reads Conductor Engine local data sources.
package client

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	"gopkg.in/yaml.v3"
)

// BuiltinCapabilities lists the capabilities shipped with Conductor Engine.
var BuiltinCapabilities = []CapabilityEntry{
	{
		Name:        "echo",
		Description: "Echo input back as output. Useful for testing task pipelines.",
		RiskLevel:   "low",
		Tags:        []string{"builtin", "test"},
		ImportPath:  "engine.capabilities.echo:EchoCapability",
	},
	{
		Name:        "filesystem",
		Description: "Read, write, list, and delete local filesystem paths.",
		RiskLevel:   "medium",
		Tags:        []string{"builtin", "io"},
		ImportPath:  "engine.capabilities.filesystem:FilesystemCapability",
	},
	{
		Name:        "http",
		Description: "Perform outbound HTTP requests (GET, POST, etc.).",
		RiskLevel:   "medium",
		Tags:        []string{"builtin", "network"},
		ImportPath:  "engine.capabilities.http:HttpCapability",
	},
	{
		Name:        "memory",
		Description: "Store and retrieve keyed values across task executions.",
		RiskLevel:   "low",
		Tags:        []string{"builtin", "optional"},
		ImportPath:  "engine.capabilities.memory:MemoryCapability",
	},
}

// StoreReader reads task records from a Conductor Engine JSON task store.
type StoreReader struct {
	Path string
}

// NewStoreReader returns a StoreReader pointed at the given path.
// If path is empty it defaults to .conductor/tasks.json relative to the
// working directory.
func NewStoreReader(path string) *StoreReader {
	if path == "" {
		path = filepath.Join(".conductor", "tasks.json")
	}
	return &StoreReader{Path: path}
}

// ReadAll reads all task records from the JSON store.
// Records are returned sorted by created_at ascending.
func (r *StoreReader) ReadAll() ([]TaskRecord, error) {
	f, err := os.Open(r.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("open store %s: %w", r.Path, err)
	}
	defer f.Close()

	raw, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("read store %s: %w", r.Path, err)
	}

	var payload map[string]json.RawMessage
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, fmt.Errorf("parse store %s: %w", r.Path, err)
	}

	records := make([]TaskRecord, 0, len(payload))
	for _, v := range payload {
		var rec TaskRecord
		if err := json.Unmarshal(v, &rec); err != nil {
			continue
		}
		records = append(records, rec)
	}

	sort.Slice(records, func(i, j int) bool {
		return records[i].CreatedAt.Before(records[j].CreatedAt)
	})
	return records, nil
}

// Exists reports whether the store file is present.
func (r *StoreReader) Exists() bool {
	_, err := os.Stat(r.Path)
	return err == nil
}

// RegistryReader reads capability configuration from a YAML file.
type RegistryReader struct {
	Path string
}

// NewRegistryReader returns a RegistryReader pointed at the given path.
// Defaults to config/conductor.capabilities.yaml.
func NewRegistryReader(path string) *RegistryReader {
	if path == "" {
		path = filepath.Join("config", "conductor.capabilities.yaml")
	}
	return &RegistryReader{Path: path}
}

// Read loads capability entries from the YAML file and merges them with
// builtins (when include_builtins is true or the file is absent).
func (r *RegistryReader) Read() ([]CapabilityEntry, error) {
	entries := make([]CapabilityEntry, 0)

	f, err := os.Open(r.Path)
	if err != nil {
		if os.IsNotExist(err) {
			// No config file — just return builtins.
			return append(entries, BuiltinCapabilities...), nil
		}
		return nil, fmt.Errorf("open registry %s: %w", r.Path, err)
	}
	defer f.Close()

	var cfg CapabilityConfig
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("parse registry %s: %w", r.Path, err)
	}

	if cfg.IncludeBuiltins {
		entries = append(entries, BuiltinCapabilities...)
	}

	for _, p := range cfg.Capabilities {
		entries = append(entries, CapabilityEntry{
			Name:       importPathToName(p.ImportPath),
			ImportPath: p.ImportPath,
			RiskLevel:  "unknown",
		})
	}
	return entries, nil
}

// importPathToName extracts the class name portion from a Python import path
// like "engine.capabilities.echo:EchoCapability" → "EchoCapability".
func importPathToName(importPath string) string {
	for i := len(importPath) - 1; i >= 0; i-- {
		if importPath[i] == ':' {
			return importPath[i+1:]
		}
	}
	return importPath
}

// LogLine is a single line emitted from the log file.
type LogLine struct {
	Index int
	Text  string
}

// LogTailer tails a log file, delivering new lines on demand.
type LogTailer struct {
	Path   string
	offset int64
}

// NewLogTailer creates a LogTailer for the given path.
func NewLogTailer(path string) *LogTailer {
	return &LogTailer{Path: path}
}

// Poll reads any new lines since the last call and appends them to lines.
// If the file has shrunk (rotated), reading restarts from the beginning.
func (t *LogTailer) Poll(max int) ([]LogLine, error) {
	f, err := os.Open(t.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("open log %s: %w", t.Path, err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return nil, err
	}

	// Detect rotation.
	if info.Size() < t.offset {
		t.offset = 0
	}

	if _, err := f.Seek(t.offset, io.SeekStart); err != nil {
		return nil, err
	}

	raw, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	t.offset += int64(len(raw))

	// Split into lines.
	var lines []LogLine
	start := 0
	idx := 0
	for i, b := range raw {
		if b == '\n' {
			line := string(raw[start:i])
			if line != "" {
				lines = append(lines, LogLine{Index: idx, Text: line})
				idx++
			}
			start = i + 1
		}
	}
	// Trailing non-newline fragment.
	if start < len(raw) {
		lines = append(lines, LogLine{Index: idx, Text: string(raw[start:])})
	}

	if max > 0 && len(lines) > max {
		lines = lines[len(lines)-max:]
	}
	return lines, nil
}

// ModTime returns the modification time of the store file for change detection.
func (r *StoreReader) ModTime() (time.Time, error) {
	info, err := os.Stat(r.Path)
	if err != nil {
		return time.Time{}, err
	}
	return info.ModTime(), nil
}
