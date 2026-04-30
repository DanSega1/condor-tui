# Project Context

- **Owner:** dans
- **Project:** Standalone Go/Bubble Tea terminal UI for monitoring and operating a running Conductor Engine instance.
- **Stack:** Go 1.24, Bubble Tea, Bubbles, Lip Gloss, YAML config, gopsutil.
- **Created:** 2026-04-30T14:14:21Z

## Learnings

### 1. ASCII Banner as Visual Anchor
The cleaner figlet-standard ASCII art (4-line logo in header.go) serves as an instant visual reference for the app. Displaying it prominently in the README (centered, in a `<div>`) helps users instantly recognize the project before reading features.

### 2. Data Source Clarity Prevents Configuration Errors
Users often misconfigure paths or forget to provide the task store. A dedicated "Data Source Expectations" section describing defaults, required files, and supported task statuses upfront reduces friction in first-time setup.

### 3. Implemented vs Planned Features Must Be Visually Distinct
Adding status indicators (✅ for implemented, phases for planned) and a separate roadmap section prevents users from expecting unbuilt features. This aligns with Conductor Engine's roadmap phases (4, 5, 6+).

### 4. Dev Setup Should Be Self-Contained
The project has zero external dependencies for testing (`go test ./...`). Highlighting this in documentation (no integration setup required) improves developer experience and encourages contributions.

### 5. Project Layout as Living Documentation
Keeping `internal/` structure documented alongside go.mod versions ensures future maintainers understand architectural intent without reading code first.

### 6. Execution Controls Display—Keep Docs Aligned with Implementation
When a feature ships (execution controls in detail pane), update feature description with concise task-oriented language: "press `enter` to view execution controls (timeout, min_interval) if defined." Avoid over-documenting speculative fields or planned integrations that aren't exposed yet. Update happens inline within the feature table, not as a separate section.
