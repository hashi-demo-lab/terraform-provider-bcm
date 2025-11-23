# Quick Start: Modernizing Legacy Testing Patterns

**Feature**: Issue #40 - Test Modernization
**For**: Developers implementing the transformation

---

## Prerequisites

### Environment Setup

```bash
# 1. Export BCM cluster credentials
export TF_ACC=1
export BCM_ENDPOINT="https://172.21.15.254:8081"
export BCM_USERNAME="root"
export BCM_PASSWORD="Hashicorp123!"

# 2. Verify Go environment
go version  # Must be 1.24.0+

# 3. Verify terraform-plugin-testing version
grep terraform-plugin-testing go.mod  # Should be v1.13.3
```

### BCM Cluster Access

```bash
# Verify cluster reachability
curl -k -X POST $BCM_ENDPOINT/json \
  -H "Content-Type: application/json" \
  -d "{\"service\":\"login\",\"username\":\"$BCM_USERNAME\",\"password\":\"$BCM_PASSWORD\"}" \
  && echo "✅ BCM cluster reachable" || echo "❌ BCM cluster offline"
```

---

## Transformation Checklist

### For Each Test File

- [ ] **Step 1**: Add required imports (if needed)
  - `statecheck`
  - `knownvalue`
  - `tfjsonpath`
  - `compare` (resources only)

- [ ] **Step 2**: Transform or remove Check blocks
  - P1 file: Convert legacy → modern
  - P2/P3 files: Remove redundant Check blocks

- [ ] **Step 3**: Add missing patterns
  - Idempotency steps (P1 file only)
  - ID tracking (P1 Basic function only)

- [ ] **Step 4**: Run tests for that file
  - `TF_ACC=1 go test -v -run TestAccFunctionName`

- [ ] **Step 5**: Verify zero legacy patterns
  - `grep -c "TestCheckResourceAttr" file.go` → 0

- [ ] **Step 6**: Run full file test suite
  - `TF_ACC=1 go test -v -run TestAccFilePrefix`

---

## Pattern Quick Reference

### String Attributes
```go
// Legacy
resource.TestCheckResourceAttr("resource", "name", networkName)

// Modern
statecheck.ExpectKnownValue(
    "resource",
    tfjsonpath.New("name"),
    knownvalue.StringExact(networkName),
)
```

### Numeric Attributes
```go
// Legacy
resource.TestCheckResourceAttr("resource", "mtu", "9000")

// Modern
statecheck.ExpectKnownValue(
    "resource",
    tfjsonpath.New("mtu"),
    knownvalue.Int64Exact(9000),  // No quotes!
)
```

### Boolean Attributes
```go
// Legacy
resource.TestCheckResourceAttr("resource", "enabled", "true")

// Modern
statecheck.ExpectKnownValue(
    "resource",
    tfjsonpath.New("enabled"),
    knownvalue.Bool(true),  // No quotes!
)
```

### Computed Fields
```go
// Legacy
resource.TestCheckResourceAttrSet("resource", "uuid")

// Modern
statecheck.ExpectKnownValue(
    "resource",
    tfjsonpath.New("uuid"),
    knownvalue.NotNull(),
)
```

### List Size (Known Count)
```go
// Legacy
resource.TestCheckResourceAttr("resource", "worker_nodes.#", "2")

// Modern
statecheck.ExpectKnownValue(
    "resource",
    tfjsonpath.New("worker_nodes"),  // No .# suffix
    knownvalue.ListSizeExact(2),
)
```

### List Size (Unknown Count)
```go
// Legacy
resource.TestCheckResourceAttrSet("data.resource", "partitions.#")

// Modern
statecheck.ExpectKnownValue(
    "data.resource",
    tfjsonpath.New("partitions"),
    knownvalue.NotNull(),  // Don't hardcode count
)
```

### Idempotency Check
```go
// Add after Create or Update step
{
    Config: testConfig(name),  // Same config
    ConfigPlanChecks: resource.ConfigPlanChecks{
        PreApply: []plancheck.PlanCheck{
            plancheck.ExpectEmptyPlan(),
        },
    },
},
```

### ID Tracking
```go
// Initialize once at function start
compareID := statecheck.CompareValue(compare.ValuesSame())

// Add to Create, Import, Update steps
ConfigStateChecks: []statecheck.StateCheck{
    compareID.AddStateValue("resource", tfjsonpath.New("id")),
},
```

---

## Testing Commands

### Run Single Test Function
```bash
TF_ACC=1 go test -v -timeout 10m ./internal/provider/ \
  -run TestAccCMNetNetwork_Basic
```

### Run All Tests for One File
```bash
# P1 file
TF_ACC=1 go test -v -timeout 30m ./internal/provider/ \
  -run TestAccCMNetNetwork

# P2 file
TF_ACC=1 go test -v -timeout 60m ./internal/provider/ \
  -run TestAccCMKubeCluster

# P3 file
TF_ACC=1 go test -v -timeout 30m ./internal/provider/ \
  -run TestAccCMPartPartitions
```

### Run Full Target Suite
```bash
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ \
  -run "TestAccCMNetNetwork|TestAccCMKubeCluster|TestAccCMPartPartitions"
```

### Verify Zero Legacy Patterns
```bash
# Should return 0
grep -c "TestCheckResourceAttr\|TestCheckResourceAttrSet" \
  internal/provider/resource_cmnet_network_test.go
```

---

## Troubleshooting

### Type Mismatch Errors

**Error**: `expected types.Int64Type, got types.StringType`

**Cause**: Using wrong type matcher

**Fix**: Change `StringExact("9000")` → `Int64Exact(9000)`

---

**Error**: `expected int64, got float64`

**Cause**: BCM API returns float64, not int64

**Fix**: Change `Int64Exact(9000)` → `Float64Exact(9000.0)`

---

### Import Errors

**Error**: `undefined: statecheck`

**Cause**: Missing import

**Fix**: Add to imports:
```go
"github.com/hashicorp/terraform-plugin-testing/statecheck"
```

---

**Error**: `undefined: knownvalue`

**Cause**: Missing import

**Fix**: Add to imports:
```go
"github.com/hashicorp/terraform-plugin-testing/knownvalue"
```

---

### Test Failures

**Error**: `unexpected non-empty plan`

**Cause**: Resource not idempotent (generates diffs on re-apply)

**Fix**: Check resource implementation - may have bug in Read() or Update()

---

**Error**: `ID mismatch detected`

**Cause**: Resource ID changed during Import or Update

**Fix**: Check Import implementation - should preserve existing ID

---

**Error**: `value type mismatch: expected bool, got string "true"`

**Cause**: API returning string instead of boolean

**Fix**: Check resource schema - should use `types.Bool` not `types.String`

---

### Path Errors

**Error**: `cannot use .# in tfjsonpath`

**Cause**: Using legacy `.#` suffix for lists

**Fix**: Remove `.#` - use `tfjsonpath.New("list")` not `tfjsonpath.New("list.#")`

---

### Legacy Pattern Remains

**Issue**: Grep shows remaining TestCheckResourceAttr

**Cause**: Missed some legacy assertions

**Fix**: Search file for all occurrences:
```bash
grep -n "TestCheckResourceAttr\|TestCheckResourceAttrSet" file.go
```

---

## File-Specific Notes

### P1: resource_cmnet_network_test.go

**Strategy**: Full conversion (legacy → modern)

**Required Imports**: Add 4 imports (statecheck, knownvalue, tfjsonpath, compare)

**Key Tasks**:
- Convert 18 legacy assertions
- Add 3 idempotency steps
- Add 1 ID tracking pattern

**Test Runtime**: ~30 seconds per function

---

### P2: resource_cmkube_cluster_test.go

**Strategy**: Remove redundant Check blocks (modern ConfigStateChecks already exist)

**Required Imports**: None (already complete)

**Key Tasks**:
- Remove 10 redundant Check blocks
- Keep all ConfigStateChecks unchanged
- Verify no regressions

**Test Runtime**: ~60 seconds per function (cluster operations slow)

---

### P3: data_source_cmpart_partitions_test.go

**Strategy**: Remove redundant Check blocks

**Required Imports**: None (already has statecheck, knownvalue, tfjsonpath)

**Key Tasks**:
- Remove 5 redundant Check blocks
- Keep all ConfigStateChecks unchanged
- No idempotency needed (data sources)

**Test Runtime**: ~10 seconds per function

---

## Performance Guidelines

**Baseline**: Establish before any changes
```bash
TF_ACC=1 time go test -v -timeout 120m ./internal/provider/ \
  -run "TestAccCMNetNetwork|TestAccCMKubeCluster|TestAccCMPartPartitions" \
  2>&1 | tee baseline.log
```

**Track After Each Phase**:
- After P1: Measure P1 file
- After P2: Measure P2 file
- After P3: Measure P3 file
- After P6: Measure full suite

**Success Threshold**: <10% increase from baseline

---

## Code Quality

### Before Committing

```bash
# Format code
make fmt

# Run linter
make lint

# Verify imports organized
go mod tidy
```

### Commit Message Format

```
feat(tests): Modernize [file] to terraform-plugin-testing v1.13.3 patterns

- Convert [N] legacy assertions to type-safe state checks
- Add [N] idempotency verification steps
- Add ID consistency tracking across operations
- Remove [N] redundant Check blocks

100% modern pattern adoption achieved.

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
```

---

## Success Criteria Verification

After all phases, verify:

- [ ] **SC-001**: Zero `TestCheckResourceAttr()` calls
- [ ] **SC-002**: Zero `TestCheckResourceAttrSet()` calls
- [ ] **SC-003**: All numeric use `Int64Exact()` not string "9000"
- [ ] **SC-004**: All boolean use `Bool(true/false)` not string
- [ ] **SC-005**: All resource tests have idempotency
- [ ] **SC-006**: All resource tests with Import have ID tracking
- [ ] **SC-007**: 100% modern patterns
- [ ] **SC-008**: All 21 tests pass
- [ ] **SC-009**: Performance within 10% baseline

---

## Reference Documentation

- **Contracts**: `specs/040-modernize-legacy-tests/contracts/` (7 files)
- **Data Model**: `specs/040-modernize-legacy-tests/data-model.md`
- **Research**: `specs/040-modernize-legacy-tests/research.md`
- **CLAUDE.md**: Modern Testing Patterns section
- **HashiCorp Docs**: https://developer.hashicorp.com/terraform/plugin/testing/testing-patterns

---

**Quick Start Guide Complete** - Ready for Implementation!
