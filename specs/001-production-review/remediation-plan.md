# Production Readiness Remediation Plan

**Generated**: 2025-11-22
**Feature**: Production-Ready Codebase Review (001-production-review)
**Analysis Type**: User Story 5 - Phased Remediation Plan
**Source Reports**: test-coverage-report.md, api-gap-analysis.md, documentation-review.md, code-consistency-report.md

## Executive Summary

This remediation plan synthesizes findings from comprehensive production-readiness analysis across test coverage, API coverage, documentation, and code consistency. The plan organizes 48 identified issues into 4 prioritized phases with clear success criteria, effort estimates, and regression testing requirements.

### Overall Assessment

**Production Readiness Score**: 85/100
- Test Coverage: 85% (missing drift/idempotency for some resources)
- API Coverage: 45% (only 3 of ~30 BCM services covered)
- Documentation: 95% (excellent example coverage, test harness functional)
- Code Consistency: 88% (excellent adherence to HashiCorp standards)

**Critical Finding**: Examples follow correct Terraform documentation patterns. Test harness properly injects provider configuration. Only missing tests are Import and Drift for some resources.

### Key Priorities

**Phase 0 (Blockers)** - 2 issues, ~2 hours
- Add missing Import test for bcm_cmpart_softwareimage
- Add missing Drift tests for bcm_cmpart_softwareimage and bcm_cmdevice_device

**Phase 1 (Production Readiness)** - 15 issues, ~18 hours
- Complete idempotency tests for all resources
- Migrate to modern testing patterns (statecheck, plancheck, compareValue)
- Add top 3 high-value BCM resources from API gap analysis

**Phase 2 (Quality & Consistency)** - 13 issues, ~10 hours
- Standardize async operation handling with exponential backoff
- Add advanced examples and missing documentation
- Implement medium-value BCM resources

**Phase 3 (Polish & Future)** - 9 issues, ~8 hours
- Code polish (field mapping docs, validator enhancements)
- Low-value API gaps
- Performance optimizations

**Total Estimated Effort**: 38 hours (4.75 person-days)

---

## Phase 0: Critical Blockers (Production Gate)

**Goal**: Address showstopper issues that prevent provider usage

**Duration**: ~2 hours

**Success Criteria**:
- ✅ All resources have Import test coverage
- ✅ All resources have Drift test coverage
- ✅ Acceptance test suite passes 100%

---

### Blocker 1: Missing Import Tests

**Source**: test-coverage-report.md
**Impact**: Resources cannot be imported into Terraform state

**Missing Import Tests**:
- ❌ bcm_cmpart_softwareimage - No ImportState test
- ✅ bcm_cmdevice_category - HAS ImportState test
- ✅ bcm_cmdevice_device - HAS ImportState test

**Fix Steps**:

1. Add ImportState test step to `resource_cmpart_softwareimage_test.go`:
   ```go
   {
       ResourceName:      "bcm_cmpart_softwareimage.test",
       ImportState:       true,
       ImportStateVerify: true,
   },
   ```

2. Verify ImportState implementation exists in resource file

**Validation**:
```bash
TF_ACC=1 go test -v -run TestAccCMPartSoftwareImage ./internal/provider/
```

**Effort**: 30 minutes
**Priority**: **HIGH - Required for production**

---

### Blocker 2: Missing Drift Detection Tests

**Source**: test-coverage-report.md
**Impact**: Provider may not detect external resource modifications

**Missing Drift Tests**:
- ❌ bcm_cmpart_softwareimage - No drift detection test
- ❌ bcm_cmdevice_device - No drift detection test
- ✅ bcm_cmdevice_category - HAS drift test (TestAccCMDeviceCategory_DriftNotes)

**Fix Steps**:

1. Create `TestAccCMPartSoftwareImage_Drift` in `resource_cmpart_softwareimage_test.go`
2. Create `TestAccCMDeviceDevice_Drift` in `resource_cmdevice_device_test.go`

**Pattern** (from test-coverage-report.md):
```go
func TestAccResourceName_Drift(t *testing.T) {
    name := generateUniqueTestName("test-drift")

    resource.Test(t, resource.TestCase{
        Steps: []resource.TestStep{
            // Step 1: Create resource
            {Config: config(name, "initial"), Check: ...},

            // Step 2: Modify externally + verify drift detected
            {
                PreConfig: func() {
                    client := createTestBCMClient(t)
                    // Modify via BCM API
                    // Wait 2s for consistency
                },
                Config: config(name, "initial"),
                ConfigPlanChecks: resource.ConfigPlanChecks{
                    PreApply: []plancheck.PlanCheck{
                        plancheck.ExpectNonEmptyPlan(),
                    },
                },
            },

            // Step 3: Terraform restores state
            {Config: config(name, "initial"), Check: ...},
        },
    })
}
```

**Validation**:
```bash
TF_ACC=1 go test -v -run "Drift" ./internal/provider/
```

**Effort**: 3 hours (2 tests × 1.5 hours)
**Priority**: **HIGH - Required for production**

---

### Blocker 4: Documentation Regeneration

**Source**: documentation-review.md
**Impact**: Generated docs don't show required configuration

**Issue**: Documentation generated from examples that lack terraform{} blocks

**Fix Steps**:

1. Complete Blocker 1 (fix all examples)
2. Regenerate documentation:
   ```bash
   make generate
   ```
3. Review generated docs/resources/ and docs/data-sources/ for completeness
4. Verify terraform{} blocks appear in generated examples

**Validation**:
```bash
git diff docs/
# Should show updated examples with terraform{} blocks
```

**Effort**: 2 hours (includes review and validation)
**Priority**: **HIGH - Required after Blocker 1**

---

## Phase 0 Summary

| Issue | Effort | Priority | Validation |
|-------|--------|----------|------------|
| Blocker 1: Fix 21 examples | 6h | CRITICAL | test-examples.sh passes |
| Blocker 2: Add Import test (softwareimage) | 1h | HIGH | ImportState test passes |
| Blocker 3: Add Drift tests (2 resources) | 3h | HIGH | Drift tests pass |
| Blocker 4: Regenerate documentation | 2h | HIGH | Docs include config blocks |

**Total Phase 0**: ~12 hours

**Regression Testing Requirements**:
```bash
# Full acceptance test suite
TF_ACC=1 go test -v -timeout 120m ./internal/provider/

# Example validation
make install
./scripts/test-examples.sh

# Documentation validation
make generate
git diff docs/  # Verify updates correct
```

**Phase 0 Exit Criteria**:
- [ ] All 21 examples include terraform{} and provider{} blocks
- [ ] All 21 examples pass terraform init/validate/plan
- [ ] test-examples.sh achieves 100% pass rate
- [ ] All 3 resources have ImportState tests
- [ ] All 3 resources have Drift detection tests
- [ ] Documentation regenerated with complete examples
- [ ] Full acceptance test suite passes (100%)

---

## Phase 1: High-Priority Production Readiness

**Goal**: Complete essential testing and add high-value API coverage

**Duration**: ~18 hours

**Success Criteria**:
- ✅ 100% resources have complete CRUD + Import + Drift + Idempotency tests
- ✅ 0% legacy TestCheckResourceAttr usage (all modern patterns)
- ✅ Top 3 high-value BCM resources implemented with full test coverage
- ✅ All schema attributes have MarkdownDescription

---

### Priority 1: Complete Idempotency Test Coverage

**Source**: test-coverage-report.md
**Impact**: Cannot verify resources don't cause unnecessary changes

**Missing Idempotency Tests**:
- ❌ bcm_cmpart_softwareimage - No idempotency verification
- ❌ bcm_cmdevice_category - No idempotency verification
- ❌ bcm_cmdevice_device - No idempotency verification

**Fix Steps**:

Add test steps verifying ExpectEmptyPlan after Create and Update:
```go
// After Create - verify idempotency
{
    Config: testAccResourceConfig(name, "initial-value"),
    ConfigPlanChecks: resource.ConfigPlanChecks{
        PreApply: []plancheck.PlanCheck{
            plancheck.ExpectEmptyPlan(),
        },
    },
},
// After Update - verify idempotency
{
    Config: testAccResourceConfig(name, "updated-value"),
    ConfigPlanChecks: resource.ConfigPlanChecks{
        PreApply: []plancheck.PlanCheck{
            plancheck.ExpectEmptyPlan(),
        },
    },
},
```

**Validation**:
```bash
TF_ACC=1 go test -v ./internal/provider/
```

**Effort**: 4 hours (3 resources × ~1.3 hours)
**Priority**: **HIGH**

---

### Priority 2: Migrate to Modern Testing Patterns

**Source**: test-coverage-report.md
**Impact**: Tests don't leverage Plugin Framework 1.13.3+ features

**Pattern Migration Required**:
- Replace `resource.TestCheckResourceAttr` → `statecheck.ExpectKnownValue`
- Add ID consistency tracking with `statecheck.CompareValue`
- Add sensitive attribute verification with `statecheck.ExpectSensitiveValue`

**Current Usage**:
- Legacy TestCheckResourceAttr: Used extensively across all test files
- Modern statecheck: Partially adopted in device resource tests

**Fix Steps**:

1. Update all test files to use modern patterns
2. Add ID consistency tracking across Create/Import/Update steps
3. Mark sensitive attributes (passwords, API keys) for verification

**Example Migration**:
```go
// OLD PATTERN
resource.TestCheckResourceAttr("bcm_resource.test", "name", "expected-value")

// NEW PATTERN
ConfigStateChecks: []statecheck.StateCheck{
    statecheck.ExpectKnownValue(
        "bcm_resource.test",
        tfjsonpath.New("name"),
        knownvalue.StringExact("expected-value"),
    ),
}
```

**Validation**:
```bash
TF_ACC=1 go test -v ./internal/provider/
```

**Effort**: 6 hours (modernize all test files)
**Priority**: **MEDIUM-HIGH**

---

### Priority 3: Implement Top 3 High-Value BCM Resources

**Source**: api-gap-analysis.md
**Impact**: Missing critical user workflows for production deployments

**Top 3 High-Value Resources** (from API gap analysis):

**3.1 bcm_cmprov_role** (CMProv Service)
- **API Methods**: `getRoles()`, `addRole()`, `updateRole()`, `removeRole()`
- **Business Value**: HIGH - Software image roles are core provisioning workflow
- **Use Case**: Assign roles (login, compute, storage) to software images
- **Effort**: 4 hours (resource + data source + tests)
- **Dependencies**: bcm_cmpart_softwareimage (exists)

**3.2 bcm_cmjob_job** (CMJob Service)
- **API Methods**: `getJobs()`, `submitJob()`, `getJobStatus()`, `cancelJob()`
- **Business Value**: HIGH - Job monitoring essential for automation
- **Use Case**: Track provisioning jobs, monitor cluster operations
- **Effort**: 5 hours (resource + data source + tests + async handling)
- **Dependencies**: None

**3.3 bcm_cmnet_network** (CMNet Service Resource)
- **API Methods**: `addNetwork()`, `updateNetwork()`, `removeNetwork()`, `getNetwork()`
- **Business Value**: HIGH - Network provisioning completes network management
- **Use Case**: Create/manage cluster networks (currently only data source exists)
- **Effort**: 4 hours (resource + tests, data source exists)
- **Dependencies**: bcm_cmnet_networks data source (exists)

**Implementation Pattern** (TDD - RED-GREEN-REFACTOR):
1. Write failing acceptance test first
2. Implement minimal CRUD operations
3. Add full API integration
4. Add examples and documentation
5. Run `make generate`

**Validation**:
```bash
TF_ACC=1 go test -v -run TestAccCMProvRole ./internal/provider/
TF_ACC=1 go test -v -run TestAccCMJobJob ./internal/provider/
TF_ACC=1 go test -v -run TestAccCMNetNetwork ./internal/provider/
```

**Effort**: 13 hours total (4h + 5h + 4h)
**Priority**: **MEDIUM**

---

## Phase 1 Summary

| Issue | Effort | Priority | Validation |
|-------|--------|----------|------------|
| Priority 1: Add Idempotency tests (3 resources) | 4h | HIGH | All tests pass |
| Priority 2: Migrate to modern testing patterns | 6h | MEDIUM-HIGH | All tests pass |
| Priority 3: Implement top 3 API gaps | 13h | MEDIUM | New tests pass |

**Total Phase 1**: ~23 hours (adjust to 18h if Priority 3 deferred)

**Regression Testing Requirements**:
```bash
# Full test suite with modern patterns
TF_ACC=1 go test -v -timeout 120m ./internal/provider/

# Verify idempotency
TF_ACC=1 go test -v -run "Idempot" ./internal/provider/

# Verify new resources
TF_ACC=1 go test -v -run "CMProvRole|CMJobJob|CMNetNetwork" ./internal/provider/

# Regenerate docs
make generate
```

**Phase 1 Exit Criteria**:
- [ ] All resources have idempotency tests passing
- [ ] All tests use modern statecheck/plancheck patterns
- [ ] Top 3 high-value resources implemented
- [ ] New resources have complete CRUD + Import + Drift + Idempotency tests
- [ ] Examples created for new resources
- [ ] Documentation regenerated
- [ ] Full acceptance test suite passes (100%)

---

## Phase 2: Quality & Consistency Improvements

**Goal**: Standardize code patterns and add comprehensive documentation

**Duration**: ~10 hours

**Success Criteria**:
- ✅ All async operations use exponential backoff
- ✅ All resources have basic + advanced examples
- ✅ Medium-value BCM resources implemented
- ✅ CheckDestroy functions use test helper pattern

---

### Quality 1: Standardize Async Operation Handling

**Source**: code-consistency-report.md
**Impact**: Inefficient API usage, slower operations

**Issues**:
- resource_cmpart_softwareimage.go: Clone polling uses fixed 2s sleep
- resource_cmdevice_device.go: Multiple async operations use fixed sleep

**Fix Steps**:

Replace fixed sleep with exponential backoff:
```go
// Current (fixed delay)
for i := 0; i < 30; i++ {
    time.Sleep(2 * time.Second)
    // Check status
}

// Improved (exponential backoff with cap)
maxRetries := 10
for i := 0; i < maxRetries; i++ {
    waitTime := time.Duration(1<<i) * time.Second
    if waitTime > 30*time.Second {
        waitTime = 30 * time.Second  // Cap at 30s
    }
    time.Sleep(waitTime)
    // Check status
}
```

**Validation**:
```bash
TF_ACC=1 go test -v -run TestAccCMPartSoftwareImage ./internal/provider/
TF_ACC=1 go test -v -run TestAccCMDeviceDevice ./internal/provider/
```

**Effort**: 2 hours
**Priority**: **MEDIUM**

---

### Quality 2: Standardize Test Helper Usage

**Source**: code-consistency-report.md
**Impact**: Code duplication in CheckDestroy functions

**Issues**:
- resource_cmdevice_category_test.go: Custom CheckDestroy retry logic
- resource_cmpart_softwareimage_test.go: Custom CheckDestroy retry logic
- resource_cmdevice_device_test.go: Custom CheckDestroy retry logic

**Fix Steps**:

Refactor CheckDestroy to use `verifyResourceDeleted()` helper:
```go
func testAccCheckResourceDestroy(s *terraform.State) error {
    client := createTestBCMClient(&testing.T{})
    ctx := context.Background()

    for _, rs := range s.RootModule().Resources {
        if rs.Type != "bcm_resource" {
            continue
        }

        // Use helper instead of custom retry logic
        deleted, err := verifyResourceDeleted(ctx, client,
            "Service", "getMethod", rs.Primary.ID, 4)

        if !deleted {
            return fmt.Errorf("resource %s still exists", rs.Primary.ID)
        }
    }
    return nil
}
```

**Validation**:
```bash
TF_ACC=1 go test -v ./internal/provider/
```

**Effort**: 2 hours
**Priority**: **LOW-MEDIUM**

---

### Quality 3: Add Advanced Examples

**Source**: documentation-review.md
**Impact**: Users lack comprehensive usage examples

**Missing Advanced Examples**:
- bcm_cmpart_softwareimages (data source) - Only 1 basic example, needs filtering examples
- All resources - Could benefit from composite workflow examples

**Fix Steps**:

1. Create `examples/data-sources/bcm_cmpart_softwareimages/filter_by_category.tf`
2. Create `examples/data-sources/bcm_cmpart_softwareimages/filter_by_role.tf`
3. Create composite workflow examples (data source → resource chains)

**Effort**: 2 hours
**Priority**: **MEDIUM**

---

### Quality 4: Implement Medium-Value API Gaps

**Source**: api-gap-analysis.md
**Impact**: Additional BCM functionality for advanced users

**Medium-Value Resources** (implement top 2):

**4.1 bcm_cmserv_setting** (CMServ Service)
- **API Methods**: `getSettings()`, `updateSetting()`
- **Business Value**: MEDIUM - Cluster configuration management
- **Use Case**: Configure cluster-wide settings (timeouts, defaults, etc.)
- **Effort**: 3 hours
- **Note**: Read-only settings may limit resource usefulness

**4.2 bcm_cmdevice_power** (CMDevice Service)
- **API Methods**: `powerOn()`, `powerOff()`, `reboot()`
- **Business Value**: MEDIUM - Remote power management
- **Use Case**: Power control for cluster nodes
- **Effort**: 4 hours
- **Note**: May be better as provider function than resource

**Effort**: 7 hours total (defer if timeline constrained)
**Priority**: **LOW-MEDIUM**

---

## Phase 2 Summary

| Issue | Effort | Priority | Validation |
|-------|--------|----------|------------|
| Quality 1: Async operation standardization | 2h | MEDIUM | Tests pass, operations faster |
| Quality 2: Test helper adoption | 2h | LOW-MEDIUM | Tests pass |
| Quality 3: Add advanced examples | 2h | MEDIUM | Examples validate |
| Quality 4: Medium-value API gaps | 7h | LOW-MEDIUM | New tests pass |

**Total Phase 2**: ~13 hours (adjust to 10h if Quality 4 deferred)

**Regression Testing Requirements**:
```bash
TF_ACC=1 go test -v -timeout 120m ./internal/provider/
./scripts/test-examples.sh
make generate
```

**Phase 2 Exit Criteria**:
- [ ] All async operations use exponential backoff
- [ ] All CheckDestroy functions use test helper
- [ ] Advanced examples created for all data sources
- [ ] Medium-value resources implemented (if timeline permits)
- [ ] Full test suite passes
- [ ] Documentation regenerated

---

## Phase 3: Polish & Future Enhancements

**Goal**: Code polish, minor improvements, and low-value API coverage

**Duration**: ~8 hours

**Success Criteria**:
- ✅ Field name mappings documented
- ✅ Enhanced validators for IP addresses
- ✅ Low-value BCM resources evaluated
- ✅ Performance optimizations implemented

---

### Polish 1: Document Field Name Mappings

**Source**: code-consistency-report.md
**Impact**: Developer experience for debugging and drift tests

**Issue**: Terraform snake_case → BCM API camelCase mappings not documented

**Fix Steps**:

Add mapping table to each resource file:
```go
// FIELD NAME MAPPINGS (Terraform snake_case → BCM API camelCase)
//
// Core Fields:
//   - id                    → uuid
//   - management_network    → managementNetwork
//   - kernel_parameters     → kernelParameters
//   - boot_loader           → bootLoader
// ... etc.
```

**Effort**: 1 hour
**Priority**: **LOW**

---

### Polish 2: Enhanced Validators

**Source**: code-consistency-report.md
**Impact**: Better input validation for users

**Enhancements**:
- Add IP address validation to `default_gateway` (resource_cmdevice_category.go)
- Add IP address validation to `name_servers` list elements

**Fix Steps**:

```go
"default_gateway": schema.StringAttribute{
    Optional: true,
    MarkdownDescription: "Default gateway IP address",
    Validators: []validator.String{
        stringvalidator.RegexMatches(
            regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}$`),
            "must be a valid IPv4 address",
        ),
    },
},
```

**Effort**: 1 hour
**Priority**: **LOW**

---

### Polish 3: Evaluate Low-Value API Gaps

**Source**: api-gap-analysis.md
**Impact**: Completeness of BCM API coverage

**Low-Value API Methods**:
- GUI-related settings (CMGui service)
- Internal monitoring (CMMon detailed metrics)
- Legacy compatibility endpoints

**Decision**: Defer low-value APIs until user demand demonstrated

**Effort**: 0 hours (decision to defer)
**Priority**: **LOW**

---

### Polish 4: Performance Optimization

**Source**: code-consistency-report.md (test execution time)
**Impact**: Faster development feedback loops

**Optimizations**:
- Review test parallelization (currently -parallel=4, could increase)
- Cache BCM client authentication across tests
- Optimize example validation script parallelization

**Effort**: 2 hours
**Priority**: **LOW**

---

## Phase 3 Summary

| Issue | Effort | Priority | Validation |
|-------|--------|----------|------------|
| Polish 1: Field mapping documentation | 1h | LOW | Code review |
| Polish 2: Enhanced validators | 1h | LOW | Tests pass |
| Polish 3: Evaluate low-value APIs | 0h | LOW | Decision documented |
| Polish 4: Performance optimization | 2h | LOW | Test time reduced |

**Total Phase 3**: ~4 hours

---

## Remediation Summary

### Total Effort by Phase

| Phase | Duration | Issues | Status |
|-------|----------|--------|--------|
| Phase 0: Critical Blockers | 12h | 4 | **REQUIRED** |
| Phase 1: Production Readiness | 18h | 3 | **REQUIRED** |
| Phase 2: Quality & Consistency | 10h | 4 | **RECOMMENDED** |
| Phase 3: Polish & Future | 4h | 4 | **OPTIONAL** |

**Total Estimated Effort**: 44 hours (5.5 person-days)

---

### Timeline Estimates

**Aggressive Timeline** (1 developer):
- Week 1: Phase 0 + Phase 1
- Week 2: Phase 2
- Total: 2 weeks

**Conservative Timeline** (1 developer):
- Week 1: Phase 0
- Week 2: Phase 1
- Week 3: Phase 2
- Week 4: Phase 3
- Total: 4 weeks

**Parallel Timeline** (2 developers):
- Week 1: Phase 0 (both) + Phase 1 start (split)
- Week 2: Phase 1 complete + Phase 2
- Total: 2 weeks

---

## Success Metrics

### Phase 0 Success Criteria

- [ ] **100%** examples pass terraform init/validate/plan
- [ ] **100%** resources have Import test coverage
- [ ] **100%** resources have Drift test coverage
- [ ] **0** test-examples.sh failures

### Phase 1 Success Criteria

- [ ] **100%** resources have Idempotency test coverage
- [ ] **0%** legacy TestCheckResourceAttr usage (all modern patterns)
- [ ] **+3** high-value resources implemented (roles, jobs, network)
- [ ] **100%** acceptance test pass rate

### Phase 2 Success Criteria

- [ ] **100%** async operations use exponential backoff
- [ ] **100%** CheckDestroy functions use test helper
- [ ] **≥2** advanced examples per data source
- [ ] **+2** medium-value resources (if timeline permits)

### Phase 3 Success Criteria

- [ ] Field name mappings documented
- [ ] Enhanced IP validation implemented
- [ ] Test execution time reduced by 20%

---

## Regression Testing Strategy

### Per-Phase Testing

**After Phase 0**:
```bash
# Example validation (critical)
make install
./scripts/test-examples.sh

# Acceptance tests
TF_ACC=1 go test -v -timeout 120m ./internal/provider/

# Documentation
make generate
git diff docs/
```

**After Phase 1**:
```bash
# Full test suite with idempotency
TF_ACC=1 go test -v -timeout 120m ./internal/provider/

# Specific idempotency tests
TF_ACC=1 go test -v -run "Idempot" ./internal/provider/

# New resources
TF_ACC=1 go test -v -run "CMProvRole|CMJobJob|CMNetNetwork" ./internal/provider/
```

**After Phase 2**:
```bash
# Full suite
TF_ACC=1 go test -v -timeout 120m ./internal/provider/

# Examples
./scripts/test-examples.sh

# Performance
time TF_ACC=1 go test -v -timeout 120m ./internal/provider/
```

**After Phase 3**:
```bash
# Final validation
TF_ACC=1 go test -v -timeout 120m ./internal/provider/
./scripts/test-examples.sh
make generate
git status  # Verify clean
```

---

## Risk Mitigation

### High-Risk Items

**Risk 1**: Example fixes break existing workflows
**Mitigation**: Test each example individually before bulk changes

**Risk 2**: Modern testing patterns introduce regressions
**Mitigation**: Migrate one test file at a time, verify before proceeding

**Risk 3**: New resources depend on unavailable BCM API versions
**Mitigation**: Verify API availability on target BCM cluster before implementation

**Risk 4**: Async operation changes affect timing-sensitive tests
**Mitigation**: Test with various BCM cluster loads, adjust backoff caps if needed

---

## Dependencies

### External Dependencies

- **BCM Cluster Access**: All phases require access to https://172.21.15.254:8081
- **BCM API Version**: Verify BCM version supports all required API methods
- **Terraform Version**: Ensure Terraform 1.5+ for required_providers syntax
- **Go Version**: Go 1.24+ required for provider compilation

### Internal Dependencies

- **Phase 0 → Phase 1**: Must fix examples before adding new resources (examples needed for docs)
- **Phase 1 → Phase 2**: Modern testing patterns must be established before new test files
- **terraform-provider-design skill**: Required for HashiCorp best practices validation

---

## Conclusion

This remediation plan provides a clear path from current 72% production readiness to 95%+ with all critical issues resolved in Phase 0-1 (30 hours). The provider demonstrates strong technical foundations (88% code consistency score) but requires documentation fixes and test coverage completion before production deployment.

**Recommended Approach**:
1. Execute Phase 0 immediately (critical blocker)
2. Execute Phase 1 for production deployment
3. Execute Phase 2 for quality and user experience
4. Defer Phase 3 until user demand demonstrated

**Production Gate**: Complete Phase 0 + Phase 1 before deploying to production (30 hours, 4 person-days).
