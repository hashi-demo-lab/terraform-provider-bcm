# Quickstart Guide: Device Interfaces Block Implementation

**Feature**: 038-device-interfaces-block
**Date**: 2025-11-25
**Audience**: Developers implementing or reviewing this feature

## Overview

This guide helps you get started with implementing the `interfaces` block enhancement for `bcm_cmdevice_device`. Follow the TDD workflow: RED (write failing tests) -> GREEN (implement) -> REFACTOR (improve).

---

## Prerequisites

### Environment Setup

```bash
# Required environment variables for acceptance testing
export TF_ACC=1
export BCM_ENDPOINT="https://172.21.15.254:8081"
export BCM_USERNAME="root"
export BCM_PASSWORD="Hashicorp123!"

# Go environment
export GOMODCACHE=/workspace/.go/pkg/mod
export GOCACHE=/workspace/.go/cache
export GOPATH=/workspace/.go
```

### Verify BCM Connectivity

```bash
# Quick connectivity test
curl -k -X POST "$BCM_ENDPOINT/json" \
  -H "Content-Type: application/json" \
  -d '{"service":"login","username":"root","password":"Hashicorp123!"}'
```

---

## Quick Reference

### Files to Modify

| File | Action | Purpose |
|------|--------|---------|
| `internal/provider/resource_cmdevice_device.go` | MODIFY | Add interfaces block to schema and CRUD |
| `internal/provider/resource_cmdevice_device_interfaces.go` | CREATE | Interface helper functions |
| `internal/provider/resource_cmdevice_device_interfaces_test.go` | CREATE | Interface acceptance tests |

### Key Functions to Implement

1. `DeviceInterfaceModel` struct
2. `InterfacesBlockSchema()` - Returns schema.ListNestedBlock
3. `buildInterfaceAPIEntity()` - Converts model to BCM entity
4. `parseInterfaceFromAPI()` - Converts BCM response to model
5. `interfaceTypeToBCMChildType()` - Type mapping

---

## TDD Workflow

### Phase 1: RED - Write Failing Tests

Create `internal/provider/resource_cmdevice_device_interfaces_test.go`:

```go
package provider

import (
    "fmt"
    "os"
    "testing"

    "github.com/hashicorp/terraform-plugin-testing/helper/resource"
    "github.com/hashicorp/terraform-plugin-testing/knownvalue"
    "github.com/hashicorp/terraform-plugin-testing/statecheck"
    "github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccCMDeviceDevice_InterfaceSingle(t *testing.T) {
    deviceName := generateUniqueTestName("tftest-iface-single")
    categoryName := generateUniqueTestName("tftest-cat-iface")
    imageName := generateUniqueTestName("tftest-img-iface")
    mac := generateUniqueMAC()

    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
        Steps: []resource.TestStep{
            {
                Config: testAccCMDeviceDeviceConfigInterfaceSingle(
                    deviceName, categoryName, imageName, mac,
                ),
                ConfigStateChecks: []statecheck.StateCheck{
                    statecheck.ExpectKnownValue(
                        "bcm_cmdevice_device.test",
                        tfjsonpath.New("interfaces").AtSliceIndex(0).AtMapKey("name"),
                        knownvalue.StringExact("eth0"),
                    ),
                    statecheck.ExpectKnownValue(
                        "bcm_cmdevice_device.test",
                        tfjsonpath.New("interfaces").AtSliceIndex(0).AtMapKey("type"),
                        knownvalue.StringExact("physical"),
                    ),
                    statecheck.ExpectKnownValue(
                        "bcm_cmdevice_device.test",
                        tfjsonpath.New("interfaces").AtSliceIndex(0).AtMapKey("uuid"),
                        knownvalue.NotNull(),
                    ),
                },
            },
        },
    })
}

func testAccCMDeviceDeviceConfigInterfaceSingle(deviceName, categoryName, imageName, mac string) string {
    return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmnet_networks" "management" {
  filter {
    name_pattern = "managementnet"
  }
}

resource "bcm_cmpart_softwareimage" "test" {
  name = %[4]q
  path = "/cm/images/%[4]s.iso"
}

resource "bcm_cmdevice_category" "test" {
  name               = %[5]q
  management_network = data.bcm_cmnet_networks.management.networks[0].id

  software_image_proxy = {
    parent_software_image = bcm_cmpart_softwareimage.test.id
  }

  depends_on = [bcm_cmpart_softwareimage.test]
}

resource "bcm_cmdevice_device" "test" {
  hostname = %[6]q
  category = bcm_cmdevice_category.test.id

  interfaces {
    name     = "eth0"
    type     = "physical"
    mac      = %[7]q
    network  = data.bcm_cmnet_networks.management.networks[0].id
    bootable = true
    dhcp     = true
  }

  depends_on = [bcm_cmdevice_category.test]
}
`,
        os.Getenv("BCM_ENDPOINT"),
        os.Getenv("BCM_USERNAME"),
        os.Getenv("BCM_PASSWORD"),
        imageName,
        categoryName,
        deviceName,
        mac,
    )
}
```

Run the test (expect failure):

```bash
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceDevice_InterfaceSingle
```

### Phase 2: GREEN - Implement Minimum Code

1. **Add DeviceInterfaceModel struct** to `resource_cmdevice_device.go`:

```go
// DeviceInterfaceModel represents a network interface in the device resource
type DeviceInterfaceModel struct {
    Name      types.String `tfsdk:"name"`
    Type      types.String `tfsdk:"type"`
    Network   types.String `tfsdk:"network"`
    MAC       types.String `tfsdk:"mac"`
    IP        types.String `tfsdk:"ip"`
    IPv6IP    types.String `tfsdk:"ipv6_ip"`
    DHCP      types.Bool   `tfsdk:"dhcp"`
    Bootable  types.Bool   `tfsdk:"bootable"`
    StartIf   types.String `tfsdk:"start_if"`
    Members   types.List   `tfsdk:"members"`
    BondMode  types.String `tfsdk:"bond_mode"`
    UUID      types.String `tfsdk:"uuid"`
    BaseType  types.String `tfsdk:"base_type"`
    ChildType types.String `tfsdk:"child_type"`
    CardType  types.String `tfsdk:"cardtype"`
}
```

2. **Add Interfaces field** to `CMDeviceDeviceResourceModel`:

```go
type CMDeviceDeviceResourceModel struct {
    // ... existing fields ...
    Interfaces []DeviceInterfaceModel `tfsdk:"interfaces"`
}
```

3. **Add interfaces block to Schema**:

```go
// In Schema() function, add to Blocks:
Blocks: map[string]schema.Block{
    "interfaces": schema.ListNestedBlock{
        MarkdownDescription: "Network interface configurations",
        NestedObject: schema.NestedBlockObject{
            Attributes: map[string]schema.Attribute{
                "name": schema.StringAttribute{
                    Required:            true,
                    MarkdownDescription: "Interface name",
                },
                "type": schema.StringAttribute{
                    Required:            true,
                    MarkdownDescription: "Interface type: physical, bond, bmc",
                },
                // ... add remaining attributes from data-model.md
            },
        },
    },
},
```

4. **Update buildDeviceAPIEntity** to include interfaces from plan:

```go
// In buildDeviceAPIEntity, replace hardcoded interface:
var interfaces []interface{}
if len(plan.Interfaces) > 0 {
    for _, iface := range plan.Interfaces {
        interfaces = append(interfaces, buildInterfaceAPIEntity(iface, ""))
    }
} else {
    // Legacy mode: create single interface from mac field
    interfaces = []interface{}{legacyInterface}
}
entity["interfaces"] = interfaces
```

5. **Update parseDeviceFromAPI** to extract interfaces:

```go
// In parseDeviceFromAPI, add:
if ifaceData, ok := data["interfaces"].([]interface{}); ok {
    model.Interfaces = make([]DeviceInterfaceModel, 0, len(ifaceData))
    for _, iface := range ifaceData {
        if ifaceMap, ok := iface.(map[string]interface{}); ok {
            model.Interfaces = append(model.Interfaces, parseInterfaceFromAPI(ifaceMap))
        }
    }
}
```

Run tests again:

```bash
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceDevice_InterfaceSingle
```

### Phase 3: REFACTOR - Improve Code Quality

1. Extract interface helpers to `resource_cmdevice_device_interfaces.go`
2. Add comprehensive validators
3. Improve error messages
4. Run linter and formatter:

```bash
make fmt
make lint
```

---

## Testing Commands

### Run All Interface Tests

```bash
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "TestAccCMDeviceDevice_Interface"
```

### Run Specific Test Categories

```bash
# Create tests
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "Interface.*Single|Interface.*Multiple|Interface.*Bond|Interface.*BMC"

# Update tests
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "Interface.*Update|Interface.*Add|Interface.*Remove"

# Import tests
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "Interface.*Import"

# Drift tests
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "Interface.*Drift"
```

### Verify Idempotency

```bash
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "Interface.*Idempotency"
```

---

## Common Patterns

### Type Mapping

```go
func interfaceTypeToBCMChildType(tfType string) string {
    switch tfType {
    case "physical":
        return "NetworkPhysicalInterface"
    case "bond":
        return "NetworkBondInterface"
    case "bmc":
        return "NetworkBMCInterface"
    default:
        return "NetworkPhysicalInterface"
    }
}
```

### UUID Generation

```go
import "github.com/google/uuid"

// For new interfaces
interfaceUUID := uuid.New().String()

// For existing interfaces (update)
existingUUID := state.Interfaces[i].UUID.ValueString()
```

### Null-Safe Field Access

```go
// Use existing helper functions
model.Name = getStringValue(data, "name")
model.DHCP = getBoolValue(data, "dhcp")
model.CardType = getStringValue(data, "cardtype")
```

### Bond Member Extraction

```go
if members, ok := data["members"].([]interface{}); ok {
    memberStrings := make([]string, len(members))
    for i, m := range members {
        memberStrings[i] = m.(string)
    }
    memberList, _ := types.ListValueFrom(ctx, types.StringType, memberStrings)
    model.Members = memberList
}
```

---

## Debugging Tips

### Enable Terraform Logging

```bash
export TF_LOG=DEBUG
export TF_LOG_PATH=./terraform.log
```

### BCM API Request Logging

Add to BCMClient:

```go
tflog.Debug(ctx, "BCM API Request", map[string]interface{}{
    "service": service,
    "method":  method,
    "entity":  entity,
})
```

### Test Helper for BCM Client

```go
client := createTestBCMClient(t)
ctx := context.Background()

// Get device directly
body, err := client.CallJSONRPC(ctx, "cmdevice", "getDevice", deviceName)
```

---

## Documentation Generation

After implementation, regenerate docs:

```bash
make generate

# Verify generated docs
cat docs/resources/bcm_cmdevice_device.md | grep -A 50 "interfaces"
```

---

## Checklist

- [ ] DeviceInterfaceModel struct defined
- [ ] Schema includes interfaces block with all attributes
- [ ] buildInterfaceAPIEntity implemented
- [ ] parseInterfaceFromAPI implemented
- [ ] Type mapping functions implemented
- [ ] Create flow handles interfaces
- [ ] Read flow populates interfaces
- [ ] Update flow preserves interface UUIDs
- [ ] Import populates interfaces
- [ ] Drift detection works for interfaces
- [ ] Validation: unique names
- [ ] Validation: bond requires members
- [ ] Backward compatibility: legacy mac field works
- [ ] Tests pass: Create, Update, Import, Drift
- [ ] Documentation generated
- [ ] Examples created

---

## References

- **Spec**: `/workspace/specs/038-device-interfaces-block/spec.md`
- **Data Model**: `/workspace/specs/038-device-interfaces-block/data-model.md`
- **Research**: `/workspace/specs/038-device-interfaces-block/research.md`
- **Contracts**: `/workspace/specs/038-device-interfaces-block/contracts/interfaces.json`
- **Existing Resource**: `/workspace/internal/provider/resource_cmdevice_device.go`
- **Interface Model Reference**: `/workspace/internal/provider/data_source_cmdevice_nodes.go`
