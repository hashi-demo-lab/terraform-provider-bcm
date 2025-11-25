# Implementation Plan: bcm_cmkube_cluster etcd_nodes Enhancement

**Branch**: `025-cmkube-etcd-nodes` | **Date**: 2025-11-25 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/025-cmkube-etcd-nodes/spec.md`
**Issue**: [#25](https://github.com/hashi-demo-lab/terraform-provider-bcm/issues/25)

## Summary

Add `etcd_nodes` attribute to `bcm_cmkube_cluster` resource for NVIDIA DGX BasePOD K8s deployments requiring explicit etcd node designation. Validate all existing optional attributes against BCM API and document their support status.

**Technical Approach**:
- Add `etcd_nodes` as optional list attribute following the same write-only pattern as `master_nodes`/`worker_nodes`
- Update schema, model, buildClusterEntity, and preserve etcd_nodes in Create/Read/Update operations
- Add acceptance tests using modern terraform-plugin-testing patterns
- Update documentation and examples

## Technical Context

**Language/Version**: Go 1.24
**Primary Dependencies**: terraform-plugin-framework v1.16.1, terraform-plugin-testing v1.13.3
**Storage**: N/A (API-based)
**Testing**: go test with TF_ACC=1 for acceptance tests
**Target Platform**: Cross-platform (Terraform Provider)
**Project Type**: Terraform Provider (single project)
**Performance Goals**: N/A (API passthrough)
**Constraints**: BCM API compatibility, write-only field pattern for node lists
**Scale/Scope**: Single attribute addition with test coverage

## Constitution Check

*GATE: Pass*

- [x] Single project structure
- [x] TDD approach (tests first)
- [x] No over-engineering
- [x] Minimal changes to existing code
- [x] Follows existing patterns

## Project Structure

### Documentation (this feature)

```text
specs/025-cmkube-etcd-nodes/
├── spec.md              # Feature specification (Phase 1 complete)
├── plan.md              # This file
├── research.md          # API research findings
└── tasks.md             # Task breakdown (Phase 3)
```

### Source Code (existing repository)

```text
internal/provider/
├── resource_cmkube_cluster.go       # MODIFY: Add etcd_nodes attribute
├── resource_cmkube_cluster_test.go  # MODIFY: Add etcd_nodes tests

examples/resources/bcm_cmkube_cluster/
└── resource.tf                       # MODIFY: Add etcd_nodes example

docs/resources/
└── cmkube_cluster.md                 # AUTO-GENERATED: make generate
```

**Structure Decision**: Minimal modifications to existing files. No new files required except documentation artifacts.

## API Contract Summary

Based on Phase 0 research (`sampleRest/cmkube-etcd-test.py`):

### etcdNodes Field

- **BCM API Field**: `etcdNodes` (camelCase)
- **Terraform Attribute**: `etcd_nodes` (snake_case)
- **Type**: List of node UUID strings
- **API Behavior**:
  - ACCEPTED by addKubeCluster/updateKubeCluster
  - NOT RETURNED by getKubeCluster (write-only)
- **Pattern**: Same as masterNodes/workerNodes

### Implementation Pattern

Follow existing pattern for write-only node list fields:

```go
// Schema
"etcd_nodes": schema.ListAttribute{
    ElementType:         types.StringType,
    Optional:            true,
    MarkdownDescription: "List of node UUIDs designated as etcd cluster members. NVIDIA recommends 3 nodes for production HA. If not specified, etcd runs on master nodes.",
}

// Model
EtcdNodes types.List `tfsdk:"etcd_nodes"` // Optional, list of UUIDs

// buildClusterEntity
if !model.EtcdNodes.IsNull() && !model.EtcdNodes.IsUnknown() {
    var etcdNodes []string
    diags.Append(model.EtcdNodes.ElementsAs(ctx, &etcdNodes, false)...)
    entity["etcdNodes"] = etcdNodes
}

// Create/Update - Preserve from plan (same as master_nodes/worker_nodes)
// Read - Preserve from state (BCM doesn't return this field)
// ImportState - Set to null (add to ImportStateVerifyIgnore)
```

## Design Decisions

### D1: etcd_nodes as Optional List

**Decision**: Make `etcd_nodes` optional with no default value.

**Rationale**:
- BCM defaults to running etcd on master nodes when not specified
- Users who don't need explicit etcd designation don't need to configure it
- Matches master_nodes/worker_nodes pattern

### D2: Write-Only Field Handling

**Decision**: Preserve `etcd_nodes` from plan/state during Read operations.

**Rationale**:
- BCM API doesn't return etcdNodes in getKubeCluster response
- Same pattern used for master_nodes and worker_nodes
- Prevents "inconsistent result after apply" errors

### D3: Import Behavior

**Decision**: Add `etcd_nodes` to `ImportStateVerifyIgnore` list.

**Rationale**:
- Cannot populate from BCM API during import
- Same treatment as master_nodes/worker_nodes
- Users can manually set after import if needed

### D4: Schema Description Updates

**Decision**: Update schema descriptions to document BCM API support status.

**Rationale**:
- Users should know which fields are write-only
- Prevents confusion about why values aren't in state after import
- Follows documentation best practices

## Test Strategy

### New Test: TestAccCMKubeClusterResource_EtcdNodes

```go
// Test etcd_nodes attribute
Steps: []resource.TestStep{
    // Create with etcd_nodes
    {
        Config: testAccCMKubeClusterResourceConfigWithEtcdNodes(name, masterUUID, []string{etcdUUID1}),
        ConfigStateChecks: []statecheck.StateCheck{
            statecheck.ExpectKnownValue("bcm_cmkube_cluster.test",
                tfjsonpath.New("etcd_nodes"), knownvalue.ListSizeExact(1)),
        },
    },
    // Idempotency check
    {
        Config: testAccCMKubeClusterResourceConfigWithEtcdNodes(name, masterUUID, []string{etcdUUID1}),
        ConfigPlanChecks: resource.ConfigPlanChecks{
            PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
        },
    },
    // Update etcd_nodes
    {
        Config: testAccCMKubeClusterResourceConfigWithEtcdNodes(name, masterUUID, []string{etcdUUID1, etcdUUID2}),
        ConfigStateChecks: []statecheck.StateCheck{
            statecheck.ExpectKnownValue("bcm_cmkube_cluster.test",
                tfjsonpath.New("etcd_nodes"), knownvalue.ListSizeExact(2)),
        },
    },
}
```

### Test Helper

```go
func getTestEtcdNodeUUID(t *testing.T, index int) string {
    // Similar to getTestWorkerNodeUUID but for etcd designation
}

func testAccCMKubeClusterResourceConfigWithEtcdNodes(name, masterUUID string, etcdUUIDs []string) string {
    // Generate config with etcd_nodes attribute
}
```

### Existing Test Modifications

- Add `etcd_nodes` to `ImportStateVerifyIgnore` in import tests
- Verify existing tests still pass without etcd_nodes

## Implementation Phases

### Phase 1: Schema and Model (RED)

1. Add `EtcdNodes` field to `CMKubeClusterResourceModel`
2. Add `etcd_nodes` attribute to schema
3. Add test configuration helper
4. Write failing acceptance test

### Phase 2: CRUD Operations (GREEN)

1. Add etcdNodes to `buildClusterEntity`
2. Handle etcdNodes in readCluster (preserve from state)
3. Preserve etcdNodes in Create/Update operations
4. Add to ImportStateVerifyIgnore

### Phase 3: Documentation and Validation (REFACTOR)

1. Update example configuration
2. Run `make generate` for documentation
3. Run full test suite
4. Verify linting passes

## Risk Assessment

### Low Risk

- Simple attribute addition following existing pattern
- No changes to API client
- No new dependencies

### Medium Risk

- Test environment needs multiple nodes for etcd designation testing
- May need to skip test if insufficient nodes available

### Mitigation

- Use dynamic node discovery in tests
- Add skip logic if fewer than 2 nodes available
- Test with single etcd node if 3 not available

## Dependencies

- BCM test environment with 2+ available nodes
- Existing test helpers (createTestBCMClient, getTestMasterNodeUUID)
- terraform-plugin-testing v1.13.3+

## Success Metrics

- [ ] `etcd_nodes` attribute added to schema
- [ ] Acceptance test passes: `TestAccCMKubeClusterResource_EtcdNodes`
- [ ] All existing CMKubeCluster tests pass
- [ ] Documentation regenerated with `make generate`
- [ ] Example updated with etcd_nodes usage
- [ ] Linting passes: `make lint`
