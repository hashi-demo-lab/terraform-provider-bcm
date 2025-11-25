# Implementation Plan: Category Dynamic Fields Schema Implementation

**Branch**: `001-category-dynamic-fields` | **Date**: 2025-11-24 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/workspace/specs/001-category-dynamic-fields/spec.md`

## Summary

Replace 5 types.Dynamic placeholders in bcm_cmdevice_category resource with properly typed ListNestedAttribute schemas. This addresses technical debt from the initial resource implementation where complex nested structures were stubbed with dynamic types. The feature enables proper type safety, validation, and documentation for static_routes, fsexports, roles, gpu_settings, and services fields.

**Technical Approach**: Incremental TDD implementation following RED-GREEN-REFACTOR cycles. Each field is implemented as a separate iteration: (1) Define schema, (2) Write failing tests, (3) Implement CRUD serialization/deserialization, (4) Refactor and optimize. Priority order: static_routes (P1) → fsexports (P1) → roles (P2) → gpu_settings (P3) → services (P3).

## Technical Context

**Language/Version**: Go 1.24.0
**Primary Dependencies**: terraform-plugin-framework v1.16.1, terraform-plugin-testing v1.13.3
**Storage**: BCM JSON-RPC API (cmdevice service) over HTTPS
**Testing**: terraform-plugin-testing with TF_ACC=1 acceptance tests
**Target Platform**: Linux provider binary (Terraform 1.5+)
**Project Type**: Terraform Provider (infrastructure as code)
**Performance Goals**: <2s per CRUD operation, acceptance test suite <120m total
**Constraints**: BCM API eventual consistency (requires retry logic with exponential backoff), field name mapping (snake_case ↔ camelCase), preserve empty lists (not null)
**Scale/Scope**: 5 nested object types, ~30 new attributes, 35+ acceptance test scenarios (7 per field)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

✅ **TDD Compliance**: All changes follow RED-GREEN-REFACTOR pattern with acceptance tests first
✅ **Test Coverage**: Each field requires 7 test scenarios (CRUD, import, idempotency, drift)
✅ **No Repository Pattern**: Direct BCM API integration (terraform-plugin-framework standard)
✅ **Parallel Execution**: Independent field implementations can proceed concurrently
✅ **Documentation**: Auto-generated via tfplugindocs after implementation

**No violations** - This is refactoring technical debt within existing resource architecture.

## Project Structure

### Documentation (this feature)

```text
specs/001-category-dynamic-fields/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output - BCM API field structures
├── data-model.md        # Phase 1 output - Nested object schemas
├── quickstart.md        # Phase 1 output - Developer guide
├── contracts/           # Phase 1 output - BCM API JSON examples
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
internal/provider/
├── resource_cmdevice_category.go       # Main resource implementation (MODIFY)
├── resource_cmdevice_category_test.go  # Acceptance tests (EXTEND)
├── test_helpers.go                     # Test utilities (USE EXISTING)
├── bcm_client.go                       # API client (NO CHANGES)
└── helpers.go                          # Field mappers (NEW - optional)

examples/resources/bcm_cmdevice_category/
├── resource.tf                         # Basic example (UPDATE)
├── static-routes.tf                    # Static routes example (NEW)
├── fsexports.tf                        # NFS exports example (NEW)
├── roles.tf                            # Role assignment example (NEW)
└── gpu-settings.tf                     # GPU config example (NEW)

sampleRest/
├── CMDevice_Complete_Documentation.md  # API reference (REFERENCE)
└── category-dynamic-fields/            # API exploration (NEW)
    ├── static-routes.json
    ├── fsexports.json
    ├── roles.json
    ├── gpu-settings.json
    └── services.json
```

**Structure Decision**: Single Terraform provider project following HashiCorp plugin framework conventions. All changes isolated to resource_cmdevice_category.go with comprehensive test coverage in corresponding _test.go file. Examples directory updated to demonstrate each field's usage patterns.

## Complexity Tracking

**No complexity violations** - This feature simplifies existing architecture by replacing dynamic types with proper schemas. Complexity reduces as type safety increases.

---

## Phase 0: Research & API Discovery

**Objective**: Document BCM API structure for all 5 dynamic fields using real API responses.

### Research Tasks

1. **Static Routes Research** (Priority: P1)
   - Query BCM API for categories with static routes configured
   - Document route object structure: destination, gateway, metric
   - Validate CIDR notation for destination, IPv4 for gateway
   - Output: `research.md` section with JSON examples

2. **FSExports Research** (Priority: P1)
   - Query BCM API for categories with NFS exports configured
   - Document export object structure: path, network, allowWrite, rootSquash, async
   - Confirm baseType field presence (BCM internal, not exposed in Terraform)
   - Output: `research.md` section with JSON examples

3. **Roles Research** (Priority: P2)
   - Query BCM API for categories with roles assigned
   - Document role object structure: name, childType, uuid, addServices
   - List known role types: HeadNodeRole, StorageRole, BackupRole, etc.
   - Confirm uuid is computed by BCM (not user-provided)
   - Output: `research.md` section with JSON examples

4. **GPU Settings Research** (Priority: P3)
   - Query BCM API for categories with GPU configurations
   - Document GPU object structure: deviceId, model, computeMode
   - List known compute modes: default, exclusive, prohibited
   - Output: `research.md` section with JSON examples

5. **Services Research** (Priority: P3)
   - Query BCM API for categories with services configured
   - Document service object structure (NEEDS CLARIFICATION - unknown structure)
   - Determine if services follow same pattern as roles
   - Output: `research.md` section with JSON examples or note if not implemented

### Research Deliverables

- **File**: `specs/001-category-dynamic-fields/research.md`
- **Content**:
  - BCM API JSON response examples for each field
  - Field type mappings (string, int64, bool)
  - Required vs optional attribute analysis
  - Validation requirements (CIDR, IPv4, UUID formats)
  - Known enumeration values (role types, compute modes)

**Research Exit Criteria**: All NEEDS CLARIFICATION items resolved, JSON structure documented for 4 confirmed fields (services may be deferred to P3).

---

## Phase 1: Design & Schema Definition

**Objective**: Define Terraform schemas for all 5 fields and document API contracts.

### Design Decisions

**1. Schema Pattern**: Use ListNestedAttribute for all fields
- Rationale: All 5 fields are arrays of nested objects in BCM API
- Alternative rejected: DynamicAttribute (insufficient type safety, no validation)

**2. Field Name Mapping**: Implement snake_case ↔ camelCase conversion
- Terraform uses snake_case: `allow_write`, `root_squash`, `add_services`
- BCM API uses camelCase: `allowWrite`, `rootSquash`, `addServices`
- Implementation: Helper functions in buildAPIEntity and readCategory methods

**3. Empty List Handling**: Preserve empty arrays as empty lists (not null)
- Rationale: Users expect `[]` in config to remain `[]` in state
- Implementation: Use types.ListValueMust(elementType, []attr.Value{}) for empty arrays

**4. UUID Fields**: Mark as Computed for role uuid, fs_export uuid
- Rationale: BCM assigns UUIDs server-side, not user-configurable
- Implementation: Computed: true in schema definition

**5. BaseType Exclusion**: Do not expose baseType in Terraform schema
- Rationale: Internal BCM field for polymorphism, not user-relevant
- Implementation: Include in API serialization, exclude from schema

### Schema Definitions

**Static Route Schema** (P1):
```go
"static_routes": schema.ListNestedAttribute{
    Optional: true,
    MarkdownDescription: "Static network routes for nodes in this category",
    NestedObject: schema.NestedAttributeObject{
        Attributes: map[string]schema.Attribute{
            "destination": schema.StringAttribute{
                Required: true,
                MarkdownDescription: "Destination network in CIDR notation (e.g., 192.168.1.0/24)",
                Validators: []validator.String{
                    stringvalidator.RegexMatches(
                        regexp.MustCompile(`^([0-9]{1,3}\.){3}[0-9]{1,3}/[0-9]{1,2}$`),
                        "must be valid CIDR notation (e.g., 192.168.1.0/24)",
                    ),
                },
            },
            "gateway": schema.StringAttribute{
                Required: true,
                MarkdownDescription: "Gateway IP address (e.g., 10.0.0.1)",
                Validators: []validator.String{
                    stringvalidator.RegexMatches(
                        regexp.MustCompile(`^([0-9]{1,3}\.){3}[0-9]{1,3}$`),
                        "must be valid IPv4 address",
                    ),
                },
            },
            "metric": schema.Int64Attribute{
                Optional: true,
                MarkdownDescription: "Route metric (priority, lower is preferred)",
            },
        },
    },
}
```

**FSExport Schema** (P1):
```go
"fsexports": schema.ListNestedAttribute{
    Optional: true,
    MarkdownDescription: "NFS filesystem exports for nodes in this category",
    NestedObject: schema.NestedAttributeObject{
        Attributes: map[string]schema.Attribute{
            "path": schema.StringAttribute{
                Required: true,
                MarkdownDescription: "Export path (e.g., /home, /shared)",
            },
            "network": schema.StringAttribute{
                Required: true,
                MarkdownDescription: "Network UUID reference for export access",
                Validators: []validator.String{
                    stringvalidator.RegexMatches(
                        regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`),
                        "must be valid RFC 4122 UUID",
                    ),
                },
            },
            "allow_write": schema.BoolAttribute{
                Optional: true,
                MarkdownDescription: "Allow write access (default: false)",
            },
            "root_squash": schema.BoolAttribute{
                Optional: true,
                MarkdownDescription: "Enable root squash security (default: false)",
            },
            "async": schema.BoolAttribute{
                Optional: true,
                MarkdownDescription: "Use async mode for writes (default: false)",
            },
        },
    },
}
```

**Role Schema** (P2):
```go
"roles": schema.ListNestedAttribute{
    Optional: true,
    MarkdownDescription: "Service role assignments for nodes in this category",
    NestedObject: schema.NestedAttributeObject{
        Attributes: map[string]schema.Attribute{
            "name": schema.StringAttribute{
                Required: true,
                MarkdownDescription: "Role name (e.g., headnode, storage, compute)",
            },
            "child_type": schema.StringAttribute{
                Required: true,
                MarkdownDescription: "Role type (e.g., HeadNodeRole, StorageRole, BackupRole)",
            },
            "uuid": schema.StringAttribute{
                Computed: true,
                MarkdownDescription: "Role UUID (assigned by BCM)",
            },
            "add_services": schema.BoolAttribute{
                Optional: true,
                MarkdownDescription: "Automatically add role services (default: false)",
            },
        },
    },
}
```

**GPU Settings Schema** (P3):
```go
"gpu_settings": schema.ListNestedAttribute{
    Optional: true,
    MarkdownDescription: "GPU hardware configuration for nodes in this category",
    NestedObject: schema.NestedAttributeObject{
        Attributes: map[string]schema.Attribute{
            "device_id": schema.StringAttribute{
                Required: true,
                MarkdownDescription: "GPU device ID (e.g., 0, 1, 2)",
            },
            "model": schema.StringAttribute{
                Optional: true,
                MarkdownDescription: "GPU model name (e.g., Tesla V100, A100)",
            },
            "compute_mode": schema.StringAttribute{
                Optional: true,
                MarkdownDescription: "Compute mode (default, exclusive, prohibited)",
            },
        },
    },
}
```

**Services Schema** (P3):
```go
// Structure TBD based on Phase 0 research
// If services structure is unknown, defer to POST-MVP
```

### API Contract Documentation

**File**: `specs/001-category-dynamic-fields/contracts/static-routes.json`
```json
{
  "terraformRequest": {
    "static_routes": [
      {
        "destination": "192.168.1.0/24",
        "gateway": "10.0.0.1",
        "metric": 100
      }
    ]
  },
  "bcmApiEntity": {
    "staticRoutes": [
      {
        "destination": "192.168.1.0/24",
        "gateway": "10.0.0.1",
        "metric": 100
      }
    ]
  }
}
```

Similar contracts for fsexports.json, roles.json, gpu-settings.json, services.json.

### Data Model Documentation

**File**: `specs/001-category-dynamic-fields/data-model.md`

```markdown
# Category Dynamic Fields Data Model

## Static Route
- **destination**: string (required) - CIDR notation network address
- **gateway**: string (required) - IPv4 gateway address
- **metric**: int64 (optional) - Route priority metric

## FS Export
- **path**: string (required) - Filesystem export path
- **network**: string (required) - Network UUID reference
- **allow_write**: bool (optional) - Write permission flag
- **root_squash**: bool (optional) - Root squash security flag
- **async**: bool (optional) - Async write mode flag

## Role
- **name**: string (required) - Role identifier
- **child_type**: string (required) - Role type classification
- **uuid**: string (computed) - BCM-assigned UUID
- **add_services**: bool (optional) - Auto-add services flag

## GPU Setting
- **device_id**: string (required) - GPU device identifier
- **model**: string (optional) - GPU model name
- **compute_mode**: string (optional) - Compute mode setting

## Service
- Structure TBD based on BCM API research
```

**Phase 1 Deliverables**:
- data-model.md with all entity schemas
- contracts/ directory with JSON examples for each field
- quickstart.md with developer setup instructions

---

## Phase 2: Implementation Planning (Not Executed by /speckit.plan)

**Note**: This section documents the implementation approach. The actual task breakdown is generated by `/speckit.tasks` command and output to `tasks.md`.

### Implementation Strategy

**Incremental Field Implementation**: Each field follows complete RED-GREEN-REFACTOR cycle before starting next field.

**Priority Order**:
1. static_routes (P1) - Highest business value, clear validation rules
2. fsexports (P1) - Critical for storage architecture
3. roles (P2) - Important but less frequently modified
4. gpu_settings (P3) - Specialized use case
5. services (P3) - May be deferred if structure unclear

### Per-Field Implementation Pattern

For each field (e.g., static_routes):

**RED Phase** (Write Failing Tests):
1. Define test config with field data
2. Write acceptance test: TestAccCMDeviceCategoryResource_StaticRoutes
3. Test scenarios:
   - Create with 2 routes
   - Read and verify routes persist
   - Update routes (add, modify, remove)
   - Import resource and verify routes
   - Idempotency check (no changes on re-apply)
   - Drift detection (external BCM API modification)
   - Empty list handling (set to [])
4. Run tests - all fail (schema not implemented)

**GREEN Phase** (Minimal Implementation):
1. Replace `types.Dynamic` with `ListNestedAttribute` in Schema method
2. Define nested object model struct (e.g., StaticRouteModel)
3. Update buildAPIEntity to serialize routes to BCM camelCase format
4. Update readCategory to deserialize routes from BCM API response
5. Handle empty list: use types.ListValueMust instead of types.ListNull
6. Run tests - all pass

**REFACTOR Phase** (Improve Quality):
1. Extract field mapping logic to helper functions if repetitive
2. Add detailed error messages for validation failures
3. Optimize serialization (avoid unnecessary allocations)
4. Add debug logging for field transformations
5. Update examples/ directory with usage patterns
6. Run tests - still pass

**Documentation Phase**:
1. Update examples/resources/bcm_cmdevice_category/*.tf
2. Run `make generate` to regenerate provider documentation
3. Verify docs/resources/bcm_cmdevice_category.md updated

### File Modifications per Field

**resource_cmdevice_category.go**:
- Line 80-96: Replace 5 types.Dynamic fields with ListNestedAttribute schemas
- Line 1207-1321: Update buildAPIEntity to handle field serialization
- Line 1323-1517: Update readCategory to handle field deserialization
- Add nested object model structs (StaticRouteModel, FSExportModel, etc.)

**resource_cmdevice_category_test.go**:
- Add 7 test functions per field:
  - TestAccCMDeviceCategoryResource_StaticRoutesCreate
  - TestAccCMDeviceCategoryResource_StaticRoutesUpdate
  - TestAccCMDeviceCategoryResource_StaticRoutesImport
  - TestAccCMDeviceCategoryResource_StaticRoutesIdempotency
  - TestAccCMDeviceCategoryResource_StaticRoutesDrift
  - TestAccCMDeviceCategoryResource_StaticRoutesEmpty
  - TestAccCMDeviceCategoryResource_StaticRoutesValidation
- Add test config helper functions per field

### Test Coverage Requirements

**Per Field** (35 test scenarios total for 5 fields):
1. Create: Resource created with field data persists correctly
2. Read: Field data readable from state matches BCM API
3. Update: Field modifications trigger plan and apply correctly
4. Delete: Resource deletion cleans up field data
5. Import: Imported resource preserves field data in state
6. Idempotency: No changes detected on subsequent applies
7. Drift: External BCM API changes detected by Terraform

### Build Sequence & Dependencies

```
Phase 0: Research (ALL fields in parallel)
├── Static Routes Research
├── FSExports Research
├── Roles Research
├── GPU Settings Research
└── Services Research
    └── research.md (consolidated output)

Phase 1: Design (depends on Phase 0)
├── Define all schemas
├── Document API contracts
└── Write data-model.md

Phase 2: Implementation (sequential per priority)
├── P1: static_routes
│   ├── RED: Tests
│   ├── GREEN: Implementation
│   ├── REFACTOR: Optimization
│   └── DOCS: Examples
├── P1: fsexports
│   ├── RED: Tests
│   ├── GREEN: Implementation
│   ├── REFACTOR: Optimization
│   └── DOCS: Examples
├── P2: roles
│   ├── RED: Tests
│   ├── GREEN: Implementation
│   ├── REFACTOR: Optimization
│   └── DOCS: Examples
├── P3: gpu_settings
│   ├── RED: Tests
│   ├── GREEN: Implementation
│   ├── REFACTOR: Optimization
│   └── DOCS: Examples
└── P3: services (optional)
    ├── RED: Tests
    ├── GREEN: Implementation
    ├── REFACTOR: Optimization
    └── DOCS: Examples
```

**Parallel Execution Opportunity**: After Phase 1, P1 fields (static_routes, fsexports) can be implemented in parallel if using separate branches.

---

## Risk Mitigation

### Technical Risks

**Risk 1: BCM API Eventual Consistency**
- **Impact**: Read after Create may not return newly set fields immediately
- **Mitigation**: Existing retry logic with exponential backoff (5 attempts, 1-16s delays)
- **Location**: resource_cmdevice_category.go lines 676-735 (Create method)
- **Test Coverage**: Idempotency tests verify fields persist correctly

**Risk 2: Field Name Mapping Errors**
- **Impact**: snake_case ↔ camelCase conversion bugs cause data loss
- **Mitigation**: Comprehensive drift detection tests catch mapping errors
- **Test Strategy**: External BCM API modifications verify bidirectional mapping
- **Example**: TestAccCMDeviceCategoryResource_StaticRoutesDrift

**Risk 3: Empty List vs Null Semantics**
- **Impact**: Empty arrays converted to null cause config drift
- **Mitigation**: Explicit empty list handling using types.ListValueMust
- **Test Coverage**: TestAccCMDeviceCategoryResource_*Empty for each field
- **Validation**: Idempotency check after setting field to []

**Risk 4: Unknown Services Structure**
- **Impact**: Cannot implement services field if BCM API structure unclear
- **Mitigation**: Mark services as P3, defer to POST-MVP if needed
- **Workaround**: Keep services as types.Dynamic if Phase 0 research inconclusive
- **Decision Point**: Phase 0 research completion

**Risk 5: Validation Complexity**
- **Impact**: Complex validators (CIDR, UUID) may have edge cases
- **Mitigation**: Dedicated validation test scenarios per field
- **Test Coverage**: TestAccCMDeviceCategoryResource_*Validation
- **Example**: Invalid CIDR notation should fail with clear error

### Process Risks

**Risk 6: Test Execution Time**
- **Impact**: 35+ new test scenarios may exceed 120m timeout
- **Mitigation**: Run field tests independently, prioritize P1/P2
- **Optimization**: Use `make testacc TESTARGS="-run TestAccCMDeviceCategory_StaticRoutes"`
- **Monitoring**: Track test execution time per field, optimize slow tests

**Risk 7: Breaking Changes to Existing Tests**
- **Impact**: Modifying CMDeviceCategoryResourceModel breaks existing tests
- **Mitigation**: Run full test suite after each field implementation
- **Validation**: All existing tests must pass before merging
- **Command**: `make testacc` (runs all provider tests)

---

## Success Criteria (from Spec)

This implementation plan addresses all success criteria from the feature specification:

✅ **SC-001**: Type safety - All dynamic types replaced with ListNestedAttribute
✅ **SC-002**: Test coverage - 7 test scenarios per field (35 total minimum)
✅ **SC-003**: API documentation - research.md documents all field structures
✅ **SC-004**: CRUD correctness - Serialization/deserialization in buildAPIEntity/readCategory
✅ **SC-005**: Empty list preservation - Explicit empty list handling in code
✅ **SC-006**: Drift detection - External modification tests for all fields
✅ **SC-007**: Field mapping - snake_case ↔ camelCase bidirectional conversion
✅ **SC-008**: Documentation - make generate updates docs after implementation
✅ **SC-009**: Null safety - Nested objects support null without runtime errors
✅ **SC-010**: Validation - CIDR and IP validators catch invalid formats

---

## Appendix: Helper Functions

### Field Mapping Helpers (Optional Enhancement)

If repetitive mapping logic emerges during implementation, consider extracting to helpers.go:

```go
// toCamelCase converts snake_case to camelCase for BCM API
func toCamelCase(s string) string {
    // Implementation TBD based on actual needs
}

// toSnakeCase converts camelCase to snake_case for Terraform
func toSnakeCase(s string) string {
    // Implementation TBD based on actual needs
}

// serializeStaticRoutes converts Terraform list to BCM API array
func serializeStaticRoutes(ctx context.Context, routes types.List) []interface{} {
    // Implementation TBD based on actual needs
}

// deserializeStaticRoutes converts BCM API array to Terraform list
func deserializeStaticRoutes(ctx context.Context, data []interface{}) (types.List, diag.Diagnostics) {
    // Implementation TBD based on actual needs
}
```

**Decision Point**: Refactor phase - extract helpers if 3+ fields share identical patterns.

---

## Next Steps

1. **Execute**: Run `.specify/scripts/bash/update-agent-context.sh copilot` to update AI context
2. **Research**: Execute Phase 0 research tasks to populate research.md
3. **Design**: Complete Phase 1 schema definitions and API contracts
4. **Tasks**: Run `/speckit.tasks` to generate detailed task breakdown in tasks.md
5. **Implement**: Execute tasks following RED-GREEN-REFACTOR TDD pattern
6. **Analyze**: Run `/speckit.analyze` to verify cross-artifact consistency
7. **Document**: Generate provider docs with `make generate`

**Estimated Timeline**:
- Phase 0 Research: 2-4 hours (API exploration)
- Phase 1 Design: 2-3 hours (schema definition)
- Phase 2 Implementation: 16-24 hours (5 fields × 4 hours average)
- Total: 20-31 hours

**Branch Strategy**: Single feature branch `001-category-dynamic-fields` for all fields, or separate branches per priority tier if parallel development desired.
