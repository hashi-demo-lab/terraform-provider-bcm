# Developer Quickstart: Fix Device Role Association Bug

**Feature**: 086-fix-device-role-association
**Date**: 2025-11-26

## Overview

This guide helps you implement the device role association fix. The goal is to allow users to specify roles by name (e.g., `roles = ["backup"]`) instead of requiring UUID lookups.

---

## Prerequisites

- Go 1.24.0 installed
- BCM cluster access (172.21.15.254:8081)
- Environment variables set:
  ```bash
  export TF_ACC=1
  export BCM_ENDPOINT="https://172.21.15.254:8081"
  export BCM_USERNAME="root"
  export BCM_PASSWORD="Hashicorp123!"
  ```

---

## Files to Modify

| File | Changes |
|------|---------|
| `internal/provider/resource_cmdevice_device.go` | Modify `lookupAndBuildRolesForEntity()`, `parseRolesFromAPI()`, add `isUUID()`, update schema docs |
| `internal/provider/resource_cmdevice_device_test.go` | Add role-by-name tests |
| `examples/resources/bcm_cmdevice_device/with_roles.tf` | Simplify to use role names |

---

## Step 1: Add isUUID Helper Function

Add this helper near the top of `resource_cmdevice_device.go` (after imports):

```go
import "regexp"

// uuidRegex matches standard UUID format (8-4-4-4-12)
var uuidRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// isUUID returns true if the string matches UUID format
func isUUID(s string) bool {
    return uuidRegex.MatchString(s)
}
```

---

## Step 2: Modify lookupAndBuildRolesForEntity()

Replace the existing function (around line 1339) with this implementation:

```go
// lookupAndBuildRolesForEntity looks up role objects by name or UUID and adds them to the device entity.
// BCM requires full role objects (not just names or UUIDs) when assigning roles to devices.
// This function queries all nodes to extract available role objects, then matches them by name or UUID.
// Role names are the preferred input format; UUIDs are supported for backward compatibility.
func (r *CMDeviceDeviceResource) lookupAndBuildRolesForEntity(ctx context.Context, plan CMDeviceDeviceResourceModel, entity map[string]interface{}) error {
    if plan.Roles.IsNull() || plan.Roles.IsUnknown() {
        return nil
    }

    var roleIdentifiers []string
    diags := plan.Roles.ElementsAs(ctx, &roleIdentifiers, false)
    if diags.HasError() {
        return fmt.Errorf("failed to extract role identifiers from plan")
    }

    // If empty list, set empty roles (explicit removal)
    if len(roleIdentifiers) == 0 {
        entity["roles"] = []interface{}{}
        return nil
    }

    // Deduplicate role identifiers and validate non-empty
    roleSet := make(map[string]struct{})
    var invalidIdentifiers []string
    for _, id := range roleIdentifiers {
        if id == "" {
            invalidIdentifiers = append(invalidIdentifiers, "(empty string)")
        } else {
            roleSet[id] = struct{}{}
        }
    }

    // Return error if any empty identifiers were found
    if len(invalidIdentifiers) > 0 {
        return fmt.Errorf("invalid role identifiers found: %v - role identifiers must be non-empty strings", invalidIdentifiers)
    }

    // Query all nodes to extract available role objects
    body, err := r.client.CallJSONRPC(ctx, "cmdevice", "getNodes")
    if err != nil {
        return fmt.Errorf("failed to query nodes for role lookup: %w", err)
    }

    var nodes []map[string]interface{}
    if err := json.Unmarshal(body, &nodes); err != nil {
        return fmt.Errorf("failed to parse nodes response: %w", err)
    }

    // Build maps of all available role objects by name and UUID
    rolesByName := make(map[string]map[string]interface{})
    rolesByUUID := make(map[string]map[string]interface{})
    for _, node := range nodes {
        if rolesData, ok := node["roles"].([]interface{}); ok {
            for _, roleData := range rolesData {
                if role, ok := roleData.(map[string]interface{}); ok {
                    if uuid, ok := role["uuid"].(string); ok && uuid != "" {
                        rolesByUUID[uuid] = role
                    }
                    if name, ok := role["name"].(string); ok && name != "" {
                        rolesByName[name] = role
                    }
                }
            }
        }
    }

    // Match requested role identifiers to full role objects
    roleObjects := make([]interface{}, 0, len(roleSet))
    var missingRoles []string

    for identifier := range roleSet {
        var role map[string]interface{}
        var found bool

        // Determine lookup method based on identifier format
        if isUUID(identifier) {
            // Lookup by UUID (backward compatibility)
            role, found = rolesByUUID[identifier]
        } else {
            // Lookup by name (preferred)
            role, found = rolesByName[identifier]
        }

        if found {
            // Create a copy of the role object for this assignment
            roleCopy := make(map[string]interface{})
            for k, v := range role {
                roleCopy[k] = v
            }
            roleObjects = append(roleObjects, roleCopy)
        } else {
            missingRoles = append(missingRoles, identifier)
        }
    }

    if len(missingRoles) > 0 {
        // Build list of available role names for helpful error message
        availableNames := make([]string, 0, len(rolesByName))
        for name := range rolesByName {
            availableNames = append(availableNames, name)
        }
        sort.Strings(availableNames)
        sort.Strings(missingRoles)

        return fmt.Errorf(
            "Roles not found in cluster: %s\nAvailable roles: %s\nUse the `bcm_cmdevice_roles` data source to discover available roles.",
            strings.Join(missingRoles, ", "),
            strings.Join(availableNames, ", "),
        )
    }

    // Sort role objects by UUID for consistent ordering
    sort.Slice(roleObjects, func(i, j int) bool {
        uuidI, _ := roleObjects[i].(map[string]interface{})["uuid"].(string)
        uuidJ, _ := roleObjects[j].(map[string]interface{})["uuid"].(string)
        return uuidI < uuidJ
    })

    entity["roles"] = roleObjects
    return nil
}
```

**Note**: Add `"strings"` to imports if not already present.

---

## Step 3: Modify parseRolesFromAPI()

Replace the existing function (around line 1557) to return role names instead of UUIDs:

```go
// parseRolesFromAPI parses BCM API roles response into a Terraform set of role names.
// BCM returns roles as an array of role objects: [{"uuid": "...", "name": "backup", ...}]
// We extract names because users configure roles by name (the recommended approach).
// Returns a Set since order of roles is not significant.
func parseRolesFromAPI(rolesData interface{}) types.Set {
    if rolesData == nil {
        return types.SetNull(types.StringType)
    }

    rolesArray, ok := rolesData.([]interface{})
    if !ok || len(rolesArray) == 0 {
        return types.SetNull(types.StringType)
    }

    // Extract role names from the array
    roleNames := make([]string, 0, len(rolesArray))
    for _, roleItem := range rolesArray {
        if role, ok := roleItem.(map[string]interface{}); ok {
            // Role object with "name" field
            if name, ok := role["name"].(string); ok && name != "" {
                roleNames = append(roleNames, name)
            }
        }
    }

    if len(roleNames) == 0 {
        return types.SetNull(types.StringType)
    }

    // Convert to Terraform set (order doesn't matter for sets)
    roleValues := make([]attr.Value, len(roleNames))
    for i, name := range roleNames {
        roleValues[i] = types.StringValue(name)
    }

    rolesSet, _ := types.SetValue(types.StringType, roleValues)
    return rolesSet
}
```

---

## Step 4: Update Schema Documentation

Find the `"roles"` attribute in the Schema() method and update the MarkdownDescription:

```go
"roles": schema.SetAttribute{
    Optional:    true,
    Computed:    true,
    ElementType: types.StringType,
    MarkdownDescription: "Set of role names assigned to this device. Roles define the device's function " +
        "in the cluster (e.g., \"backup\", \"provisioning\", \"boot\"). Use the `bcm_cmdevice_roles` " +
        "data source to discover available roles. Role names are case-sensitive. For backward " +
        "compatibility, role UUIDs are also accepted but role names are recommended for readability.\n\n" +
        "Example usage:\n\n" +
        "```hcl\n" +
        "resource \"bcm_cmdevice_device\" \"node\" {\n" +
        "  # ... other configuration ...\n" +
        "  roles = [\"backup\", \"provisioning\"]\n" +
        "}\n" +
        "```",
},
```

---

## Step 5: Update Example File

Replace `/workspace/examples/resources/bcm_cmdevice_device/with_roles.tf`:

```hcl
# Example: Device with role assignments
#
# This example demonstrates how to create a device with specific roles assigned.
# Roles define the device's function in the cluster (backup, provisioning, etc.).
# Simply specify role names directly - no UUID lookups needed!

# Generate unique suffix for this test run
locals {
  test_suffix = formatdate("YYYYMMDDhhmmss", timestamp())
}

# Lookup management network
data "bcm_cmnet_networks" "management" {
  filter {
    name_pattern = "managementnet"
  }
}

# Create a unique software image for each test run
resource "bcm_cmpart_softwareimage" "test_image" {
  name = "citest-image-${local.test_suffix}"
  path = "/cm/images/ubuntu-22.04-server-amd64.iso"
}

# Create a unique category for each test run
resource "bcm_cmdevice_category" "roles_category" {
  name               = "citest-category-${local.test_suffix}"
  management_network = data.bcm_cmnet_networks.management.networks[0].id

  software_image_proxy = {
    parent_software_image = bcm_cmpart_softwareimage.test_image.id
  }

  notes = "Test category for device with roles (run: ${local.test_suffix})"

  depends_on = [bcm_cmpart_softwareimage.test_image]
}

# Create device with roles assigned BY NAME (not UUID!)
resource "bcm_cmdevice_device" "with_roles" {
  hostname           = "citest-roles-${local.test_suffix}"
  mac                = "00:11:22:33:44:BB"
  category           = bcm_cmdevice_category.roles_category.id
  management_network = data.bcm_cmnet_networks.management.networks[0].id

  # Simply specify role names - the provider resolves them automatically
  roles = ["backup", "provisioning"]

  notes = "Device with backup and provisioning roles (run: ${local.test_suffix})"

  depends_on = [bcm_cmdevice_category.roles_category]
}

# Outputs
output "device_id" {
  value       = bcm_cmdevice_device.with_roles.id
  description = "UUID of the created device"
}

output "device_hostname" {
  value       = bcm_cmdevice_device.with_roles.hostname
  description = "Hostname of the created device"
}

output "device_roles" {
  value       = bcm_cmdevice_device.with_roles.roles
  description = "Role names assigned to the device"
}
```

---

## Step 6: Write Acceptance Tests (TDD - RED Phase)

Add these tests to `internal/provider/resource_cmdevice_device_test.go`:

```go
func TestAccCMDeviceDevice_RolesByName(t *testing.T) {
    hostname := generateUniqueTestName("test-role-name")

    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
        Steps: []resource.TestStep{
            {
                Config: testAccCMDeviceDeviceConfigWithRoles(hostname, []string{"backup", "provisioning"}),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("bcm_cmdevice_device.test", "hostname", hostname),
                    resource.TestCheckResourceAttr("bcm_cmdevice_device.test", "roles.#", "2"),
                ),
            },
        },
    })
}

func TestAccCMDeviceDevice_InvalidRoleName(t *testing.T) {
    hostname := generateUniqueTestName("test-invalid-role")

    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            {
                Config:      testAccCMDeviceDeviceConfigWithRoles(hostname, []string{"nonexistent-role"}),
                ExpectError: regexp.MustCompile(`Roles not found in cluster: nonexistent-role`),
            },
        },
    })
}

func testAccCMDeviceDeviceConfigWithRoles(hostname string, roles []string) string {
    rolesHCL := "["
    for i, r := range roles {
        if i > 0 {
            rolesHCL += ", "
        }
        rolesHCL += fmt.Sprintf(`"%s"`, r)
    }
    rolesHCL += "]"

    return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

# ... (category and network setup - reuse existing test helpers)

resource "bcm_cmdevice_device" "test" {
  hostname           = %[4]q
  mac                = "00:11:22:33:44:FF"
  category           = "..." # Use test category
  management_network = "..." # Use test network
  roles              = %[5]s
}
`,
        os.Getenv("BCM_ENDPOINT"),
        os.Getenv("BCM_USERNAME"),
        os.Getenv("BCM_PASSWORD"),
        hostname,
        rolesHCL,
    )
}
```

---

## Step 7: Run Tests

```bash
# Run the new tests (should pass after implementation)
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "TestAccCMDeviceDevice_RolesByName"
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "TestAccCMDeviceDevice_InvalidRoleName"

# Run lint
make lint

# Generate documentation
make generate
```

---

## Verification Checklist

- [ ] `isUUID()` helper function added
- [ ] `lookupAndBuildRolesForEntity()` supports name and UUID lookup
- [ ] `parseRolesFromAPI()` returns role names (not UUIDs)
- [ ] Schema documentation updated
- [ ] Example file simplified (no UUID lookups)
- [ ] `TestAccCMDeviceDevice_RolesByName` passes
- [ ] `TestAccCMDeviceDevice_InvalidRoleName` passes
- [ ] `make lint` passes
- [ ] `make generate` succeeds
- [ ] `pre-commit run --all-files` passes

---

## Common Issues

### "regexp" import not found

Add `"regexp"` to the import block at the top of the file.

### "strings" import not found

Add `"strings"` to the import block for `strings.Join()`.

### Tests fail with "role not found"

Ensure the BCM cluster has the roles you're testing with (e.g., "backup", "provisioning"). Use the `bcm_cmdevice_roles` data source to verify available roles.

### State shows UUIDs after upgrade

This is expected for existing configurations. Run `terraform apply` once to migrate state to role names.
