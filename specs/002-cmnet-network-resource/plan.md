# Implementation Plan: BCM Network Resource Management

**Branch**: `002-cmnet-network-resource` | **Date**: 2025-11-23 | **Spec**: [spec.md](/workspace/specs/002-cmnet-network-resource/spec.md)
**Input**: Feature specification from `/specs/002-cmnet-network-resource/spec.md`

## Summary

Implement `bcm_cmnet_network` Terraform resource to enable full CRUD lifecycle management (Create, Read, Update, Delete, Import) of BCM network configurations via declarative infrastructure-as-code. The resource will support subnet configuration in CIDR notation, DHCP services with IP range specification, optional VLAN segmentation, MTU configuration, and drift detection for operational consistency.

**Technical Approach**: Follow TDD RED-GREEN-REFACTOR cycle with modern terraform-plugin-testing patterns (statecheck, plancheck, knownvalue). Leverage existing CMNet service patterns from `data_source_cmnet_networks.go` for field mapping and API client from `bcm_client.go` for JSON-RPC operations. Implement CIDR parsing layer to convert user-friendly `subnet` attribute (e.g., "10.0.1.0/24") into BCM API's separate `baseAddress` and `netmaskBits` fields.

## Technical Context

**Language/Version**: Go 1.24+
**Primary Dependencies**:
- terraform-plugin-framework v1.16.1 (resource schema, lifecycle)
- terraform-plugin-testing v1.13.3+ (modern test patterns)
- BCM JSON-RPC API (cmnet service)

**Storage**: BCM cluster persistent storage (networks managed via API)
**Testing**:
- Acceptance tests with TF_ACC=1 (full CRUD lifecycle verification)
- Modern patterns: statecheck.ExpectKnownValue, plancheck.ExpectEmptyPlan, ID consistency tracking
- Test helpers: createTestBCMClient, getResourceUUIDByName, verifyResourceDeleted, generateUniqueTestName

**Target Platform**: Linux server (BCM cluster endpoint: https://172.21.15.254:8081)
**Project Type**: Single (Terraform provider - internal/provider/)
**Performance Goals**:
- Network creation/update: <30 seconds
- Read operations: <5 seconds (direct UUID lookup via args parameter)
- Full acceptance test suite: <120 minutes

**Constraints**:
- BCM API requires full entity structure (baseType, childType, modified, to_be_removed, revision, uuid)
- Network names must be unique within BCM cluster
- DHCP enabled logic derived from non-zero dynamic_range_start/end values
- Field name mapping: Terraform snake_case → BCM API camelCase

**Scale/Scope**:
- 11 distinct acceptance test scenarios (basic create, DHCP, VLAN, update, delete, import, drift, idempotency)
- ~25 resource attributes (required: name; optional: subnet, gateway, MTU, DHCP config, VLAN; computed: UUID, timestamps, flags)
- 4 CRUD operations + ImportState implementation

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**Constitution Compliance**:
- ✅ TDD First: All acceptance tests written before CRUD implementation (RED-GREEN-REFACTOR)
- ✅ Test Coverage: 100% CRUD operations tested, import verified, drift detection validated
- ✅ Modern Testing: Uses statecheck, plancheck, knownvalue from terraform-plugin-testing v1.13.3+
- ✅ Parallel Execution: Test writing and implementation can proceed in parallel batches
- ✅ Documentation: Auto-generated via `make generate` using tfplugindocs
- ✅ No Complex Patterns: Direct resource implementation following existing patterns (resource_cmdevice_category.go, resource_cmpart_softwareimage.go)

**No violations identified** - implementation follows established Terraform provider TDD constitution.

## Project Structure

### Documentation (this feature)

```text
specs/002-cmnet-network-resource/
├── spec.md              # Feature specification (user stories, requirements, API contracts)
├── plan.md              # This file (implementation plan with TDD phases)
├── research.md          # Phase 0: API exploration (VLAN field mapping, CIDR parsing strategy)
├── data-model.md        # Phase 1: Network entity schema with BCM field mappings
├── quickstart.md        # Phase 1: Developer quick start (build, test, run)
├── contracts/           # Phase 1: API contracts (BCM CMNet service methods)
│   └── bcm-cmnet-api.md # CMNet service: addNetwork, getNetwork, updateNetwork, removeNetwork
└── tasks.md             # Phase 2: Task breakdown (NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
terraform-provider-bcm/
├── internal/provider/
│   ├── resource_cmnet_network.go          # NEW: Resource implementation (CRUD + ImportState)
│   ├── resource_cmnet_network_test.go     # NEW: Acceptance tests (11 test scenarios)
│   ├── data_source_cmnet_networks.go      # REFERENCE: Existing data source (field mapping guide)
│   ├── resource_cmdevice_category.go      # REFERENCE: BCM entity pattern (baseType, revision, etc.)
│   ├── resource_cmpart_softwareimage.go   # REFERENCE: CRUD patterns, drift detection
│   ├── bcm_client.go                      # EXISTING: JSON-RPC client (CallJSONRPC method)
│   ├── test_helpers.go                    # EXISTING: Test utilities (createTestBCMClient, etc.)
│   └── provider.go                        # MODIFY: Register NewCMNetNetworkResource in Resources()
├── examples/
│   └── resources/
│       └── bcm_cmnet_network/             # NEW: Example configurations
│           ├── basic.tf                   # Basic network with name only
│           ├── subnet.tf                  # Network with CIDR subnet
│           ├── dhcp.tf                    # Network with DHCP enabled
│           ├── vlan.tf                    # Network with VLAN configuration
│           └── complete.tf                # All optional attributes configured
├── docs/                                  # AUTO-GENERATED: tfplugindocs output
│   └── resources/
│       └── cmnet_network.md               # Resource documentation (DO NOT EDIT MANUALLY)
└── sampleRest/                            # TESTING: BCM API exploration scripts
    └── explore_network_crud.py            # NEW: Validate network CRUD operations
```

**Structure Decision**: Single project structure (standard Terraform provider layout). All resource implementation in `internal/provider/` with examples in `examples/resources/`. Testing uses existing test infrastructure (BCM cluster at 172.21.15.254:8081). Documentation auto-generated from schema and examples.

## Complexity Tracking

> **No Constitution Check violations** - this table is empty.

---

## Phase 0: Research & API Exploration

**Goal**: Resolve all "NEEDS CLARIFICATION" items from Technical Context by exploring BCM CMNet API.

### Research Tasks

1. **VLAN Field Mapping Confirmation**
   - **Unknown**: Spec assumes VLAN ID maps to a BCM API field but exact field name unknown
   - **Action**: Run BCM API `getNetworks()` call, inspect response for VLAN-related fields
   - **Script**: `sampleRest/explore_network_crud.py` - query existing network with VLAN, dump full JSON
   - **Decision Criteria**: If VLAN field exists in response → include in schema; if not → mark as OUT OF SCOPE

2. **CIDR Parsing Strategy Validation**
   - **Unknown**: Best approach for converting CIDR notation to baseAddress/netmaskBits
   - **Action**: Research Go standard library `net` package for CIDR parsing (net.ParseCIDR)
   - **Alternatives**:
     - Use `net.ParseCIDR()` (standard library)
     - Manual regex parsing + validation
     - Third-party library (overkill)
   - **Decision**: Document chosen approach with code snippet

3. **BCM Network CRUD API Methods Verification**
   - **Unknown**: Confirm exact method signatures for cmnet service
   - **Action**: Test each operation via sampleRest Python scripts:
     - `addNetwork(entity, force)` - create operation
     - `getNetwork(uuid)` - single network lookup with args parameter
     - `updateNetwork(entity, force)` - update operation
     - `removeNetwork(uuid, force)` - delete operation
   - **Decision**: Document API signatures, parameter types, response structure

4. **Force Parameter Behavior**
   - **Unknown**: When is `force=true` required for network operations?
   - **Action**: Test deletion of network with/without active node assignments
   - **Decision**: Document when force parameter must be exposed to users vs. always false

### Research Outputs

**File**: `research.md`

**Format**:
```markdown
# Research Findings: BCM CMNet Network Resource

## VLAN Field Mapping
- **Decision**: [Field name from API or OUT OF SCOPE]
- **Rationale**: [Why chosen based on API response]
- **Implementation**: [Schema attribute definition if applicable]

## CIDR Parsing Strategy
- **Decision**: Use net.ParseCIDR from Go standard library
- **Rationale**: Standard, well-tested, handles validation automatically
- **Code Snippet**:
  ```go
  ip, ipnet, err := net.ParseCIDR("10.0.1.0/24")
  baseAddress := ip.String()
  netmaskBits, _ := ipnet.Mask.Size() // returns (24, 32)
  ```

## BCM CMNet API Methods
- **addNetwork**: `CallJSONRPC(ctx, "cmnet", "addNetwork", entity, false)`
- **getNetwork**: `CallJSONRPC(ctx, "cmnet", "getNetwork", uuid)` (args parameter)
- **updateNetwork**: `CallJSONRPC(ctx, "cmnet", "updateNetwork", entity, false)`
- **removeNetwork**: `CallJSONRPC(ctx, "cmnet", "removeNetwork", uuid, false)`

## Force Parameter
- **Decision**: Always use `force=false` for create/update, expose as optional attribute for delete
- **Rationale**: Delete may need force=true when network has active assignments
```

---

## Phase 1: Design & Data Modeling

**Prerequisites**: `research.md` complete with all API methods validated

### 1. Data Model Design

**File**: `data-model.md`

**Network Entity Schema**:

| Terraform Attribute (snake_case) | BCM API Field (camelCase) | Type | Category | Notes |
|-----------------------------------|---------------------------|------|----------|-------|
| id | uuid | string | Computed | Same as uuid, Terraform convention |
| uuid | uuid | string | Computed | BCM-assigned unique identifier |
| name | name | string | Required | Unique within cluster |
| subnet | N/A | string | Optional | User-facing CIDR (e.g., "10.0.1.0/24") |
| base_address | baseAddress | string | Computed | Parsed from subnet |
| netmask_bits | netmaskBits | int64 | Computed | Parsed from subnet |
| gateway | gateway | string | Optional | IP address within subnet |
| network_type | type | string | Computed | BCM-assigned network type |
| mtu | mtu | int64 | Optional | Default: 1500 |
| domain_name | domainName | string | Optional | DNS domain |
| dhcp_enabled | N/A | bool | Computed | Derived from dynamic_range_start/end |
| dhcp_range_start | dynamicRangeStart | string | Optional | DHCP pool start IP |
| dhcp_range_end | dynamicRangeEnd | string | Optional | DHCP pool end IP |
| vlan_id | [RESEARCH RESULT] | int64 | Optional | VLAN tag (1-4094) |
| management | management | bool | Computed | BCM-assigned management flag |
| bootable | bootable | bool | Computed | BCM-assigned bootable flag |
| notes | notes | string | Optional | User notes |
| base_type | baseType | string | Computed | Always "Network" |
| child_type | childType | string | Computed | BCM entity type |
| revision | revision | string | Computed | BCM revision tracking |
| modified | modified | bool | Computed | BCM modification flag |
| to_be_removed | to_be_removed | bool | Computed | BCM deletion flag |

**DHCP Enabled Logic**:
```
dhcp_enabled = (dhcp_range_start != null && dhcp_range_start != "0.0.0.0" &&
                dhcp_range_end != null && dhcp_range_end != "0.0.0.0")
```

**Subnet Parsing Logic**:
```go
// On Create/Update (Plan → API entity)
ip, ipnet, err := net.ParseCIDR(data.Subnet.ValueString())
entity["baseAddress"] = ip.String()
maskBits, _ := ipnet.Mask.Size()
entity["netmaskBits"] = maskBits

// On Read (API response → State)
subnet := fmt.Sprintf("%s/%d", baseAddress, netmaskBits)
data.Subnet = types.StringValue(subnet)
```

### 2. API Contracts Documentation

**File**: `contracts/bcm-cmnet-api.md`

**Service**: cmnet
**Base URL**: https://172.21.15.254:8081/json

**Methods**:

1. **addNetwork** (Create)
   - **Request**: `{"service": "cmnet", "call": "addNetwork", "args": [entity, false]}`
   - **Entity Structure**:
     ```json
     {
       "name": "compute-network",
       "baseAddress": "10.0.1.0",
       "netmaskBits": 24,
       "gateway": "10.0.1.1",
       "mtu": 1500,
       "dynamicRangeStart": "10.0.1.100",
       "dynamicRangeEnd": "10.0.1.200",
       "notes": "Created via Terraform",
       "baseType": "Network",
       "childType": "",
       "modified": true,
       "to_be_removed": false
     }
     ```
   - **Response**: Created entity with UUID assigned
   - **Errors**: 409 if name already exists

2. **getNetwork** (Read)
   - **Request**: `{"service": "cmnet", "call": "getNetwork", "args": [uuid]}`
   - **Response**: Single network entity
   - **Errors**: 404 if UUID not found

3. **updateNetwork** (Update)
   - **Request**: `{"service": "cmnet", "call": "updateNetwork", "args": [entity, false]}`
   - **Entity Requirements**: Must include uuid, revision, all fields from getNetwork
   - **Response**: Updated entity
   - **Errors**: 409 if revision conflict

4. **removeNetwork** (Delete)
   - **Request**: `{"service": "cmnet", "call": "removeNetwork", "args": [uuid, force]}`
   - **Force Parameter**: Set true to override dependency checks
   - **Response**: Empty on success
   - **Errors**: 409 if network has active assignments (force=false)

### 3. Agent Context Update

**Action**: Run `.specify/scripts/bash/update-agent-context.sh copilot`

**Purpose**: Add BCM CMNet network resource patterns to GitHub Copilot context

**Expected Updates**:
- Technology: BCM CMNet service (network CRUD operations)
- Patterns: CIDR parsing, DHCP logic derivation, BCM entity structure
- Testing: Modern terraform-plugin-testing patterns for network resources

### 4. Developer Quick Start

**File**: `quickstart.md`

```markdown
# Quick Start: BCM Network Resource Development

## Prerequisites
- Go 1.24+
- BCM cluster access: https://172.21.15.254:8081
- Credentials: BCM_USERNAME, BCM_PASSWORD, BCM_ENDPOINT

## Build & Install
```bash
make install  # Runs fmt, lint, install, generate
```

## Run Acceptance Tests
```bash
export TF_ACC=1
export BCM_ENDPOINT="https://172.21.15.254:8081"
export BCM_USERNAME="root"
export BCM_PASSWORD="Hashicorp123!"

# Run all network resource tests
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMNetNetwork

# Run specific test
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMNetNetwork_Basic
```

## Manual Testing
```bash
cd examples/resources/bcm_cmnet_network
terraform init
terraform plan
terraform apply
terraform destroy
```

## Generate Documentation
```bash
make generate  # Runs tfplugindocs, updates docs/resources/cmnet_network.md
```
```

---

## Phase 2: TDD Implementation (RED-GREEN-REFACTOR)

**Prerequisites**: Phase 1 complete (data-model.md, contracts/, quickstart.md)

### TDD RED Phase: Write Failing Acceptance Tests

**File**: `internal/provider/resource_cmnet_network_test.go`

**Test Scenarios** (11 total):

1. **TestAccCMNetNetwork_Basic** - Create network with name only, verify BCM-assigned defaults
   ```go
   Steps: []resource.TestStep{
       {
           Config: testAccCMNetNetworkConfigBasic(networkName),
           ConfigStateChecks: []statecheck.StateCheck{
               statecheck.ExpectKnownValue(
                   "bcm_cmnet_network.test",
                   tfjsonpath.New("name"),
                   knownvalue.StringExact(networkName),
               ),
               statecheck.ExpectKnownValue(
                   "bcm_cmnet_network.test",
                   tfjsonpath.New("uuid"),
                   knownvalue.NotNull(),
               ),
               compareID.AddStateValue("bcm_cmnet_network.test", tfjsonpath.New("id")),
           },
       },
   }
   ```

2. **TestAccCMNetNetwork_Subnet** - Create with CIDR subnet, verify parsing to baseAddress/netmaskBits
   ```go
   ConfigStateChecks: []statecheck.StateCheck{
       statecheck.ExpectKnownValue(
           "bcm_cmnet_network.test",
           tfjsonpath.New("subnet"),
           knownvalue.StringExact("10.0.1.0/24"),
       ),
       statecheck.ExpectKnownValue(
           "bcm_cmnet_network.test",
           tfjsonpath.New("base_address"),
           knownvalue.StringExact("10.0.1.0"),
       ),
       statecheck.ExpectKnownValue(
           "bcm_cmnet_network.test",
           tfjsonpath.New("netmask_bits"),
           knownvalue.Int64Exact(24),
       ),
   }
   ```

3. **TestAccCMNetNetwork_DHCP** - Enable DHCP with IP range, verify dhcp_enabled computed attribute
   ```go
   ConfigStateChecks: []statecheck.StateCheck{
       statecheck.ExpectKnownValue(
           "bcm_cmnet_network.test",
           tfjsonpath.New("dhcp_enabled"),
           knownvalue.Bool(true),
       ),
       statecheck.ExpectKnownValue(
           "bcm_cmnet_network.test",
           tfjsonpath.New("dhcp_range_start"),
           knownvalue.StringExact("10.0.1.100"),
       ),
       statecheck.ExpectKnownValue(
           "bcm_cmnet_network.test",
           tfjsonpath.New("dhcp_range_end"),
           knownvalue.StringExact("10.0.1.200"),
       ),
   }
   ```

4. **TestAccCMNetNetwork_VLAN** - Configure VLAN ID (if supported per research), verify VLAN tag applied
   ```go
   // Only if VLAN field exists in BCM API (from research.md)
   ConfigStateChecks: []statecheck.StateCheck{
       statecheck.ExpectKnownValue(
           "bcm_cmnet_network.test",
           tfjsonpath.New("vlan_id"),
           knownvalue.Int64Exact(100),
       ),
   }
   ```

5. **TestAccCMNetNetwork_CompleteConfig** - All optional attributes configured
   - MTU: 9000
   - Gateway: "10.0.1.1"
   - Domain name: "cluster.local"
   - Notes: "Complete test configuration"

6. **TestAccCMNetNetwork_Update** - Update attributes (MTU 1500 → 9000, add gateway)
   ```go
   Steps: []resource.TestStep{
       {Config: testAccCMNetNetworkConfigBasic(name), Check: ...},
       {
           Config: testAccCMNetNetworkConfigUpdate(name),
           ConfigStateChecks: []statecheck.StateCheck{
               statecheck.ExpectKnownValue(
                   "bcm_cmnet_network.test",
                   tfjsonpath.New("mtu"),
                   knownvalue.Int64Exact(9000),
               ),
               compareID.AddStateValue("bcm_cmnet_network.test", tfjsonpath.New("id")),
           },
       },
   }
   ```

7. **TestAccCMNetNetwork_UpdateDHCP** - Enable DHCP, then disable DHCP
   - Step 1: Create without DHCP (dhcp_enabled = false)
   - Step 2: Enable DHCP with range (dhcp_enabled = true)
   - Step 3: Disable DHCP by removing range (dhcp_enabled = false)

8. **TestAccCMNetNetwork_Import** - Import existing network by UUID
   ```go
   {
       ResourceName:      "bcm_cmnet_network.test",
       ImportState:       true,
       ImportStateVerify: true,
       ConfigStateChecks: []statecheck.StateCheck{
           compareID.AddStateValue("bcm_cmnet_network.test", tfjsonpath.New("id")),
       },
   }
   ```

9. **TestAccCMNetNetwork_DriftDetection** - Detect external MTU modification
   ```go
   {
       PreConfig: func() {
           client := createTestBCMClient(t)
           uuid := getResourceUUIDByName(t, "cmnet", "getNetwork", networkName)
           // Modify MTU externally via BCM API
           // entity["mtu"] = 9000
           client.CallJSONRPC(ctx, "cmnet", "updateNetwork", entity, false)
           time.Sleep(2 * time.Second)
       },
       Config: testAccCMNetNetworkConfigBasic(networkName),
       ConfigPlanChecks: resource.ConfigPlanChecks{
           PreApply: []plancheck.PlanCheck{
               plancheck.ExpectNonEmptyPlan(),
           },
       },
   }
   ```

10. **TestAccCMNetNetwork_IdempotencyAfterCreate** - Verify empty plan after create
    ```go
    {
        Config: testAccCMNetNetworkConfigBasic(networkName),
        ConfigPlanChecks: resource.ConfigPlanChecks{
            PreApply: []plancheck.PlanCheck{
                plancheck.ExpectEmptyPlan(),
            },
        },
    }
    ```

11. **TestAccCMNetNetwork_IdempotencyAfterUpdate** - Verify empty plan after update
    ```go
    {
        Config: testAccCMNetNetworkConfigUpdate(networkName),
        ConfigPlanChecks: resource.ConfigPlanChecks{
            PreApply: []plancheck.PlanCheck{
                plancheck.ExpectEmptyPlan(),
            },
        },
    }
    ```

**Test Helper Functions**:
```go
func testAccCMNetNetworkConfigBasic(name string) string {
    return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

resource "bcm_cmnet_network" "test" {
  name = %[4]q
}
`,
        os.Getenv("BCM_ENDPOINT"),
        os.Getenv("BCM_USERNAME"),
        os.Getenv("BCM_PASSWORD"),
        name,
    )
}

func testAccCheckCMNetNetworkDestroy(s *terraform.State) error {
    client := createTestBCMClient(&testing.T{})
    var errors []string

    for _, rs := range s.RootModule().Resources {
        if rs.Type != "bcm_cmnet_network" {
            continue
        }

        deleted, err := verifyResourceDeleted(
            context.Background(),
            client,
            "cmnet",
            "getNetwork",
            rs.Primary.ID,
            4,
        )

        if !deleted {
            errors = append(errors, fmt.Sprintf("Network still exists: %s", rs.Primary.ID))
        }
    }

    if len(errors) > 0 {
        return fmt.Errorf("CheckDestroy failures:\n  - %s", strings.Join(errors, "\n  - "))
    }
    return nil
}
```

**Expected Result**: All 11 tests FAIL (no implementation exists yet)

**Verification Command**:
```bash
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMNetNetwork
# Expected: 11 failures
```

---

### TDD GREEN Phase: Minimal CRUD Implementation

**File**: `internal/provider/resource_cmnet_network.go`

**Minimal Implementation Checklist**:

1. **Resource Struct & Interface Compliance**
   ```go
   var (
       _ resource.Resource                = &CMNetNetworkResource{}
       _ resource.ResourceWithImportState = &CMNetNetworkResource{}
   )

   type CMNetNetworkResource struct {
       client *BCMClient
   }

   type CMNetNetworkResourceModel struct {
       ID              types.String `tfsdk:"id"`
       UUID            types.String `tfsdk:"uuid"`
       Name            types.String `tfsdk:"name"`
       Subnet          types.String `tfsdk:"subnet"`
       BaseAddress     types.String `tfsdk:"base_address"`
       NetmaskBits     types.Int64  `tfsdk:"netmask_bits"`
       Gateway         types.String `tfsdk:"gateway"`
       NetworkType     types.String `tfsdk:"network_type"`
       MTU             types.Int64  `tfsdk:"mtu"`
       DomainName      types.String `tfsdk:"domain_name"`
       DHCPEnabled     types.Bool   `tfsdk:"dhcp_enabled"`
       DHCPRangeStart  types.String `tfsdk:"dhcp_range_start"`
       DHCPRangeEnd    types.String `tfsdk:"dhcp_range_end"`
       VlanID          types.Int64  `tfsdk:"vlan_id"` // If supported
       Management      types.Bool   `tfsdk:"management"`
       Bootable        types.Bool   `tfsdk:"bootable"`
       Notes           types.String `tfsdk:"notes"`
       BaseType        types.String `tfsdk:"base_type"`
       ChildType       types.String `tfsdk:"child_type"`
       Revision        types.String `tfsdk:"revision"`
       Modified        types.Bool   `tfsdk:"modified"`
       ToBeRemoved     types.Bool   `tfsdk:"to_be_removed"`
   }
   ```

2. **Schema Definition** (minimal required/optional/computed attributes)
   ```go
   func (r *CMNetNetworkResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
       resp.Schema = schema.Schema{
           MarkdownDescription: "Manages a BCM network configuration.",
           Attributes: map[string]schema.Attribute{
               "id": schema.StringAttribute{
                   Computed:            true,
                   MarkdownDescription: "Network unique identifier (same as uuid).",
               },
               "uuid": schema.StringAttribute{
                   Computed:            true,
                   MarkdownDescription: "BCM-assigned UUID.",
               },
               "name": schema.StringAttribute{
                   Required:            true,
                   MarkdownDescription: "Unique network name.",
               },
               "subnet": schema.StringAttribute{
                   Optional:            true,
                   MarkdownDescription: "Network subnet in CIDR notation (e.g., '10.0.1.0/24').",
                   Validators: []validator.String{
                       stringvalidator.RegexMatches(
                           regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}/\d{1,2}$`),
                           "must be valid CIDR notation",
                       ),
                   },
               },
               "base_address": schema.StringAttribute{
                   Computed:            true,
                   MarkdownDescription: "Network base IP address (parsed from subnet).",
               },
               "netmask_bits": schema.Int64Attribute{
                   Computed:            true,
                   MarkdownDescription: "Network mask bits (parsed from subnet).",
               },
               "dhcp_enabled": schema.BoolAttribute{
                   Computed:            true,
                   MarkdownDescription: "Whether DHCP is enabled (derived from dhcp_range_start/end).",
               },
               "dhcp_range_start": schema.StringAttribute{
                   Optional:            true,
                   MarkdownDescription: "DHCP pool start IP address.",
               },
               "dhcp_range_end": schema.StringAttribute{
                   Optional:            true,
                   MarkdownDescription: "DHCP pool end IP address.",
               },
               // ... remaining attributes (mtu, gateway, notes, computed fields)
           },
       }
   }
   ```

3. **Create Method** (minimal: hardcoded UUID for initial green phase)
   ```go
   func (r *CMNetNetworkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
       var data CMNetNetworkResourceModel
       resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
       if resp.Diagnostics.HasError() {
           return
       }

       // Minimal implementation: just set computed fields
       data.ID = types.StringValue("network-test-uuid")
       data.UUID = types.StringValue("network-test-uuid")
       data.DHCPEnabled = types.BoolValue(false)
       data.BaseType = types.StringValue("Network")

       tflog.Trace(ctx, "created network resource")
       resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
   }
   ```

4. **Read Method** (minimal: no-op, return existing state)
   ```go
   func (r *CMNetNetworkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
       var data CMNetNetworkResourceModel
       resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
       if resp.Diagnostics.HasError() {
           return
       }

       // Minimal implementation: state already has data
       resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
   }
   ```

5. **Update Method** (minimal: accept new values)
   ```go
   func (r *CMNetNetworkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
       var data CMNetNetworkResourceModel
       resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
       if resp.Diagnostics.HasError() {
           return
       }

       tflog.Trace(ctx, "updated network resource")
       resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
   }
   ```

6. **Delete Method** (minimal: no-op)
   ```go
   func (r *CMNetNetworkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
       var data CMNetNetworkResourceModel
       resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
       if resp.Diagnostics.HasError() {
           return
       }

       tflog.Trace(ctx, "deleted network resource")
   }
   ```

7. **ImportState Method**
   ```go
   func (r *CMNetNetworkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
       resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
   }
   ```

8. **Register Resource in Provider**
   ```go
   // File: internal/provider/provider.go
   func (p *BCMProvider) Resources(ctx context.Context) []func() resource.Resource {
       return []func() resource.Resource{
           NewCMDeviceCategoryResource,
           NewCMPartSoftwareImageResource,
           NewCMNetNetworkResource, // NEW
           // ...
       }
   }
   ```

**Expected Result**: Tests still FAIL but with different errors (missing API calls, wrong values)

**Verification Command**:
```bash
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMNetNetwork_Basic
# Expected: Test runs but fails validation checks
```

---

### TDD REFACTOR Phase: Full CRUD Implementation

**Goal**: Implement full BCM API integration, CIDR parsing, DHCP logic, error handling

**Refactoring Tasks**:

1. **CIDR Parsing Helper**
   ```go
   // parseCIDR converts CIDR notation to baseAddress and netmaskBits
   func parseCIDR(cidr string) (baseAddress string, netmaskBits int, err error) {
       ip, ipnet, err := net.ParseCIDR(cidr)
       if err != nil {
           return "", 0, fmt.Errorf("invalid CIDR notation: %w", err)
       }

       baseAddress = ip.String()
       maskBits, _ := ipnet.Mask.Size()
       return baseAddress, maskBits, nil
   }

   // formatCIDR reconstructs CIDR from baseAddress and netmaskBits
   func formatCIDR(baseAddress string, netmaskBits int64) string {
       return fmt.Sprintf("%s/%d", baseAddress, netmaskBits)
   }
   ```

2. **DHCP Enabled Derivation**
   ```go
   func isDHCPEnabled(rangeStart, rangeEnd string) bool {
       return rangeStart != "" && rangeStart != "0.0.0.0" &&
              rangeEnd != "" && rangeEnd != "0.0.0.0"
   }
   ```

3. **Build API Entity Helper**
   ```go
   func buildNetworkAPIEntity(ctx context.Context, data CMNetNetworkResourceModel) (map[string]interface{}, diag.Diagnostics) {
       var diags diag.Diagnostics
       entity := make(map[string]interface{})

       // Required fields
       entity["name"] = data.Name.ValueString()

       // Parse subnet if provided
       if !data.Subnet.IsNull() && !data.Subnet.IsUnknown() {
           baseAddr, maskBits, err := parseCIDR(data.Subnet.ValueString())
           if err != nil {
               diags.AddError("Invalid Subnet", err.Error())
               return nil, diags
           }
           entity["baseAddress"] = baseAddr
           entity["netmaskBits"] = maskBits
       }

       // Optional fields
       if !data.Gateway.IsNull() {
           entity["gateway"] = data.Gateway.ValueString()
       }
       if !data.MTU.IsNull() {
           entity["mtu"] = data.MTU.ValueInt64()
       }
       if !data.DHCPRangeStart.IsNull() {
           entity["dynamicRangeStart"] = data.DHCPRangeStart.ValueString()
       }
       if !data.DHCPRangeEnd.IsNull() {
           entity["dynamicRangeEnd"] = data.DHCPRangeEnd.ValueString()
       }
       if !data.Notes.IsNull() {
           entity["notes"] = data.Notes.ValueString()
       }

       // BCM entity structure requirements
       entity["baseType"] = "Network"
       entity["childType"] = ""
       entity["modified"] = true
       entity["to_be_removed"] = false

       // Include UUID for updates
       if !data.UUID.IsNull() {
           entity["uuid"] = data.UUID.ValueString()
       }
       if !data.Revision.IsNull() {
           entity["revision"] = data.Revision.ValueString()
       }

       return entity, diags
   }
   ```

4. **Map API Response to State**
   ```go
   func mapNetworkAPIResponseToState(ctx context.Context, apiData map[string]interface{}, data *CMNetNetworkResourceModel) diag.Diagnostics {
       var diags diag.Diagnostics

       // Identity fields
       if uuid, ok := apiData["uuid"].(string); ok {
           data.UUID = types.StringValue(uuid)
           data.ID = types.StringValue(uuid)
       }
       if name, ok := apiData["name"].(string); ok {
           data.Name = types.StringValue(name)
       }

       // Network addressing
       baseAddr := getStringValue(apiData, "baseAddress")
       netmaskBits := getInt64Value(apiData, "netmaskBits")

       data.BaseAddress = baseAddr
       data.NetmaskBits = netmaskBits

       // Reconstruct subnet for user-facing attribute
       if !baseAddr.IsNull() && !netmaskBits.IsNull() {
           subnet := formatCIDR(baseAddr.ValueString(), netmaskBits.ValueInt64())
           data.Subnet = types.StringValue(subnet)
       }

       // DHCP configuration
       data.DHCPRangeStart = getStringValue(apiData, "dynamicRangeStart")
       data.DHCPRangeEnd = getStringValue(apiData, "dynamicRangeEnd")

       // Derive DHCP enabled
       dhcpEnabled := isDHCPEnabled(
           data.DHCPRangeStart.ValueString(),
           data.DHCPRangeEnd.ValueString(),
       )
       data.DHCPEnabled = types.BoolValue(dhcpEnabled)

       // Optional fields
       data.Gateway = getStringValue(apiData, "gateway")
       data.MTU = getInt64Value(apiData, "mtu")
       data.DomainName = getStringValue(apiData, "domainName")
       data.Notes = getStringValue(apiData, "notes")

       // Computed BCM fields
       data.NetworkType = getStringValue(apiData, "type")
       data.Management = getBoolValue(apiData, "management")
       data.Bootable = getBoolValue(apiData, "bootable")
       data.BaseType = getStringValue(apiData, "baseType")
       data.ChildType = getStringValue(apiData, "childType")
       data.Revision = getStringValue(apiData, "revision")
       data.Modified = getBoolValue(apiData, "modified")
       data.ToBeRemoved = getBoolValue(apiData, "to_be_removed")

       return diags
   }
   ```

5. **Full Create Implementation**
   ```go
   func (r *CMNetNetworkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
       var data CMNetNetworkResourceModel
       resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
       if resp.Diagnostics.HasError() {
           return
       }

       // Build API entity
       entity, diags := buildNetworkAPIEntity(ctx, data)
       resp.Diagnostics.Append(diags...)
       if resp.Diagnostics.HasError() {
           return
       }

       // Call BCM API to create network
       body, err := r.client.CallJSONRPC(ctx, "cmnet", "addNetwork", entity, false)
       if err != nil {
           resp.Diagnostics.AddError(
               "Error Creating Network",
               "Could not create network via BCM API: "+err.Error(),
           )
           return
       }

       // Parse response
       var createdNetwork map[string]interface{}
       if err := json.Unmarshal(body, &createdNetwork); err != nil {
           resp.Diagnostics.AddError(
               "Error Parsing Create Response",
               "Could not parse BCM API response: "+err.Error(),
           )
           return
       }

       // Map response to state
       resp.Diagnostics.Append(mapNetworkAPIResponseToState(ctx, createdNetwork, &data)...)
       if resp.Diagnostics.HasError() {
           return
       }

       tflog.Trace(ctx, "created network resource", map[string]interface{}{
           "uuid": data.UUID.ValueString(),
           "name": data.Name.ValueString(),
       })

       resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
   }
   ```

6. **Full Read Implementation** (with args parameter for efficient lookup)
   ```go
   func (r *CMNetNetworkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
       var data CMNetNetworkResourceModel
       resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
       if resp.Diagnostics.HasError() {
           return
       }

       // Call BCM API to get network by UUID (args parameter)
       body, err := r.client.CallJSONRPC(ctx, "cmnet", "getNetwork", data.UUID.ValueString())
       if err != nil {
           // Network not found = removed outside Terraform
           resp.State.RemoveResource(ctx)
           return
       }

       // Parse response
       var network map[string]interface{}
       if err := json.Unmarshal(body, &network); err != nil {
           resp.Diagnostics.AddError(
               "Error Parsing Read Response",
               "Could not parse BCM API response: "+err.Error(),
           )
           return
       }

       // Map response to state
       resp.Diagnostics.Append(mapNetworkAPIResponseToState(ctx, network, &data)...)
       if resp.Diagnostics.HasError() {
           return
       }

       resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
   }
   ```

7. **Full Update Implementation**
   ```go
   func (r *CMNetNetworkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
       var data CMNetNetworkResourceModel
       resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
       if resp.Diagnostics.HasError() {
           return
       }

       // Build API entity (includes UUID and revision from state)
       entity, diags := buildNetworkAPIEntity(ctx, data)
       resp.Diagnostics.Append(diags...)
       if resp.Diagnostics.HasError() {
           return
       }

       // Call BCM API to update network
       body, err := r.client.CallJSONRPC(ctx, "cmnet", "updateNetwork", entity, false)
       if err != nil {
           resp.Diagnostics.AddError(
               "Error Updating Network",
               "Could not update network via BCM API: "+err.Error(),
           )
           return
       }

       // Parse response
       var updatedNetwork map[string]interface{}
       if err := json.Unmarshal(body, &updatedNetwork); err != nil {
           resp.Diagnostics.AddError(
               "Error Parsing Update Response",
               "Could not parse BCM API response: "+err.Error(),
           )
           return
       }

       // Map response to state
       resp.Diagnostics.Append(mapNetworkAPIResponseToState(ctx, updatedNetwork, &data)...)
       if resp.Diagnostics.HasError() {
           return
       }

       tflog.Trace(ctx, "updated network resource", map[string]interface{}{
           "uuid": data.UUID.ValueString(),
       })

       resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
   }
   ```

8. **Full Delete Implementation**
   ```go
   func (r *CMNetNetworkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
       var data CMNetNetworkResourceModel
       resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
       if resp.Diagnostics.HasError() {
           return
       }

       // Call BCM API to delete network (force=false for safety)
       _, err := r.client.CallJSONRPC(ctx, "cmnet", "removeNetwork", data.UUID.ValueString(), false)
       if err != nil {
           resp.Diagnostics.AddError(
               "Error Deleting Network",
               "Could not delete network via BCM API: "+err.Error()+
               "\nIf network has active assignments, manual cleanup may be required.",
           )
           return
       }

       tflog.Trace(ctx, "deleted network resource", map[string]interface{}{
           "uuid": data.UUID.ValueString(),
       })
   }
   ```

**Expected Result**: All 11 acceptance tests PASS

**Verification Command**:
```bash
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMNetNetwork
# Expected: 11 passes
```

---

## Phase 3: Examples & Documentation

**Prerequisites**: Phase 2 complete, all tests passing

### 1. Create Example Configurations

**File**: `examples/resources/bcm_cmnet_network/basic.tf`
```hcl
resource "bcm_cmnet_network" "example" {
  name = "basic-network"
}
```

**File**: `examples/resources/bcm_cmnet_network/subnet.tf`
```hcl
resource "bcm_cmnet_network" "example" {
  name    = "compute-network"
  subnet  = "10.0.1.0/24"
  gateway = "10.0.1.1"
  mtu     = 9000
  notes   = "High-performance compute network with jumbo frames"
}
```

**File**: `examples/resources/bcm_cmnet_network/dhcp.tf`
```hcl
resource "bcm_cmnet_network" "example" {
  name    = "dhcp-network"
  subnet  = "192.168.100.0/24"
  gateway = "192.168.100.1"

  dhcp_range_start = "192.168.100.100"
  dhcp_range_end   = "192.168.100.200"
}
```

**File**: `examples/resources/bcm_cmnet_network/complete.tf`
```hcl
resource "bcm_cmnet_network" "example" {
  name        = "production-network"
  subnet      = "10.10.0.0/16"
  gateway     = "10.10.0.1"
  mtu         = 9000
  domain_name = "prod.cluster.local"

  dhcp_range_start = "10.10.10.100"
  dhcp_range_end   = "10.10.10.200"

  notes = "Production cluster network - managed by Terraform"
}
```

### 2. Generate Documentation

**Command**:
```bash
make generate
```

**Expected Output**:
- `docs/resources/cmnet_network.md` - Auto-generated from schema + examples
- Formatted examples in `examples/resources/bcm_cmnet_network/`
- Copyright headers added via copywrite

**Manual Verification**:
```bash
# Verify documentation matches schema
cat docs/resources/cmnet_network.md

# Verify examples are formatted
terraform fmt -check examples/resources/bcm_cmnet_network/
```

---

## Phase 4: Integration Testing & Validation

**Prerequisites**: Phase 3 complete, documentation generated

### Validation Checklist

- [ ] All 11 acceptance tests pass with 100% success rate
- [ ] Example configurations validate successfully (`terraform validate`)
- [ ] Import functionality tested with real BCM network
- [ ] Drift detection verified with external modification
- [ ] Documentation accurately reflects schema attributes
- [ ] No golangci-lint warnings (`make lint`)
- [ ] Code formatted correctly (`make fmt`)
- [ ] Pre-commit hooks pass (`pre-commit run --all-files`)

### Real-World Testing Scenarios

1. **Scenario: Create Basic Network**
   ```bash
   cd examples/resources/bcm_cmnet_network
   terraform init
   terraform apply -auto-approve
   terraform show
   terraform destroy -auto-approve
   ```

2. **Scenario: Import Existing Network**
   ```bash
   # Manually create network via BCM UI/API, note UUID
   terraform import bcm_cmnet_network.imported <uuid>
   terraform plan  # Should show no changes if config matches
   ```

3. **Scenario: Drift Detection**
   ```bash
   terraform apply -auto-approve
   # Manually modify MTU via BCM UI
   terraform plan  # Should detect drift and propose correction
   terraform apply -auto-approve  # Restore configuration
   ```

4. **Scenario: DHCP Configuration Change**
   ```bash
   # Create network without DHCP
   terraform apply -auto-approve
   # Update config to enable DHCP
   terraform apply -auto-approve
   # Update config to disable DHCP
   terraform apply -auto-approve
   ```

---

## Success Criteria Validation

After Phase 4 completion, verify all spec success criteria:

- [x] **SC-001**: Network creation completes in <30 seconds
- [x] **SC-002**: All CRUD operations pass with 100% acceptance test rate
- [x] **SC-003**: Import functionality works for all existing networks
- [x] **SC-004**: Drift detection identifies external changes within one plan cycle
- [x] **SC-005**: Idempotency verified via plancheck.ExpectEmptyPlan (100% pass rate)
- [x] **SC-006**: Documentation auto-generated successfully via `make generate`
- [x] **SC-007**: DHCP changes apply within 60 seconds (tested in acceptance tests)
- [x] **SC-008**: All 11 test scenarios pass with 100% success rate

---

## Troubleshooting Guide

### Common Issues

**Issue**: Tests fail with "invalid CIDR notation"
- **Cause**: Subnet validation regex not matching input
- **Fix**: Verify `net.ParseCIDR()` accepts the format, update regex in validator

**Issue**: Drift detection doesn't detect external changes
- **Cause**: Read method not fetching latest state from BCM API
- **Fix**: Ensure `getNetwork(uuid)` called on every Read, not cached

**Issue**: DHCP enabled stays false despite range configuration
- **Cause**: `isDHCPEnabled()` logic not correctly evaluating conditions
- **Fix**: Debug logic, ensure `dynamicRangeStart/End` parsed from API response

**Issue**: Import fails with "resource not found"
- **Cause**: ImportState using wrong identifier or API call failing
- **Fix**: Verify UUID format, check BCM API response for `getNetwork(uuid)`

**Issue**: Update shows plan diff even when no changes
- **Cause**: Computed fields changing on Update or state drift
- **Fix**: Use planmodifier.UseStateForUnknown() for stable computed fields

---

## Risk Mitigation

| Risk | Impact | Mitigation |
|------|--------|------------|
| VLAN field not in BCM API | Medium | Phase 0 research confirms availability; if missing, document as OUT OF SCOPE |
| CIDR parsing edge cases | Low | Use Go standard library `net.ParseCIDR` (well-tested, handles validation) |
| BCM API timeout on operations | Medium | Implement retry logic in BCM client for transient failures |
| Network deletion with active nodes | High | Use `force=false` by default; provide clear error message for dependency conflicts |
| Concurrent network modifications | Medium | BCM API handles via revision/locking; tests verify behavior |
| Test environment unavailability | High | Use existing test infrastructure (172.21.15.254); fallback to mock client if needed |

---

## Definition of Done

Implementation is complete when:

1. ✅ All 11 acceptance tests pass with TF_ACC=1
2. ✅ Resource registered in provider.go Resources() method
3. ✅ Examples created for all major use cases (basic, subnet, DHCP, complete)
4. ✅ Documentation auto-generated via `make generate`
5. ✅ Code passes `make lint` and `make fmt` checks
6. ✅ Pre-commit hooks pass without warnings
7. ✅ Import functionality verified with real BCM network
8. ✅ Drift detection tested with external modification
9. ✅ CIDR parsing logic tested for edge cases (invalid formats, boundary values)
10. ✅ Error messages provide actionable guidance for users

---

## Next Steps for Autonomous Implementation

**Command**: `/speckit.tasks`

This command will generate `tasks.md` with granular task breakdown suitable for `/speckit.implement` autonomous execution.

**Then**: `/speckit.implement`

This command will execute all tasks in RED-GREEN-REFACTOR order with parallel execution where possible.
