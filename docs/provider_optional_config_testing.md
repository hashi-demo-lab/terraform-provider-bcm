# Provider Optional Configuration Testing

This document describes the test scenarios for optional provider configuration fields in the BCM Terraform Provider.

## Overview

The BCM provider has two optional configuration fields:

1. **`insecure_skip_verify`** (bool) - Skip TLS certificate verification
2. **`timeout`** (int64) - API timeout in seconds

These fields are defined in `/workspace/internal/provider/provider.go` and tested in:
- `/workspace/internal/provider/provider_optional_config_unit_test.go` - Unit tests (no BCM required)
- `/workspace/internal/provider/provider_optional_config_test.go` - Acceptance tests (requires BCM)

## Implementation Details

### insecure_skip_verify

**Location**: `provider.go` lines 130-133

```go
insecureSkipVerify := false
if !data.InsecureSkipVerify.IsNull() {
    insecureSkipVerify = data.InsecureSkipVerify.ValueBool()
}
```

**Behavior**:
- **Default**: `false` (when not set)
- **Type**: `types.Bool` (schema) → `bool` (client parameter)
- **Passed to**: `NewBCMClient()` line 146
- **Client usage**: `bcm_client.go` line 54 - Controls `tls.Config.InsecureSkipVerify`

### timeout

**Location**: `provider.go` lines 135-138, 147

```go
timeout := int64(30) // Default 30 seconds
if !data.Timeout.IsNull() {
    timeout = data.Timeout.ValueInt64()
}
// ...
client, err := NewBCMClient(ctx, endpoint, username, password, insecureSkipVerify, int(timeout))
```

**Behavior**:
- **Default**: `30` seconds (when not set)
- **Type**: `types.Int64` (schema) → `int64` (provider) → `int` (client parameter)
- **Conversion**: Line 147 converts `int64` to `int` for `NewBCMClient()`
- **Client usage**: `bcm_client.go` line 61 - Sets `http.Client.Timeout`

## Test Coverage

### Unit Tests (`provider_optional_config_unit_test.go`)

These tests run without requiring a BCM server connection. They verify the conversion logic in `provider.go`.

#### TestProviderConfig_ParameterTypeConversion

Tests all combinations of null/non-null values and type conversions:

| Test Case | insecure_skip_verify | timeout | Expected Insecure | Expected Timeout |
|-----------|---------------------|---------|------------------|-----------------|
| BothDefaults_NotSet | null | null | false | 30 |
| InsecureTrue_TimeoutDefault | true | null | true | 30 |
| InsecureFalse_Timeout60 | false | 60 | false | 60 |
| BothExplicitlySet | true | 120 | true | 120 |
| Int64ToIntConversion_SmallValue | null | 5 | false | 5 |
| Int64ToIntConversion_LargeValue | null | 2147483647 | false | 2147483647 |
| InsecureFalseExplicit_TimeoutDefault | false | null | false | 30 |
| InsecureDefault_TimeoutExplicit30 | null | 30 | false | 30 |

**Run Tests**:
```bash
go test -v ./internal/provider/ -run "TestProviderConfig_ParameterTypeConversion"
```

**Expected Output**: All 8 sub-tests pass, showing the conversion logic works correctly.

#### TestProviderConfig_DefaultValues

Verifies that default values match documentation:
- `insecure_skip_verify`: `false`
- `timeout`: `30` seconds

#### TestProviderConfig_NewBCMClientParameterSignature

Documents the parameter types expected by `NewBCMClient()`:
- `insecureSkipVerify`: `bool` (NOT `types.Bool`)
- `timeout`: `int` (NOT `int64` or `types.Int64`)

#### TestProviderConfig_EdgeCases

Documents edge case behavior:
- **Zero timeout**: Behavior undefined (http.Client may use no timeout)
- **Negative timeout**: Behavior undefined (time.Duration may accept negative)
- **Very large timeout**: Technically valid but impractical

#### TestProviderConfig_BoolHandling

Verifies boolean field handling for all cases:
- Not set (null) → uses default `false`
- Explicitly `true`
- Explicitly `false`

### Acceptance Tests (`provider_optional_config_test.go`)

These tests require a BCM server connection (set `TF_ACC=1` and BCM environment variables).

#### insecure_skip_verify Tests

| Test | Config | Verifies |
|------|--------|----------|
| TestAccProviderConfig_InsecureSkipVerify_Default | Not set | Defaults to false, data source works |
| TestAccProviderConfig_InsecureSkipVerify_ExplicitTrue | `true` | TLS verification skipped, data source works |
| TestAccProviderConfig_InsecureSkipVerify_ExplicitFalse | `false` | Explicit false same as default |

#### timeout Tests

| Test | Config | Verifies |
|------|--------|----------|
| TestAccProviderConfig_Timeout_Default | Not set | Defaults to 30s, data source works |
| TestAccProviderConfig_Timeout_Custom60Seconds | `60` | Custom timeout applied |
| TestAccProviderConfig_Timeout_Custom120Seconds | `120` | Extended timeout applied |
| TestAccProviderConfig_Timeout_Short5Seconds | `5` | Short timeout applied |
| TestAccProviderConfig_Timeout_Int64ToIntConversion | `2147483647` | int64→int conversion works |

#### Combined Tests

| Test | Config | Verifies |
|------|--------|----------|
| TestAccProviderConfig_CombinedOptionalFields | Both set | Multiple optional fields work together |
| TestAccProviderConfig_CombinedDefaults | Both not set | Both defaults applied correctly |

#### Edge Case Tests

| Test | Config | Status |
|------|--------|--------|
| TestAccProviderConfig_Timeout_ZeroValue | `0` | Skipped - undefined behavior |
| TestAccProviderConfig_Timeout_NegativeValue | `-1` | Skipped - undefined behavior |

**Run Acceptance Tests**:
```bash
export TF_ACC=1
export BCM_ENDPOINT="https://172.21.15.254:8081"
export BCM_USERNAME="root"
export BCM_PASSWORD="Hashicorp123!"

go test -v -timeout 120m ./internal/provider/ -run "TestAccProviderConfig_"
```

## Test Configuration Helpers

All acceptance test configurations use these helper functions:

- `testAccProviderConfigInsecureSkipVerifyNotSet()` - insecure_skip_verify not set
- `testAccProviderConfigInsecureSkipVerifyTrue()` - insecure_skip_verify = true
- `testAccProviderConfigInsecureSkipVerifyFalse()` - insecure_skip_verify = false
- `testAccProviderConfigTimeoutNotSet()` - timeout not set
- `testAccProviderConfigTimeout(timeout int64)` - timeout = custom value
- `testAccProviderConfigCombined(insecure bool, timeout int64)` - both fields set
- `testAccProviderConfigBothOptionalNotSet()` - neither field set

All helpers use environment variables for credentials:
```go
os.Getenv("BCM_ENDPOINT")
os.Getenv("BCM_USERNAME")
os.Getenv("BCM_PASSWORD")
```

## Type Conversion Flow

```
Schema Definition (provider.go:63-70)
  ↓
types.Bool / types.Int64
  ↓
Provider Configure (provider.go:130-138)
  ↓
bool / int64 (with defaults)
  ↓
NewBCMClient Call (provider.go:147)
  ↓
bool / int (int64 → int conversion)
  ↓
BCM Client (bcm_client.go:44)
  ↓
TLS Config / HTTP Client Timeout
```

## Test Results

### Unit Tests

```
$ go test -v ./internal/provider/ -run "TestProviderConfig_"

=== RUN   TestProviderConfig_ParameterTypeConversion
--- PASS: TestProviderConfig_ParameterTypeConversion (0.00s)
    (8 sub-tests passed)

=== RUN   TestProviderConfig_DefaultValues
--- PASS: TestProviderConfig_DefaultValues (0.00s)

=== RUN   TestProviderConfig_NewBCMClientParameterSignature
--- PASS: TestProviderConfig_NewBCMClientParameterSignature (0.00s)

=== RUN   TestProviderConfig_EdgeCases
--- PASS: TestProviderConfig_EdgeCases (0.00s)
    (3 sub-tests passed)

=== RUN   TestProviderConfig_BoolHandling
--- PASS: TestProviderConfig_BoolHandling (0.00s)
    (3 sub-tests passed)

PASS
ok      github.com/hashi-demo-lab/terraform-provider-bcm/internal/provider    0.008s
```

### Acceptance Tests

Acceptance tests require:
1. `TF_ACC=1` environment variable
2. Access to BCM server at `BCM_ENDPOINT`
3. Valid credentials in `BCM_USERNAME` and `BCM_PASSWORD`

Expected behavior:
- All tests create a `bcm_cmpart_softwareimages` data source
- Each test verifies the data source works (proves provider configured correctly)
- Tests use `statecheck.ExpectKnownValue()` for type-safe assertions

## Key Insights

1. **Default Values Matter**: When a field is not set (`IsNull()` returns `true`), the default value is used. This is not the same as explicitly setting the field to its default value.

2. **Type Conversion Required**: The schema uses framework types (`types.Bool`, `types.Int64`), but `NewBCMClient()` expects Go primitive types (`bool`, `int`). The provider must convert these types.

3. **int64 to int Conversion**: The `timeout` field is defined as `types.Int64` in the schema but must be converted to `int` for `NewBCMClient()`. This conversion happens at line 147.

4. **No Validation**: The provider currently does not validate `timeout` values. Zero and negative values are passed through to `NewBCMClient()` and may have undefined behavior.

5. **Boolean Null Handling**: For boolean fields, there's a distinction between:
   - Not set (null) → uses default
   - Explicitly false → same result as default, but explicit

## Related Documentation

- **Provider Implementation**: `/workspace/internal/provider/provider.go`
- **Client Implementation**: `/workspace/internal/provider/bcm_client.go`
- **Main Documentation**: `/workspace/CLAUDE.md`
- **TDD Guidelines**: `/workspace/AGENTS.md`

## Future Enhancements

Potential improvements to consider:

1. **Timeout Validation**: Add validation to reject zero/negative timeout values
2. **Timeout Range**: Document recommended timeout ranges based on typical API response times
3. **Error Messages**: Improve error messages when TLS verification fails (currently generic)
4. **Timeout Units**: Consider supporting different time units (not just seconds)

## References

- [Terraform Plugin Framework - Provider Schema](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/schemas)
- [Terraform Plugin Testing](https://developer.hashicorp.com/terraform/plugin/testing)
- [Go http.Client Timeout](https://pkg.go.dev/net/http#Client)
- [Go tls.Config](https://pkg.go.dev/crypto/tls#Config)
