# Implementation Plan: BCM Device Resource Management

**Branch**: `008-cmdevice-device-resource` | **Date**: 2025-11-22 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/workspace/specs/008-cmdevice-device-resource/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Implement a Terraform resource (`bcm_cmdevice_device`) for managing individual devices (compute nodes) in BCM clusters. This resource enables Infrastructure as Code management of cluster topology, supporting full CRUD operations, import functionality, and drift detection. The implementation follows TDD RED-GREEN-REFACTOR principles with modern terraform-plugin-testing patterns (v1.13.3+), leveraging the BCM JSON-RPC API with efficient direct lookup via the `getDevice` method with args parameter support.

**Primary Value**: Allow cluster administrators to programmatically define and manage cluster infrastructure using Terraform, replacing manual device creation in BCM UI with declarative configuration.

**Technical Approach**: Build on established provider patterns (bcm_client.go, test_helpers.go) using terraform-plugin-framework v1.16.1 with comprehensive schema validation (RFC 1123 hostname, MAC address format, UUID references), full CRUD lifecycle with import state support, and drift detection using external BCM API modifications.

## Technical Context

**Language/Version**: Go 1.24.0
**Primary Dependencies**: terraform-plugin-framework v1.16.1, terraform-plugin-testing v1.13.3, terraform-plugin-log
**Storage**: BCM cluster state via JSON-RPC API (stateless provider, state in Terraform)
**Testing**: terraform-plugin-testing acceptance tests (TF_ACC=1), Go unit tests with table-driven patterns
**Target Platform**: Linux (provider binary), BCM cluster API endpoint (https://172.21.15.254:8081)
**Project Type**: Terraform Provider (single binary, plugin-based architecture)
**Performance Goals**: <2s resource Create/Read/Update operations, <120m total acceptance test suite execution
**Constraints**: BCM API eventual consistency (2s wait after external modifications), cookie-based authentication session management, self-signed TLS certificates (insecure_skip_verify=true)
**Scale/Scope**: Manage hundreds of devices per cluster, support parallel test execution (4-8 concurrent acceptance tests), enterprise production clusters

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**Constitution Principle**: Test-Driven Development (RED-GREEN-REFACTOR)

**Gate Evaluation**:

| Principle | Status | Evidence | Action Required |
|-----------|--------|----------|-----------------|
| Tests before implementation | ✅ PASS | Spec includes comprehensive test patterns (lines 177-391), acceptance test structure defined, drift detection test pattern documented | Write failing acceptance tests in Phase 0 before implementation |
| RED-GREEN-REFACTOR cycle | ✅ PASS | TDD workflow documented in AGENTS.md, parallel execution patterns established, existing resources follow this pattern | Follow established TDD cycle: 1) Write failing tests, 2) Minimal implementation, 3) Refactor with tests green |
| Comprehensive test coverage | ✅ PASS | Required coverage defined: Basic CRUD, Import, Update, Idempotency checks, Drift detection, CheckDestroy, Validation errors | Implement all required test scenarios from spec |
| Modern testing patterns | ✅ PASS | Uses terraform-plugin-testing v1.13.3+ with statecheck.ExpectKnownValue, plancheck.ExpectEmptyPlan, compare.ValuesSame for ID tracking | Follow modern patterns throughout implementation |

**Gate Decision**: ✅ PROCEED - All constitutional requirements satisfied. Feature aligns with TDD principles and established testing patterns.

**Complexity Justification**: None required - implementation follows existing provider patterns without introducing architectural complexity.

## Project Structure

### Documentation (this feature)

```text
specs/008-cmdevice-device-resource/
├── spec.md              # Feature specification (input)
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output - API research and validation patterns
├── data-model.md        # Phase 1 output - Device entity schema and relationships
├── quickstart.md        # Phase 1 output - Developer quick start guide
├── contracts/           # Phase 1 output - OpenAPI/JSON-RPC contracts
│   └── device-api.json  # BCM JSON-RPC API contract for cmdevice service
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
terraform-provider-bcm/
├── internal/
│   └── provider/
│       ├── provider.go                      # Provider registration (add device resource)
│       ├── bcm_client.go                    # JSON-RPC client (already supports args parameter)
│       ├── test_helpers.go                  # Test utilities (already supports device operations)
│       │
│       ├── resource_cmdevice_device.go      # NEW: Device resource implementation
│       ├── resource_cmdevice_device_test.go # NEW: Device acceptance tests
│       │
│       ├── resource_cmdevice_category.go    # REFERENCE: Similar resource pattern
│       ├── resource_cmpart_softwareimage.go # REFERENCE: Similar resource pattern
│       └── data_source_cmdevice_nodes.go    # REFERENCE: getNodes API usage
│
├── examples/
│   └── resources/
│       └── bcm_cmdevice_device/             # NEW: Example configurations
│           ├── resource.tf                  # Basic device example
│           ├── compute-node.tf              # Compute node example
│           └── head-node.tf                 # Head node example
│
├── docs/
│   └── resources/
│       └── bcm_cmdevice_device.md           # GENERATED: Resource documentation
│
├── sampleRest/
│   ├── DeviceEntity.md                      # REFERENCE: BCM API documentation
│   └── CMDevice_Complete_Documentation.md   # REFERENCE: Complete API reference
│
└── specs/
    └── 008-cmdevice-device-resource/        # This feature's documentation
```

**Structure Decision**: Single Terraform Provider project following Go plugin architecture. Implementation adds new resource to existing `internal/provider/` package, leveraging established BCM client and test helper patterns. No structural changes required - purely additive implementation within existing provider framework.

## Complexity Tracking

No violations detected - implementation follows established patterns without introducing architectural complexity.

---

## Phase 0: Research & Discovery

**Goal**: Resolve unknowns from Technical Context, research BCM device API patterns, document validation approaches.

### Research Tasks

#### RT-001: BCM Device API Validation
- **Objective**: Confirm API methods (addDevice, updateDevice, removeDevice, getDevice) and args parameter support
- **Method**: Test API calls using existing BCM client with test cluster
- **Output**: Confirmed API signatures, error response patterns, async operation behavior

#### RT-002: Device Type (childType) Assignment Logic
- **Objective**: Understand how BCM determines childType (HeadNode vs ComputeNode vs PhysicalNode)
- **Method**: Examine existing devices in test cluster, test device creation with different role configurations
- **Output**: Document childType assignment rules (likely based on roles assigned)

#### RT-003: Force Parameter Scenarios
- **Objective**: Identify specific scenarios requiring force=true parameter
- **Method**: Test device updates/deletions in various states (active, provisioning, dependencies)
- **Output**: Document force parameter requirements for Create/Update/Delete operations

#### RT-004: Network Interface Configuration Patterns
- **Objective**: Research network interface ordering, required vs optional fields, validation rules
- **Method**: Examine existing device configurations, test interface creation patterns
- **Output**: Document interface schema requirements, ordering constraints

#### RT-005: Validation Pattern Best Practices
- **Objective**: Research Terraform Plugin Framework validators for hostname (RFC 1123), MAC address, UUID
- **Method**: Review terraform-plugin-framework-validators package, existing resource patterns
- **Output**: Document validator implementations for all validation requirements (FR-003 through FR-005)

### Research Output: research.md

Document structure:
```markdown
# BCM Device Resource Research

## API Method Signatures
- addDevice([device_entity, force]) - Confirmed signature and response format
- getDevice([identifier]) - Efficient direct lookup with args parameter
- updateDevice([device_entity, force]) - Update behavior and response
- removeDevice([uuid, force]) - Deletion behavior and cleanup

## Device Type Assignment
- childType determination logic (roles-based vs explicit)
- Common device types: HeadNode, ComputeNode, PhysicalNode, CloudNode, StorageNode

## Force Parameter Requirements
- Create scenarios requiring force=true
- Update scenarios requiring force=true
- Delete scenarios requiring force=true

## Network Interface Patterns
- Interface ordering requirements
- Required vs optional fields
- Validation rules and constraints

## Validation Implementations
- RFC 1123 hostname validation (stringvalidator.RegexMatches)
- MAC address validation pattern
- UUID validation (RFC 4122)
- Network/category reference validation

## Error Handling Patterns
- BCM API error response parsing
- Terraform diagnostic message formatting
- Async operation polling strategy (if needed)
```

---

## Phase 1: Design & Contracts

**Goal**: Generate data model, API contracts, quickstart guide, and update agent context.

### Design Artifacts

#### D-001: data-model.md

Define complete Device entity schema with:

**Core Entities**:
- **Device**: Primary resource (hostname, uuid, mac, category, managementNetwork)
- **NetworkInterface**: Nested interface configuration (name, mac, ip, network reference)
- **Role**: Service role assignments (HeadNodeRole, ComputeRole, StorageRole, etc.)
- **BMCSettings**: Hardware management configuration (username, password, powerControl)

**Entity Relationships**:
- Device → Category (UUID reference, required)
- Device → ManagementNetwork (UUID reference, required)
- Device → NetworkInterface[] (nested objects, optional)
- Device → Role[] (nested objects, optional)
- NetworkInterface → Network (UUID reference)

**Field Mappings** (Terraform snake_case → BCM camelCase):
```
hostname → hostname (no change)
mac → mac (no change)
category → category (UUID reference)
management_network → managementNetwork
kernel_parameters → kernelParameters
boot_loader → bootLoader
bmc_settings → bmcSettings
```

**Validation Rules**:
- hostname: RFC 1123 DNS label (lowercase, alphanumeric + hyphens, 1-63 chars)
- mac: Six groups of two hex digits with colons (00:11:22:33:44:55)
- category: Valid UUID (RFC 4122)
- management_network: Valid UUID (RFC 4122)

#### D-002: contracts/device-api.json

JSON-RPC API contract:
```json
{
  "service": "cmdevice",
  "methods": {
    "addDevice": {
      "args": ["device_entity", "force"],
      "returns": "device_uuid",
      "description": "Create new device in BCM cluster"
    },
    "getDevice": {
      "args": ["hostname_or_uuid"],
      "returns": "device_entity",
      "description": "Retrieve device by hostname or UUID"
    },
    "updateDevice": {
      "args": ["device_entity", "force"],
      "returns": "success",
      "description": "Update existing device configuration"
    },
    "removeDevice": {
      "args": ["uuid", "force"],
      "returns": "success",
      "description": "Remove device from cluster"
    }
  }
}
```

#### D-003: quickstart.md

Developer quick start guide:
```markdown
# BCM Device Resource Quick Start

## Prerequisites
- Go 1.24+
- terraform-plugin-framework v1.16.1
- BCM test cluster access

## Implementation Workflow

### Phase 1: RED (Failing Tests)
1. Create resource_cmdevice_device_test.go
2. Write TestAccCMDeviceDeviceResource_Basic
3. Write TestAccCMDeviceDevice_DriftDetection
4. Run: TF_ACC=1 go test -v ./internal/provider/ -run TestAccCMDeviceDevice
5. Verify: All tests fail (resource not implemented)

### Phase 2: GREEN (Minimal Implementation)
1. Create resource_cmdevice_device.go
2. Implement Schema() with required fields only
3. Implement Create() with minimal BCM API call
4. Implement Read() with getDevice direct lookup
5. Implement Update() with updateDevice
6. Implement Delete() with removeDevice
7. Implement ImportState() with UUID passthrough
8. Run tests: Verify all pass

### Phase 3: REFACTOR (Full Implementation)
1. Add optional fields to schema
2. Add validation (hostname, MAC, UUID)
3. Add error handling and diagnostics
4. Add computed field management
5. Implement buildDeviceEntity helper
6. Run tests: Verify still passing

## Testing
BCM_ENDPOINT="https://..." TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceDevice

## Example Configuration
resource "bcm_cmdevice_device" "test" {
  hostname           = "compute001"
  mac                = "00:11:22:33:44:55"
  category           = bcm_cmdevice_category.default.uuid
  management_network = data.bcm_cmnet_network.mgmt.uuid
}
```

#### D-004: Update Agent Context

Run update script:
```bash
.specify/scripts/bash/update-agent-context.sh copilot
```

Add to agent context:
- Device resource schema pattern
- BCM device API usage (cmdevice service)
- Validation patterns (hostname RFC 1123, MAC address, UUID)
- Test helper usage (createTestBCMClient, getResourceUUIDByName, verifyResourceDeleted)

---

## Phase 2: Implementation Architecture

### Schema Design

**Resource Model Structure**:
```go
type CMDeviceDeviceResourceModel struct {
    // Identity (Required)
    ID       types.String `tfsdk:"id"`        // Computed, same as UUID
    UUID     types.String `tfsdk:"uuid"`      // Computed, BCM-assigned
    Hostname types.String `tfsdk:"hostname"`  // Required, RFC 1123 validation
    MAC      types.String `tfsdk:"mac"`       // Required, MAC address validation

    // References (Required)
    Category          types.String `tfsdk:"category"`           // Required, UUID validation
    ManagementNetwork types.String `tfsdk:"management_network"` // Required, UUID validation

    // Boot Configuration (Optional)
    BootLoader         types.String `tfsdk:"boot_loader"`          // Optional
    BootLoaderProtocol types.String `tfsdk:"boot_loader_protocol"` // Optional
    KernelParameters   types.String `tfsdk:"kernel_parameters"`    // Optional

    // Network Configuration (Optional)
    Interfaces types.List `tfsdk:"interfaces"` // Optional, list of NetworkInterfaceModel

    // Hardware Management (Optional)
    BMCSettings types.Object `tfsdk:"bmc_settings"` // Optional, nested BMCSettingsModel
    PowerControl types.String `tfsdk:"power_control"` // Optional

    // Role Assignments (Optional)
    Roles types.List `tfsdk:"roles"` // Optional, list of RoleModel

    // Metadata (Optional)
    Notes types.String `tfsdk:"notes"` // Optional

    // Force Parameter (Optional)
    Force types.Bool `tfsdk:"force"` // Optional, default: false

    // Computed Fields
    CreationTime types.Int64  `tfsdk:"creation_time"` // Computed
    BaseType     types.String `tfsdk:"base_type"`     // Computed, always "Device"
    ChildType    types.String `tfsdk:"child_type"`    // Computed, BCM-determined
}
```

**Nested Object Models**:
```go
type NetworkInterfaceModel struct {
    Name    types.String `tfsdk:"name"`    // Required
    MAC     types.String `tfsdk:"mac"`     // Required
    IP      types.String `tfsdk:"ip"`      // Optional
    Network types.String `tfsdk:"network"` // Required, UUID reference
    DHCP    types.Bool   `tfsdk:"dhcp"`    // Optional
}

type RoleModel struct {
    Name        types.String `tfsdk:"name"`         // Required
    UUID        types.String `tfsdk:"uuid"`         // Computed
    AddServices types.Bool   `tfsdk:"add_services"` // Optional
}

type BMCSettingsModel struct {
    UserName  types.String `tfsdk:"user_name"`  // Optional
    Password  types.String `tfsdk:"password"`   // Optional, sensitive
    Privilege types.String `tfsdk:"privilege"`  // Optional
}
```

### Schema Validators

**Hostname Validation** (FR-003):
```go
stringvalidator.RegexMatches(
    regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`),
    "hostname must be RFC 1123 DNS label (lowercase alphanumeric and hyphens, 1-63 chars)",
)
```

**MAC Address Validation** (FR-004):
```go
stringvalidator.RegexMatches(
    regexp.MustCompile(`^([0-9A-Fa-f]{2}:){5}[0-9A-Fa-f]{2}$`),
    "mac must be six groups of two hexadecimal digits separated by colons",
)
```

**UUID Validation** (FR-005):
```go
stringvalidator.RegexMatches(
    regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`),
    "must be valid UUID (RFC 4122)",
)
```

### CRUD Implementation Strategy

#### Create Operation
1. **Validate Plan**: Extract plan data, validate required fields
2. **Build Entity**: Construct BCM device entity JSON with required fields
3. **API Call**: `client.CallJSONRPC(ctx, "cmdevice", "addDevice", deviceEntity, forceValue)`
4. **Parse Response**: Extract UUID from response
5. **Read-After-Create**: Call getDevice to populate computed fields
6. **Set State**: Populate all attributes including computed fields

**Error Handling**:
- Duplicate hostname: BCM returns error, add diagnostic with clear message
- Invalid references: BCM returns error, add diagnostic with validation context
- API timeout: Return diagnostic with retry suggestion

#### Read Operation
1. **Get Current State**: Extract ID/UUID from state
2. **API Call**: `client.CallJSONRPC(ctx, "cmdevice", "getDevice", identifier)` (efficient direct lookup)
3. **Handle Not Found**: Empty response → remove from state (resource deleted externally)
4. **Parse Response**: Unmarshal device entity JSON
5. **Map to State**: Convert BCM camelCase to Terraform snake_case, handle null values
6. **Set State**: Update all attributes, preserve plan values for force-only fields

**Drift Detection**: Read operation detects external modifications by comparing BCM response to current state.

#### Update Operation
1. **Get Plan**: Extract updated configuration from plan
2. **Build Entity**: Construct complete device entity with UUID from state
3. **API Call**: `client.CallJSONRPC(ctx, "cmdevice", "updateDevice", deviceEntity, forceValue)`
4. **Read-After-Update**: Call getDevice to refresh state
5. **Set State**: Update all attributes with post-update values

**Force Parameter**: If BCM rejects update (e.g., device has active provisioning), error message suggests setting `force = true`.

#### Delete Operation
1. **Get State**: Extract UUID from state
2. **API Call**: `client.CallJSONRPC(ctx, "cmdevice", "removeDevice", uuid, forceValue)`
3. **Handle Errors**: BCM rejection (dependencies exist) → diagnostic with force parameter suggestion
4. **Verify Deletion**: Optional exponential backoff check (if needed for eventual consistency)

### Import State Implementation

```go
func (r *CMDeviceDeviceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
```

**Import Workflow**:
1. User runs: `terraform import bcm_cmdevice_device.test <uuid>`
2. Provider calls Read with UUID
3. Read operation populates all attributes from BCM API
4. State written with complete device configuration

### Helper Functions

**buildDeviceAPIEntity**:
```go
func buildDeviceAPIEntity(plan CMDeviceDeviceResourceModel, uuid string) map[string]interface{} {
    entity := map[string]interface{}{
        "baseType":      "Device",
        "childType":     plan.ChildType.ValueString(), // Empty for new devices
        "hostname":      plan.Hostname.ValueString(),
        "mac":           plan.MAC.ValueString(),
        "category":      plan.Category.ValueString(),
        "managementNetwork": plan.ManagementNetwork.ValueString(),
        "modified":      true,
        "to_be_removed": false,
    }

    if uuid != "" {
        entity["uuid"] = uuid
    }

    // Add optional fields if present
    if !plan.KernelParameters.IsNull() {
        entity["kernelParameters"] = plan.KernelParameters.ValueString()
    }

    // Add interfaces if configured
    if !plan.Interfaces.IsNull() {
        interfaces := buildInterfacesList(plan.Interfaces)
        entity["interfaces"] = interfaces
    }

    return entity
}
```

**parseDeviceFromAPI**:
```go
func parseDeviceFromAPI(data map[string]interface{}) CMDeviceDeviceResourceModel {
    model := CMDeviceDeviceResourceModel{}

    // Required fields
    model.UUID = types.StringValue(data["uuid"].(string))
    model.ID = model.UUID
    model.Hostname = types.StringValue(data["hostname"].(string))
    model.MAC = types.StringValue(data["mac"].(string))

    // Handle optional fields with null safety
    if val, ok := data["kernelParameters"].(string); ok {
        model.KernelParameters = types.StringValue(val)
    } else {
        model.KernelParameters = types.StringNull()
    }

    // Parse computed fields
    if val, ok := data["creationTime"].(float64); ok {
        model.CreationTime = types.Int64Value(int64(val))
    }

    return model
}
```

---

## Phase 3: Test Strategy

### Test Structure

**File**: `internal/provider/resource_cmdevice_device_test.go`

**Required Imports**:
```go
import (
    "context"
    "encoding/json"
    "fmt"
    "os"
    "testing"
    "time"

    "github.com/hashicorp/terraform-plugin-testing/helper/resource"
    "github.com/hashicorp/terraform-plugin-testing/plancheck"
    "github.com/hashicorp/terraform-plugin-testing/statecheck"
    "github.com/hashicorp/terraform-plugin-testing/knownvalue"
    "github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
    "github.com/hashicorp/terraform-plugin-testing/compare"
)
```

### Test Cases

#### TC-001: TestAccCMDeviceDeviceResource_Basic
**Purpose**: Full CRUD lifecycle with import and idempotency checks

**Test Steps**:
1. Create device with minimal required fields
2. Verify idempotency after Create (empty plan)
3. Import device by UUID
4. Update device (modify notes)
5. Verify idempotency after Update (empty plan)
6. Delete device (automatic via TestCase)

**State Checks**:
- hostname: `knownvalue.StringExact(deviceName)`
- uuid: `knownvalue.NotNull()`
- category: `knownvalue.StringExact(categoryUUID)`
- ID consistency: `compareID.AddStateValue()`

#### TC-002: TestAccCMDeviceDevice_DriftHostname
**Purpose**: Detect external modifications to hostname

**Test Steps**:
1. Create device with hostname "test-device-001"
2. Modify hostname externally via BCM API (set to "modified-hostname")
3. Run terraform plan → expect non-empty plan (drift detected)
4. Apply terraform → hostname restored to "test-device-001"

#### TC-003: TestAccCMDeviceDevice_DriftKernelParameters
**Purpose**: Detect external modifications to kernel parameters

**Test Steps**:
1. Create device with kernel_parameters "quiet splash"
2. Modify externally via BCM API (set to "nomodeset")
3. Run terraform plan → expect non-empty plan
4. Apply terraform → kernel_parameters restored

#### TC-004: TestAccCMDeviceDevice_ValidationInvalidHostname
**Purpose**: Test hostname RFC 1123 validation

**Test Cases**:
- Empty hostname → expect error
- Uppercase letters → expect error
- Leading hyphen → expect error
- Over 63 characters → expect error
- Special characters → expect error

#### TC-005: TestAccCMDeviceDevice_ValidationInvalidMAC
**Purpose**: Test MAC address format validation

**Test Cases**:
- Invalid format "00-11-22-33-44-55" → expect error
- Missing octets "00:11:22:33:44" → expect error
- Non-hex characters "ZZ:11:22:33:44:55" → expect error

#### TC-006: TestAccCMDeviceDevice_ValidationInvalidUUID
**Purpose**: Test UUID reference validation

**Test Cases**:
- Invalid category UUID → expect error
- Invalid management_network UUID → expect error

### Test Helper Functions

**testAccCMDeviceDevicePreCheck**:
```go
func testAccCMDeviceDevicePreCheck(t *testing.T, deviceNames ...string) {
    client := createTestBCMClient(t)
    ctx := context.Background()

    // Clean up leftover test devices
    for _, name := range deviceNames {
        deleted, _ := verifyResourceDeleted(ctx, client, "cmdevice", "getDevice", name, 5)
        if !deleted {
            t.Logf("Warning: Could not verify deletion of %s", name)
        }
    }
}
```

**testAccCheckCMDeviceDeviceDestroy**:
```go
func testAccCheckCMDeviceDeviceDestroy(s *terraform.State) error {
    client := createTestBCMClient(&testing.T{})
    ctx := context.Background()

    var errors []string
    for _, rs := range s.RootModule().Resources {
        if rs.Type != "bcm_cmdevice_device" {
            continue
        }

        uuid := rs.Primary.ID
        deleted, err := verifyResourceDeleted(ctx, client, "cmdevice", "getDevice", uuid, 4)

        if err != nil || !deleted {
            errors = append(errors, fmt.Sprintf(
                "Device still exists: UUID=%s", uuid,
            ))
        }
    }

    if len(errors) > 0 {
        return fmt.Errorf("CheckDestroy failures: %v", errors)
    }
    return nil
}
```

**testAccCMDeviceDeviceResourceConfig_Basic**:
```go
func testAccCMDeviceDeviceResourceConfig_Basic(hostname string) string {
    return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmdevice_category" "default" {
  name = "default"
}

data "bcm_cmnet_network" "mgmt" {
  name = "management"
}

resource "bcm_cmdevice_device" "test" {
  hostname           = %[4]q
  mac                = "00:11:22:33:44:55"
  category           = data.bcm_cmdevice_category.default.uuid
  management_network = data.bcm_cmnet_network.mgmt.uuid
}
`,
        os.Getenv("BCM_ENDPOINT"),
        os.Getenv("BCM_USERNAME"),
        os.Getenv("BCM_PASSWORD"),
        hostname,
    )
}
```

### Test Execution

**Run All Tests**:
```bash
TF_ACC=1 BCM_ENDPOINT="https://172.21.15.254:8081" BCM_USERNAME="root" BCM_PASSWORD="Hashicorp123!" \
  go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceDevice
```

**Run Specific Test**:
```bash
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceDeviceResource_Basic
```

---

## Error Handling & Edge Cases

### API Error Scenarios

**Duplicate Hostname**:
```go
if strings.Contains(errorMsg, "already exists") {
    resp.Diagnostics.AddError(
        "Device Hostname Already Exists",
        fmt.Sprintf("A device with hostname '%s' already exists in BCM. "+
            "Hostnames must be unique across the cluster.", plan.Hostname.ValueString()),
    )
    return
}
```

**Invalid Category Reference**:
```go
if strings.Contains(errorMsg, "category not found") {
    resp.Diagnostics.AddError(
        "Invalid Category Reference",
        fmt.Sprintf("Category UUID '%s' does not exist in BCM. "+
            "Ensure the category resource exists before creating devices.",
            plan.Category.ValueString()),
    )
    return
}
```

**Force Parameter Required**:
```go
if strings.Contains(errorMsg, "device has active provisioning") {
    resp.Diagnostics.AddError(
        "Device Update Requires Force",
        "Cannot update device with active provisioning operations. "+
            "Set 'force = true' in the configuration to override this constraint.",
    )
    return
}
```

### External Deletion Detection

```go
func (r *CMDeviceDeviceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    var state CMDeviceDeviceResourceModel
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

    body, err := r.client.CallJSONRPC(ctx, "cmdevice", "getDevice", state.UUID.ValueString())

    // Resource deleted externally
    if err != nil || len(body) == 0 {
        tflog.Warn(ctx, "Device not found in BCM, removing from state", map[string]interface{}{
            "uuid": state.UUID.ValueString(),
        })
        resp.State.RemoveResource(ctx)
        return
    }

    // Continue with normal read...
}
```

### Async Operation Handling

**If device creation is asynchronous** (determined in Phase 0 research):
```go
func waitForDeviceReady(ctx context.Context, client *BCMClient, uuid string, maxRetries int) error {
    waitTime := 2 * time.Second

    for retry := 0; retry < maxRetries; retry++ {
        time.Sleep(waitTime)

        body, err := client.CallJSONRPC(ctx, "cmdevice", "getDevice", uuid)
        if err != nil {
            return fmt.Errorf("failed to check device status: %w", err)
        }

        var device map[string]interface{}
        json.Unmarshal(body, &device)

        // Check if provisioning complete
        if inProgress, ok := device["fileOperationInProgress"].(bool); ok && !inProgress {
            return nil // Device ready
        }

        waitTime *= 2 // Exponential backoff
    }

    return fmt.Errorf("device creation timed out after %d retries", maxRetries)
}
```

---

## Documentation Requirements

### Generated Documentation

**File**: `docs/resources/bcm_cmdevice_device.md` (auto-generated by tfplugindocs)

**Sections**:
- Resource description
- Example Usage (basic, compute node, head node)
- Argument Reference (required, optional, computed fields)
- Attributes Reference
- Import instructions

### Example Configurations

**Basic Device**:
```hcl
resource "bcm_cmdevice_device" "compute001" {
  hostname           = "compute001"
  mac                = "00:11:22:33:44:55"
  category           = bcm_cmdevice_category.compute.uuid
  management_network = data.bcm_cmnet_network.mgmt.uuid

  notes = "Compute node for ML workloads"
}
```

**Device with Network Interfaces**:
```hcl
resource "bcm_cmdevice_device" "storage001" {
  hostname           = "storage001"
  mac                = "00:11:22:33:44:66"
  category           = bcm_cmdevice_category.storage.uuid
  management_network = data.bcm_cmnet_network.mgmt.uuid

  interfaces = [
    {
      name    = "ens33"
      mac     = "00:11:22:33:44:66"
      ip      = "172.21.15.100"
      network = data.bcm_cmnet_network.mgmt.uuid
      dhcp    = false
    },
    {
      name    = "ens34"
      mac     = "00:11:22:33:44:77"
      ip      = "10.0.0.100"
      network = data.bcm_cmnet_network.data.uuid
      dhcp    = false
    }
  ]

  bmc_settings = {
    user_name = "admin"
    password  = "secret123"
    privilege = "ADMINISTRATOR"
  }
}
```

**Import Example**:
```bash
terraform import bcm_cmdevice_device.compute001 9f885869-a146-4cd6-af1f-f9b6c674a84c
```

---

## Success Criteria Verification

| Criterion | Verification Method | Expected Outcome |
|-----------|---------------------|------------------|
| SC-001: Basic device creation | TestAccCMDeviceDeviceResource_Basic | Device created with minimal config in single apply |
| SC-002: Import functionality | ImportState test step | All fields populated correctly on import |
| SC-003: Drift detection | TestAccCMDeviceDevice_Drift* tests | External modifications detected in plan |
| SC-004: Full CRUD lifecycle | Complete test suite | All CRUD operations pass with proper state management |
| SC-005: 100% test pass rate | CI pipeline execution | All acceptance tests pass |
| SC-006: Documentation | `make generate` output | Auto-generated docs with working examples |
| SC-007: Schema validation | Validation test cases | Invalid configs caught before API calls |

---

## Implementation Checklist

### Phase 0: Research (Complete Before Implementation)
- [ ] Verify BCM device API methods (addDevice, getDevice, updateDevice, removeDevice)
- [ ] Document childType assignment logic
- [ ] Identify force parameter scenarios
- [ ] Research network interface patterns
- [ ] Document validation implementations
- [ ] Create research.md with findings

### Phase 1: Design (Generate Artifacts)
- [ ] Create data-model.md with complete schema
- [ ] Create contracts/device-api.json
- [ ] Create quickstart.md with TDD workflow
- [ ] Run update-agent-context.sh to update copilot/cursor

### Phase 2: RED Phase (Failing Tests)
- [ ] Create resource_cmdevice_device_test.go
- [ ] Write TestAccCMDeviceDeviceResource_Basic (full CRUD)
- [ ] Write TestAccCMDeviceDevice_DriftHostname
- [ ] Write TestAccCMDeviceDevice_DriftKernelParameters
- [ ] Write validation test cases (hostname, MAC, UUID)
- [ ] Run tests: Verify all fail (resource not implemented)

### Phase 3: GREEN Phase (Minimal Implementation)
- [ ] Create resource_cmdevice_device.go
- [ ] Implement Schema() with required fields
- [ ] Implement Create() with minimal API call
- [ ] Implement Read() with getDevice direct lookup
- [ ] Implement Update() with updateDevice
- [ ] Implement Delete() with removeDevice
- [ ] Implement ImportState() with UUID passthrough
- [ ] Run tests: Verify all pass

### Phase 4: REFACTOR Phase (Full Implementation)
- [ ] Add optional fields to schema (interfaces, roles, BMC settings)
- [ ] Add validation (hostname RFC 1123, MAC address, UUID)
- [ ] Add error handling and diagnostics
- [ ] Implement buildDeviceAPIEntity helper
- [ ] Implement parseDeviceFromAPI helper
- [ ] Add computed field management
- [ ] Run tests: Verify still passing

### Phase 5: Documentation & Polish
- [ ] Create examples/resources/bcm_cmdevice_device/
- [ ] Add basic example configuration
- [ ] Add compute node example
- [ ] Add head node example
- [ ] Run `make generate` to create docs
- [ ] Verify documentation accuracy
- [ ] Update CLAUDE.md with any new patterns

### Phase 6: Final Verification
- [ ] Run full acceptance test suite
- [ ] Verify 100% test pass rate
- [ ] Test import functionality manually
- [ ] Test drift detection manually
- [ ] Verify idempotency (multiple applies with no changes)
- [ ] Review error messages for clarity
