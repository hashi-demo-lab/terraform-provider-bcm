# Implementation Plan: BCM CMDevice Category Resource

**Branch**: `001-cmdevice-category` | **Date**: 2025-11-21 | **Spec**: [spec.md](/workspace/specs/001-cmdevice-category/spec.md)
**Input**: Feature specification from `/specs/001-cmdevice-category/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Implement a Terraform resource (`bcm_cmdevice_category`) for managing BCM device categories following the same patterns as the existing `bcm_cmpart_softwareimage` resource. Categories define node configuration templates including boot configuration, kernel parameters, disk layouts, network settings, and filesystem mounts. The resource will support full CRUD operations plus import, using the BCM CMDevice JSON-RPC API with efficient direct lookup for Read operations via `getCategory(name)`.

## Technical Context

**Language/Version**: Go 1.24.0
**Primary Dependencies**: terraform-plugin-framework (v1.16.1), terraform-plugin-testing (v1.13.3), terraform-plugin-log
**Storage**: BCM API backend (JSON-RPC over HTTPS), cookie-based session authentication
**Testing**: terraform-plugin-testing (acceptance tests with TF_ACC=1), Go native testing framework
**Target Platform**: Linux server (Terraform provider binary)
**Project Type**: Terraform Provider (single project structure with internal/provider/)
**Performance Goals**:
- Create operation: <10 seconds for typical category configuration
- Read operation: <3 seconds using efficient getCategory(name) API
- Update operation: <10 seconds for configuration changes
- Import operation: <5 seconds for category with full configuration

**Constraints**:
- Must follow existing provider patterns (resource_cmpart_softwareimage.go as reference)
- BCM API session cookies remain valid for 30+ minutes
- Category entities with all fields must remain under 500KB JSON size
- Disksetup XML limited to 10KB
- Exclude list fields limited to 50KB each

**Scale/Scope**:
- 60+ attributes in resource schema (including nested objects)
- Support for complex nested objects (BMCSettings, FSMount, KernelModule arrays)
- 6 API methods (getCategories, getCategory, addCategory, updateCategory, validateCategory, removeCategory)
- Comprehensive acceptance test coverage (5+ test scenarios)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Rule 1: Test-Driven Development (TDD)
**Status**: PASS
**Rationale**: Implementation will follow strict RED-GREEN-REFACTOR cycle:
- RED: Write failing acceptance tests for CRUD operations
- GREEN: Implement minimal code to pass tests
- REFACTOR: Improve code quality while maintaining passing tests
- Pattern follows existing resource_cmpart_softwareimage.go implementation

### Rule 2: Prefer Simplicity
**Status**: PASS
**Rationale**:
- Using existing proven patterns from bcm_cmpart_softwareimage resource
- No new abstractions or frameworks introduced
- Direct JSON-RPC API calls via established BCMClient
- Reusing helper functions (getStringValue, getBoolValue, getInt64Value)

### Rule 3: Implementation Before Architecture
**Status**: PASS
**Rationale**:
- Architecture already established by existing provider structure
- Following concrete reference implementation (resource_cmpart_softwareimage.go)
- No speculative design - implementing specific Category resource requirements

### Rule 4: Parallel Test Execution
**Status**: PASS
**Rationale**:
- Acceptance tests will be written in parallel batches
- CRUD operations implemented concurrently where independent
- Multiple test scenarios executed in parallel during CI
- Following terraform-plugin-testing parallel execution patterns

### Overall Constitution Compliance
**PASS** - All constitutional rules satisfied. No complexity violations to justify.

## Project Structure

### Documentation (this feature)

```text
specs/[###-feature]/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
terraform-provider-bcm/
├── internal/provider/
│   ├── provider.go                           # Provider registration (update Resources() method)
│   ├── bcm_client.go                         # JSON-RPC client (existing, no changes)
│   ├── resource_cmdevice_category.go         # NEW: Category resource implementation
│   ├── resource_cmdevice_category_test.go    # NEW: Acceptance tests
│   └── data_source_cmpart_softwareimages.go  # Existing (contains helper functions)
│
├── examples/
│   └── resources/
│       └── bcm_cmdevice_category/
│           └── resource.tf                    # NEW: Example configurations
│
├── docs/
│   └── resources/
│       └── bcm_cmdevice_category.md          # AUTO-GENERATED by tfplugindocs
│
└── specs/001-cmdevice-category/              # This feature's planning docs
    ├── spec.md                                # Existing specification
    ├── plan.md                                # This file
    ├── research.md                            # Phase 0 output
    ├── data-model.md                          # Phase 1 output
    ├── quickstart.md                          # Phase 1 output
    ├── contracts/                             # Phase 1 output
    │   └── cmdevice-category-api.md          # API contract documentation
    └── tasks.md                               # Phase 2 output (created by /speckit.tasks)
```

**Structure Decision**: Standard Terraform Provider framework structure. All implementation code goes in `internal/provider/`, following the established pattern of `resource_<service>_<entity>.go` for resources and `resource_<service>_<entity>_test.go` for acceptance tests. Helper functions will be reused from existing data source implementations. Documentation is auto-generated by tfplugindocs from schema definitions and example files.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

**N/A** - No complexity violations. Implementation follows existing patterns.

---

## Phase 0: Research & Outline

**Status**: COMPLETED
**Output**: `/workspace/specs/001-cmdevice-category/research.md`

### Research Summary

All technical unknowns resolved through analysis of existing patterns and BCM API documentation:

1. **API Methods**: Verified all 6 CMDevice API methods (getCategories, getCategory, addCategory, updateCategory, validateCategory, removeCategory)
2. **Nested Objects**: Defined strategy using terraform-plugin-framework nested attributes with dedicated model structs
3. **Long Text Fields**: Will store as plain strings with length validation (disksetup 10KB, excludeLists 50KB)
4. **Validation**: Hybrid approach with client-side schema validators + server-side validateCategory API
5. **Import**: Two-phase approach (getCategories to find name, then getCategory for full data)
6. **Eventual Consistency**: Not needed (category operations are synchronous, unlike image cloning)
7. **Force Parameter**: Single optional boolean attribute applying to all operations
8. **Helper Functions**: Reusing getStringValue, getBoolValue, getInt64Value from data_source_cmpart_softwareimages.go
9. **Testing**: 5 comprehensive acceptance test scenarios following TDD RED-GREEN-REFACTOR

**Key Decisions**:
- Use efficient `getCategory(name)` for Read operations (not list+filter)
- Mark `bmc_settings.password` as sensitive
- No special handling for eventual consistency (operations are synchronous)
- Reuse proven patterns from `resource_cmpart_softwareimage.go`

---

## Phase 1: Design & Contracts

**Status**: COMPLETED
**Outputs**:
- `/workspace/specs/001-cmdevice-category/data-model.md`
- `/workspace/specs/001-cmdevice-category/contracts/cmdevice-category-api.md`
- `/workspace/specs/001-cmdevice-category/quickstart.md`
- Agent context updated: `.github/agents/copilot-instructions.md`

### Design Summary

**Data Model**:
- Primary entity: `CMDeviceCategoryResourceModel` with 60+ attributes
- Nested entities: `SoftwareImageProxyModel`, `BMCSettingsModel`, `FSMountModel`, `KernelModuleModel`
- Complete field mappings between Terraform schema and BCM API entities
- Validation rules for all user-provided fields
- Transformation examples for common scenarios

**API Contracts**:
- Complete specification for all 6 BCM CMDevice API methods
- Request/response formats with examples
- Error handling patterns and response codes
- Enum values for all constrained fields
- Performance characteristics and rate limiting guidance
- Security considerations for sensitive data

**Developer Quickstart**:
- Step-by-step TDD implementation guide (RED-GREEN-REFACTOR)
- Phase 0: Write failing tests (45-60 min)
- Phase 1: Minimal implementation (60-90 min)
- Phase 2: Full API integration (90-120 min)
- Phase 3: Additional tests (60-90 min)
- Phase 4: Documentation (30-45 min)
- Common issues & solutions
- Verification checklist

**Constitution Re-Check**: PASS (all rules satisfied after design completion)

---

## Phase 2: Task Generation

**Status**: NOT STARTED (requires `/speckit.tasks` command)
**Expected Output**: `/workspace/specs/001-cmdevice-category/tasks.md`

This phase will generate actionable, dependency-ordered tasks based on the design artifacts created in Phase 0 and Phase 1. The tasks will follow TDD workflow with clear RED-GREEN-REFACTOR cycles.

**Note**: Per the command workflow, `/speckit.plan` stops after Phase 1. Run `/speckit.tasks` to generate the task breakdown for implementation.
