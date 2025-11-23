# Testing Pattern Modernization Guide

**Created**: 2025-11-23
**Status**: Analysis Complete
**Scope**: Migrate legacy testing patterns to modern terraform-plugin-testing v1.13.3+ patterns

## Executive Summary

**Current State**: 13/16 acceptance test files use modern patterns (81% adoption)
**Action Required**: Modernize 3 files with legacy patterns

### Files Requiring Modernization

| Priority | File | Modern | Legacy | Adoption | Impact |
|----------|------|--------|--------|----------|--------|
| 🔴 **P1** | `resource_cmnet_network_test.go` | 1 | 18 | 5% | HIGH - Critical inconsistency |
| 🟡 **P2** | `resource_cmkube_cluster_test.go` | 43 | 38 | 53% | MEDIUM - Mixed patterns |
| 🟢 **P3** | `data_source_cmpart_partitions_test.go` | 27 | 10 | 73% | LOW - Mostly compliant |

## Legacy vs Modern Patterns

### Legacy Pattern (terraform-plugin-sdk)

```go
// ❌ String-based, no type safety
Check: resource.ComposeAggregateTestCheckFunc(
    resource.TestCheckResourceAttr("bcm_cmnet_network.test", "name", networkName),
    resource.TestCheckResourceAttr("bcm_cmnet_network.test", "mtu", "9000"),
    resource.TestCheckResourceAttr("bcm_cmnet_network.test", "dhcp_enabled", "true"),
    resource.TestCheckResourceAttrSet("bcm_cmnet_network.test", "uuid"),
),
```

**Problems**:
- No type safety (everything is a string)
- No validation of attribute types
- Cannot track value consistency across steps
- No idempotency verification
- Poor error messages

### Modern Pattern (terraform-plugin-testing v1.13.3+)

```go
// ✅ Type-safe with knownvalue matchers
ConfigStateChecks: []statecheck.StateCheck{
    statecheck.ExpectKnownValue(
        "bcm_cmnet_network.test",
        tfjsonpath.New("name"),
        knownvalue.StringExact(networkName),
    ),
    statecheck.ExpectKnownValue(
        "bcm_cmnet_network.test",
        tfjsonpath.New("mtu"),
        knownvalue.Int64Exact(9000),
    ),
    statecheck.ExpectKnownValue(
        "bcm_cmnet_network.test",
        tfjsonpath.New("dhcp_enabled"),
        knownvalue.Bool(true),
    ),
    statecheck.ExpectKnownValue(
        "bcm_cmnet_network.test",
        tfjsonpath.New("uuid"),
        knownvalue.NotNull(),
    ),
},
```

**Benefits**:
- ✅ Type-safe verification (String, Int64, Bool, etc.)
- ✅ Better error messages showing expected vs actual types
- ✅ Can track consistency with `CompareValue`
- ✅ Supports plan checks for idempotency
- ✅ Aligns with HashiCorp best practices

## Migration Examples

### Example 1: Basic String Attribute

**Before (Legacy)**:
```go
resource.TestCheckResourceAttr("bcm_cmnet_network.test", "name", networkName)
```

**After (Modern)**:
```go
statecheck.ExpectKnownValue(
    "bcm_cmnet_network.test",
    tfjsonpath.New("name"),
    knownvalue.StringExact(networkName),
)
```

### Example 2: Numeric Attribute

**Before (Legacy)**:
```go
resource.TestCheckResourceAttr("bcm_cmnet_network.test", "mtu", "9000")
```

**After (Modern)**:
```go
statecheck.ExpectKnownValue(
    "bcm_cmnet_network.test",
    tfjsonpath.New("mtu"),
    knownvalue.Int64Exact(9000),
)
```

### Example 3: Boolean Attribute

**Before (Legacy)**:
```go
resource.TestCheckResourceAttr("bcm_cmnet_network.test", "dhcp_enabled", "true")
```

**After (Modern)**:
```go
statecheck.ExpectKnownValue(
    "bcm_cmnet_network.test",
    tfjsonpath.New("dhcp_enabled"),
    knownvalue.Bool(true),
)
```

### Example 4: Existence Check

**Before (Legacy)**:
```go
resource.TestCheckResourceAttrSet("bcm_cmnet_network.test", "uuid")
```

**After (Modern)**:
```go
statecheck.ExpectKnownValue(
    "bcm_cmnet_network.test",
    tfjsonpath.New("uuid"),
    knownvalue.NotNull(),
)
```

### Example 5: List Count (Complex)

**Before (Legacy)**:
```go
resource.TestCheckResourceAttr("bcm_cmkube_cluster.test", "worker_nodes.#", "2")
```

**After (Modern)**:
```go
statecheck.ExpectKnownValue(
    "bcm_cmkube_cluster.test",
    tfjsonpath.New("worker_nodes"),
    knownvalue.ListSizeExact(2),
)
```

## Required Import Additions

Add these imports to files being modernized:

```go
import (
    "github.com/hashicorp/terraform-plugin-testing/helper/resource"
    "github.com/hashicorp/terraform-plugin-testing/plancheck"
    "github.com/hashicorp/terraform-plugin-testing/statecheck"
    "github.com/hashicorp/terraform-plugin-testing/knownvalue"
    "github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)
```

## File-Specific Migration Plans

### Priority 1: resource_cmnet_network_test.go

**Current State**: 5% modern (1/19 checks)
**Estimated Effort**: 2-3 hours
**Impact**: HIGH - Critical for consistency

**Legacy Usages to Replace**:
1. Line 30: `name` attribute (string)
2. Line 31-32: `id`, `uuid` attributes (existence)
3. Line 33: `domain_name` attribute (string)
4. Line 68-77: Complete configuration (8 attributes + 1 existence)
5. Line 96, 103-105: Update test assertions

**Migration Steps**:
1. Add required imports
2. Replace `Check:` with `ConfigStateChecks:` in test steps
3. Convert each `TestCheckResourceAttr` to `ExpectKnownValue`
4. Add type-specific matchers (StringExact, Int64Exact, Bool)
5. Replace `TestCheckResourceAttrSet` with `NotNull()`
6. Add idempotency checks with `plancheck.ExpectEmptyPlan()`
7. Run tests to verify: `TF_ACC=1 go test -v -run TestAccCMNetNetwork ./internal/provider/`

### Priority 2: resource_cmkube_cluster_test.go

**Current State**: 53% modern (43/81 checks)
**Estimated Effort**: 3-4 hours
**Impact**: MEDIUM - Inconsistent patterns

**Legacy Usages to Replace** (38 occurrences):
- Lines 179, 228, 270, 340: Basic attributes (name, version)
- Lines 374, 392, 410: List counts (`worker_nodes.#`)
- Lines 487-491: Multiple attributes in single test
- Lines 546-547, 688-690, 723-724, 757: Update and configuration tests

**Migration Steps**:
1. Identify all `TestCheckResourceAttr` calls
2. Group by attribute type (string, list, computed)
3. Replace with appropriate `knownvalue.*` matchers
4. Add `CompareValue` tracking for `id` attribute across steps
5. Run tests: `TF_ACC=1 go test -v -run TestAccCMKubeCluster ./internal/provider/`

### Priority 3: data_source_cmpart_partitions_test.go

**Current State**: 73% modern (27/37 checks)
**Estimated Effort**: 1-2 hours
**Impact**: LOW - Minor cleanup

**Legacy Usages to Replace** (10 occurrences):
- Likely filter validation and edge cases

**Migration Steps**:
1. Locate remaining `TestCheckResourceAttr` calls
2. Replace with `ExpectKnownValue` + appropriate matchers
3. Run tests: `TF_ACC=1 go test -v -run TestAccCMPartPartitions ./internal/provider/`

## Validation Checklist

After modernizing each file:

- [ ] All imports added (statecheck, knownvalue, tfjsonpath, plancheck)
- [ ] No remaining `resource.TestCheckResourceAttr` calls
- [ ] No remaining `resource.TestCheckResourceAttrSet` calls
- [ ] Type-specific matchers used (StringExact, Int64Exact, Bool, NotNull)
- [ ] Idempotency checks added with `ExpectEmptyPlan()`
- [ ] All tests pass: `TF_ACC=1 go test -v -run TestAcc{ResourceName} ./internal/provider/`
- [ ] Test coverage maintained or improved
- [ ] No breaking changes to test behavior

## Benefits of Modernization

1. **Type Safety**: Catches type mismatches at test time
2. **Better Errors**: Clear messages showing expected vs actual types
3. **Consistency**: Uniform testing patterns across all files
4. **Future-Proof**: Aligns with HashiCorp's latest recommendations
5. **Maintainability**: Easier to understand and modify tests
6. **Idempotency**: Built-in plan checks prevent regressions

## References

- HashiCorp Testing Patterns: https://developer.hashicorp.com/terraform/plugin/testing/testing-patterns
- terraform-plugin-testing v1.13.3: https://github.com/hashicorp/terraform-plugin-testing
- Project CLAUDE.md: /workspace/CLAUDE.md (Modern Testing Patterns section)
- Terraform Provider Design Skill: /workspace/.claude/skills/terraform-provider-design/

## Rollout Plan

1. **Phase 1** (Week 1): Modernize `resource_cmnet_network_test.go` (P1)
2. **Phase 2** (Week 2): Modernize `resource_cmkube_cluster_test.go` (P2)
3. **Phase 3** (Week 3): Modernize `data_source_cmpart_partitions_test.go` (P3)
4. **Phase 4** (Week 4): Update CLAUDE.md with migration examples

## Success Metrics

- [ ] 100% modern pattern adoption (16/16 files)
- [ ] All acceptance tests passing
- [ ] Zero legacy pattern usage
- [ ] Documentation updated with modern examples
