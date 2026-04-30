# Project Context

- **Owner:** dans
- **Project:** Standalone Go/Bubble Tea terminal UI for monitoring and operating a running Conductor Engine instance.
- **Stack:** Go 1.24, Bubble Tea, Bubbles, Lip Gloss, YAML config, gopsutil.
- **Created:** 2026-04-30T14:14:21Z

## Learnings

<!-- Append new learnings below. Each entry is something lasting about the project. -->

### 2026-04-30: Test Coverage Enhancement

**Context:** Added comprehensive edge-case tests across reader, config, and UI packages.

**Key Findings:**

1. **Reader Package Resilience:**
   - StoreReader gracefully handles partially corrupt JSON by skipping invalid records (line 86-91, reader.go)
   - RegistryReader returns builtins on missing config file (line 128-130, reader.go)
   - LogTailer detects file rotation by comparing size to offset (line 200-202, reader.go)
   - Empty YAML files cause EOF error in yaml.Decode, not graceful handling

2. **Config Merge Semantics:**
   - Zero-value flags (empty string, 0 duration) mean "not set" and don't override file config
   - Flag precedence: CLI flags > config file > built-in defaults
   - Nested structs (Stats, Engine, Tools) merge as complete units, not field-by-field

3. **UI Rendering Edge Cases:**
   - Header logo omitted when terminal width < 20 cols (line 85-92, header.go)
   - Store path truncated to 40 chars with leading ellipsis (line 36-38, header.go)
   - Stats color thresholds: >80% red, >60% yellow, else green (line 158-165, header.go)
   - Engine kind auto-detection from URL keywords: "k8s", "kubernetes", "docker" (line 141-147, header.go)

4. **Task Detail Display:**
   - Retry info shows `attempt / (max_retries+1)` format (line 236-240, tasks.go)
   - Retry eligibility shown for failed tasks with remaining attempts (line 243-247, tasks.go)
   - Audit trail includes optional metadata display (line 286-291, tasks.go)
   - Result timestamps (started_at, completed_at) rendered in 15:04:05 format (line 265-269, tasks.go)

5. **Command Palette Behavior:**
   - Commands are case-sensitive (line 101, cmdpalette.go)
   - Only first word validated; arguments passed through to app layer
   - No built-in validation for invalid durations or stats args in palette

**Testing Decisions:**

- Created `ui/testing.go` helper to expose internal msg types (recordsLoadedMsg, etc.) for deterministic testing
- Avoided full-screen snapshot tests in favor of targeted string matching
- Tests focus on edge cases (empty files, rotation, truncation) not happy paths
- Config tests verify merge semantics with zero-values and nested structs

**Test Count:** Increased from ~30 to 78 tests (48 new tests added)

**Files Modified:**
- `internal/client/reader_edge_test.go` (new: 12 edge-case tests)
- `internal/config/config_edge_test.go` (new: 11 edge-case tests)
- `internal/ui/ui_edge_test.go` (new: 23 UI rendering tests)
- `internal/ui/testing.go` (new: test helper for msg types)
- `internal/ui/app_test.go` (updated: use new test helper)

**Regression Coverage:** All 78 tests pass. Key areas now covered:
- Malformed JSON/YAML handling
- File rotation detection
- Config merge precedence
- Header rendering at various widths
- Task detail retry/audit/result display
- Command palette case sensitivity

### 2026-04-30: Execution Controls & Compact Status Tests

**Context:** Added tests for execution controls parsing from YAML and compact task status visibility feature that Fry is implementing. Tests are deterministic and avoid brittle full-screen snapshots.

**Key Findings:**

1. **Execution Controls Parsing (reader.go):**
   - Execution controls map by capability name from YAML `execution_controls:` section (line 143-162, reader.go)
   - Controls merge into builtins when `include_builtins: true` (line 141-149, reader.go)
   - Plugin capabilities also receive controls when name matches YAML key (line 152-163, reader.go)
   - Optional fields: `timeout_seconds`, `min_interval_seconds` (both float64 pointers)
   - Empty `execution_controls: {}` block is valid and results in nil controls for all capabilities

2. **Execution Controls Display (registry.go):**
   - Controls shown in detail pane under "Execution Controls:" header (line 214-223, registry.go)
   - Only displays non-nil fields (timeout and/or min_interval)
   - Format: "timeout: X.Xs" and "min_interval: X.Xs" (1 decimal place)
   - Detail pane toggles with space/enter, closes with ESC

3. **Compact Status Bar (tasks.go):**
   - `renderStatusBar()` aggregates task counts by status (line 320-337, tasks.go)
   - Returns horizontal badge layout: "● running 2  ○ pending 1  ✓ completed 3  ✗ failed 1"
   - Status display order: running, pending, completed, failed, awaiting_approval, approved, policy_denied, cancelled
   - Only shows statuses with count > 0
   - Returns empty string when no tasks present
   - Function not yet integrated into main view (Fry is working on surfacing it)

4. **Test Coverage Strategy:**
   - Used `export_test.go` to expose `RenderStatusBar()` for testing without modifying main code
   - Tested execution controls with all combinations: single field, both fields, empty, builtins, plugins
   - Verified UI detail pane shows controls correctly via message simulation (not snapshot)
   - Status bar tests cover edge cases: empty, single status, multiple counts, unknown statuses

**Testing Decisions:**

- Execution controls tests verify YAML parsing correctness at reader layer
- UI tests verify display rendering via targeted string matching (no brittle snapshots)
- Status bar tests are deterministic (no timing issues, no full renders)
- Export pattern used to test internal functions without polluting public API

**Test Count:** Increased from 78 to 91 tests (13 new tests added)

**Files Modified:**
- `internal/client/reader_edge_test.go` (+5 execution controls parsing tests)
- `internal/ui/ui_edge_test.go` (+6 status bar tests, +3 execution controls display tests)
- `internal/ui/export_test.go` (+1 export wrapper for RenderStatusBar)

**Regression Coverage:** All 91 tests pass. New areas covered:
- Execution controls parsing from YAML (single, partial, builtin merge, empty)
- Execution controls UI rendering in detail pane
- Compact status bar aggregation and display
- Edge cases: nil controls, partial controls, empty records

**Ready for Integration:**
- Execution controls fully parsed and displayed ✓
- Status bar function tested and ready (needs Fry to surface in main view)
- All existing tests still passing (no regressions)


