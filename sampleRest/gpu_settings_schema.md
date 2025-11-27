# BCM GPU Settings Entity Schema

Extracted from BCM API documentation JavaScript bundle.

## Entity Hierarchy

```
Entity
  └── GPUSettings (base)
        ├── NvidiaGPUSettings
        └── AMDGPUSettings
```

## GPUSettings (Base Type)

- **parent**: Entity
- **plural**: GPUSettings

### Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes (unique) | GPU range for which these settings apply (e.g., "0", "0-3") |

### BCM API Structure

```json
{
  "baseType": "GPUSettings",
  "childType": "NvidiaGPUSettings",  // or "AMDGPUSettings"
  "uuid": "...",
  "name": "0-3",
  // ... child type specific fields
}
```

## NvidiaGPUSettings (Child Type)

- **parent**: GPUSettings
- **simple**: "Nvidia"

### Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `name` | string | - | GPU range (inherited from GPUSettings) |
| `powerLimit` | uint64 | - | Power limit in Watts |
| `eccMode` | enum (EccMode) | NONE | ECC mode: DISABLED, ENABLED, NONE |
| `computeMode` | enum (ComputeMode) | NONE | Compute mode |
| `clockSyncBoostMode` | enum (ClockSyncBoostMode) | NONE | Clock sync boost among GPUs |
| `multiProcessorClockSpeed` | uint64 | - | Streaming MP clock speed (Hz) |
| `memoryClockSpeed` | uint64 | - | Memory clock speed (Hz) |
| `migProfiles` | array[string] | - | MIG profiles (regex: `^(\d+\*)?(\d+|\d+g\.\d+gb(\+me)?)(:(\d+c?))*$`) |
| `workloadPowerProfile` | enum (WorkloadPowerProfile) | UNDEFINED | Workload power profile |
| `secondaryWorkloadPowerProfile` | enum (WorkloadPowerProfile) | UNDEFINED | Secondary workload power profile |

### BCM API Example

```json
{
  "baseType": "GPUSettings",
  "childType": "NvidiaGPUSettings",
  "uuid": "a1b2c3d4-...",
  "name": "0-1",
  "powerLimit": 300,
  "eccMode": "ENABLED",
  "computeMode": "EXCLUSIVE_PROCESS",
  "clockSyncBoostMode": "NONE",
  "multiProcessorClockSpeed": 1800000000,
  "memoryClockSpeed": 1500000000,
  "migProfiles": ["1g.5gb"],
  "workloadPowerProfile": "UNDEFINED",
  "secondaryWorkloadPowerProfile": "UNDEFINED"
}
```

## AMDGPUSettings (Child Type)

- **parent**: GPUSettings
- **simple**: "AMD"

### Fields

| Field | Type | Range | Description |
|-------|------|-------|-------------|
| `name` | string | - | GPU range (inherited from GPUSettings) |
| `gpuClockLevel` | uint64 | 0-7 | GPU clock frequency level |
| `memoryClockLevel` | uint64 | 0-3 | Memory clock frequency level |
| `powerPlay` | enum (PowerPlay) | - | Power play mode (default: DEFAULT) |
| `gpuOverDrive` | double | 0-0.2 | GPU overdrive percentage |
| `memoryOverDrive` | double | 0-0.2 | Memory overdrive percentage (hidden) |
| `fanSpeed` | uint64 | 0-255 | Fan speed value |
| `minimalGPUClock` | uint64 | - | Minimum GPU clock speed (Hz) |
| `minimalMemoryClock` | uint64 | - | Minimum memory clock speed (Hz) |
| `activityThreshold` | double | 0-1 | Workload threshold before clock change (%) |
| `hysteresisUp` | double | - | Delay before clock level increase (s) |
| `hysteresisDown` | double | - | Delay before clock level decrease (s) |

### BCM API Example

```json
{
  "baseType": "GPUSettings",
  "childType": "AMDGPUSettings",
  "uuid": "e5f6g7h8-...",
  "name": "0",
  "gpuClockLevel": 5,
  "memoryClockLevel": 2,
  "powerPlay": "DEFAULT",
  "gpuOverDrive": 0.1,
  "fanSpeed": 128,
  "minimalGPUClock": 800000000,
  "minimalMemoryClock": 400000000,
  "activityThreshold": 0.5,
  "hysteresisUp": 1.0,
  "hysteresisDown": 2.0
}
```

## Enums

### EccMode (Nvidia)
- `DISABLED`
- `ENABLED`
- `NONE` (default)

### ComputeMode (Nvidia)
- Values not fully extracted, typically includes:
  - `DEFAULT`
  - `EXCLUSIVE_THREAD`
  - `EXCLUSIVE_PROCESS`
  - `PROHIBITED`
  - `NONE` (default)

### ClockSyncBoostMode (Nvidia)
- `NONE` (default)
- Other values TBD

### WorkloadPowerProfile (Nvidia)
- `UNDEFINED` (default)
- Other values TBD

### PowerPlay (AMD)
- `DEFAULT` (default)
- Other values TBD

## Terraform Schema Design

```hcl
gpu_settings = [
  {
    # Base GPUSettings fields
    name        = "0-1"     # GPU range (required)
    child_type  = "nvidia"  # or "amd" (required)
    uuid        = "..."     # Computed

    # Nvidia-specific (only when child_type = "nvidia")
    power_limit                    = 300
    ecc_mode                       = "ENABLED"
    compute_mode                   = "EXCLUSIVE_PROCESS"
    clock_sync_boost_mode          = "NONE"
    multiprocessor_clock_speed     = 1800000000
    memory_clock_speed             = 1500000000
    mig_profiles                   = ["1g.5gb"]
    workload_power_profile         = "UNDEFINED"
    secondary_workload_power_profile = "UNDEFINED"

    # AMD-specific (only when child_type = "amd")
    gpu_clock_level      = 5
    memory_clock_level   = 2
    power_play           = "DEFAULT"
    gpu_overdrive        = 0.1
    fan_speed            = 128
    minimal_gpu_clock    = 800000000
    minimal_memory_clock = 400000000
    activity_threshold   = 0.5
    hysteresis_up        = 1.0
    hysteresis_down      = 2.0
  }
]
```

## Notes

1. The `name` field is the GPU range selector, NOT device_id
2. `childType` must be "NvidiaGPUSettings" or "AMDGPUSettings"
3. UUID is BCM-assigned (computed)
4. Fields are vendor-specific - Nvidia fields won't apply to AMD and vice versa
5. Empty `gpuSettings: []` is the default for categories
