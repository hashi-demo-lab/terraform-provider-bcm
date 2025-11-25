# Data Model: BCM CMDevice Power Action

**Feature**: bcm_cmdevice_power Action
**Branch**: `069-bcm-cmdevice-power`
**Date**: 2025-11-26

## Overview

Actions in Terraform do not maintain persistent state. This data model defines:
1. The configuration model passed to the action
2. Internal data structures for API interaction
3. No state is persisted after action completion

## Entity Diagram

```
+---------------------------+
|   CMDevicePowerAction     |  (Terraform Action)
+---------------------------+
| - client: *BCMClient      |  (injected via Configure)
+---------------------------+
            |
            | uses
            v
+---------------------------+
|  CMDevicePowerActionModel |  (Configuration from HCL)
+---------------------------+
| - DeviceID: string        |  (required)
| - PowerAction: string     |  (required)
| - WaitForCompletion: bool |  (optional, default: false)
| - Timeout: string         |  (optional, default: "5m")
+---------------------------+
            |
            | transforms to
            v
+---------------------------+
|    BCM JSON-RPC Request   |  (API Call)
+---------------------------+
| - service: "cmdevice"     |
| - call: "{powerMethod}"   |
| - args: ["{deviceID}"]    |
+---------------------------+
```

## Configuration Model

### CMDevicePowerActionModel

The Terraform configuration model for action inputs.

```go
type CMDevicePowerActionModel struct {
    DeviceID          types.String `tfsdk:"device_id"`
    PowerAction       types.String `tfsdk:"power_action"`
    WaitForCompletion types.Bool   `tfsdk:"wait_for_completion"`
    Timeout           types.String `tfsdk:"timeout"`
}
```

**Fields**:

| Field | Type | HCL Attribute | Required | Default |
|-------|------|---------------|----------|---------|
| DeviceID | types.String | device_id | Yes | - |
| PowerAction | types.String | power_action | Yes | - |
| WaitForCompletion | types.Bool | wait_for_completion | No | false |
| Timeout | types.String | timeout | No | "5m" |

### DeviceID

Identifies the target BCM device for the power operation.

**Valid Formats**:
- UUID (RFC 4122): `"2870c0b0-6fda-4026-9b8f-28be4c372fee"`
- Hostname: `"node001"`

**Validation**:
- Non-empty string
- No format validation at schema level (BCM accepts both)

### PowerAction

Specifies the power operation to execute.

**Valid Values**:

| Value | BCM Method | Description |
|-------|------------|-------------|
| `power_on` | `powerOn` | Power on device via BMC |
| `power_off` | `powerOff` | Power off device via BMC |
| `reboot` | `reboot` | Graceful device reboot |
| `power_cycle` | `powerCycle` | Hard power cycle |

**Validation**:
- Schema validator: `stringvalidator.OneOf(...)`

### WaitForCompletion

Controls whether the action blocks until power state is confirmed.

**Behavior**:
- `false` (default): Return immediately after API call succeeds
- `true`: Poll device status until power state matches expected or timeout

**State Expectations by Operation**:

| PowerAction | Expected Final State |
|-------------|---------------------|
| power_on | "on" |
| power_off | "off" |
| reboot | "on" (after brief off) |
| power_cycle | "on" (after brief off) |

### Timeout

Duration to wait when `WaitForCompletion = true`.

**Format**: Go duration string
**Default**: `"5m"` (5 minutes)
**Range**: 10s to 30m

**Parsing**:
```go
duration, err := time.ParseDuration(config.Timeout.ValueString())
```

---

## Internal Data Structures

### Power Method Mapping

Internal mapping from Terraform values to BCM API methods:

```go
var powerMethodMapping = map[string]string{
    "power_on":    "powerOn",
    "power_off":   "powerOff",
    "reboot":      "reboot",
    "power_cycle": "powerCycle",
}
```

### Power State Constants

Expected power states for wait verification:

```go
const (
    PowerStateOn      = "on"
    PowerStateOff     = "off"
    PowerStateUnknown = "unknown"
)
```

### Expected State by Operation

```go
var expectedPowerState = map[string]string{
    "power_on":    PowerStateOn,
    "power_off":   PowerStateOff,
    "reboot":      PowerStateOn,
    "power_cycle": PowerStateOn,
}
```

---

## BCM Device Power Fields

Relevant fields from BCM Device entity for power operations:

```go
type BCMDevicePowerInfo struct {
    UUID         string `json:"uuid"`
    Hostname     string `json:"hostname"`
    PowerControl string `json:"powerControl"` // "none", "ipmi", "redfish", "pdu"
    PowerStatus  string `json:"powerStatus"`  // "on", "off", "unknown" (needs verification)
}
```

**Note**: The `powerStatus` field existence needs Phase 0 verification.

---

## API Request/Response Models

### Power Operation Request

```go
type PowerOperationRequest struct {
    Service string        `json:"service"` // "cmdevice"
    Call    string        `json:"call"`    // powerOn, powerOff, reboot, powerCycle
    Args    []interface{} `json:"args"`    // [deviceID]
}
```

### Power Operation Response

**Success Response** (expected):
```json
true
```

**Error Response**:
```json
{
  "error": "Device not found",
  "code": "NOT_FOUND"
}
```

---

## State Machine (No Persistent State)

Actions do not persist state. The execution flow:

```
[Start]
    |
    v
[Read Configuration]
    |
    v
[Validate Configuration]
    |-- Invalid --> [Return Error]
    v
[Execute Power API Call]
    |-- Error --> [Return Error]
    v
[Wait for Completion?]
    |-- No --> [Return Success]
    |-- Yes
    v
[Poll Device Status]
    |-- Timeout --> [Return Warning]
    |-- State Changed --> [Return Success]
    |-- Error --> [Return Error]
    v
[End]
```

---

## Relationships to Existing Entities

### BCM Device (bcm_cmdevice_device)

The power action operates on devices managed by the `bcm_cmdevice_device` resource.

**Reference Pattern**:
```hcl
resource "bcm_cmdevice_device" "worker" {
  hostname           = "worker-01"
  mac                = "00:11:22:33:44:55"
  category           = bcm_cmdevice_category.compute.uuid
  management_network = bcm_cmnet_network.mgmt.uuid
}

action "bcm_cmdevice_power" "boot_worker" {
  device_id    = bcm_cmdevice_device.worker.uuid  # Reference device UUID
  power_action = "power_on"
}
```

### BCM Category (bcm_cmdevice_category)

Power operations apply to devices in categories. Category determines device characteristics but not power control.

### BCM Network (bcm_cmnet_network)

Power operations are independent of network configuration. BMC/IPMI typically uses separate management network.

---

## Validation Rules

### Configuration Validation (Plan Time)

| Rule | Field | Validator |
|------|-------|-----------|
| Required | device_id | Required attribute |
| Non-empty | device_id | LengthAtLeast(1) |
| Required | power_action | Required attribute |
| Allowed values | power_action | OneOf("power_on", "power_off", "reboot", "power_cycle") |
| Valid duration | timeout | Custom duration validator |

### Invoke-Time Validation

| Rule | Error Type |
|------|------------|
| Device exists | Error: Device not found |
| Device has power control | Warning: No power control configured |
| BMC reachable | Error: BMC connection failed |
| Timeout | Warning: Operation sent but not confirmed |

---

## No State Persistence

Unlike resources, actions do not:
- Store state in terraform.tfstate
- Support drift detection
- Support import
- Have computed attributes

Each invocation is independent and executes the power operation fresh.
