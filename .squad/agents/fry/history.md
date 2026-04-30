# Project Context

- **Owner:** dans
- **Project:** Standalone Go/Bubble Tea terminal UI for monitoring and operating a running Conductor Engine instance.
- **Stack:** Go 1.24, Bubble Tea, Bubbles, Lip Gloss, YAML config, gopsutil.
- **Created:** 2026-04-30T14:14:21Z

## Learnings

<!-- Append new learnings below. Each entry is something lasting about the project. -->

### 2026-04-30: Logo Update and TUI Enhancements

- **Logo Change:** Replaced the smslant figlet font with the cleaner standard figlet font in `internal/ui/header.go`. The new logo is more compact (4 lines vs 4 lines but narrower characters).
- **Retry Information Display:** Enhanced task detail view to prominently show retry attempts with visual indicators when tasks have been retried (`(retried)` suffix in error style). Added retry eligibility indicator for failed tasks that are eligible for retry.
- **Audit Trail Metadata:** Enhanced audit trail display to show metadata fields when present, providing visibility into approval decisions, policy checks, and other audit context.
- **Improved Empty States:** Added more helpful guidance in empty state messages across all views (tasks, workflows, registry) with multiple hints about what actions to take and what to check.
- **Status Constants:** When checking task status, use fully-qualified `client.StatusFailed` and other constants from the client package rather than unqualified names.

### 2026-04-30: Task Status Summary & Execution Controls

- **Task Status Summary in Header:** Added compact task status summary line in the header showing counts for running, pending, completed, and failed tasks using mini-badges (RUN, PND, OK, FAIL). The summary is built from loaded task records and only displays when tasks exist. This provides at-a-glance visibility without adding new API calls.
- **Execution Controls in Registry:** Extended `CapabilityEntry` to include `ExecutionControls` field (pointer to `ExecControls`). The `RegistryReader.Read()` now maps execution controls from `conductor.capabilities.yaml` to both plugin entries and builtins when present in the YAML. The registry detail view displays timeout and min_interval controls when available. This is safe and deterministic — controls are only shown when explicitly configured in YAML; no synthetic data is generated.
- **Header Signature Change:** Updated `renderHeader` to accept `[]client.TaskRecord` parameter so it can build the task summary. The exported `RenderHeader` test helper passes `nil` for backward compatibility.
- **ExecutionControls Mapping:** Controls are keyed by capability name (not import path). For builtins like "echo" or "filesystem", if `execution_controls.echo.timeout_seconds` exists in YAML, it's now surfaced. For plugins, the name is extracted from the import path and matched against the controls map.

