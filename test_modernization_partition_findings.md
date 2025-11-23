# Partition Test Modernization Findings

## Summary

The partition resource tests (`resource_cmpart_partition_test.go`) cannot follow the standard CREATE → UPDATE → DELETE test pattern due to BCM architectural constraints.

## BCM Partition Constraints

1. **Single Partition Only**: BCM requires exactly one partition named "base"
2. **Partition Exists**: The "base" partition already exists in the test environment
   - UUID: `ddd19eb5-f04a-48dc-9cc6-f160b704a7dd`
   - Name: `base`
   - ClusterName: `BCM 11.0 Cluster`
3. **Cannot Delete**: The base partition cannot be deleted (system partition)
4. **Cannot Create Duplicate**: Attempting to create a partition named "base" fails with: `A partition with that name already exists`

## Test Errors Encountered

### TestAccCMPartPartition_Basic
```
Error: Error Creating Partition

Could not create partition 'base': validation errors: [
  name: A partition with that name already exists
  uuid: Zero UUID Partition:base
  Node basename: The node basename can only contain a-z, A-Z, 0-9 and dashes (-).
    It can not start or end with a dash, or contain only numbers.
  timeZone: is not set to a predefined time zone
  timeZone: Illegal value for: timeZone
]
```

## Configuration Requirements

### Required Fields
All partition configurations must include:
- `name` - Must be "base"
- `cluster_name` - Any string
- `timezone_settings` - IANA timezone (e.g., "America/New_York", NOT "UTC")
- `primary_head_node` - UUID of head node
- `external_network` - UUID of external network
- `default_category` - UUID of default category
- `management_network` - UUID of management network

### Invalid Timezone Values
❌ `"UTC"` - Not recognized by BCM
✅ `"America/New_York"` - Valid IANA timezone
✅ `"America/Los_Angeles"` - Valid IANA timezone
✅ `"Europe/London"` - Valid IANA timezone

## Recommended Test Strategy

### Import-Update-Restore Pattern

```go
func TestAccCMPartPartition_Basic(t *testing.T) {
    // Get existing partition UUID
    uuid := getResourceUUIDByName(t, "cmpart", "getPartitions", "base")

    // Store original state for cleanup
    originalState := getPartitionState(t, uuid)
    defer restorePartitionState(t, uuid, originalState)

    resource.Test(t, resource.TestCase{
        PreCheck: func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        // Skip CheckDestroy - partition cannot be deleted
        Steps: []resource.TestStep{
            // Step 1: Import existing partition
            {
                Config: testAccPartitionConfig("Test Cluster"),
                ResourceName: "bcm_cmpart_partition.test",
                ImportState: true,
                ImportStateId: uuid,
                ImportStateVerify: true,
            },
            // Step 2: Update cluster_name
            {
                Config: testAccPartitionConfig("Updated Cluster"),
                Check: resource.TestCheckResourceAttr(
                    "bcm_cmpart_partition.test",
                    "cluster_name",
                    "Updated Cluster",
                ),
            },
            // Step 3: Idempotency check
            {
                Config: testAccPartitionConfig("Updated Cluster"),
                PlanOnly: true,
            },
        },
    })
}
```

## Files Created

1. **resource_cmpart_partition_test_minimal.go** - Working minimal test implementation
   - TestAccCMPartPartition_BasicMinimal - Import and update test
   - TestAccCMPartPartition_UpdateMinimal - Update with state restoration

## Next Steps

1. **Update TestAccCMPartPartition_Basic**:
   - Change from CREATE to IMPORT as first step
   - Remove CheckDestroy
   - Add state restoration in cleanup

2. **Update TestAccCMPartPartition_Update**:
   - Start with import
   - Test updates to various fields
   - Restore original state after test

3. **Skip or Remove Tests**:
   - TestAccCMPartPartition_NetworkSettings - May need import-based approach
   - TestAccCMPartPartition_DriftDetection - Can work with import approach
   - TestAccCMPartPartition_SlaveNaming - Can work with import approach
   - TestAccCMPartPartition_IDConsistency - Can work with import approach
   - TestAccCMPartPartition_ValidationErrors - Can test validation without BCM API

## Test Config Template

```hcl
provider "bcm" {
  endpoint             = "<endpoint>"
  username             = "<username>"
  password             = "<password>"
  insecure_skip_verify = true
}

# Lookup required resources
data "bcm_cmdevice_nodes" "headnode" {
  filter {
    child_type = "HeadNode"
  }
}

data "bcm_cmnet_networks" "external" {
  filter {
    name_pattern = "external"
  }
}

data "bcm_cmnet_networks" "mgmt" {
  filter {
    name_pattern = "provisioning"
  }
}

data "bcm_cmdevice_categories" "default" {
  name = "default"
}

resource "bcm_cmpart_partition" "test" {
  name               = "base"
  cluster_name       = var.cluster_name
  timezone_settings  = "America/New_York"
  primary_head_node  = data.bcm_cmdevice_nodes.headnode.nodes[0].uuid
  external_network   = data.bcm_cmnet_networks.external.networks[0].uuid
  default_category   = data.bcm_cmdevice_categories.default.categories[0].uuid
  management_network = data.bcm_cmnet_networks.mgmt.networks[0].uuid
}
```

## Validation

All other resource tests (software image, network, cluster, category) follow standard CREATE → UPDATE → DELETE pattern and work correctly. The partition resource is unique in BCM's architecture as a singleton system resource.
