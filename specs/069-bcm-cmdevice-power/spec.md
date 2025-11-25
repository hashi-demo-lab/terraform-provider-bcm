# Feature Specification: BCM CMDevice Power Action

**Feature Branch**: `069-bcm-cmdevice-power`
**Created**: 2025-11-26
**Status**: Draft
**Input**: User description: "Implement bcm_cmdevice_power Terraform Action for node power operations (power on/off/reboot/cycle)"
**GitHub Issue**: #69

## Overview

This feature implements a new Terraform Action `bcm_cmdevice_power` to manage node power operations in NVIDIA Bright Cluster Manager (BCM). Actions are a new Terraform feature (v1.14+) designed specifically for imperative, non-CRUD operations that do not fit the traditional resource lifecycle model.

### Why Actions Instead of Resources?

Power operations are imperative commands that execute immediately and do not maintain state:
- No "desired state" to reconcile - power commands execute once
- No drift detection needed - power state changes externally
- Lifecycle triggers enable "power on after device creation" workflows
- Direct invocation via `terraform apply -invoke="action.<name>"`

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Direct Power Control (Priority: P1)

As an infrastructure operator, I want to power on, power off, reboot, or power cycle BCM-managed nodes directly from Terraform so that I can control node power state as part of my infrastructure automation workflow.

**Why this priority**: This is the core functionality - without direct power control, the action provides no value. It enables the primary use case of the NVIDIA DGX BasePOD Deployment Guide: powering on nodes after configuration for PXE boot provisioning.

**Independent Test**: Can be fully tested by creating a test device in BCM, invoking the power action with each operation type, and verifying the operation was called via BCM API logs. Delivers immediate value for manual power operations.

**Acceptance Scenarios**:

1. **Given** a BCM device with a known UUID, **When** I invoke the power action with `power_action = "power_on"`, **Then** the BCM API `powerOn` method is called with the device UUID
2. **Given** a BCM device with a known UUID, **When** I invoke the power action with `power_action = "power_off"`, **Then** the BCM API `powerOff` method is called with the device UUID
3. **Given** a BCM device with a known UUID, **When** I invoke the power action with `power_action = "reboot"`, **Then** the BCM API `reboot` method is called with the device UUID
4. **Given** a BCM device with a known UUID, **When** I invoke the power action with `power_action = "power_cycle"`, **Then** the BCM API `powerCycle` method is called with the device UUID

---

### User Story 2 - Lifecycle Triggered Power On (Priority: P2)

As a DevOps engineer deploying bare-metal Kubernetes nodes, I want nodes to automatically power on after Terraform creates them so that PXE boot provisioning begins immediately without manual intervention.

**Why this priority**: This is the key automation use case that differentiates Actions from manual scripts. It enables fully automated provisioning workflows where device creation triggers power-on for PXE boot.

**Independent Test**: Can be tested by creating a `bcm_cmdevice_device` resource with a lifecycle action trigger configured to invoke the power action on `after_create`. Verify the power action is automatically invoked when the device is created.

**Acceptance Scenarios**:

1. **Given** a bcm_cmdevice_device resource with an `action_trigger` configured for `after_create`, **When** `terraform apply` creates the device, **Then** the power action is automatically invoked after device creation
2. **Given** multiple devices with lifecycle triggers, **When** `terraform apply` creates them, **Then** each device's power action is invoked in the correct order

---

### User Story 3 - Wait for Power State Change (Priority: P3)

As an automation engineer, I want to optionally wait for the power operation to complete before continuing so that subsequent operations can depend on the node being in the expected power state.

**Why this priority**: Optional enhancement that enables sequenced operations. Not required for basic functionality but important for complex orchestration scenarios where subsequent steps depend on power state.

**Independent Test**: Can be tested by invoking the power action with `wait_for_completion = true` and verifying that the action blocks until the power state changes or timeout occurs.

**Acceptance Scenarios**:

1. **Given** a device and `wait_for_completion = true`, **When** the power action is invoked, **Then** the action blocks until the power state change is confirmed or timeout occurs
2. **Given** a device and `wait_for_completion = true` with `timeout = "30s"`, **When** the power state does not change within 30 seconds, **Then** the action fails with a timeout error
3. **Given** a device and `wait_for_completion = false` (default), **When** the power action is invoked, **Then** the action returns immediately after sending the API command

---

### Edge Cases

- What happens when the device UUID does not exist in BCM?
  - The action MUST fail with a clear error message indicating the device was not found
- What happens when the device is already in the requested power state?
  - The action MUST succeed (idempotent behavior) - BCM API handles this gracefully
- What happens when the BMC/IPMI is unreachable?
  - The action MUST fail with a clear error message from BCM API
- What happens when an invalid power_action value is provided?
  - The action MUST fail at plan time with a validation error listing valid options
- What happens when wait_for_completion times out?
  - The action MUST fail with a timeout error indicating the operation was sent but completion could not be confirmed

## Requirements *(mandatory)*

### Phase 0: API Verification Requirements

Before implementation, the following BCM API methods MUST be verified:

- **P0-001**: Verify `cmdevice` service exposes `powerOn` method accepting device UUID
- **P0-002**: Verify `cmdevice` service exposes `powerOff` method accepting device UUID
- **P0-003**: Verify `cmdevice` service exposes `reboot` method accepting device UUID
- **P0-004**: Verify `cmdevice` service exposes `powerCycle` method accepting device UUID
- **P0-005**: Document the expected API request/response format for each power method
- **P0-006**: Verify error response format when device UUID is invalid
- **P0-007**: Identify any additional API methods for querying current power state (for wait_for_completion)

### Functional Requirements

- **FR-001**: Provider MUST implement `provider.ProviderWithActions` interface to register actions
- **FR-002**: Action MUST implement `action.Action` interface with `Schema`, `Metadata`, and `Invoke` methods
- **FR-003**: Action MUST implement `action.ActionWithConfigure` interface to receive BCM client
- **FR-004**: Action MUST accept `device_id` (string, required) attribute containing the BCM device UUID
- **FR-005**: Action MUST accept `power_action` (string, required) attribute with allowed values: `power_on`, `power_off`, `reboot`, `power_cycle`
- **FR-006**: Action MUST validate `power_action` values at plan time using schema validators
- **FR-007**: Action MUST accept `wait_for_completion` (bool, optional) attribute defaulting to `false`
- **FR-008**: Action MUST accept `timeout` (string, optional) attribute defaulting to `"5m"` (only used when wait_for_completion is true)
- **FR-009**: Action MUST report progress during invocation using Terraform's progress reporting mechanism
- **FR-010**: Action MUST return clear error messages when BCM API calls fail
- **FR-011**: Action MUST support lifecycle triggers for `after_create` events on devices

### Key Entities

- **BCM Device**: A managed compute node in Bright Cluster Manager identified by UUID. Has power state (on/off) controllable via BMC/IPMI.
- **Power Action**: An imperative command sent to BCM to change device power state. Does not persist in Terraform state.
- **Terraform Action**: A new Terraform primitive (v1.14+) for imperative operations that can be invoked directly or triggered by resource lifecycle events.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: All four power operations (power_on, power_off, reboot, power_cycle) successfully invoke the corresponding BCM API method
- **SC-002**: Invalid power_action values are rejected at plan time with clear validation errors
- **SC-003**: Invalid device_id values result in clear error messages from BCM API
- **SC-004**: Action can be invoked directly via `terraform apply -invoke="action.bcm_cmdevice_power.<name>"`
- **SC-005**: Action can be triggered automatically via resource lifecycle `action_trigger` configuration
- **SC-006**: Progress reporting provides user feedback during action invocation
- **SC-007**: Documentation and examples enable users to adopt the action without support requests

## Technical Context

### Terraform Actions Pattern (New for This Provider)

This is the first Terraform Action in this provider. Actions differ from resources:

| Aspect | Resource | Action |
|--------|----------|--------|
| State | Maintains state in terraform.tfstate | No state - executes and completes |
| Lifecycle | CRUD (Create, Read, Update, Delete) | Invoke only |
| Drift | Detects and reconciles drift | N/A - no state to drift |
| Invocation | Automatic on apply | Direct (`-invoke`) or lifecycle trigger |
| Use Case | Stateful infrastructure | Imperative operations |

### Provider Registration Pattern

```go
// provider.go - Add to BCMProvider
var _ provider.ProviderWithActions = &BCMProvider{}

func (p *BCMProvider) Actions(ctx context.Context) []func() action.Action {
    return []func() action.Action{
        NewCMDevicePowerAction,
    }
}
```

### Action Interface Implementation

```go
// action_cmdevice_power.go
type CMDevicePowerAction struct {
    client *BCMClient
}

var _ action.Action = &CMDevicePowerAction{}
var _ action.ActionWithConfigure = &CMDevicePowerAction{}

func NewCMDevicePowerAction() action.Action {
    return &CMDevicePowerAction{}
}

func (a *CMDevicePowerAction) Metadata(ctx context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
    resp.TypeName = "bcm_cmdevice_power"
}

func (a *CMDevicePowerAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
    // Define device_id, power_action, wait_for_completion, timeout attributes
}

func (a *CMDevicePowerAction) Configure(ctx context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
    // Receive BCM client from provider
}

func (a *CMDevicePowerAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
    // Execute power operation via BCM API
}
```

### Example Configurations

**Direct Invocation:**
```hcl
action "bcm_cmdevice_power" "shutdown_worker" {
  device_id    = bcm_cmdevice_device.worker.uuid
  power_action = "power_off"
}
```

```bash
terraform apply -invoke="action.bcm_cmdevice_power.shutdown_worker"
```

**Lifecycle Trigger (Auto-boot after creation):**
```hcl
resource "bcm_cmdevice_device" "knode" {
  hostname = "knode-01"
  category = bcm_cmdevice_category.k8s_control_plane.uuid

  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.bcm_cmdevice_power.boot_node]
    }
  }
}

action "bcm_cmdevice_power" "boot_node" {
  device_id    = bcm_cmdevice_device.knode.uuid
  power_action = "power_on"
}
```

## Version Requirements

| Component | Minimum | Current Status |
|-----------|---------|----------------|
| Terraform CLI | 1.14.0 | Beta (as of Nov 2025) |
| terraform-plugin-framework | 1.16.0 | 1.16.1 (supported) |
| Go | 1.24.0 | 1.24.0 (supported) |

## TDD Requirements

### Testing Strategy

Due to Terraform 1.14 being in beta, testing will be structured in phases:

**Phase 1: Unit Tests (Immediate)**
- Test schema validation (valid/invalid power_action values)
- Test action metadata
- Mock BCM client to test Invoke logic without real API calls

**Phase 2: Integration Tests (When Terraform 1.14 GA)**
- Full acceptance tests using `terraform-plugin-testing`
- Test all four power operations against real BCM API
- Test lifecycle trigger integration with device resources
- Test wait_for_completion behavior

### Test File Structure

```
internal/provider/
  action_cmdevice_power.go          # Action implementation
  action_cmdevice_power_test.go     # Unit and acceptance tests
examples/actions/bcm_cmdevice_power/
  main.tf                           # Example configuration
  README.md                         # Usage documentation
```

## Assumptions

- BCM API power methods (`powerOn`, `powerOff`, `reboot`, `powerCycle`) accept device UUID as the primary argument
- BCM API returns immediately after sending the power command to BMC/IPMI
- Power state can be queried via the device's `powerStatus` or similar field for wait_for_completion
- Terraform 1.14+ will be available before this feature is released to production
- The terraform-plugin-framework `action` package provides the interfaces documented in v1.16.0 release notes

## Dependencies

- GitHub Issue #69 tracks this feature
- NVIDIA DGX BasePOD Deployment Guide requirements
- Terraform 1.14 beta for action invocation testing
- Existing `bcm_cmdevice_device` resource for lifecycle trigger integration

## Out of Scope

- Bulk power operations (multiple devices in one action) - use terraform `for_each` instead
- Power scheduling (timed power operations) - out of scope for this feature
- Power state monitoring/alerting - use BCM native monitoring
- Integration with other power management systems (non-BCM)
