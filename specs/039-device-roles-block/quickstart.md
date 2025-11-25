# Quick Start: BCM Device Roles Block

**Feature**: 039-device-roles-block
**Date**: 2025-11-26

## Overview

This guide helps developers implement and test the device roles feature for the BCM Terraform provider.

## Prerequisites

- Go 1.24+
- Access to BCM cluster (172.21.15.254:8081)
- BCM credentials configured
- Terraform CLI installed

## Environment Setup

```bash
# Set required environment variables
export BCM_ENDPOINT="https://172.21.15.254:8081"
export BCM_USERNAME="root"
export BCM_PASSWORD="Hashicorp123!"
export TF_ACC=1

# Go environment (if needed)
export GOMODCACHE=/workspace/.go/pkg/mod
export GOCACHE=/workspace/.go/cache
export GOPATH=/workspace/.go
```

## Files to Modify

| File | Changes |
|------|---------|
| `internal/provider/resource_cmdevice_device.go` | Add roles field, schema, build/parse logic |
| `internal/provider/resource_cmdevice_device_test.go` | Add roles acceptance tests |
| `examples/resources/bcm_cmdevice_device/resource_with_roles.tf` | Add example |

## Implementation Steps

### Step 1: Add Model Field

In `resource_cmdevice_device.go`, add to `CMDeviceDeviceResourceModel`:

```go
// After Interfaces field (~line 72)
Roles types.List `tfsdk:"roles"` // List of StringType
```

### Step 2: Add Schema Attribute

In the `Schema()` method, add to `Attributes`:

```go
"roles": schema.ListAttribute{
    ElementType:         types.StringType,
    Optional:            true,
    MarkdownDescription: "List of role names to assign to the device " +
        "(e.g., 'control-plane', 'worker', 'etcd'). " +
        "Use data.bcm_cmdevice_roles to discover available role names.",
},
```

### Step 3: Add Build Logic

In `buildDeviceAPIEntityWithExisting()`, add after interfaces handling:

```go
// Add roles if specified
if !plan.Roles.IsNull() && !plan.Roles.IsUnknown() {
    var roleNames []string
    plan.Roles.ElementsAs(ctx, &roleNames, false)

    // Deduplicate
    seen := make(map[string]bool)
    rolesArray := make([]map[string]interface{}, 0)
    for _, name := range roleNames {
        if !seen[name] && name != "" {
            seen[name] = true
            rolesArray = append(rolesArray, map[string]interface{}{
                "baseType":      "Role",
                "name":          name,
                "modified":      true,
                "to_be_removed": false,
                "revision":      "",
            })
        }
    }
    entity["roles"] = rolesArray
}
```

### Step 4: Add Parse Logic

In `parseDeviceFromAPI()`, add after interfaces parsing:

```go
// Parse roles from BCM response
if rolesData, ok := data["roles"].([]interface{}); ok && len(rolesData) > 0 {
    roleNames := make([]string, 0, len(rolesData))
    for _, roleData := range rolesData {
        if role, ok := roleData.(map[string]interface{}); ok {
            if name, ok := role["name"].(string); ok && name != "" {
                roleNames = append(roleNames, name)
            }
        }
    }
    sort.Strings(roleNames)
    model.Roles, _ = types.ListValueFrom(ctx, types.StringType, roleNames)
} else {
    model.Roles = types.ListNull(types.StringType)
}
```

### Step 5: Handle State Preservation

In Create() and Update(), after reading back device:

```go
// Preserve null roles if not specified in plan
if plan.Roles.IsNull() && !newState.Roles.IsNull() {
    // Check if BCM returned empty roles
    var stateRoles []string
    newState.Roles.ElementsAs(ctx, &stateRoles, false)
    if len(stateRoles) == 0 {
        newState.Roles = types.ListNull(types.StringType)
    }
}
```

## Running Tests

### Run Specific Role Tests

```bash
# Run all roles-related tests
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "Roles"

# Run single test
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceDevice_RolesCreate
```

### Test Order (TDD)

1. Write failing test
2. Run test - verify it fails
3. Implement minimal code
4. Run test - verify it passes
5. Refactor if needed
6. Repeat for next test

## Acceptance Test Template

```go
func TestAccCMDeviceDevice_RolesCreate(t *testing.T) {
    hostname := generateUniqueTestName("roles-test")
    mac := generateUniqueMAC()
    categoryName := generateUniqueTestName("roles-cat")
    imageName := generateUniqueTestName("roles-img")

    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
        Steps: []resource.TestStep{
            {
                Config: testAccCMDeviceDeviceConfigWithRoles(
                    hostname, mac, categoryName, imageName,
                    []string{"headnode", "storage"},
                ),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr(
                        "bcm_cmdevice_device.test", "hostname", hostname),
                    resource.TestCheckResourceAttr(
                        "bcm_cmdevice_device.test", "roles.#", "2"),
                ),
            },
        },
    })
}
```

## Example Configuration

```hcl
# examples/resources/bcm_cmdevice_device/resource_with_roles.tf
resource "bcm_cmdevice_device" "kubernetes_control" {
  hostname           = "control-01"
  mac                = "00:11:22:33:44:55"
  category           = bcm_cmdevice_category.kube.uuid
  management_network = data.bcm_cmnet_networks.mgmt.networks[0].uuid

  roles = ["control-plane", "master", "etcd"]
}

resource "bcm_cmdevice_device" "kubernetes_worker" {
  hostname           = "worker-01"
  mac                = "00:11:22:33:44:66"
  category           = bcm_cmdevice_category.kube.uuid
  management_network = data.bcm_cmnet_networks.mgmt.networks[0].uuid

  roles = ["worker"]
}
```

## Debugging Tips

### Enable Terraform Logging

```bash
export TF_LOG=DEBUG
export TF_LOG_PATH=./terraform.log
```

### Check BCM API Response

```bash
# Login
curl -k -X POST https://172.21.15.254:8081/json \
  -H "Content-Type: application/json" \
  -d '{"service":"login","username":"root","password":"Hashicorp123!"}' \
  -c cookies.txt

# Get device with roles
curl -k -X POST https://172.21.15.254:8081/json \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d '{"service":"cmdevice","call":"getDevice","args":["<uuid>"]}' | jq '.roles'
```

### Common Issues

| Issue | Solution |
|-------|----------|
| Role not found | Use `data.bcm_cmdevice_roles` to verify role exists |
| Drift detected | Check role name sorting - must be alphabetical |
| Empty roles in state | Verify parseDeviceFromAPI handles empty array |
| Import fails | Ensure Read() handles import case (isImport check) |

## Build and Generate Docs

```bash
# Build provider
make build

# Generate documentation
make generate

# Run linter
make lint
```

## Reference Files

- Implementation: `/workspace/internal/provider/resource_cmdevice_device.go`
- Tests: `/workspace/internal/provider/resource_cmdevice_device_test.go`
- Roles data source: `/workspace/internal/provider/data_source_cmdevice_roles.go`
- Spec: `/workspace/specs/039-device-roles-block/spec.md`
- Plan: `/workspace/specs/039-device-roles-block/plan.md`
