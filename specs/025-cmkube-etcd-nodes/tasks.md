# Tasks: bcm_cmkube_cluster etcd_nodes Enhancement

**Feature**: Add etcd_nodes attribute to bcm_cmkube_cluster
**Issue**: [#25](https://github.com/hashi-demo-lab/terraform-provider-bcm/issues/25)
**Plan**: [plan.md](./plan.md)

## Task Overview

| Phase | Description | Tasks |
|-------|-------------|-------|
| 1 | Schema & Model (RED) | T1-T4 |
| 2 | CRUD Operations (GREEN) | T5-T8 |
| 3 | Documentation (REFACTOR) | T9-T12 |

## Phase 1: Schema & Model (RED)

### T1: Add EtcdNodes field to model

**File**: `internal/provider/resource_cmkube_cluster.go`
**Status**: [ ] Pending

Add `EtcdNodes` field to `CMKubeClusterResourceModel` struct:

```go
// Node configuration
MasterNodes types.List `tfsdk:"master_nodes"` // Required, list of UUIDs
WorkerNodes types.List `tfsdk:"worker_nodes"` // Optional, list of UUIDs
EtcdNodes   types.List `tfsdk:"etcd_nodes"`   // Optional, list of UUIDs for etcd cluster members
```

**Location**: After line 48 (WorkerNodes field)

---

### T2: Add etcd_nodes schema attribute

**File**: `internal/provider/resource_cmkube_cluster.go`
**Status**: [ ] Pending

Add schema attribute in `Schema()` function after `worker_nodes`:

```go
"etcd_nodes": schema.ListAttribute{
    ElementType:         types.StringType,
    Optional:            true,
    MarkdownDescription: "List of node UUIDs designated as etcd cluster members. NVIDIA recommends 3 nodes for production high availability. If not specified, etcd runs on master nodes.",
    PlanModifiers: []planmodifier.List{
        listplanmodifier.UseStateForUnknown(),
    },
},
```

**Location**: After `worker_nodes` attribute (around line 146)

---

### T3: Add test configuration helper

**File**: `internal/provider/resource_cmkube_cluster_test.go`
**Status**: [ ] Pending

Add helper function for etcd_nodes test configuration:

```go
func getTestEtcdNodeUUID(t *testing.T, index int) string {
    client := createTestBCMClient(t)
    ctx := context.Background()

    body, err := client.CallJSONRPC(ctx, "cmdevice", "getNodes")
    if err != nil {
        t.Fatalf("Failed to get nodes: %v", err)
    }

    var nodes []map[string]interface{}
    if err := json.Unmarshal(body, &nodes); err != nil {
        t.Fatalf("Failed to parse nodes: %v", err)
    }

    // Use different index from master (0) - etcd can overlap with workers
    nodeIndex := index + 1
    if len(nodes) <= nodeIndex {
        t.Skipf("Not enough nodes for etcd index %d (need at least %d nodes)", index, nodeIndex+1)
    }

    if uuid, ok := nodes[nodeIndex]["uuid"].(string); ok {
        return uuid
    }

    t.Fatalf("Node at index %d has invalid UUID format", nodeIndex)
    return ""
}

func testAccCMKubeClusterResourceConfigWithEtcdNodes(name, masterNodeUUID string, etcdNodeUUIDs []string) string {
    var etcdStr string
    if len(etcdNodeUUIDs) > 0 {
        etcdNodes := make([]string, len(etcdNodeUUIDs))
        for i, uuid := range etcdNodeUUIDs {
            etcdNodes[i] = fmt.Sprintf("%q", uuid)
        }
        etcdStr = fmt.Sprintf("\n  etcd_nodes = [%s]", strings.Join(etcdNodes, ", "))
    }

    return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

resource "bcm_cmkube_cluster" "test" {
  name         = %[4]q
  master_nodes = [%[5]q]%[6]s
}
`,
        os.Getenv("BCM_ENDPOINT"),
        os.Getenv("BCM_USERNAME"),
        os.Getenv("BCM_PASSWORD"),
        name,
        masterNodeUUID,
        etcdStr,
    )
}
```

---

### T4: Write failing acceptance test

**File**: `internal/provider/resource_cmkube_cluster_test.go`
**Status**: [ ] Pending

Add acceptance test for etcd_nodes:

```go
// TestAccCMKubeClusterResource_EtcdNodes tests etcd node designation
func TestAccCMKubeClusterResource_EtcdNodes(t *testing.T) {
    clusterName := generateUniqueTestName("tftest-cluster-etcd")
    masterNodeUUID := getTestMasterNodeUUID(t)
    etcdNodeUUID := getTestEtcdNodeUUID(t, 0)

    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheckCMKubeCluster(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        CheckDestroy:             testAccCheckCMKubeClusterDestroy,
        Steps: []resource.TestStep{
            // Create with etcd_nodes
            {
                Config: testAccCMKubeClusterResourceConfigWithEtcdNodes(
                    clusterName,
                    masterNodeUUID,
                    []string{etcdNodeUUID},
                ),
                ConfigStateChecks: []statecheck.StateCheck{
                    statecheck.ExpectKnownValue(
                        "bcm_cmkube_cluster.test",
                        tfjsonpath.New("name"),
                        knownvalue.StringExact(clusterName),
                    ),
                    statecheck.ExpectKnownValue(
                        "bcm_cmkube_cluster.test",
                        tfjsonpath.New("etcd_nodes"),
                        knownvalue.ListSizeExact(1),
                    ),
                },
            },
            // Idempotency check
            {
                Config: testAccCMKubeClusterResourceConfigWithEtcdNodes(
                    clusterName,
                    masterNodeUUID,
                    []string{etcdNodeUUID},
                ),
                ConfigPlanChecks: resource.ConfigPlanChecks{
                    PreApply: []plancheck.PlanCheck{
                        plancheck.ExpectEmptyPlan(),
                    },
                },
            },
        },
    })
}
```

**Expected Result**: Test fails because etcd_nodes attribute doesn't exist yet.

---

## Phase 2: CRUD Operations (GREEN)

### T5: Add etcdNodes to buildClusterEntity

**File**: `internal/provider/resource_cmkube_cluster.go`
**Status**: [ ] Pending

Add etcdNodes handling in `buildClusterEntity()` function:

```go
// Etcd nodes (optional)
if !model.EtcdNodes.IsNull() && !model.EtcdNodes.IsUnknown() {
    var etcdNodes []string
    diags.Append(model.EtcdNodes.ElementsAs(ctx, &etcdNodes, false)...)
    entity["etcdNodes"] = etcdNodes
}
```

**Location**: After workerNodes handling (around line 780)

---

### T6: Handle etcdNodes in readCluster

**File**: `internal/provider/resource_cmkube_cluster.go`
**Status**: [ ] Pending

Handle etcdNodes in `readCluster()` - preserve from existing state since BCM doesn't return it:

```go
// Etcd nodes (write-only field, preserve from state)
// NOTE: BCM cmkube API behavior for node lists:
// - etcdNodes is write-only (used during create/update but not returned in read)
// - We preserve the state value to maintain consistency
// - ImportState ignores this field (see ImportStateVerifyIgnore)
if model.EtcdNodes.IsNull() || model.EtcdNodes.IsUnknown() {
    model.EtcdNodes = types.ListNull(types.StringType)
}
```

**Location**: After workerNodes handling in readCluster (around line 527)

---

### T7: Preserve etcdNodes in Create/Update

**File**: `internal/provider/resource_cmkube_cluster.go`
**Status**: [ ] Pending

Add preservation of etcd_nodes in Create() and Update() operations.

**Create() - around line 353**:
```go
// Preserve etcd_nodes from plan
planEtcdNodes := plan.EtcdNodes
// ... after readCluster ...
if !planEtcdNodes.IsUnknown() && !planEtcdNodes.IsNull() {
    plan.EtcdNodes = planEtcdNodes
}
```

**Update() - around line 656**:
```go
// Preserve etcd_nodes from plan
planEtcdNodes := plan.EtcdNodes
// ... after readCluster ...
if !planEtcdNodes.IsUnknown() && !planEtcdNodes.IsNull() {
    plan.EtcdNodes = planEtcdNodes
}
```

---

### T8: Update ImportStateVerifyIgnore

**File**: `internal/provider/resource_cmkube_cluster_test.go`
**Status**: [ ] Pending

Update the Import test to include etcd_nodes in ImportStateVerifyIgnore:

```go
// Import
{
    ResourceName:      "bcm_cmkube_cluster.test",
    ImportState:       true,
    ImportStateVerify: true,
    // BCM cmkube API Limitation: getKubeCluster does NOT return master_nodes/worker_nodes/etcd_nodes
    // These fields are write-only (used during create/update but not returned in read)
    ImportStateVerifyIgnore: []string{"master_nodes", "worker_nodes", "etcd_nodes"},
},
```

**Location**: Update existing import test steps (around line 217)

---

## Phase 3: Documentation (REFACTOR)

### T9: Update example configuration

**File**: `examples/resources/bcm_cmkube_cluster/resource.tf`
**Status**: [ ] Pending

Update example to show etcd_nodes usage:

```hcl
# Basic cluster with etcd nodes
resource "bcm_cmkube_cluster" "example" {
  name         = "production-cluster"
  master_nodes = [data.bcm_cmdevice_nodes.masters.nodes[0].uuid]
  worker_nodes = [
    data.bcm_cmdevice_nodes.workers.nodes[0].uuid,
    data.bcm_cmdevice_nodes.workers.nodes[1].uuid,
  ]

  # Optional: Explicit etcd node designation for HA
  # NVIDIA recommends 3 etcd nodes for production deployments
  etcd_nodes = [
    data.bcm_cmdevice_nodes.etcd.nodes[0].uuid,
    data.bcm_cmdevice_nodes.etcd.nodes[1].uuid,
    data.bcm_cmdevice_nodes.etcd.nodes[2].uuid,
  ]

  version = "1.28.0"
}
```

---

### T10: Regenerate documentation

**Command**: `make generate`
**Status**: [ ] Pending

Run documentation generation:

```bash
make generate
```

This will update `docs/resources/cmkube_cluster.md` with the new `etcd_nodes` attribute.

---

### T11: Run full test suite

**Command**: `TF_ACC=1 go test -v ./internal/provider/ -run CMKubeCluster`
**Status**: [ ] Pending

Run all CMKubeCluster tests:

```bash
TF_ACC=1 BCM_ENDPOINT="https://172.21.15.254:8081" \
  BCM_USERNAME="root" BCM_PASSWORD="Hashicorp123!" \
  go test -v -timeout 120m ./internal/provider/ -run CMKubeCluster
```

**Expected**: All tests pass.

---

### T12: Run linting

**Command**: `make lint`
**Status**: [ ] Pending

Run linting to ensure code quality:

```bash
make lint
```

**Expected**: No linting errors.

---

## Execution Order

```
Phase 1 (RED):
  T1 → T2 → T3 → T4 (run test - expect failure)

Phase 2 (GREEN):
  T5 → T6 → T7 → T8 (run test - expect pass)

Phase 3 (REFACTOR):
  T9 → T10 → T11 → T12
```

## Verification Checklist

- [ ] T1: EtcdNodes field added to model
- [ ] T2: etcd_nodes schema attribute added
- [ ] T3: Test helper functions added
- [ ] T4: Failing test written
- [ ] T5: buildClusterEntity handles etcdNodes
- [ ] T6: readCluster preserves etcdNodes
- [ ] T7: Create/Update preserve etcdNodes from plan
- [ ] T8: ImportStateVerifyIgnore updated
- [ ] T9: Example configuration updated
- [ ] T10: Documentation regenerated
- [ ] T11: All tests pass
- [ ] T12: Linting passes

## Dependencies

- Test environment: 2+ available nodes
- BCM credentials configured
- terraform-plugin-testing v1.13.3+

## Notes

- etcd_nodes follows same pattern as master_nodes/worker_nodes
- Field is write-only (BCM doesn't return in getKubeCluster)
- ImportState cannot populate this field
- NVIDIA recommends 3 etcd nodes for production HA
