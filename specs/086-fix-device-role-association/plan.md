# Implementation Plan: Fix Device Role Association Bug

**Branch**: `086-fix-device-role-association` | **Date**: 2025-11-26 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/086-fix-device-role-association/spec.md`

## Summary

Fix the device role association to accept role names (e.g., "backup", "provisioning") directly instead of requiring UUID lookups. The current implementation forces users to query role UUIDs via the `bcm_cmdevice_roles` data source, which is verbose and error-prone. The fix will resolve role names to full role objects internally, with client-side validation to provide clear error messages for invalid role names. Backward compatibility with UUID input is maintained.

## Technical Context

**Language/Version**: Go 1.24.0
**Primary Dependencies**: terraform-plugin-framework v1.16.1, terraform-plugin-testing v1.13.3
**Storage**: N/A (BCM API handles persistence)
**Testing**: TF_ACC=1 acceptance tests with real BCM cluster (172.21.15.254:8081)
**Target Platform**: Linux/macOS/Windows (Terraform provider binary)
**Project Type**: Single project (Terraform provider)
**Performance Goals**: Role validation and resolution <500ms during plan/apply
**Constraints**: Must maintain backward compatibility with existing UUID-based configurations
**Scale/Scope**: Single resource modification (`bcm_cmdevice_device`), single data source (`bcm_cmdevice_roles`)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Gate | Status | Notes |
|------|--------|-------|
| TDD-first approach | PASS | Tests written before implementation per CLAUDE.md |
| Acceptance tests required | PASS | Drift, import, CRUD tests planned |
| Parallel execution patterns | PASS | Single resource modification, no parallelism needed |
| Backward compatibility | PASS | UUID input remains supported |
| Documentation updates | PASS | Example and schema documentation will be updated |
| Pre-commit hooks | PASS | Will run before commits |

## Project Structure

### Documentation (this feature)

```text
specs/086-fix-device-role-association/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output - role resolution patterns
├── data-model.md        # Phase 1 output - role input/output model
├── quickstart.md        # Phase 1 output - developer implementation guide
├── contracts/           # Phase 1 output - validation contract
│   └── role-validation.md
└── tasks.md             # Phase 2 output (/speckit.tasks command)
```

### Source Code (repository root)

```text
internal/provider/
├── resource_cmdevice_device.go      # MODIFY: lookupAndBuildRolesForEntity(), parseRolesFromAPI()
├── resource_cmdevice_device_test.go # ADD: acceptance tests for role name input
├── data_source_cmdevice_roles.go    # READ-ONLY: used for role discovery

examples/resources/bcm_cmdevice_device/
├── with_roles.tf                    # MODIFY: simplify to use role names directly
└── basic.tf                         # UNCHANGED

docs/resources/cmdevice_device.md    # AUTO-GENERATED via make generate
```

**Structure Decision**: This is a modification to an existing resource within a single-project Terraform provider. No new files are needed except for the specification artifacts.

## Complexity Tracking

> No constitution violations detected.

---

## Phase 0: Research & Unknowns Resolution

### Research Tasks

1. **Role Name Uniqueness**: Confirm BCM role names are unique within a cluster
2. **UUID Detection Pattern**: Determine reliable regex to distinguish UUID from role name
3. **API Query Efficiency**: Confirm getNodes is the only way to discover roles
4. **Role Object Structure**: Document the required role object fields for assignment

### Findings (see research.md for details)

Completed research documented in [research.md](./research.md).

---

## Phase 1: Design

### 1.1 Role Input Resolution Strategy

**Decision**: Implement a "name-first, UUID-fallback" resolution strategy.

**Algorithm**:
1. For each input string in the `roles` set:
   a. Check if string matches UUID format (regex: `^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
   b. If UUID format: look up role by UUID (existing behavior)
   c. If NOT UUID format: look up role by name (new behavior)
2. Build role-name-to-object map from getNodes response for efficient lookups
3. Return clear error if role not found (by name or UUID)

**Rationale**: This approach provides full backward compatibility while making name-based input the primary (and recommended) method.

### 1.2 Validation Strategy

**Client-Side Pre-flight Validation**:
- Query all roles from BCM via `getNodes` (already done in `lookupAndBuildRolesForEntity`)
- Build both name->role and uuid->role lookup maps
- Validate all input role identifiers exist BEFORE any API call
- Return user-friendly error with specific invalid role identifier(s)

**Error Message Format**:
```
Role 'invalid-role-name' does not exist in the BCM cluster.
Available roles: backup, provisioning, monitoring, boot
Use the `bcm_cmdevice_roles` data source to discover available roles.
```

### 1.3 State Representation

**Decision**: Store role NAMES in Terraform state (not UUIDs).

**Rationale**:
- Users configure roles by name (new approach)
- State should reflect what users configured
- Import should show human-readable role names
- Eliminates UUID lookup complexity for users reading state

**Implementation**:
- `parseRolesFromAPI()` returns role names (from `name` field)
- `lookupAndBuildRolesForEntity()` accepts both names and UUIDs but resolves to full role objects
- State always contains role names after Read

### 1.4 Import Behavior

When importing a device with roles:
1. Read device from BCM API
2. Extract role objects from response
3. Return role names (not UUIDs) in state
4. User can then manage roles by name in their configuration

### 1.5 Example Update

**Before** (verbose UUID lookup):
```hcl
data "bcm_cmdevice_roles" "all" {}

locals {
  backup_role = [for r in data.bcm_cmdevice_roles.all.roles : r.uuid if r.name == "backup"][0]
}

resource "bcm_cmdevice_device" "node" {
  roles = [local.backup_role]
}
```

**After** (simple name-based):
```hcl
resource "bcm_cmdevice_device" "node" {
  roles = ["backup", "provisioning"]
}
```

### 1.6 Schema Documentation Update

Update `roles` attribute MarkdownDescription to:
```
Set of role names assigned to this device. Roles define the device's function in the cluster
(e.g., "backup", "provisioning", "boot"). Use the `bcm_cmdevice_roles` data source to
discover available roles. Role names are case-sensitive. For backward compatibility,
role UUIDs are also accepted but role names are recommended for readability.
```

---

## Phase 2: Implementation Approach

### 2.1 Modify `lookupAndBuildRolesForEntity()`

Current function signature and behavior:
```go
func (r *CMDeviceDeviceResource) lookupAndBuildRolesForEntity(
    ctx context.Context,
    plan CMDeviceDeviceResourceModel,
    entity map[string]interface{},
) error
```

Changes needed:
1. Build both `rolesByName` and `rolesByUUID` maps from getNodes response
2. Add `isUUID()` helper function
3. Resolve each input identifier via name or UUID lookup
4. Update error messages to include available role names

### 2.2 Modify `parseRolesFromAPI()`

Current behavior: Returns role UUIDs as `types.Set`

New behavior: Returns role NAMES as `types.Set`

```go
func parseRolesFromAPI(rolesData interface{}) types.Set {
    // Extract role names (not UUIDs) from array
    for _, roleItem := range rolesArray {
        if role, ok := roleItem.(map[string]interface{}); ok {
            if name, ok := role["name"].(string); ok && name != "" {
                roleNames = append(roleNames, name)
            }
        }
    }
    // Return set of role names
}
```

### 2.3 Add Helper Function

```go
// isUUID returns true if the string matches UUID format
func isUUID(s string) bool {
    uuidRegex := regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
    return uuidRegex.MatchString(s)
}
```

### 2.4 Update Example File

Simplify `/workspace/examples/resources/bcm_cmdevice_device/with_roles.tf` to demonstrate name-based role assignment without UUID lookups.

### 2.5 Update Schema Documentation

Modify the `roles` attribute description in the Schema() method.

---

## Phase 3: Testing Strategy (TDD)

### 3.1 New Acceptance Tests

| Test Name | Purpose |
|-----------|---------|
| `TestAccCMDeviceDevice_RolesByName` | Create device with roles specified by name |
| `TestAccCMDeviceDevice_RolesByUUID` | Verify backward compatibility with UUID input |
| `TestAccCMDeviceDevice_RolesByNameAndUUID` | Mixed input (both names and UUIDs) |
| `TestAccCMDeviceDevice_InvalidRoleName` | Verify clear error for non-existent role name |
| `TestAccCMDeviceDevice_RolesImport` | Verify imported device shows role names in state |
| `TestAccCMDeviceDevice_RolesUpdate` | Update roles from one set to another |
| `TestAccCMDeviceDevice_RolesDrift` | Detect drift when roles changed externally |

### 3.2 Test Configuration Pattern

```go
func testAccCMDeviceDeviceConfigWithRoleNames(hostname string, roles []string) string {
    rolesStr := fmt.Sprintf(`[%s]`, strings.Join(
        func() []string {
            quoted := make([]string, len(roles))
            for i, r := range roles {
                quoted[i] = fmt.Sprintf(`"%s"`, r)
            }
            return quoted
        }(),
        ", ",
    ))

    return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

resource "bcm_cmdevice_device" "test" {
  hostname           = %[4]q
  mac                = "00:11:22:33:44:FF"
  category           = "..." // test category UUID
  management_network = "..." // test network UUID
  roles              = %[5]s
}
`,
        os.Getenv("BCM_ENDPOINT"),
        os.Getenv("BCM_USERNAME"),
        os.Getenv("BCM_PASSWORD"),
        hostname,
        rolesStr,
    )
}
```

### 3.3 Error Case Test

```go
func TestAccCMDeviceDevice_InvalidRoleName(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            {
                Config:      testAccCMDeviceDeviceConfigWithRoleNames("test-node", []string{"nonexistent-role"}),
                ExpectError: regexp.MustCompile(`Role 'nonexistent-role' does not exist`),
            },
        },
    })
}
```

---

## Success Metrics

| Metric | Target | Verification |
|--------|--------|--------------|
| Role assignment by name | Works | Acceptance test passes |
| Invalid role error | Clear message | Error test passes |
| Backward UUID compatibility | Works | UUID acceptance test passes |
| Import shows names | Role names in state | Import test passes |
| Example simplified | No UUID lookups | Manual review |
| Documentation updated | Reflects name input | make generate succeeds |

---

## Implementation Order

1. **Tests First (RED)**
   - Write `TestAccCMDeviceDevice_RolesByName` (expected to fail initially)
   - Write `TestAccCMDeviceDevice_InvalidRoleName` (expected to fail initially)

2. **Implementation (GREEN)**
   - Add `isUUID()` helper
   - Modify `lookupAndBuildRolesForEntity()` for name resolution
   - Modify `parseRolesFromAPI()` to return names
   - Update schema documentation

3. **Example & Docs (REFACTOR)**
   - Update `with_roles.tf` example
   - Run `make generate` for documentation
   - Verify all tests pass

4. **Final Verification**
   - Run full acceptance test suite
   - Run `make lint`
   - Run `pre-commit run --all-files`
