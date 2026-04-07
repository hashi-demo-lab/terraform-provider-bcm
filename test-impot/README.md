# Test: Import and Config Generation

Demonstrates importing an existing BCM device into Terraform using
`terraform plan -generate-config-out` (Terraform 1.5+).

## Prerequisites

```bash
# Provider must be built and installed
make install

# dev_overrides must be configured in ~/.terraformrc:
# provider_installation {
#   dev_overrides {
#     "hashicorp/bcm" = "/Users/simon.lynch/go/bin"
#   }
#   direct {}
# }
```

## Step 1: Find the device UUID

```bash
cd test-impot

# Create a lookup config
cat > lookup.tf <<'EOF'
data "bcm_cmdevice_nodes" "all" {}

output "devices" {
  value = [for n in data.bcm_cmdevice_nodes.all.nodes : "${n.id}  ${n.hostname}  ${n.child_type}"]
}
EOF

terraform apply -auto-approve
# Pick a UUID from the output, then clean up:
rm lookup.tf terraform.tfstate
```

## Step 2: Create the import block

```bash
cat > import.tf <<'EOF'
import {
  to = bcm_cmdevice_device.example
  id = "<device-uuid-from-step-1>"
}
EOF
```

## Step 3: Generate config

```bash
terraform plan -generate-config-out=generated.tf
```

This connects to BCM, reads the device, and generates `generated.tf` with
the full resource configuration. Review the output — BCM sentinel values
(`0.0.0.0`, `00:00:00:00:00:00`, zero UUID) are mapped to null automatically.

## Step 4: Apply the import

```bash
terraform apply
```

The device is now in Terraform state with 0 changes.

## Step 5: Clean up and use

```bash
# Remove the import block (no longer needed)
rm import.tf

# Edit generated.tf — remove null lines, use data sources for UUIDs:
#   category = data.bcm_cmdevice_categories.example.categories[0].id

# Verify idempotency
terraform plan
# Should show: No changes. Your infrastructure matches the configuration.
```

## Notes

- The `provider.tf` uses `hashicorp/bcm` source to match `dev_overrides`.
  For production, use `hashi-demo-lab/bcm` with a version constraint.
- Generated config includes `= null` for all unset Optional attributes.
  This is standard Terraform behavior — remove them for cleaner config.
- BCM returns `"CATEGORY"` for `boot_loader` and `boot_loader_protocol`
  on devices inheriting from their category. The provider maps these to
  null so they don't appear in generated config or cause validator errors.
