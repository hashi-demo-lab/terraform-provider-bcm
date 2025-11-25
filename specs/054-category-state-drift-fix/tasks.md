# Tasks: Category Resource State Drift Fix

## Issue Reference
- **GitHub Issue**: [#54](https://github.com/hashi-demo-lab/terraform-provider-bcm/issues/54)
- **Specification**: `specs/054-category-state-drift-fix/spec.md`
- **Plan**: `specs/054-category-state-drift-fix/plan.md`

## Task Breakdown

### Phase 1: Add Plan Modifiers to Schema

#### T001: Add Plan Modifier Imports
**File**: `internal/provider/resource_cmdevice_category.go`
**Status**: [ ] Not Started
**Dependencies**: None

**Changes**:
```go
// Add after line 22 (existing imports)
import (
    "github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)
```

**Verification**: `go build ./...` succeeds

---

#### T002: Add UseStateForUnknown to `id` Field
**File**: `internal/provider/resource_cmdevice_category.go`
**Status**: [ ] Not Started
**Dependencies**: T001

**Location**: Schema() function, `id` attribute (line ~190)

**Before**:
```go
"id": schema.StringAttribute{
    Computed:            true,
    MarkdownDescription: "Resource identifier (same as UUID)",
},
```

**After**:
```go
"id": schema.StringAttribute{
    Computed:            true,
    MarkdownDescription: "Resource identifier (same as UUID)",
    PlanModifiers: []planmodifier.String{
        stringplanmodifier.UseStateForUnknown(),
    },
},
```

**Verification**: `go build ./...` succeeds

---

#### T003: Add UseStateForUnknown to `uuid` Field
**File**: `internal/provider/resource_cmdevice_category.go`
**Status**: [ ] Not Started
**Dependencies**: T001

**Location**: Schema() function, `uuid` attribute (line ~194)

**Before**:
```go
"uuid": schema.StringAttribute{
    Computed:            true,
    MarkdownDescription: "Unique identifier assigned by BCM",
},
```

**After**:
```go
"uuid": schema.StringAttribute{
    Computed:            true,
    MarkdownDescription: "Unique identifier assigned by BCM",
    PlanModifiers: []planmodifier.String{
        stringplanmodifier.UseStateForUnknown(),
    },
},
```

---

#### T004: Add UseStateForUnknown to Computed Metadata Fields
**File**: `internal/provider/resource_cmdevice_category.go`
**Status**: [ ] Not Started
**Dependencies**: T001

**Fields to update** (lines ~544-567):
- `parent_uuid` - String
- `revision` - String
- `modified` - Bool
- `to_be_removed` - Bool
- `base_type` - String
- `child_type` - String

**Example for Bool field**:
```go
"modified": schema.BoolAttribute{
    Computed:            true,
    MarkdownDescription: "Modified flag",
    PlanModifiers: []planmodifier.Bool{
        boolplanmodifier.UseStateForUnknown(),
    },
},
```

---

#### T005: Add UseStateForUnknown to `software_image_proxy` Computed Fields
**File**: `internal/provider/resource_cmdevice_category.go`
**Status**: [ ] Not Started
**Dependencies**: T001

**Location**: Schema() function, `software_image_proxy` nested attribute (lines ~219-236)

**Fields to update**:
- `uuid` (line ~223) - String
- `revision_id` (line ~230) - Int64

**Example**:
```go
"uuid": schema.StringAttribute{
    Computed:            true,
    MarkdownDescription: "Unique identifier",
    PlanModifiers: []planmodifier.String{
        stringplanmodifier.UseStateForUnknown(),
    },
},
"revision_id": schema.Int64Attribute{
    Computed:            true,
    MarkdownDescription: "Revision identifier",
    PlanModifiers: []planmodifier.Int64{
        int64planmodifier.UseStateForUnknown(),
    },
},
```

---

### Phase 2: Preserve `software_image_proxy` in Read

#### T006: Preserve `software_image_proxy` from State in Read Function
**File**: `internal/provider/resource_cmdevice_category.go`
**Status**: [ ] Not Started
**Dependencies**: T001-T005

**Location**: Read() function, after readCategory() call (line ~800)

**Pattern Reference**: `resource_cmpart_softwareimage.go` lines 466-482

**Implementation**:
```go
// In Read() function, before resp.State.Set():
// Preserve software_image_proxy.parent_software_image from prior state
// BCM API may return different reference on subsequent reads
originalSoftwareImageProxy := state.SoftwareImageProxy
// ... after readCategory() call ...
if !originalSoftwareImageProxy.IsNull() && !originalSoftwareImageProxy.IsUnknown() {
    state.SoftwareImageProxy = originalSoftwareImageProxy
    tflog.Debug(ctx, "Preserved software_image_proxy from state", map[string]interface{}{
        "preserved": true,
    })
}
```

---

#### T007: Preserve `software_image_proxy` in Create After Read
**File**: `internal/provider/resource_cmdevice_category.go`
**Status**: [ ] Not Started
**Dependencies**: T006

**Location**: Create() function, after readCategory() retry loop (line ~774)

**Implementation**: Similar to T006 - preserve plan's software_image_proxy after readCategory()

---

#### T008: Preserve `software_image_proxy` in Update After Read
**File**: `internal/provider/resource_cmdevice_category.go`
**Status**: [ ] Not Started
**Dependencies**: T006

**Location**: Update() function, after readCategory() retry loop (line ~1045)

**Implementation**: Similar to T006 - preserve plan's software_image_proxy after readCategory()

---

### Phase 3: Testing

#### T009: Run Basic Category Test
**Status**: [ ] Not Started
**Dependencies**: T001-T008

**Command**:
```bash
TF_ACC=1 BCM_ENDPOINT="https://172.21.15.254:8081" \
BCM_USERNAME="root" BCM_PASSWORD="Hashicorp123!" \
go test -v -timeout 30m ./internal/provider/ -run "^TestAccCMDeviceCategoryResource_Basic$"
```

**Success Criteria**: All 5 steps pass, including idempotency checks

---

#### T010: Run Full Category Test Suite
**Status**: [ ] Not Started
**Dependencies**: T009

**Command**:
```bash
TF_ACC=1 BCM_ENDPOINT="https://172.21.15.254:8081" \
BCM_USERNAME="root" BCM_PASSWORD="Hashicorp123!" \
go test -v -timeout 120m ./internal/provider/ -run "CMDeviceCategory"
```

**Success Criteria**: All 16 category tests pass

---

#### T011: Run Lint and Unit Tests
**Status**: [ ] Not Started
**Dependencies**: T001-T008

**Commands**:
```bash
make lint
make test
```

**Success Criteria**: No lint errors, all unit tests pass

---

### Phase 4: Documentation and Cleanup

#### T012: Update CLAUDE.md if Needed
**Status**: [ ] Not Started
**Dependencies**: T010, T011

**Action**: Document any new patterns discovered during implementation

---

#### T013: Generate Provider Documentation
**Status**: [ ] Not Started
**Dependencies**: T010, T011

**Command**:
```bash
make generate
```

**Success Criteria**: Documentation regenerated without errors

---

## Task Summary

| Task | Description | Status | Dependencies |
|------|-------------|--------|--------------|
| T001 | Add plan modifier imports | [ ] | None |
| T002 | Add UseStateForUnknown to `id` | [ ] | T001 |
| T003 | Add UseStateForUnknown to `uuid` | [ ] | T001 |
| T004 | Add UseStateForUnknown to metadata fields | [ ] | T001 |
| T005 | Add UseStateForUnknown to software_image_proxy fields | [ ] | T001 |
| T006 | Preserve software_image_proxy in Read | [ ] | T001-T005 |
| T007 | Preserve software_image_proxy in Create | [ ] | T006 |
| T008 | Preserve software_image_proxy in Update | [ ] | T006 |
| T009 | Run Basic Category Test | [ ] | T001-T008 |
| T010 | Run Full Category Test Suite | [ ] | T009 |
| T011 | Run Lint and Unit Tests | [ ] | T001-T008 |
| T012 | Update CLAUDE.md | [ ] | T010, T011 |
| T013 | Generate Documentation | [ ] | T010, T011 |

## Execution Order

1. T001 (imports)
2. T002, T003, T004, T005 (schema changes - can be done in one edit)
3. T006, T007, T008 (Read preservation - can be done in one edit)
4. T011 (lint/test)
5. T009 (basic acceptance test)
6. T010 (full acceptance test suite)
7. T012, T013 (documentation)
