# Squad Decisions

## Active Decisions

### 1. Roadmap Fit Analysis: Phase 4 TUI Scope (2026-04-30)
**Author:** Leela (Lead/Architect)  
**Status:** Assessment complete; candidates ready for implementation

**Summary:** The condor-tui app already implements all four Phase 4 features (Tasks, Workflows, Registry, Logs). Six roadmap-fit enhancement candidates identified, prioritized by effort and value. Read-only, stateless architecture eliminates backend API coupling.

**Key Decisions:**
- ✅ Audit Trail Details, Task Retry Metadata, and Risk Levels are Priority 1 (ready now)
- ❌ Remote HTTP API, Approval flows, and Backpressure UI explicitly out of scope (belong in Engine or Phase 5)
- ✅ Keybinding space tight; new toggles should use sub-commands (`:audit <id>`)

**Files:** `.squad/decisions/inbox/leela-roadmap-fit.md`

---

### 2. TUI Enhancements - Roadmap Fit Strategy (2026-04-30)
**Author:** Fry (TUI Engineer)  
**Status:** Implemented

**Summary:** Implemented three focused enhancements surfacing existing data from task store without API changes: retry information visibility, audit trail metadata display, improved empty state guidance.

**Key Decisions:**
- ✅ Display "N / M (retried)" badge and retry eligibility for failed tasks
- ✅ Show audit metadata key-value pairs below main action line
- ✅ Add multi-line helpful hints to empty states (tasks, workflows, registry)
- ✅ Update figlet logo to cleaner standard font in header.go

**Files Modified:** `internal/ui/header.go`, `internal/ui/tasks.go`, `internal/ui/workflow.go`, `internal/ui/registry.go`  
**Files:** `.squad/decisions/inbox/fry-roadmap-fit.md`

---

### 3. Documentation & Roadmap Decisions – condor-tui (2026-04-30)
**Author:** Amy (Technical Writer)  
**Status:** Implemented

**Summary:** Enhanced README.md with clearer structure, data source expectations, development guidance, and roadmap section distinguishing implemented vs planned features.

**Key Decisions:**
- ✅ Display cleaner figlet ASCII art in centered `<div>` at README top
- ✅ Add "Data Source Expectations" section before installation (clarifies file paths, requirements, supported statuses)
- ✅ Split features table into Core (✅ implemented) and Planned (roadmap phases 5-7)
- ✅ Reorganize development section: "Run Tests", "Build", "Lint" with explanatory text
- ✅ Document Go 1.24.13 version and internal structure in project layout

**Trade-offs:** README uses non-standard markdown (HTML div), but widely supported by GitHub.  
**Files:** `.squad/decisions/inbox/amy-docs-roadmap.md`

---

### 4. Test Scope Decision: UI Message Type Test Helpers (2026-04-30)
**Author:** Zoidberg (QA Engineer)  
**Status:** Implemented

**Summary:** Created `internal/ui/export_test.go` with test-only exported helper functions to wrap internal message types (`recordsLoadedMsg`, `capabilitiesLoadedMsg`, `logLinesMsg`) for deterministic UI testing.

**Key Decisions:**
- ✅ Provide test helpers: `RecordsLoadedMsgForTest()`, `CapabilitiesLoadedMsgForTest()`, `LogLinesMsgForTest()`
- ✅ Keep helpers test-only in `export_test.go` to maintain encapsulation without polluting production code
- ✅ Enable 20+ new UI behavior tests for retry, audit, result formatting
- ✅ Added 48 edge-case tests across reader, config, and UI packages (160% test increase)

**Why This Works:** Test helpers are clearly marked as test-only; they wrap internal types without exposing them to package consumers; they enable fast, deterministic tests without I/O or timing issues.

**Files Created:** `internal/ui/export_test.go`, `internal/client/reader_edge_test.go`, `internal/config/config_edge_test.go`, `internal/ui/ui_edge_test.go`  
**Files Modified:** `internal/ui/app_test.go`  
**Files:** `.squad/decisions/inbox/zoidberg-test-helpers.md`

---

### 5. Task Status Summary & Execution Controls Display (2026-04-30)
**Author:** Fry (TUI Engineer)  
**Status:** Implemented

**Summary:** Added compact task status summary in header (mini-badges for all known Conductor task statuses) and surfaced execution controls (timeout, min_interval) in registry detail view when configured.

**Key Decisions:**
- ✅ Header displays task status summary with mini-badges below engine/store/version info (only when tasks exist)
- ✅ Registry detail shows timeout and min_interval from `conductor.capabilities.yaml` ExecutionControls
- ✅ Extended `CapabilityEntry` struct with `ExecutionControls` field; `RegistryReader.Read()` maps controls to entries
- ✅ No new backend API calls; uses existing task records and registry config

**Rationale:** Operators need queue health visibility and control transparency without tab switching or file inspection.

**Files Modified:** `internal/client/types.go`, `internal/client/reader.go`, `internal/ui/header.go`, `internal/ui/app.go`, `internal/ui/registry.go`

---

### 6. QA Test Patterns for Execution Controls & Status Visibility (2026-04-30)
**Author:** Zoidberg (QA Engineer)  
**Status:** Implemented

**Summary:** Adopted edge-case test file pattern (`*_edge_test.go`), test-only exports (`export_test.go`), and message-simulation approach for deterministic UI testing.

**Key Decisions:**
- ✅ Edge test files focus on boundary/error cases; regular files cover happy paths
- ✅ Use `export_test.go` to expose internal functions without polluting production code
- ✅ Test UI via message passing (RecordsLoadedMsg, CapabilitiesLoadedMsg) not snapshots
- ✅ Test primitives even if feature not yet surfaced (documents expected behavior)

**Test Impact:**
- **Before:** 78 tests
- **After:** 91 tests (+13 new tests for status bar and execution controls)
- **Files Created:** `internal/client/reader_edge_test.go`, `internal/config/config_edge_test.go`, `internal/ui/ui_edge_test.go`, `internal/ui/export_test.go`

---

### 7. Documentation: Execution Controls Display & Status Awareness (2026-04-30)
**Author:** Amy (Technical Writer)  
**Status:** Implemented

**Summary:** Updated README feature table to document compact task status counts and execution controls visibility.

**Key Decisions:**
- ✅ Added Tasks description noting compact task status counts in the header.
- ✅ Added Registry description: "Select a capability and press `enter` to view execution controls (timeout, min_interval) if defined."
- ✅ Avoids separate section bloat; documents only what's implemented

**Rationale:** Align shipped features with user-facing documentation; sets pattern for future docs updates when features ship.

---

## Governance

- All meaningful changes require team consensus
- Document architectural decisions here
- Keep history focused on work, decisions focused on direction
- Decision inbox files are temporary drop-box entries; Scribe merges them here and removes them after merge.
