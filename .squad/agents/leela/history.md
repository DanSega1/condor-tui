# Project Context

- **Owner:** dans
- **Project:** Standalone Go/Bubble Tea terminal UI for monitoring and operating a running Conductor Engine instance.
- **Stack:** Go 1.24, Bubble Tea, Bubbles, Lip Gloss, YAML config, gopsutil.
- **Created:** 2026-04-30T14:14:21Z

## Learnings

### 1. Core Roadmap Alignment (2026-04-30)

The app **already ships the four Phase 4 TUI features**: Tasks, Workflows, Registry, Logs. No major roadmap gaps exist.

**Key insight:** The architecture is read-only and stateless. All four tabs consume local JSON/YAML/log files via `client.{StoreReader, RegistryReader, LogTailer}`. This eliminates temptation to add backend APIs and keeps scope crisp.

**Implication:** Future enhancements should respect this read-only boundary. If new features require writes or state (e.g., approval flows), defer to Phase 5 and the Engine.

### 2. Six Implementation-Ready Micro-Features Identified (2026-04-30)

Existing data structures contain rendered fields. Candidates for small, non-breaking additions:
- Audit trail rendering (fields: `AuditTrail[]` in `TaskRecord`)
- Retry attempt display (fields: `Attempt`, `MaxRetries` in `TaskRecord`)
- Risk level + execution control display (fields: `RiskLevel`, `ExecControls` in `CapabilityEntry`)
- Task status distribution in header (aggregation of existing records)
- Workflow timeline visualization (using `StartedAt`, `CompletedAt` in `TaskResult`)
- Config hot-reload (nice-to-have; not critical)

**Implication:** Prioritize Priorities 1–4 (low risk, high value). Save 5–6 for later sprint.

### 3. Test-Driven Project Pattern (2026-04-30)

All packages except `sysinfo` and root `main` have `_test.go` files. Config and client readers are well-tested. UI models have happy-path tests.

**Implication:** New features must follow the same pattern. Reviewers should reject PRs without accompanying tests.

### 4. Keybinding Space is Tight (2026-04-30)

Current bindings:
- Tab navigation: `1234`, `Tab`, `←/→`, `hjkl`
- List nav: `↑/↓`, `gG`
- Details: `enter` / `space`, `esc`, `e`, `d` (diff), `r` (refresh)
- Logs only: `/`, `f`, `c` (clear)
- Global: `:`, `?`, `q`

**Implication:** New task-detail toggles (audit, retry) should use lower-value keys or sub-commands (`:audit <task_id>`). Avoid polluting the global namespace.

### 5. Theme System is Extensible (2026-04-30)

Five themes defined in `internal/ui/theme.go`. Each has a `ThemeConfig` struct with colors, styles. Adding new themes or modifying existing colors is low-friction.

**Implication:** Risk badges and timeline visualization can use theme colors without creating new infrastructure.
