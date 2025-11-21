# Implementation Plan: Comprehensive Test Review - Drift Detection and Destroy Testing

**Branch**: `006-test-review` | **Date**: 2025-11-21 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/workspace/specs/006-test-review/spec.md`

## Summary

This feature enhances the Terraform Provider for Nvidia BCM test infrastructure to achieve comprehensive drift detection and destroy testing coverage. The primary requirement is to add missing drift detection tests for all resources (currently NO drift tests exist) and enhance destroy verification with edge case coverage. The technical approach follows TDD RED-GREEN-REFACTOR cycles, adding test infrastructure (helper functions, BCM client access in tests), implementing drift detection test patterns for 80% of non-computed attributes, and strengthening CheckDestroy and PreCheck cleanup implementations.

## Technical Context

**Language/Version**: Go 1.24
**Primary Dependencies**: terraform-plugin-framework v1.16.1, terraform-plugin-testing v1.13.3
**Storage**: BCM JSON-RPC API at 172.21.15.254:8081
**Testing**: Go testing framework with TF_ACC=1 acceptance tests
**Target Platform**: Linux server (acceptance tests against BCM cluster)
**Project Type**: Single project (Terraform Provider)
**Performance Goals**: Acceptance test suite completes in <5 minutes total, PreCheck cleanup completes in <30 seconds per resource
**Constraints**: Tests must use unique resource names for parallel execution, must handle BCM eventual consistency (30s window), must verify deletion via BCM API
**Scale/Scope**: 2 existing resource types (bcm_cmpart_softwareimage, bcm_cmdevice_category), 28 functional requirements, 10 success criteria

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**Constitution Compliance Status**: PASS (No violations)

- **TDD Adherence**: This feature is 100% test-focused with RED-GREEN-REFACTOR cycles
- **Simplicity**: No new resources or data sources - only test infrastructure improvements
- **Attribute Coverage**: Plan achieves 80% drift coverage target for non-computed attributes
- **Destroy Verification**: Enhanced CheckDestroy will verify ALL resources in state, not just test resources
- **PreCheck Cleanup**: Exponential backoff retry logic already exists, will be standardized

**No complexity justification needed** - This feature enhances existing test infrastructure without adding new complexity.

## Project Structure

### Documentation (this feature)

```text
specs/006-test-review/
├── plan.md              # This file (/speckit.plan command output)
├── spec.md              # Feature specification (already exists)
├── research.md          # Phase 0 output (test pattern research)
├── data-model.md        # Phase 1 output (test infrastructure entities)
├── quickstart.md        # Phase 1 output (developer quick start for drift tests)
├── contracts/           # Phase 1 output (test helper function signatures)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
terraform-provider-bcm/
├── internal/provider/
│   ├── test_helpers.go                           # NEW: Shared test helper functions
│   ├── resource_cmpart_softwareimage_test.go     # ENHANCE: Add drift tests
│   ├── resource_cmdevice_category_test.go        # ENHANCE: Add drift tests
│   ├── data_source_cmpart_softwareimages_test.go # EXISTING: No changes needed
│   ├── data_source_cmdevice_categories_test.go   # EXISTING: No changes needed
│   ├── data_source_cmdevice_nodes_test.go        # EXISTING: No changes needed
│   ├── bcm_client.go                              # EXISTING: Already supports CallJSONRPC
│   └── provider_test.go                           # EXISTING: testAccProtoV6ProviderFactories
├── CLAUDE.md                                      # UPDATE: Add drift test patterns
└── AGENTS.md                                      # UPDATE: Add CheckDestroy/PreCheck patterns
```

**Structure Decision**: Single project structure maintained. All test enhancements are additions to existing `*_test.go` files in `internal/provider/`. New shared test helper functions will be centralized in `internal/provider/test_helpers.go` to avoid duplication across test files.

## Complexity Tracking

> No violations - this section intentionally left blank.

## Phase 0: Outline & Research

**Goal**: Research Terraform provider drift detection patterns, BCM API test interaction patterns, and best practices for CheckDestroy/PreCheck implementations.

### Research Tasks

1. **Drift Detection Test Patterns**
   - **Question**: What are the recommended patterns for testing drift detection in Terraform providers?
   - **Research Target**: HashiCorp terraform-plugin-testing documentation, provider development best practices
   - **Expected Output**:
     - Standard pattern: Create resource → Manually modify via API (PreConfig) → Verify drift detected (ExpectNonEmptyPlan)
     - Use of `ExpectNonEmptyPlan: true` to verify plan shows changes after external modification
     - PreConfig function pattern for invoking BCM API between test steps

2. **BCM Client Test Access Pattern**
   - **Question**: How should tests access BCM client to simulate external modifications?
   - **Research Target**: Existing test infrastructure in this provider
   - **Expected Output**:
     - Pattern: Create BCM client in test using same credentials as provider (os.Getenv)
     - Use `client.CallJSONRPC(ctx, service, method, args...)` to modify resources
     - Example: `client.CallJSONRPC(ctx, "CMPart", "updateSoftwareImage", uuid, modifiedFields)`

3. **CheckDestroy Enhancement Patterns**
   - **Question**: What are best practices for robust CheckDestroy implementations?
   - **Research Target**: terraform-plugin-testing documentation, existing CheckDestroy functions
   - **Expected Output**:
     - MUST iterate ALL resources in state: `for _, rs := range s.RootModule().Resources`
     - MUST verify via API read, not just state check
     - MUST check both error AND empty response (BCM returns empty object for not found)
     - Logging recommendations for debugging

4. **PreCheck Cleanup Best Practices**
   - **Question**: How should PreCheck cleanup handle async operations and verify completion?
   - **Research Target**: Existing PreCheck implementations in this provider
   - **Expected Output**:
     - Exponential backoff pattern (already implemented in softwareimage test)
     - Retry logic: 5 retries, starting at 1s, doubling each time
     - Verify deletion via API read after each retry
     - Warning logging if deletion not confirmed after max retries

5. **Attribute Coverage Strategy**
   - **Question**: How to systematically achieve 80% drift coverage for all attribute types?
   - **Research Target**: Resource schema definitions
   - **Expected Output**:
     - List ALL non-computed attributes per resource
     - Group by type: string, bool, int64, list, nested object
     - Prioritize: P1=frequently modified (kernel_parameters, notes), P2=configuration (enable_sol), P3=structural (modules)

### Research Deliverable: `research.md`

Document structure:
```markdown
# Test Infrastructure Research

## Drift Detection Pattern
- Decision: Use PreConfig with ExpectNonEmptyPlan
- Rationale: Terraform standard pattern for external change detection
- Implementation: [code example]

## BCM Client Access in Tests
- Decision: Create standalone client in test functions
- Rationale: Isolate test from provider instance
- Implementation: [helper function signature]

## CheckDestroy Enhancement
- Decision: Iterate all resources, verify via API, check error + empty response
- Rationale: Comprehensive verification prevents resource leaks
- Implementation: [enhanced pattern]

## PreCheck Cleanup
- Decision: Standardize exponential backoff pattern
- Rationale: Handles BCM eventual consistency
- Implementation: [retry logic template]

## Attribute Coverage Plan
- bcm_cmpart_softwareimage: [list of attributes with priority]
- bcm_cmdevice_category: [list of attributes with priority]
```

**Output**: `specs/006-test-review/research.md` with all NEEDS CLARIFICATION resolved

---

## Phase 1: Design & Contracts

**Prerequisites**: `research.md` complete

### 1.1 Data Model: Test Infrastructure Entities

Generate `data-model.md` describing test infrastructure components:

**Entities**:
- **TestBCMClient**: Helper function to create BCM client for test use
  - Fields: endpoint, username, password (from environment)
  - Relationships: Used by drift tests, CheckDestroy, PreCheck
  - Validation: Credentials must be set

- **DriftTestStep**: Structure for drift detection test pattern
  - Fields: resourceName, attributeToModify, originalValue, modifiedValue
  - Relationships: Uses TestBCMClient in PreConfig
  - State Transitions: Create → Modify (external) → Detect Drift → Restore

- **CheckDestroyFunction**: Per-resource verification function
  - Fields: resourceType, stateResources
  - Relationships: Iterates state, calls TestBCMClient
  - Validation: Error if any resource still exists

- **PreCheckCleanupFunction**: Per-resource cleanup function
  - Fields: resourceNames (varargs), retryConfig
  - Relationships: Uses TestBCMClient, exponential backoff
  - Validation: Warning if deletion not confirmed after retries

### 1.2 API Contracts: Test Helper Function Signatures

Generate contracts in `contracts/test_helpers.go.md`:

```markdown
# Test Helper Contracts

## createTestBCMClient
**Purpose**: Create authenticated BCM client for test use
**Signature**: `func createTestBCMClient(t *testing.T) *BCMClient`
**Parameters**:
  - t: testing.T for error reporting
**Returns**: Configured BCMClient with authentication
**Environment Variables**: BCM_ENDPOINT, BCM_USERNAME, BCM_PASSWORD
**Error Handling**: t.Fatalf if credentials missing or login fails

## testAccCheckResourceDestroy (Pattern)
**Purpose**: Verify all resources of a type are deleted
**Signature**: `func testAccCheck<Resource>Destroy(s *terraform.State) error`
**Parameters**:
  - s: Terraform state to verify
**Returns**: error if any resources still exist
**Implementation**:
  - Iterate s.RootModule().Resources
  - Filter by rs.Type
  - Call BCM API to verify deletion
  - Return error if resource exists

## testAcc<Resource>PreCheck (Pattern)
**Purpose**: Clean up leftover resources before test
**Signature**: `func testAcc<Resource>PreCheck(t *testing.T, names ...string)`
**Parameters**:
  - t: testing.T for logging
  - names: Resource names to clean up
**Returns**: void (logs warnings on failure)
**Implementation**:
  - For each name, check if exists
  - If exists, delete with retry logic
  - Verify deletion with exponential backoff
  - Log success/warning

## verifyResourceDeleted
**Purpose**: Poll BCM API until resource is deleted
**Signature**: `func verifyResourceDeleted(client *BCMClient, service, method, identifier string, maxRetries int) error`
**Parameters**:
  - client: BCM client
  - service: BCM service name (e.g., "CMPart")
  - method: BCM method name (e.g., "getSoftwareImage")
  - identifier: Resource identifier (name or UUID)
  - maxRetries: Maximum retry attempts (default: 5)
**Returns**: error if resource still exists after retries
**Implementation**: Exponential backoff: 1s, 2s, 4s, 8s, 16s
```

### 1.3 Quickstart Guide

Generate `quickstart.md` with developer workflow:

```markdown
# Drift Detection Testing Quickstart

## Adding Drift Test to Resource

1. **Identify Attribute to Test**
   - Choose non-computed attribute (e.g., kernel_parameters, notes)
   - Verify attribute is modifiable via BCM API

2. **Create Test Function**
   ```go
   func TestAccResource_Drift<Attribute>(t *testing.T) {
       resourceName := generateUniqueTestName("test-drift")
       // ... test implementation
   }
   ```

3. **Test Steps**
   - Step 1: Create resource with initial config
   - Step 2: Modify via BCM API (PreConfig), verify drift (ExpectNonEmptyPlan)
   - Step 3: Apply to restore, verify attribute matches config

4. **PreConfig Example**
   ```go
   PreConfig: func() {
       client := createTestBCMClient(t)
       // Modify resource via BCM API
       _, err := client.CallJSONRPC(ctx, "CMPart", "updateSoftwareImage", uuid, fields)
       if err != nil {
           t.Fatalf("Failed to simulate drift: %v", err)
       }
   }
   ```

5. **Run Test**
   ```bash
   TF_ACC=1 BCM_ENDPOINT="https://..." BCM_USERNAME="..." BCM_PASSWORD="..." \
     go test -v -timeout 120m ./internal/provider/ -run TestAccResource_Drift
   ```

## Enhancing CheckDestroy

1. **Verify All Resources in State**
   ```go
   for _, rs := range s.RootModule().Resources {
       if rs.Type != "bcm_resource_type" {
           continue
       }
       // Verify deletion
   }
   ```

2. **Check Both Error and Empty Response**
   ```go
   body, err := client.CallJSONRPC(ctx, service, method, identifier)
   if err == nil {
       var data map[string]interface{}
       if json.Unmarshal(body, &data) == nil && len(data) > 0 {
           return fmt.Errorf("resource still exists")
       }
   }
   ```
```

### 1.4 Agent Context Update

Run agent context update script:
```bash
.specify/scripts/bash/update-agent-context.sh copilot
```

This updates `.github/copilot-instructions.md` with:
- New test helper patterns
- Drift detection test template
- Enhanced CheckDestroy pattern
- PreCheck cleanup standardization

**Output**:
- `data-model.md` with test infrastructure entities
- `contracts/test_helpers.go.md` with function signatures
- `quickstart.md` with developer workflow
- Updated agent context files

---

## Phase 2: Implementation Plan (This Document)

**This phase is complete** - this document serves as the implementation plan output.

**Key Design Decisions**:

### 2.1 Test Infrastructure Architecture

**Component**: Shared Test Helpers (`internal/provider/test_helpers.go`)

**Purpose**: Centralize common test utilities to avoid duplication across resource test files.

**Functions**:
1. `createTestBCMClient(t *testing.T) *BCMClient`
   - Creates authenticated BCM client using environment variables
   - Uses same NewBCMClient as provider
   - Handles error reporting via t.Fatalf

2. `verifyResourceDeleted(ctx context.Context, client *BCMClient, service, method, identifier string, maxRetries int) (bool, error)`
   - Polls BCM API with exponential backoff
   - Returns (true, nil) if deleted
   - Returns (false, nil) if still exists after maxRetries
   - Returns (false, error) on API errors

**Rationale**:
- DRY principle - CheckDestroy and PreCheck share retry logic
- Consistent error handling across all test files
- Easier to maintain and update test patterns

### 2.2 Drift Detection Test Pattern

**Structure**: Per-attribute test functions

**Example**:
```go
func TestAccCMPartSoftwareImage_DriftKernelParameters(t *testing.T) {
    imageName := generateUniqueTestName("test-drift-kernel")
    imagePath := fmt.Sprintf("/cm/images/%s", imageName)

    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccCMPartSoftwareImagePreCheck(t, imageName) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        CheckDestroy:             testAccCheckCMPartSoftwareImageDestroy,
        Steps: []resource.TestStep{
            // Step 1: Create with initial kernel_parameters
            {
                Config: testAccCMPartSoftwareImageResourceConfig_DriftKernel(imageName, imagePath, "quiet splash"),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("bcm_cmpart_softwareimage.test", "kernel_parameters", "quiet splash"),
                    resource.TestCheckResourceAttrSet("bcm_cmpart_softwareimage.test", "uuid"),
                ),
            },
            // Step 2: Simulate external modification via BCM API
            {
                PreConfig: func() {
                    // Create BCM client
                    client := createTestBCMClient(t)

                    // Get resource UUID from state
                    uuid := getResourceUUID(t, "bcm_cmpart_softwareimage.test")

                    // Modify kernel_parameters via BCM API
                    updateFields := map[string]interface{}{
                        "kernelParameters": "quiet splash nomodeset", // BCM API field name
                    }
                    _, err := client.CallJSONRPC(context.Background(), "CMPart", "updateSoftwareImage", uuid, updateFields)
                    if err != nil {
                        t.Fatalf("Failed to simulate external modification: %v", err)
                    }
                },
                Config: testAccCMPartSoftwareImageResourceConfig_DriftKernel(imageName, imagePath, "quiet splash"),
                Check: resource.ComposeAggregateTestCheckFunc(
                    // After refresh, state should reflect BCM's current value
                    resource.TestCheckResourceAttr("bcm_cmpart_softwareimage.test", "kernel_parameters", "quiet splash nomodeset"),
                ),
                ExpectNonEmptyPlan: true, // Plan should show diff to restore "quiet splash"
            },
            // Step 3: Apply to restore desired state
            {
                Config: testAccCMPartSoftwareImageResourceConfig_DriftKernel(imageName, imagePath, "quiet splash"),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("bcm_cmpart_softwareimage.test", "kernel_parameters", "quiet splash"),
                ),
            },
        },
    })
}
```

**Rationale**:
- 3-step pattern clearly shows: initial state → drift → restoration
- ExpectNonEmptyPlan verifies Terraform detects the change
- PreConfig simulates real-world scenario of external modification

### 2.3 Attribute Coverage Strategy

**bcm_cmpart_softwareimage** (Target: 80% of non-computed attributes)

Total non-computed attributes: ~15
Target drift tests: 12 (80%)

**Priority 1 (Must Test - 6 attributes)**:
1. `kernel_parameters` (string) - Frequently modified
2. `notes` (string) - User documentation
3. `enable_sol` (bool) - Serial over LAN toggle
4. `sol_speed` (string) - Configuration parameter
5. `sol_flow_control` (bool) - Configuration parameter
6. `modules` (list of objects) - Complex list structure

**Priority 2 (Should Test - 4 attributes)**:
7. `kernel_output_console` (string) - Console configuration
8. `sol_port` (string) - Port configuration
9. `kernel_version` (string) - Version field (may be computed after clone)
10. Resource deletion detection (external delete)

**Priority 3 (Nice to Have - 2 attributes)**:
11. `path` (string) - Filesystem path
12. Nested module field modifications

**Excluded from drift testing**:
- `uuid` - Computed, never changes
- `id` - Computed, never changes
- `creation_time` - Computed, never changes
- `original_image` - Reset by BCM after clone (in ImportStateVerifyIgnore)

**bcm_cmdevice_category** (Target: 80% of non-computed attributes)

Total non-computed attributes: ~12
Target drift tests: 10 (80%)

**Priority 1 (Must Test - 5 attributes)**:
1. `notes` (string) - User documentation
2. `kernel_parameters` (string) - Frequently modified
3. `install_boot_record` (bool) - Configuration flag
4. `allow_networking_restart` (bool) - Configuration flag
5. `software_image_proxy.parent_software_image` (nested string) - Complex nested object

**Priority 2 (Should Test - 4 attributes)**:
6. `management_network` (string) - Network reference
7. `boot_loader` (string) - Boot configuration
8. `bmc_settings` (nested object) - Complex nested structure
9. Resource deletion detection (external delete)

**Priority 3 (Nice to Have - 1 attribute)**:
10. Multiple nested field modifications in single drift test

**Excluded from drift testing**:
- `uuid` - Computed, never changes
- `id` - Computed, never changes
- `base_type` - Computed, never changes
- `force` - Not persisted in BCM (in ImportStateVerifyIgnore)

### 2.4 CheckDestroy Enhancement Pattern

**Current Implementation** (Good foundation):
```go
func testAccCheckCMPartSoftwareImageDestroy(s *terraform.State) error {
    client, err := NewBCMClient(...)
    if err != nil {
        return fmt.Errorf("failed to create BCM client: %w", err)
    }

    for _, rs := range s.RootModule().Resources {
        if rs.Type != "bcm_cmpart_softwareimage" {
            continue
        }

        name := rs.Primary.Attributes["name"]
        body, err := client.CallJSONRPC(context.Background(), "CMPart", "getSoftwareImage", name)

        if err == nil {
            var imageData map[string]interface{}
            if json.Unmarshal(body, &imageData) == nil && len(imageData) > 0 {
                return fmt.Errorf("software image %s still exists", name)
            }
        }
    }

    return nil
}
```

**Enhancements to Add**:
1. **Logging for debugging**
   - Log number of resources checked
   - Log each resource verification (success/failure)

2. **Timeout handling**
   - Add context with timeout to CallJSONRPC
   - Return clear error if API call times out

3. **Detailed error messages**
   - Include resource UUID in error message
   - Include BCM API response in error for debugging

**Enhanced Implementation**:
```go
func testAccCheckCMPartSoftwareImageDestroy(s *terraform.State) error {
    client, err := NewBCMClient(context.Background(),
        os.Getenv("BCM_ENDPOINT"),
        os.Getenv("BCM_USERNAME"),
        os.Getenv("BCM_PASSWORD"),
        true, 30)
    if err != nil {
        return fmt.Errorf("failed to create BCM client for destroy verification: %w", err)
    }

    resourcesChecked := 0
    for _, rs := range s.RootModule().Resources {
        if rs.Type != "bcm_cmpart_softwareimage" {
            continue
        }

        resourcesChecked++
        name := rs.Primary.Attributes["name"]
        uuid := rs.Primary.Attributes["uuid"]

        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()

        body, err := client.CallJSONRPC(ctx, "CMPart", "getSoftwareImage", name)

        // Resource should NOT exist
        if err == nil {
            var imageData map[string]interface{}
            if json.Unmarshal(body, &imageData) == nil && len(imageData) > 0 {
                return fmt.Errorf("software image still exists after destroy: name=%s uuid=%s response=%s",
                    name, uuid, string(body))
            }
        }
        // Error or empty response = deleted (expected)
    }

    // Log completion
    if resourcesChecked == 0 {
        return fmt.Errorf("no resources found to verify destroy")
    }

    return nil
}
```

**Rationale**:
- Logging aids debugging test failures
- Timeouts prevent tests hanging on API issues
- Detailed errors help diagnose why destroy verification failed

### 2.5 PreCheck Cleanup Standardization

**Current Implementation** (Already good in softwareimage):
- Exponential backoff with 5 retries
- Deletion verification after each retry
- Warning logging if not confirmed

**Enhancements**:
1. **Extract to shared helper function**
   - `verifyResourceDeleted()` in test_helpers.go
   - Reusable across all resource types

2. **Standardize retry configuration**
   - maxRetries: 5
   - initialWait: 1 second
   - backoff factor: 2x
   - total max wait: ~31 seconds

3. **Consistent logging**
   - Log cleanup attempts
   - Log successful cleanup with timing
   - Log warnings with retry count

**Standardized PreCheck Pattern**:
```go
func testAccCMPartSoftwareImagePreCheck(t *testing.T, names ...string) {
    testAccPreCheck(t) // Base precheck

    client, err := createTestBCMClient(t)
    if err != nil {
        t.Logf("Failed to create BCM client for cleanup: %v", err)
        return
    }

    for _, name := range names {
        // Check if resource exists
        body, err := client.CallJSONRPC(context.Background(), "CMPart", "getSoftwareImage", name)
        if err == nil {
            var imageData map[string]interface{}
            if json.Unmarshal(body, &imageData) == nil && len(imageData) > 0 {
                uuid := imageData["uuid"].(string)

                // Delete resource
                _, err := client.CallJSONRPC(context.Background(), "CMPart", "removeSoftwareImage", uuid, false, false, false)
                if err != nil {
                    t.Logf("Failed to delete leftover resource %s: %v", name, err)
                    continue
                }

                // Verify deletion with shared helper
                deleted, err := verifyResourceDeleted(context.Background(), client, "CMPart", "getSoftwareImage", name, 5)
                if err != nil {
                    t.Logf("Error verifying deletion of %s: %v", name, err)
                } else if deleted {
                    t.Logf("✓ Cleaned up leftover test image: %s", name)
                } else {
                    t.Logf("⚠ Warning: Resource %s may not be fully deleted after retries", name)
                }
            }
        }
    }
}
```

**Rationale**:
- Shared helper reduces code duplication
- Standardized configuration ensures consistent behavior
- Clear logging aids debugging test failures

### 2.6 Test File Organization

**File**: `internal/provider/test_helpers.go` (NEW)
- `createTestBCMClient(t *testing.T) *BCMClient`
- `verifyResourceDeleted(ctx, client, service, method, identifier, maxRetries) (bool, error)`

**File**: `internal/provider/resource_cmpart_softwareimage_test.go` (ENHANCE)
- Existing tests: Basic, FullConfig, UpdateKernelConfig, UpdateModules, UpdateSOL, etc.
- **NEW drift tests** (6-10 new test functions):
  - `TestAccCMPartSoftwareImage_DriftKernelParameters`
  - `TestAccCMPartSoftwareImage_DriftNotes`
  - `TestAccCMPartSoftwareImage_DriftEnableSOL`
  - `TestAccCMPartSoftwareImage_DriftSOLSpeed`
  - `TestAccCMPartSoftwareImage_DriftSOLFlowControl`
  - `TestAccCMPartSoftwareImage_DriftModules` (list modification)
  - `TestAccCMPartSoftwareImage_DriftKernelOutputConsole`
  - `TestAccCMPartSoftwareImage_DriftSOLPort`
  - `TestAccCMPartSoftwareImage_DriftExternalDelete` (resource deletion detection)
- **Enhanced**:
  - `testAccCheckCMPartSoftwareImageDestroy` (add logging, timeout, detailed errors)
  - `testAccCMPartSoftwareImagePreCheck` (use shared helper)

**File**: `internal/provider/resource_cmdevice_category_test.go` (ENHANCE)
- Existing tests: Basic, Import, ForceParameter
- **NEW drift tests** (5-8 new test functions):
  - `TestAccCMDeviceCategory_DriftNotes`
  - `TestAccCMDeviceCategory_DriftKernelParameters`
  - `TestAccCMDeviceCategory_DriftInstallBootRecord`
  - `TestAccCMDeviceCategory_DriftAllowNetworkingRestart`
  - `TestAccCMDeviceCategory_DriftSoftwareImageProxy` (nested object)
  - `TestAccCMDeviceCategory_DriftManagementNetwork`
  - `TestAccCMDeviceCategory_DriftBootLoader`
  - `TestAccCMDeviceCategory_DriftExternalDelete`
- **Enhanced**:
  - `testAccCheckCMDeviceCategoryDestroy` (add logging, timeout, detailed errors)
  - `testAccCMDeviceCategoryPreCheck` (use shared helper)

**Rationale**:
- Separate test_helpers.go keeps shared code centralized
- Each resource test file contains all tests for that resource
- Drift tests grouped together in each file for easy maintenance

### 2.7 Test Execution Strategy

**Parallel Execution**: Tests use unique names (timestamp-based), safe for parallel execution

**Test Grouping**:
- Run all drift tests together: `go test -run "Drift"`
- Run all destroy tests: Already covered by CheckDestroy in all tests
- Run specific resource: `go test -run "CMPartSoftwareImage"`

**CI Integration**:
- Acceptance tests run on every commit: `TF_ACC=1 make testacc`
- Drift tests included in standard acceptance test suite
- No separate drift test suite needed

**Rationale**:
- Drift tests follow same pattern as existing acceptance tests
- No special infrastructure needed for drift testing
- Integrated into existing CI/CD pipeline

---

## Implementation Roadmap

This plan follows the 5-phase roadmap from the specification:

### Phase 1: Drift Detection Infrastructure (Priority: P1)

**Deliverables**:
1. Create `internal/provider/test_helpers.go` with:
   - `createTestBCMClient(t *testing.T) *BCMClient`
   - `verifyResourceDeleted(ctx, client, service, method, identifier, maxRetries) (bool, error)`

2. Add first drift detection test for bcm_cmpart_softwareimage:
   - `TestAccCMPartSoftwareImage_DriftKernelParameters`
   - Verify PreConfig pattern works
   - Verify ExpectNonEmptyPlan correctly detects drift

3. Add first drift detection test for bcm_cmdevice_category:
   - `TestAccCMDeviceCategory_DriftNotes`
   - Verify pattern works for category resources

4. Document drift test pattern in `CLAUDE.md`

**Validation**: 2 passing drift detection tests (one per resource), test_helpers.go functions working

**Estimated Effort**: 1-2 days

---

### Phase 2: Comprehensive Drift Coverage (Priority: P1)

**Deliverables**:
1. **bcm_cmpart_softwareimage** drift tests (6-10 tests):
   - String attributes: `DriftNotes`, `DriftKernelOutputConsole`, `DriftSOLPort`
   - Bool attributes: `DriftEnableSOL`, `DriftSOLFlowControl`
   - String (config): `DriftSOLSpeed`
   - List attribute: `DriftModules` (add/remove/modify elements)
   - Resource deletion: `DriftExternalDelete`

2. **bcm_cmdevice_category** drift tests (5-8 tests):
   - String attributes: `DriftKernelParameters`, `DriftManagementNetwork`, `DriftBootLoader`
   - Bool attributes: `DriftInstallBootRecord`, `DriftAllowNetworkingRestart`
   - Nested object: `DriftSoftwareImageProxy` (modify parent_software_image)
   - Resource deletion: `DriftExternalDelete`

3. Test configuration functions for drift scenarios

**Validation**:
- 12+ passing drift detection tests total
- 80% attribute coverage achieved for both resources
- All tests pass consistently

**Estimated Effort**: 2-3 days

---

### Phase 3: Enhanced Destroy Testing (Priority: P2)

**Deliverables**:
1. **Enhance CheckDestroy functions**:
   - Add logging (resources checked, verification results)
   - Add timeout handling (10s context per API call)
   - Add detailed error messages (include UUID, API response)
   - Apply to both `testAccCheckCMPartSoftwareImageDestroy` and `testAccCheckCMDeviceCategoryDestroy`

2. **Standardize PreCheck cleanup**:
   - Refactor to use shared `verifyResourceDeleted` helper
   - Apply to `testAccCMPartSoftwareImagePreCheck` and `testAccCMDeviceCategoryPreCheck`
   - Ensure consistent retry configuration (5 retries, exponential backoff)

3. **Add destroy edge case tests** (3-5 new tests):
   - `TestAccCMPartSoftwareImage_DestroyIdempotent` (run destroy twice)
   - `TestAccCMDeviceCategory_DestroyWithForce` (force=true scenario)
   - `TestAccCMPartSoftwareImage_DestroyDuringClone` (if applicable)

**Validation**:
- Enhanced CheckDestroy functions with logging
- Standardized PreCheck cleanup
- 3+ new destroy edge case tests passing

**Estimated Effort**: 1-2 days

---

### Phase 4: Documentation and Patterns (Priority: P3)

**Deliverables**:
1. **Update `CLAUDE.md`** with:
   - Drift detection test pattern (3-step: create → modify → restore)
   - PreConfig example for external modifications
   - ExpectNonEmptyPlan usage
   - Test helper function reference

2. **Update `AGENTS.md`** with:
   - Enhanced CheckDestroy pattern
   - Standardized PreCheck cleanup pattern
   - Drift test TDD cycle example

3. **Create test coverage matrix**:
   - Spreadsheet or markdown table
   - Columns: Resource, Attribute, Type, Drift Test, Destroy Test, Status
   - Track 80% coverage metric

4. **Document BCM-specific test considerations**:
   - Async operations (image cloning)
   - Eventual consistency (30s window)
   - Field resets (original_image after clone)
   - ImportStateVerifyIgnore rationale

**Validation**:
- Complete documentation updates
- Test coverage matrix showing 80%+ coverage
- Developer quickstart guide functional

**Estimated Effort**: 1 day

---

### Phase 5: Continuous Improvement (Priority: P3)

**Deliverables**:
1. **Add test coverage CI checks**:
   - Script to calculate drift test coverage percentage
   - Fail CI if coverage drops below 80%
   - Report coverage metrics in CI output

2. **Add test performance monitoring**:
   - Track test execution time
   - Alert if acceptance suite takes >5 minutes
   - Identify slow tests for optimization

3. **Periodic test review process**:
   - Monthly review of test failures
   - Identify flaky tests
   - Update test patterns as BCM API evolves

**Validation**:
- CI checks implemented and passing
- Test performance baseline established
- Review process documented

**Estimated Effort**: 1-2 days (initial setup), ongoing maintenance

---

## TDD Cycle Breakdown

Each phase follows RED-GREEN-REFACTOR cycles:

### Example: Phase 1 - First Drift Test

**RED Phase** (Write failing test):
```go
// internal/provider/resource_cmpart_softwareimage_test.go
func TestAccCMPartSoftwareImage_DriftKernelParameters(t *testing.T) {
    // Test implementation that will fail initially
    // Expected failure: No PreConfig, no drift detection
}
```

**GREEN Phase** (Minimal implementation):
- Add PreConfig function to modify resource via BCM API
- Add ExpectNonEmptyPlan: true
- Verify test passes with drift detection

**REFACTOR Phase** (Improve code quality):
- Extract BCM client creation to test_helpers.go
- Standardize config function naming
- Add comprehensive assertions

### Example: Phase 2 - Multiple Drift Tests

**RED Phase** (Parallel test writing):
- Write 5 drift tests simultaneously for different attributes
- All tests should fail initially (no drift detection)

**GREEN Phase** (Parallel implementation):
- Add PreConfig for each test
- Add config functions for each scenario
- Verify all tests pass

**REFACTOR Phase** (Parallel code improvement):
- Extract common patterns to helper functions
- Standardize test naming
- Add detailed assertions

### Example: Phase 3 - Enhanced CheckDestroy

**RED Phase** (Write enhanced test):
- Create test that verifies detailed logging
- Create test that verifies timeout handling
- Expected: Current CheckDestroy doesn't have these features

**GREEN Phase** (Add minimal features):
- Add logging statements
- Add context with timeout
- Verify tests pass

**REFACTOR Phase** (Improve quality):
- Standardize log format
- Extract timeout constant
- Apply pattern to all CheckDestroy functions

---

## Risk Mitigation

### High-Impact Risks

**Risk 1: Incomplete CheckDestroy verification**
- **Mitigation**: Already implemented - CheckDestroy iterates ALL resources in state
- **Validation**: Enhanced with logging to confirm resources checked

**Risk 2: Race conditions in async operations**
- **Mitigation**: Use exponential backoff in drift tests before verification
- **Example**: After modifying kernel_parameters, wait for BCM to process change
- **Implementation**: Add sleep or polling in PreConfig

**Risk 3: Insufficient attribute coverage**
- **Mitigation**: Track coverage in matrix, fail CI if below 80%
- **Review**: Monthly audit of uncovered attributes

### Medium-Impact Risks

**Risk 4: Test environment cleanup failures**
- **Mitigation**: Already implemented - PreCheck cleanup with retry logic
- **Enhancement**: Use shared helper, standardize retry configuration

**Risk 5: BCM API inconsistencies**
- **Mitigation**: Document known quirks in CLAUDE.md
- **Example**: original_image reset after clone (in ImportStateVerifyIgnore)

**Risk 6: Network timeouts during tests**
- **Mitigation**: Add timeouts to all API calls (10s per call)
- **Fallback**: Retry logic with exponential backoff

### Low-Impact Risks

**Risk 7: Test name collisions**
- **Mitigation**: Already implemented - generateUniqueTestName with timestamp
- **Validation**: Parallel test execution works without conflicts

**Risk 8: Missing ImportStateVerifyIgnore**
- **Mitigation**: Document all expected exceptions in test comments
- **Review**: Check ImportStateVerifyIgnore matches BCM behavior

---

## Success Criteria Mapping

| Success Criterion | Implementation Plan |
|-------------------|---------------------|
| SC-001: 100% CheckDestroy implementations | Phase 3: Enhance existing CheckDestroy functions |
| SC-002: 100% PreCheck cleanup | Phase 3: Standardize existing PreCheck functions |
| SC-003: Drift detection test per resource | Phase 1: Add first drift test per resource |
| SC-004: CheckDestroy verifies ALL resources | Already implemented, Phase 3 adds logging |
| SC-005: Destroy tests pass consistently | Phase 3: Add idempotent destroy tests |
| SC-006: 80% attribute coverage | Phase 2: Systematic drift test creation |
| SC-007: Unique resource names | Already implemented (generateUniqueTestName) |
| SC-008: PreCheck completes <30s | Phase 3: Standardize retry config (max 31s) |
| SC-009: 3+ drift scenarios per resource | Phase 2: 6-10 drift tests per resource |
| SC-010: 3+ destroy scenarios per resource | Phase 3: Add destroy edge case tests |

---

## Acceptance Checklist

This feature is complete when:

- [x] All 2 existing resource types have drift detection infrastructure (Phase 1)
- [ ] All 2 existing resource types have 6+ drift detection tests (Phase 2)
- [ ] All 2 existing resource types have enhanced CheckDestroy with logging (Phase 3)
- [ ] All 2 existing resource types have standardized PreCheck cleanup (Phase 3)
- [ ] Test coverage achieves 80% of non-computed attributes (Phase 2)
- [ ] At least 12 drift tests total across both resources (Phase 2)
- [ ] At least 3 destroy edge case tests (Phase 3)
- [ ] All drift tests use ExpectNonEmptyPlan (Phase 2)
- [ ] All CheckDestroy functions verify via BCM API with logging (Phase 3)
- [ ] All PreCheck functions use exponential backoff (Phase 3)
- [ ] Documentation updated in CLAUDE.md and AGENTS.md (Phase 4)
- [ ] Test coverage matrix created (Phase 4)
- [ ] All acceptance tests pass consistently without manual cleanup (Phases 1-3)
- [ ] Code review completed with focus on test quality (Post-implementation)

---

## Next Steps

1. **Run `/speckit.tasks`** to generate actionable task list from this plan
2. **Execute Phase 1** (Drift Detection Infrastructure) using RED-GREEN-REFACTOR cycles
3. **Review Phase 1 deliverables** before proceeding to Phase 2
4. **Execute remaining phases** in order, validating each before proceeding
5. **Final review** against acceptance checklist before marking feature complete

---

## Appendix: Test Helper Function Reference

### createTestBCMClient

```go
// createTestBCMClient creates an authenticated BCM client for test use
// Environment variables: BCM_ENDPOINT, BCM_USERNAME, BCM_PASSWORD
// Errors: t.Fatalf if credentials missing or login fails
func createTestBCMClient(t *testing.T) *BCMClient {
    endpoint := os.Getenv("BCM_ENDPOINT")
    username := os.Getenv("BCM_USERNAME")
    password := os.Getenv("BCM_PASSWORD")

    if endpoint == "" || username == "" || password == "" {
        t.Fatalf("BCM credentials not set (BCM_ENDPOINT, BCM_USERNAME, BCM_PASSWORD)")
    }

    client, err := NewBCMClient(context.Background(), endpoint, username, password, true, 30)
    if err != nil {
        t.Fatalf("Failed to create BCM client: %v", err)
    }

    return client
}
```

### verifyResourceDeleted

```go
// verifyResourceDeleted polls BCM API to verify resource deletion with exponential backoff
// Returns (true, nil) if resource is deleted
// Returns (false, nil) if resource still exists after maxRetries
// Returns (false, error) on API errors
func verifyResourceDeleted(ctx context.Context, client *BCMClient, service, method, identifier string, maxRetries int) (bool, error) {
    waitTime := 1 * time.Second

    for retry := 0; retry < maxRetries; retry++ {
        time.Sleep(waitTime)

        body, err := client.CallJSONRPC(ctx, service, method, identifier)

        // Check if resource is gone
        if err != nil || len(body) == 0 {
            return true, nil // Deleted
        }

        var data map[string]interface{}
        if json.Unmarshal(body, &data) == nil && len(data) == 0 {
            return true, nil // Deleted
        }

        // Resource still exists, wait longer
        waitTime *= 2
    }

    return false, nil // Not deleted after retries
}
```

### Drift Test Config Pattern

```go
// Test configuration function for drift detection
// Returns HCL config with specified attribute value
func testAccCMPartSoftwareImageResourceConfig_DriftKernel(name, path, kernelParams string) string {
    return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmpart_softwareimages" "default" {}

locals {
  default_image = [for img in data.bcm_cmpart_softwareimages.default.images : img if img.name == "default-image"][0]
}

resource "bcm_cmpart_softwareimage" "test" {
  name               = %[4]q
  path               = %[5]q
  kernel_parameters  = %[6]q
  original_image     = local.default_image.uuid
}
`,
        os.Getenv("BCM_ENDPOINT"),
        os.Getenv("BCM_USERNAME"),
        os.Getenv("BCM_PASSWORD"),
        name,
        path,
        kernelParams,
    )
}
```

---

**End of Implementation Plan**
