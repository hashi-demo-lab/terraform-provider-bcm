# Feature Specification: BCM Device Roles Block

**Feature Branch**: `039-device-roles-block`
**Created**: 2025-11-26
**Status**: Draft
**Input**: User description: "Enhance bcm_cmdevice_device resource to add roles support for Kubernetes cluster topology and service role assignment"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Assign Kubernetes Roles to Control Plane Nodes (Priority: P1)

Infrastructure engineers deploying Kubernetes on BCM need to assign control-plane, master, and etcd roles to designated nodes. This enables BCM to configure the node correctly for Kubernetes control plane functions.

**Why this priority**: This is the primary use case blocking DGX deployment. Without role assignment, engineers cannot define cluster topology in Terraform.

**Independent Test**: Can be fully tested by creating a device with `roles = ["control-plane", "master", "etcd"]` and verifying BCM assigns the roles correctly. Delivers immediate value for Kubernetes deployments.

**Acceptance Scenarios**:

1. **Given** a BCM cluster with Kubernetes role types available, **When** user creates a device with `roles = ["control-plane", "master", "etcd"]`, **Then** device is created with all three roles assigned
2. **Given** an existing device without roles, **When** user updates device to add `roles = ["control-plane"]`, **Then** BCM assigns the control-plane role to the device
3. **Given** a device with roles assigned, **When** user reads the device, **Then** all assigned roles are returned in state

---

### User Story 2 - Assign Worker Roles to Compute Nodes (Priority: P1)

Infrastructure engineers need to designate nodes as Kubernetes workers by assigning the worker role. This completes the cluster topology definition alongside control plane nodes.

**Why this priority**: Worker nodes are equally critical for a functional Kubernetes cluster. Without worker role assignment, compute nodes cannot join the cluster.

**Independent Test**: Can be fully tested by creating a device with `roles = ["worker"]` and verifying BCM configures the node as a Kubernetes worker.

**Acceptance Scenarios**:

1. **Given** a BCM cluster with worker role type available, **When** user creates a device with `roles = ["worker"]`, **Then** device is created with worker role assigned
2. **Given** multiple worker nodes need configuration, **When** user creates devices with identical `roles = ["worker"]` configuration, **Then** each device is independently assigned the worker role

---

### User Story 3 - Update Device Roles (Priority: P2)

Infrastructure engineers may need to modify role assignments during cluster lifecycle events such as promoting a worker to control-plane or removing a role during maintenance.

**Why this priority**: Role updates are important for cluster operations but secondary to initial deployment capability.

**Independent Test**: Can be tested by creating a device with one role set, then updating to a different role set and verifying the change.

**Acceptance Scenarios**:

1. **Given** a device with `roles = ["worker"]`, **When** user updates to `roles = ["control-plane", "master"]`, **Then** worker role is removed and new roles are assigned
2. **Given** a device with multiple roles, **When** user updates to empty `roles = []`, **Then** all roles are removed from the device
3. **Given** a device without roles block, **When** user adds `roles = ["worker"]`, **Then** the role is assigned without recreating the device

---

### User Story 4 - Import Device with Existing Roles (Priority: P2)

Infrastructure engineers migrating to Terraform need to import existing devices with their current role assignments preserved.

**Why this priority**: Import capability enables Terraform adoption for existing BCM infrastructure.

**Independent Test**: Can be tested by importing an existing device with roles and verifying the roles appear in state.

**Acceptance Scenarios**:

1. **Given** an existing device with roles assigned in BCM, **When** user imports the device, **Then** all existing roles are populated in Terraform state
2. **Given** an imported device, **When** user runs `terraform plan`, **Then** no changes are detected for existing role assignments

---

### User Story 5 - Drift Detection for Role Changes (Priority: P3)

When roles are modified outside of Terraform (via BCM UI or API), Terraform should detect the drift and propose corrections.

**Why this priority**: Drift detection ensures configuration consistency but is less critical than core CRUD operations.

**Independent Test**: Can be tested by modifying roles via BCM API, then running `terraform plan` to verify drift is detected.

**Acceptance Scenarios**:

1. **Given** a device with `roles = ["worker"]` in Terraform, **When** an admin adds "control-plane" role via BCM UI, **Then** `terraform plan` detects the drift
2. **Given** detected drift, **When** user runs `terraform apply`, **Then** Terraform restores the desired role configuration

---

### Edge Cases

- What happens when a role name does not exist in BCM?
  - BCM API returns validation error; provider surfaces this as Terraform error before apply
- What happens when duplicate role names are specified?
  - Provider validates and removes duplicates before API call
- What happens when roles field is omitted vs. empty array?
  - Omitted: BCM uses defaults (no roles); Empty array: explicitly removes all roles
- How does system handle role assignment on device creation failure?
  - Atomic operation - if device creation fails, no partial state is created

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Resource MUST accept an optional `roles` attribute as a list of role names (strings)
- **FR-002**: Resource MUST support role names like "control-plane", "master", "worker", "etcd", "headnode", "storage", etc.
- **FR-003**: System MUST assign specified roles to the device during CREATE operation
- **FR-004**: System MUST update role assignments during UPDATE operation (add/remove roles as needed)
- **FR-005**: System MUST read current role assignments from BCM and reflect them in state
- **FR-006**: System MUST detect drift when roles are modified outside Terraform
- **FR-007**: System MUST preserve existing roles during import
- **FR-008**: System MUST handle empty roles list (`roles = []`) as "no roles assigned"
- **FR-009**: System MUST validate role names are non-empty strings
- **FR-010**: System MUST deduplicate role names before sending to BCM API

### Key Entities

- **Role**: A service role that can be assigned to a device
  - **Attributes**: uuid (BCM-generated), name (user-specified), childType (e.g., "HeadNodeRole", "ComputeRole"), baseType (always "Role")
  - **Relationship**: Embedded in Device entity as `roles` array
  - **Assignment**: Roles are specified by name; BCM resolves name to UUID internally

- **Device.roles**: Array field on Device entity containing assigned Role objects
  - **On Create**: Provider includes roles array in device entity sent to `addDevice`
  - **On Update**: Provider includes roles array in device entity sent to `updateDevice`
  - **On Read**: Provider extracts roles from `getDevice` response

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can assign roles to devices in under 30 seconds (single API call)
- **SC-002**: Users can define complete Kubernetes cluster topology (control-plane + workers) in single Terraform apply
- **SC-003**: Role changes are detected within standard Terraform plan cycle
- **SC-004**: 100% of supported role types can be assigned via Terraform
- **SC-005**: Zero role assignment failures due to provider bugs (API validation errors are user errors)
- **SC-006**: Import preserves all existing role assignments accurately

## Assumptions

- **ASSUME-001**: BCM API accepts role assignments as part of addDevice/updateDevice operations (roles embedded in device entity)
- **ASSUME-002**: Role names are resolved to UUIDs by BCM server-side (provider sends role names, not UUIDs)
- **ASSUME-003**: BCM returns assigned roles in getDevice response with full role objects (uuid, name, childType, baseType)
- **ASSUME-004**: Role assignment is atomic - either all specified roles are assigned or operation fails
- **ASSUME-005**: Standard BCM role types are available: HeadNodeRole, StorageRole, BackupRole, MonitoringRole, ProvisioningRole, BootRole, ComputeRole
- **ASSUME-006**: Kubernetes-specific roles (control-plane, master, worker, etcd) are available in BCM clusters configured for Kubernetes

## API Contract

### BCM API Pattern

The `roles` field is embedded in the Device entity. Roles are assigned by including them in the device object sent to addDevice/updateDevice.

**Service**: `cmdevice`
**Methods**: `addDevice`, `updateDevice`, `getDevice`

### Role Object Structure

Based on existing BCM API patterns (from data_source_cmdevice_roles.go):

```json
{
  "baseType": "Role",
  "childType": "HeadNodeRole",
  "name": "headnode",
  "uuid": "a458760f-3898-4870-bd7d-8127a5ea44b8",
  "addServices": true,
  "modified": false,
  "to_be_removed": false,
  "revision": ""
}
```

### Device Entity with Roles

```json
{
  "baseType": "Device",
  "childType": "PhysicalNode",
  "hostname": "control-01",
  "uuid": "...",
  "roles": [
    {
      "baseType": "Role",
      "name": "control-plane",
      "modified": true,
      "to_be_removed": false
    },
    {
      "baseType": "Role",
      "name": "master",
      "modified": true,
      "to_be_removed": false
    }
  ],
  ...
}
```

**Key Implementation Notes**:
- When creating/updating, include roles array with role name and standard entity fields
- BCM resolves role name to UUID and fills in childType
- On read, parse full role objects from response to populate state

### Known Role Types

| Name Pattern | childType             | Description                    |
| ------------ | --------------------- | ------------------------------ |
| headnode     | HeadNodeRole          | Cluster management node        |
| storage      | StorageRole           | NFS storage services           |
| backup       | BackupRole            | Backup services                |
| monitoring   | MonitoringRole        | Cluster monitoring agent       |
| provisioning | ProvisioningRole      | Node provisioning services     |
| boot         | BootRole              | PXE boot services              |
| compute      | ComputeRole           | Compute workload execution     |
| control-plane| KubeControlPlaneRole  | Kubernetes control plane       |
| master       | KubeMasterRole        | Kubernetes master node         |
| worker       | KubeWorkerRole        | Kubernetes worker node         |
| etcd         | KubeEtcdRole          | Kubernetes etcd node           |

## Terraform Schema

### Resource: `bcm_cmdevice_device` (Updated)

**New Attribute**:

```hcl
roles = optional(list(string))
```

**Schema Definition**:

```go
"roles": schema.ListAttribute{
    ElementType:         types.StringType,
    Optional:            true,
    MarkdownDescription: "List of role names to assign to the device (e.g., 'control-plane', 'worker', 'etcd'). Use data.bcm_cmdevice_roles to discover available role names.",
},
```

**Behavior**:
- When omitted: Device has no roles (BCM defaults)
- When empty (`[]`): Device has no roles (explicit)
- When populated: Device is assigned specified roles

## Example Usage

### Kubernetes Control Plane Node

```hcl
resource "bcm_cmdevice_device" "control" {
  hostname           = "control-01"
  mac                = "00:11:22:33:44:55"
  category           = bcm_cmdevice_category.kube_control.uuid
  management_network = data.bcm_cmnet_networks.mgmt.networks[0].uuid

  roles = ["control-plane", "master", "etcd"]
}
```

### Kubernetes Worker Node

```hcl
resource "bcm_cmdevice_device" "worker" {
  count = 3

  hostname           = "worker-0${count.index + 1}"
  mac                = var.worker_macs[count.index]
  category           = bcm_cmdevice_category.kube_worker.uuid
  management_network = data.bcm_cmnet_networks.mgmt.networks[0].uuid

  roles = ["worker"]
}
```

### Multi-Role Assignment

```hcl
resource "bcm_cmdevice_device" "head" {
  hostname           = "head-01"
  mac                = "00:11:22:33:44:66"
  category           = bcm_cmdevice_category.head.uuid
  management_network = data.bcm_cmnet_networks.mgmt.networks[0].uuid

  roles = ["headnode", "storage", "monitoring"]
}
```

### Dynamic Role Discovery

```hcl
# Discover available roles
data "bcm_cmdevice_roles" "kube" {
  name_pattern = "*"
  child_type   = "KubeWorkerRole"
}

# Use discovered role names
resource "bcm_cmdevice_device" "worker" {
  hostname           = "worker-01"
  mac                = "00:11:22:33:44:77"
  category           = bcm_cmdevice_category.worker.uuid
  management_network = data.bcm_cmnet_networks.mgmt.networks[0].uuid

  roles = [data.bcm_cmdevice_roles.kube.roles[0].name]
}
```

## Implementation Notes

### State Model Changes

Add to `CMDeviceDeviceResourceModel`:

```go
// Roles - list of role names assigned to this device
Roles types.List `tfsdk:"roles"` // List of StringType
```

### API Entity Building

In `buildDeviceAPIEntity`:

```go
// Add roles if specified
if !plan.Roles.IsNull() && !plan.Roles.IsUnknown() {
    var roleNames []string
    plan.Roles.ElementsAs(ctx, &roleNames, false)

    rolesArray := make([]map[string]interface{}, 0, len(roleNames))
    for _, roleName := range roleNames {
        role := map[string]interface{}{
            "baseType":      "Role",
            "name":          roleName,
            "modified":      true,
            "to_be_removed": false,
            "revision":      "",
        }
        rolesArray = append(rolesArray, role)
    }
    entity["roles"] = rolesArray
}
```

### Response Parsing

In `parseDeviceFromAPI`:

```go
// Parse roles from BCM response
if rolesData, ok := data["roles"].([]interface{}); ok {
    roleNames := make([]string, 0, len(rolesData))
    for _, roleData := range rolesData {
        if role, ok := roleData.(map[string]interface{}); ok {
            if name, ok := role["name"].(string); ok && name != "" {
                roleNames = append(roleNames, name)
            }
        }
    }
    // Sort for consistent state comparison
    sort.Strings(roleNames)
    model.Roles, _ = types.ListValueFrom(ctx, types.StringType, roleNames)
} else {
    model.Roles = types.ListNull(types.StringType)
}
```

### Drift Detection

The standard Read operation will detect drift because:
1. Read calls `getDevice` which returns current roles
2. `parseDeviceFromAPI` extracts role names from response
3. Terraform compares state (from last apply) with refreshed state (from Read)
4. If roles differ, plan shows changes

## Testing Strategy

### Acceptance Tests

1. **TestAccCMDeviceDevice_RolesCreate**: Create device with roles, verify roles assigned
2. **TestAccCMDeviceDevice_RolesUpdate**: Create device, update roles, verify change
3. **TestAccCMDeviceDevice_RolesRemove**: Create device with roles, update to empty roles, verify removal
4. **TestAccCMDeviceDevice_RolesImport**: Create device with roles via API, import, verify state
5. **TestAccCMDeviceDevice_RolesDrift**: Create device, modify roles via API, verify drift detection
6. **TestAccCMDeviceDevice_RolesMultiple**: Create device with multiple roles, verify all assigned
7. **TestAccCMDeviceDevice_RolesIdempotent**: Create device with roles, re-apply same config, verify no changes

### Test Configuration Pattern

```go
func testAccCMDeviceDeviceConfigWithRoles(hostname, mac, categoryUUID, networkUUID string, roles []string) string {
    rolesHCL := ""
    if len(roles) > 0 {
        roleStrings := make([]string, len(roles))
        for i, r := range roles {
            roleStrings[i] = fmt.Sprintf("%q", r)
        }
        rolesHCL = fmt.Sprintf("roles = [%s]", strings.Join(roleStrings, ", "))
    }

    return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

resource "bcm_cmdevice_device" "test" {
  hostname           = %[4]q
  mac                = %[5]q
  category           = %[6]q
  management_network = %[7]q
  %[8]s
}
`,
        os.Getenv("BCM_ENDPOINT"),
        os.Getenv("BCM_USERNAME"),
        os.Getenv("BCM_PASSWORD"),
        hostname,
        mac,
        categoryUUID,
        networkUUID,
        rolesHCL,
    )
}
```

## Dependencies

- **Existing Implementation**: `internal/provider/resource_cmdevice_device.go` - Device resource to enhance
- **Reference Implementation**: `internal/provider/data_source_cmdevice_roles.go` - Role data extraction pattern
- **Test Helpers**: `internal/provider/test_helpers.go` - Test utilities
- **BCM API**: `cmdevice.addDevice`, `cmdevice.updateDevice`, `cmdevice.getDevice`

## Out of Scope

- Role creation/deletion (roles are managed by BCM, not Terraform)
- Role metadata management (addServices, childType configuration)
- Role-based RBAC (outside BCM provider scope)
- Role ordering/priority (BCM does not support role ordering)
