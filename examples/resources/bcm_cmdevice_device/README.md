# bcm_cmdevice_device Resource Examples

This directory contains comprehensive examples for the `bcm_cmdevice_device` resource, demonstrating various device configurations for the Nvidia BCM (Bright Cluster Manager) provider.

## Examples

### basic.tf
Minimal device configuration demonstrating:
- Basic compute device setup
- Required fields: hostname, mac, category, management_network
- Data source lookups for category and network

**Use case:** Quick device provisioning with minimal configuration.

### ipmi.tf
IPMI-enabled device with power control:
- Power control via IPMI
- Network gateway configuration
- Hardware identifiers (serial number, part number)
- Custom kernel parameters for IPMI

**Use case:** Servers with BMC/IPMI support requiring remote power management.

### import.tf
Device import example:
- Demonstrates importing existing BCM devices into Terraform state
- Shows import block syntax
- Useful for adopting existing infrastructure

**Use case:** Migrating manually-created devices to Terraform management.

### resource.tf
Comprehensive examples showcasing all device features:
1. **compute_basic** - Minimal configuration
2. **compute_custom** - Custom boot loader and kernel parameters
3. **compute_ipmi** - IPMI with power control and network configuration
4. **gpu_node** - GPU compute node with NVIDIA-specific parameters
5. **storage_node** - Storage node for Ceph clusters

**Use case:** Reference implementation demonstrating all available fields and configurations.

## Testing

### Quick Validation (Plan Only)

Test configuration syntax and planning without creating resources:

```bash
cd examples/resources/bcm_cmdevice_device/
terraform init
terraform validate
terraform plan
```

### Full Integration Testing

Run complete plan/apply/destroy cycle with real BCM infrastructure:

```bash
# Set credentials
export BCM_ENDPOINT="https://172.21.15.254:8081"
export BCM_USERNAME="root"
export BCM_PASSWORD="your-password"

# Run full test for basic.tf
./scripts/test-basic-full.sh
```

The full integration test script validates:
1. ✓ terraform init
2. ✓ terraform validate
3. ✓ terraform plan
4. ✓ terraform apply (creates real device in BCM)
5. ✓ Resource creation verification (BCM API)
6. ✓ Idempotency verification (plan shows no changes)
7. ✓ terraform destroy
8. ✓ Resource deletion verification (BCM API)

### Automated Testing

All examples are automatically tested via the test infrastructure:

```bash
# Test all examples
./scripts/test-examples.sh

# Test only resource examples (including bcm_cmdevice_device)
./scripts/test-examples.sh --resources-only

# Test with detailed output
./scripts/test-examples.sh --resources-only --verbose
```

## Resource Naming Convention

All test resources use the `citest-` prefix to:
- Clearly identify test infrastructure
- Enable automated cleanup
- Prevent conflicts with production resources

Example hostnames:
- `citest-compute-basic`
- `citest-compute-node-01`
- `citest-gpu-node-01`

## Prerequisites

### Required Data Sources

All examples require these data sources to be available in your BCM cluster:

1. **bcm_cmdevice_categories** - At least one category (typically "default")
2. **bcm_cmnet_networks** - At least one management network (e.g., "DefaultEthernet")

### Required Resources (for resource.tf)

The comprehensive `resource.tf` example creates supporting resources:
- Category: `bcm_cmdevice_category.compute`
- Software Image: `bcm_cmpart_softwareimage.ubuntu_compute`

These are created first, then used by device resources.

## Common Patterns

### MAC Address Management

Each device requires a unique MAC address:
```hcl
mac = "00:11:22:33:44:55"  # Unique per device
```

For testing, use the AA-FF range to avoid conflicts:
```hcl
mac = "00:11:22:33:44:AA"  # Test device 1
mac = "00:11:22:33:44:AB"  # Test device 2
```

### Category Selection

Use data source to look up existing categories:
```hcl
data "bcm_cmdevice_categories" "default" {
  name = "default"
}

resource "bcm_cmdevice_device" "example" {
  category = data.bcm_cmdevice_categories.default.categories[0].id
  # ...
}
```

Or reference a Terraform-managed category:
```hcl
resource "bcm_cmdevice_category" "compute" {
  name = "my-category"
  # ...
}

resource "bcm_cmdevice_device" "example" {
  category = bcm_cmdevice_category.compute.id
  # ...
}
```

### Network Configuration

Management network is required:
```hcl
data "bcm_cmnet_networks" "management" {
  filter {
    name_pattern = "DefaultEthernet"
  }
}

resource "bcm_cmdevice_device" "example" {
  management_network = data.bcm_cmnet_networks.management.networks[0].id
  # ...
}
```

Optional gateway configuration for advanced networking:
```hcl
resource "bcm_cmdevice_device" "example" {
  default_gateway        = "192.168.1.1"
  default_gateway_metric = 100
  # ...
}
```

## Troubleshooting

### Provider Not Found

If you see:
```
Error: Invalid resource type
│ The provider hashicorp/bcm does not support resource type "bcm_cmdevice_device".
```

**Solution:** Rebuild the provider binary:
```bash
make build
# or
go build -o terraform-provider-bcm_v0.1.0 .
```

### BCM Connection Timeout

If you see:
```
Error: Unable to Create BCM Client
│ BCM Client Error: login API call failed: context deadline exceeded
```

**Solution:** Verify BCM connectivity:
```bash
curl -k "https://172.21.15.254:8081/json" \
  -H "Content-Type: application/json" \
  -d '{"service":"login","username":"root","password":"your-password"}'
```

### Resource Already Exists

If a device with the same hostname or MAC already exists:

**Solution 1:** Import existing resource:
```bash
terraform import bcm_cmdevice_device.example <device-uuid>
```

**Solution 2:** Use unique names:
```hcl
hostname = "citest-compute-${formatdate("YYYYMMDDhhmmss", timestamp())}"
```

## Documentation

For complete resource documentation, see:
- [Resource Documentation](../../../docs/resources/cmdevice_device.md)
- [BCM Provider Documentation](../../../docs/index.md)

## Contributing

When adding new examples:
1. Use `citest-` prefix for all resource names
2. Include descriptive comments explaining the use case
3. Test with both `--no-cleanup` for debugging and cleanup verification
4. Document any special prerequisites or dependencies
