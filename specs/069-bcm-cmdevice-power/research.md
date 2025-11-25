# Phase 0 Research: BCM CMDevice Power Action

**Feature**: bcm_cmdevice_power Terraform Action
**Branch**: `069-bcm-cmdevice-power`
**Date**: 2025-11-26
**Status**: Complete

## Research Tasks

### R1: Terraform Action Interface Requirements

**Task**: Research terraform-plugin-framework action package interfaces for v1.16.0+

**Decision**: Use terraform-plugin-framework v1.16.1 action package with full interface implementation

**Rationale**:
- terraform-plugin-framework v1.16.0 introduced the `action` package for Terraform 1.14+ support
- The provider already uses terraform-plugin-framework v1.16.1
- Actions are designed specifically for imperative, side-effect operations like power control

**Key Interfaces Required**:

1. **`action.Action`** (required) - Base interface with three methods:
   - `Metadata(context.Context, MetadataRequest, *MetadataResponse)` - Returns action type name
   - `Schema(context.Context, SchemaRequest, *SchemaResponse)` - Returns action schema
   - `Invoke(context.Context, InvokeRequest, *InvokeResponse)` - Executes action logic

2. **`action.ActionWithConfigure`** (required for BCM client) - Extends Action:
   - `Configure(context.Context, ConfigureRequest, *ConfigureResponse)` - Receives BCM client from provider

3. **`action.ActionWithValidateConfig`** (recommended) - Extends Action:
   - `ValidateConfig(context.Context, ValidateConfigRequest, *ValidateConfigResponse)` - Validates power_action values

4. **`provider.ProviderWithActions`** (required on provider) - Registers actions:
   - `Actions(context.Context) []func() action.Action` - Returns list of action constructors

**Alternatives Considered**:
- Null resource with local-exec provisioner - Rejected: Not declarative, poor UX, no lifecycle integration
- Custom resource with triggers - Rejected: Forces resource lifecycle semantics on non-stateful operation
- External data source + webhook - Rejected: Complex, requires external infrastructure

**Sources**:
- [Writing a Terraform Action](https://danielmschmidt.de/posts/2025-09-26-writing-a-terraform-action/)
- [Terraform Plugin Framework Releases](https://github.com/hashicorp/terraform-plugin-framework/releases)
- [action package - pkg.go.dev](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-framework/action)

---

### R2: BCM Power Operation API Methods

**Task**: Verify BCM cmdevice service power operation methods and their signatures

**Decision**: Use cmdevice service with powerOn, powerOff, reboot, powerCycle methods accepting device identifier

**Findings from BCM API Documentation**:

The `cmdevice` service exposes power control operations:

| Method | Description | Documented |
|--------|-------------|------------|
| `reboot` | Reboot a node | Yes (verified) |
| `powerOn` | Power on node | Likely exists (per API patterns) |
| `powerOff` | Power off node | Likely exists (per API patterns) |
| `powerCycle` | Power cycle node | Likely exists (per API patterns) |

**API Request Pattern** (from documented `reboot` method):
```json
{
  "service": "cmdevice",
  "call": "reboot",
  "args": ["node001"]
}
```

**Argument Options**:
- Hostname (string): e.g., "node001", "master"
- UUID (string): e.g., "2870c0b0-6fda-4026-9b8f-28be4c372fee"

**Phase 0 Verification Required**:
Before implementation, these API methods MUST be verified:
1. `powerOn` - Verify method exists and accepts device UUID/hostname
2. `powerOff` - Verify method exists and accepts device UUID/hostname
3. `powerCycle` - Verify method exists and accepts device UUID/hostname
4. Document response format for all four methods
5. Document error response format for invalid device

**Rationale**: The BCM API documentation confirms `reboot` exists. Based on standard cluster management patterns and the documented powerControl field on devices, the other power methods almost certainly exist but need verification.

**Alternatives Considered**:
- Using BMC/IPMI directly - Rejected: BCM abstracts power control, consistent with other provider patterns
- Using SSH command execution - Rejected: Requires SSH credentials, bypasses BCM audit logging

---

### R3: Device Power State Query

**Task**: Research how to query current device power state for wait_for_completion feature

**Decision**: Use device status/state field polling if available, otherwise timeout-based completion

**Findings**:
The BCM API documentation shows devices have power management fields:
- `powerControl` - Type of power control ("none", "ipmi", "redfish", "pdu")
- `powerDistributionUnits` - Associated PDUs
- `customPowerScript` - Custom power control script

**Potential State Query Methods**:
1. `getNode` with device identifier - May include power status in response
2. `powerStatus` - Likely exists based on API patterns (needs verification)

**Implementation Strategy**:
1. **Default (wait_for_completion = false)**: Return immediately after API call succeeds
2. **With wait (wait_for_completion = true)**:
   - Poll device status using `getNode` or `powerStatus`
   - Check for power state change
   - Timeout after configurable duration (default 5m)
   - Report progress during wait

**Alternatives Considered**:
- Websocket subscription for state changes - Rejected: BCM may not support, adds complexity
- Fixed delay after command - Rejected: Unreliable, wastes time or fails prematurely

---

### R4: Action Schema Package

**Task**: Research action/schema package differences from resource/schema

**Decision**: Use `github.com/hashicorp/terraform-plugin-framework/action/schema` for action schema definition

**Rationale**:
- Actions use a specialized schema package separate from resources
- This allows Terraform to evolve action capabilities independently
- Schema supports standard attribute types (string, bool) needed for power action

**Key Differences from Resource Schema**:
- No Computed attributes (actions don't persist state)
- No plan modifiers for computed values
- Focus on input validation only

**Schema Structure for Power Action**:
```go
import "github.com/hashicorp/terraform-plugin-framework/action/schema"

schema.Schema{
    MarkdownDescription: "Execute power operations on BCM devices",
    Attributes: map[string]schema.Attribute{
        "device_id": schema.StringAttribute{
            Required: true,
            Description: "BCM device UUID or hostname",
        },
        "power_action": schema.StringAttribute{
            Required: true,
            Description: "Power operation: power_on, power_off, reboot, power_cycle",
            Validators: []validator.String{
                stringvalidator.OneOf("power_on", "power_off", "reboot", "power_cycle"),
            },
        },
        "wait_for_completion": schema.BoolAttribute{
            Optional: true,
            Description: "Wait for power state change (default: false)",
        },
        "timeout": schema.StringAttribute{
            Optional: true,
            Description: "Timeout duration when waiting (default: 5m)",
        },
    },
}
```

---

### R5: Provider Registration Pattern

**Task**: Determine how to register actions with the BCM provider

**Decision**: Implement `provider.ProviderWithActions` interface on BCMProvider

**Current Provider Interfaces** (from /workspace/internal/provider/provider.go):
```go
var _ provider.Provider = &BCMProvider{}
var _ provider.ProviderWithFunctions = &BCMProvider{}
var _ provider.ProviderWithEphemeralResources = &BCMProvider{}
```

**Required Addition**:
```go
var _ provider.ProviderWithActions = &BCMProvider{}

func (p *BCMProvider) Actions(ctx context.Context) []func() action.Action {
    return []func() action.Action{
        NewCMDevicePowerAction,
    }
}
```

**Configure Integration**:
The provider's `Configure` method already sets:
- `resp.DataSourceData = client`
- `resp.ResourceData = client`

Need to add:
- `resp.ActionData = client`

This makes the BCM client available to actions via `ActionWithConfigure`.

---

### R6: Testing Strategy for Terraform 1.14 Beta

**Task**: Research testing approaches given Terraform 1.14 is in beta

**Decision**: Implement phased testing strategy

**Rationale**:
- Terraform 1.14 is currently in beta (as of November 2025)
- The terraform-plugin-testing framework may not fully support action testing yet
- Unit tests provide immediate coverage, acceptance tests when TF 1.14 GA

**Phased Testing Approach**:

**Phase 1: Unit Tests (Immediate)**
- Test schema validation logic
- Test action metadata
- Mock BCM client for Invoke logic testing
- Test power_action value validation
- Test timeout parsing

**Phase 2: Manual Integration Tests (During Beta)**
- Test against real BCM cluster with Terraform 1.14 beta
- Document expected behavior
- Create example configurations

**Phase 3: Acceptance Tests (When TF 1.14 GA)**
- Full terraform-plugin-testing acceptance tests
- Test all four power operations
- Test lifecycle trigger integration
- Test wait_for_completion behavior
- Test error handling

**Test File Structure**:
```
internal/provider/
  action_cmdevice_power.go          # Action implementation
  action_cmdevice_power_test.go     # Unit tests (immediate)
  action_cmdevice_power_acc_test.go # Acceptance tests (TF 1.14 GA)
```

**Alternatives Considered**:
- Wait for TF 1.14 GA before implementing - Rejected: Delays feature, beta testing provides value
- Skip testing entirely - Rejected: Violates TDD principles, risky

---

### R7: Progress Reporting Pattern

**Task**: Research progress reporting during action invocation

**Decision**: Use `InvokeResponse.SendProgress()` for status updates

**Rationale**:
- Actions support progress reporting via `resp.SendProgress(action.InvokeProgressEvent{Message: "..."})`
- Important for long-running operations like wait_for_completion
- Provides user feedback during execution

**Implementation Pattern**:
```go
func (a *CMDevicePowerAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
    // Report start
    resp.SendProgress(action.InvokeProgressEvent{
        Message: "Initiating power operation...",
    })

    // Execute API call
    err := a.client.CallJSONRPC(ctx, "cmdevice", powerMethod, deviceID)

    // Report progress during wait
    if waitForCompletion {
        resp.SendProgress(action.InvokeProgressEvent{
            Message: "Waiting for power state change...",
        })
        // Poll loop with progress updates
    }

    // Report completion
    resp.SendProgress(action.InvokeProgressEvent{
        Message: fmt.Sprintf("Power operation '%s' completed", powerAction),
    })
}
```

---

### R8: Error Handling Patterns

**Task**: Research error handling for action failures

**Decision**: Use Diagnostics for structured error reporting

**Rationale**:
- Actions use the same `diag.Diagnostics` pattern as resources
- Provides consistent error messaging
- Supports warnings for non-fatal issues

**Error Categories**:

1. **Configuration Errors** (fail at plan time):
   - Invalid power_action value
   - Invalid device_id format
   - Invalid timeout format

2. **API Errors** (fail at invoke time):
   - Device not found
   - BMC/IPMI unreachable
   - Authentication failure
   - Network timeout

3. **Wait Errors** (fail during wait):
   - Timeout exceeded
   - Power state change failed

**Implementation Pattern**:
```go
// Configuration validation
if !validPowerAction(plan.PowerAction) {
    resp.Diagnostics.AddError(
        "Invalid Power Action",
        fmt.Sprintf("power_action must be one of: power_on, power_off, reboot, power_cycle. Got: %s",
            plan.PowerAction.ValueString()),
    )
    return
}

// API error
if err != nil {
    resp.Diagnostics.AddError(
        "Power Operation Failed",
        fmt.Sprintf("Failed to execute %s on device %s: %s",
            powerAction, deviceID, err.Error()),
    )
    return
}

// Timeout warning (operation sent but completion not confirmed)
if timedOut {
    resp.Diagnostics.AddWarning(
        "Wait Timeout",
        fmt.Sprintf("Power command sent successfully but confirmation timed out after %s", timeout),
    )
}
```

---

## Summary of Decisions

| Area | Decision | Confidence |
|------|----------|------------|
| Framework | terraform-plugin-framework action package v1.16.1 | High |
| Interfaces | Action, ActionWithConfigure, ActionWithValidateConfig | High |
| BCM API | cmdevice service: powerOn, powerOff, reboot, powerCycle | Medium (needs verification) |
| Device ID | Accept both UUID and hostname | High |
| Wait Feature | Poll-based with configurable timeout | Medium |
| Testing | Phased: Unit (now), Acceptance (TF 1.14 GA) | High |
| Progress | SendProgress for status updates | High |
| Errors | Diagnostics with structured messages | High |

## Outstanding Verification Tasks

These must be completed before implementation begins:

1. **[P0-001]** Verify `powerOn` method exists via BCM API test
2. **[P0-002]** Verify `powerOff` method exists via BCM API test
3. **[P0-003]** Verify `powerCycle` method exists via BCM API test
4. **[P0-004]** Document API response format for all power methods
5. **[P0-005]** Verify error response format for invalid device
6. **[P0-006]** Identify power state query method (for wait_for_completion)
7. **[P0-007]** Verify BCM client can be passed to actions via ActionData
