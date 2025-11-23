# BCM Partition Resource Test Pattern Investigation

## Problem Statement

Partition resource tests are failing due to:
1. **Timezone validation errors** - "America/Los_Angeles" is being rejected
2. **Name constraints** - Tests try to create unique partition names, but BCM requires "base"
3. **Long unique names rejected** - `generateUniqueTestName()` produces names like "test-partition-20251123-150032" which exceed BCM's limits

## Test Code Analysis

### Current Test Pattern (resource_cmpart_partition_test.go)

**TestAccCMPartPartition_Basic (lines 26-89):**
```go
partitionName := "base" // BCM constraint: partition name must be "base"
```
- Uses hardcoded "base" name (works!)
- Uses `timezone_settings = "America/Los_Angeles"`
- Tests Create, Read, Import, Idempotency

**TestAccCMPartPartition_Update (lines 92-179):**
```go
partitionName := generateUniqueTestName("test-partition")
```
- Uses unique name (FAILS - BCM rejects)
- Uses `timezone_settings = "America/Los_Angeles"`

**All Other Tests:**
- All use `generateUniqueTestName("test-partition")`
- All use `timezone_settings = "America/Los_Angeles"`

### Working Validation Test (lines 440-452)

```go
func TestAccCMPartPartition_ValidationErrors(t *testing.T) {
    resource.Test(t, resource.TestCase{
        Steps: []resource.TestStep{
            {
                Config:      testAccPartitionConfigBasic("", "Test Cluster"),
                ExpectError: regexp.MustCompile(`Attribute name string length must be between 1 and`),
            },
        },
    })
}
```

**Why it works:**
- Tests **validation errors** (ExpectError pattern)
- Never actually creates a resource
- Only checks schema validation, not API constraints

## Key Findings

### 1. Timezone Issue
All test configs use: `timezone_settings = "America/Los_Angeles"`

**Hypothesis:** BCM may require specific timezone format or different value.
- Need to check BCM API for valid timezone values
- May need to use "UTC" or different IANA timezone
- Schema validator may accept format but BCM API rejects

### 2. Partition Name Constraints
Tests show TWO patterns:
1. `partitionName := "base"` - **WORKS** (TestAccCMPartPartition_Basic)
2. `partitionName := generateUniqueTestName("test-partition")` - **FAILS** (all other tests)

**Implications:**
- BCM may only allow ONE partition named "base" (like default partition)
- "base" partition may already exist in test environment
- Tests trying to CREATE "base" will fail if it exists
- Tests using unique names fail because BCM rejects non-"base" names

### 3. Test Strategy Patterns from Other Resources

**Pattern 1: Work with Existing Resources (Categories)**
From `resource_cmdevice_category_test.go:69-100`:
```go
func testAccCMDeviceCategoryPreCheck(t *testing.T, names ...string) {
    client := createTestBCMClient(t)

    // Clean up any leftover test categories
    for _, name := range names {
        body, err := client.CallJSONRPC(context.Background(), "cmdevice", "getCategory", name)
        if err == nil {
            var categoryData map[string]interface{}
            if json.Unmarshal(body, &categoryData) == nil {
                if uuid, ok := categoryData["uuid"].(string); ok && uuid != "" {
                    // Category exists, try to delete it
                    client.CallJSONRPC(context.Background(), "cmdevice", "removeCategory", uuid, true)
                    deleted := verifyResourceDeleted(context.Background(), client, "cmdevice", "getCategory", name, 5)
                }
            }
        }
    }
}
```

**Pattern 2: Import-Only Tests (No Create)**
Could test Import/Read/Update on existing "base" partition without Create.

**Pattern 3: Skip Problematic Tests**
Use `t.Skip()` for tests that require creating partitions.

## Recommended Test Strategy

### Option A: Import-Only Pattern (RECOMMENDED)

Test the existing "base" partition without creating it:

```go
func TestAccCMPartPartition_ImportAndUpdate(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            // Step 1: Import existing "base" partition
            {
                Config: testAccPartitionConfigBasic("base", "Existing Cluster"),
                ResourceName:      "bcm_cmpart_partition.test",
                ImportState:       true,
                ImportStateId:     getBasePartitionUUID(t),
                ImportStateVerify: true,
            },
            // Step 2: Update imported partition
            {
                Config: testAccPartitionConfigBasic("base", "Updated Cluster"),
                ConfigStateChecks: []statecheck.StateCheck{
                    statecheck.ExpectKnownValue(
                        "bcm_cmpart_partition.test",
                        tfjsonpath.New("cluster_name"),
                        knownvalue.StringExact("Updated Cluster"),
                    ),
                },
            },
        },
    })
}
```

**Pros:**
- Works with existing "base" partition
- Tests Read, Import, Update, Delete operations
- No creation constraints
- Real-world scenario (importing existing infrastructure)

**Cons:**
- Doesn't test Create operation
- Requires "base" partition to exist

### Option B: Conditional Create with Cleanup

Pre-check to delete "base" partition if it exists, then create fresh:

```go
func testAccPartitionPreCheck(t *testing.T) {
    client := createTestBCMClient(t)

    // Try to delete existing "base" partition
    body, err := client.CallJSONRPC(context.Background(), "cmpart", "getPartition", "base")
    if err == nil {
        var data map[string]interface{}
        if json.Unmarshal(body, &data) == nil {
            if uuid, ok := data["uuid"].(string); ok && uuid != "" {
                client.CallJSONRPC(context.Background(), "cmpart", "removePartition", uuid, true)
                verifyResourceDeleted(context.Background(), client, "cmpart", "getPartition", "base", 5)
            }
        }
    }
}
```

**Pros:**
- Can test full Create operation
- Clean test environment

**Cons:**
- Destructive (deletes production "base" partition)
- May fail if partition has dependencies (nodes, etc.)
- Risky for shared test environments

### Option C: Use Valid Timezone

Research BCM's accepted timezone values and use one that works:

**Possible values to try:**
- `"UTC"`
- `"America/New_York"`
- `"Europe/London"`
- `"US/Pacific"` (instead of "America/Los_Angeles")

**Investigation needed:**
- Check BCM API documentation for timezone field
- Query existing "base" partition to see what timezone it uses
- Try different IANA timezone database values

### Option D: Validation-Only Tests

Focus on testing validation without actual API calls:

```go
func TestAccCMPartPartition_ValidationErrors(t *testing.T) {
    resource.Test(t, resource.TestCase{
        Steps: []resource.TestStep{
            // Test empty name
            {
                Config:      testAccPartitionConfigBasic("", "Cluster"),
                ExpectError: regexp.MustCompile(`string length must be between 1 and`),
            },
            // Test invalid timezone format
            {
                Config:      testAccPartitionConfigInvalidTimezone("base", "Invalid/Timezone"),
                ExpectError: regexp.MustCompile(`Invalid timezone`),
            },
        },
    })
}
```

**Pros:**
- No BCM API constraints
- Fast execution
- Tests schema validation

**Cons:**
- Doesn't test actual CRUD operations
- Limited value for acceptance tests

## Next Steps

1. **Investigate timezone constraint:**
   - Query existing "base" partition to see its timezone value
   - Try different timezone formats in tests
   - Update schema validator if needed

2. **Determine partition naming constraints:**
   - Can BCM support multiple partitions?
   - Is "base" a special/required name?
   - What name length/format limits exist?

3. **Choose test strategy:**
   - Implement Option A (Import-Only) for immediate testing
   - Add Option C (Valid Timezone) if timezone is the only issue
   - Consider Option D (Validation-Only) as supplementary tests

4. **Document findings:**
   - Update CLAUDE.md with partition constraints
   - Add comments to test files explaining limitations
   - Create helper functions for partition testing patterns

## Recommended Implementation

**Immediate (High Priority):**
1. Query existing "base" partition to get working timezone value
2. Create `TestAccCMPartPartition_ImportOnly` following Option A
3. Update validation tests following Option D

**Short-term (Medium Priority):**
1. Research BCM partition naming constraints
2. Add timezone validation to schema if needed
3. Document constraints in CLAUDE.md

**Long-term (Low Priority):**
1. Add multiple partition support if BCM allows
2. Create comprehensive partition testing helpers
3. Add pre-check cleanup if safe
