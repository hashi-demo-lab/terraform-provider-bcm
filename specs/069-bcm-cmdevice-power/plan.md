# Implementation Plan: BCM CMDevice Power Action

**Branch**: `069-bcm-cmdevice-power` | **Date**: 2025-11-26 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/069-bcm-cmdevice-power/spec.md`

## Summary

Implement the first Terraform Action in the BCM provider for device power operations (power_on, power_off, reboot, power_cycle). Actions are a new Terraform 1.14+ feature designed for imperative, non-CRUD operations that do not maintain state. This implementation adds `provider.ProviderWithActions` interface support and creates `bcm_cmdevice_power` action using the terraform-plugin-framework v1.16.1 action package.

## Technical Context

**Language/Version**: Go 1.24.0
**Primary Dependencies**: terraform-plugin-framework v1.16.1, terraform-plugin-framework-validators
**Storage**: N/A (actions do not persist state)
**Testing**: go test (unit tests immediate, acceptance tests when TF 1.14 GA)
**Target Platform**: Terraform 1.14+ (currently beta)
**Project Type**: Terraform Provider (single)
**Performance Goals**: Sub-second API call execution, configurable timeout for wait_for_completion
**Constraints**: Terraform 1.14 beta limits acceptance testing; BCM API methods require Phase 0 verification
**Scale/Scope**: Single action type supporting 4 power operations on individual devices

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| TDD Required | PASS | Unit tests immediate, acceptance tests deferred to TF 1.14 GA |
| Single Responsibility | PASS | Action performs one task: device power control |
| No State Management | PASS | Actions don't persist state by design |
| Provider Pattern Compliance | PASS | Follows terraform-plugin-framework action patterns |
| API Verification | PENDING | Phase 0 must verify powerOn/powerOff/powerCycle methods |

## Project Structure

### Documentation (this feature)

```text
specs/069-bcm-cmdevice-power/
├── plan.md              # This file
├── research.md          # Phase 0 output - API and framework research
├── data-model.md        # Phase 1 output - configuration and internal models
├── quickstart.md        # Phase 1 output - developer guide
├── contracts/           # Phase 1 output - API and schema contracts
│   ├── bcm-power-api.md      # BCM JSON-RPC API contract
│   └── action-schema.md      # Terraform action schema contract
└── tasks.md             # Phase 2 output (/speckit.tasks command)
```

### Source Code (repository root)

```text
internal/provider/
├── provider.go                      # MODIFY: Add ProviderWithActions interface
├── action_cmdevice_power.go         # NEW: Action implementation
├── action_cmdevice_power_test.go    # NEW: Unit tests
└── action_cmdevice_power_acc_test.go # NEW: Acceptance tests (TF 1.14 GA)

examples/actions/bcm_cmdevice_power/
├── main.tf              # Example configurations
├── variables.tf         # Variable definitions
└── README.md            # Usage documentation
```

**Structure Decision**: Single project layout following existing provider patterns. New `action_*.go` naming convention for actions parallel to existing `resource_*.go` and `data_source_*.go` patterns.

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| New primitive type (action) | Terraform 1.14 introduces actions for imperative operations | Resource/data source patterns force CRUD semantics on non-stateful operations |
| Deferred acceptance tests | TF 1.14 beta status limits terraform-plugin-testing support | Full testing requires stable TF release |

---

## Phase 0: Research Completed

**Status**: Complete - See [research.md](./research.md)

### Key Decisions

1. **Framework**: Use terraform-plugin-framework v1.16.1 action package
2. **Interfaces**: Implement `action.Action`, `action.ActionWithConfigure`, `action.ActionWithValidateConfig`
3. **Provider**: Add `provider.ProviderWithActions` interface to BCMProvider
4. **API Methods**: Use cmdevice service with powerOn, powerOff, reboot, powerCycle (verification required)
5. **Testing**: Phased approach - unit tests immediate, acceptance tests when TF 1.14 GA

### Outstanding Verification Tasks

| ID | Task | Status |
|----|------|--------|
| P0-001 | Verify `powerOn` method exists | Pending |
| P0-002 | Verify `powerOff` method exists | Pending |
| P0-003 | Verify `powerCycle` method exists | Pending |
| P0-004 | Document API response format | Pending |
| P0-005 | Verify error response format | Pending |
| P0-006 | Identify power state query method | Pending |
| P0-007 | Verify ActionData in provider Configure | Pending |

---

## Phase 1: Design Completed

**Status**: Complete - See artifacts below

### Generated Artifacts

| Artifact | Location | Description |
|----------|----------|-------------|
| data-model.md | [data-model.md](./data-model.md) | Configuration model, internal structures, no state persistence |
| contracts/bcm-power-api.md | [contracts/bcm-power-api.md](./contracts/bcm-power-api.md) | BCM JSON-RPC API contract |
| contracts/action-schema.md | [contracts/action-schema.md](./contracts/action-schema.md) | Terraform action schema definition |
| quickstart.md | [quickstart.md](./quickstart.md) | Developer implementation guide |

### Action Schema Summary

```hcl
action "bcm_cmdevice_power" "example" {
  device_id           = "uuid-or-hostname"  # Required
  power_action        = "power_on"          # Required: power_on|power_off|reboot|power_cycle
  wait_for_completion = false               # Optional, default: false
  timeout             = "5m"                # Optional, default: 5m
}
```

### API Method Mapping

| Terraform power_action | BCM API Method |
|------------------------|----------------|
| `power_on` | `powerOn` |
| `power_off` | `powerOff` |
| `reboot` | `reboot` |
| `power_cycle` | `powerCycle` |

### Key Implementation Files

| File | Purpose |
|------|---------|
| `internal/provider/provider.go` | Add ProviderWithActions, ActionData |
| `internal/provider/action_cmdevice_power.go` | Action implementation |
| `internal/provider/action_cmdevice_power_test.go` | Unit tests |
| `examples/actions/bcm_cmdevice_power/main.tf` | Example configurations |

---

## Constitution Re-Check (Post-Design)

| Principle | Status | Notes |
|-----------|--------|-------|
| TDD Required | PASS | Test-first approach documented in quickstart |
| Single Responsibility | PASS | One action, one purpose |
| No State Management | PASS | Actions by design have no state |
| Provider Pattern Compliance | PASS | Follows framework action patterns |
| API Verification | PENDING | Must complete before implementation |
| Documentation | PASS | Schema, API contracts, quickstart documented |

---

## Implementation Guidance

### Provider Modification

```go
// provider.go additions
var _ provider.ProviderWithActions = &BCMProvider{}

func (p *BCMProvider) Actions(ctx context.Context) []func() action.Action {
    return []func() action.Action{
        NewCMDevicePowerAction,
    }
}

// In Configure method, add:
resp.ActionData = client
```

### Action Interface Implementation

```go
// action_cmdevice_power.go
var (
    _ action.Action              = &CMDevicePowerAction{}
    _ action.ActionWithConfigure = &CMDevicePowerAction{}
)

type CMDevicePowerAction struct {
    client *BCMClient
}

func (a *CMDevicePowerAction) Metadata(...)  // Returns type name
func (a *CMDevicePowerAction) Schema(...)    // Returns action schema
func (a *CMDevicePowerAction) Configure(...) // Receives BCM client
func (a *CMDevicePowerAction) Invoke(...)    // Executes power operation
```

### Testing Strategy

**Phase 1: Unit Tests (Immediate)**
- Schema validation
- Method mapping
- Configuration parsing

**Phase 2: Manual Testing (TF 1.14 Beta)**
- Direct invocation: `terraform apply -invoke="action.bcm_cmdevice_power.name"`
- Lifecycle triggers with device resource

**Phase 3: Acceptance Tests (TF 1.14 GA)**
- Full terraform-plugin-testing integration
- All power operations tested
- Error handling validated

---

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| BCM API methods don't exist | Low | High | Phase 0 verification before implementation |
| TF 1.14 delayed or changed | Medium | Medium | Unit tests provide immediate coverage |
| Action interface changes | Low | Medium | Pin to framework v1.16.1 |
| wait_for_completion complexity | Medium | Low | Optional feature, can defer |

---

## Next Steps

1. **Run /speckit.tasks** to generate implementation task list
2. **Complete Phase 0 verification** of BCM API power methods
3. **Begin TDD RED phase** with unit tests
4. **Implement action** following quickstart guide
5. **Manual test** with Terraform 1.14 beta
6. **Add acceptance tests** when TF 1.14 reaches GA

---

## References

- [Writing a Terraform Action](https://danielmschmidt.de/posts/2025-09-26-writing-a-terraform-action/)
- [Terraform Action Patterns and Guidelines](https://danielmschmidt.de/posts/2025-09-26-terraform-action-patterns-and-guidelines/)
- [terraform-plugin-framework v1.16.0 Release Notes](https://github.com/hashicorp/terraform-plugin-framework/releases/tag/v1.16.0)
- [action package - pkg.go.dev](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-framework/action)
- [BCM CMDevice API Documentation](../../sampleRest/CMDevice_Complete_Documentation.md)
