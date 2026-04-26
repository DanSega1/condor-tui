// Package client provides types and readers for Conductor Engine data sources.
package client

import "time"

// TaskStatus mirrors engine.interfaces.task.TaskStatus.
type TaskStatus string

const (
	StatusPending          TaskStatus = "pending"
	StatusRunning          TaskStatus = "running"
	StatusCompleted        TaskStatus = "completed"
	StatusFailed           TaskStatus = "failed"
	StatusAwaitingApproval TaskStatus = "awaiting_approval"
	StatusApproved         TaskStatus = "approved"
	StatusPolicyDenied     TaskStatus = "policy_denied"
	StatusCancelled        TaskStatus = "cancelled"
)

// TaskResult mirrors engine.interfaces.task.TaskResult.
type TaskResult struct {
	Success     bool           `json:"success"`
	Output      interface{}    `json:"output,omitempty"`
	Error       *string        `json:"error,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	StartedAt   *time.Time     `json:"started_at,omitempty"`
	CompletedAt *time.Time     `json:"completed_at,omitempty"`
}

// AuditEntry mirrors engine.interfaces.task.AuditEntry.
type AuditEntry struct {
	Timestamp  time.Time  `json:"timestamp"`
	Actor      string     `json:"actor"`
	Action     string     `json:"action"`
	FromStatus *string    `json:"from_status,omitempty"`
	ToStatus   *string    `json:"to_status,omitempty"`
	Metadata   map[string]any `json:"metadata,omitempty"`
}

// TaskRecord mirrors engine.interfaces.task.TaskRecord.
type TaskRecord struct {
	TaskID     string         `json:"task_id"`
	Name       string         `json:"name"`
	Capability string         `json:"capability"`
	Input      map[string]any `json:"input,omitempty"`
	Metadata   map[string]any `json:"metadata,omitempty"`
	Status     TaskStatus     `json:"status"`
	Result     *TaskResult    `json:"result,omitempty"`
	Attempt    int            `json:"attempt"`
	MaxRetries int            `json:"max_retries"`
	WorkflowID *string        `json:"workflow_id,omitempty"`
	ArchivedAt *time.Time     `json:"archived_at,omitempty"`
	AuditTrail []AuditEntry   `json:"audit_trail,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

// CapabilityEntry represents a single registered capability.
type CapabilityEntry struct {
	Name        string
	Description string
	RiskLevel   string
	Tags        []string
	ImportPath  string
}

// CapabilityConfig mirrors the conductor.capabilities.yaml structure.
type CapabilityConfig struct {
	IncludeBuiltins bool                       `yaml:"include_builtins"`
	Capabilities    []CapabilityPluginEntry    `yaml:"capabilities"`
	ExecutionControls map[string]ExecControls  `yaml:"execution_controls,omitempty"`
}

// CapabilityPluginEntry is one entry under `capabilities:` in the YAML.
type CapabilityPluginEntry struct {
	ImportPath string         `yaml:"import_path"`
	Config     map[string]any `yaml:"config,omitempty"`
}

// ExecControls mirrors CapabilityExecutionControls.
type ExecControls struct {
	TimeoutSeconds     *float64 `yaml:"timeout_seconds,omitempty"`
	MinIntervalSeconds *float64 `yaml:"min_interval_seconds,omitempty"`
}
