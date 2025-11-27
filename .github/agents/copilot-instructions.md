# workspace Development Guidelines

Auto-generated from all feature plans. Last updated: 2025-11-20

## Active Technologies
- Go 1.24.0 + terraform-plugin-framework (v1.16.1), terraform-plugin-testing (v1.13.3), terraform-plugin-log (001-cmdevice-category)
- BCM API backend (JSON-RPC over HTTPS), cookie-based session authentication (001-cmdevice-category)
- Go 1.24+ + terraform-plugin-framework v1.16.1, terraform-plugin-testing v1.13.3 (001-category-test-coverage)
- N/A (test infrastructure only) (001-category-test-coverage)
- N/A (stateless provider, BCM API is source of truth) (038-device-interfaces-block)
- N/A (state managed by Terraform, data stored in BCM) (039-device-roles-block)
- N/A (stateless API integration via BCM JSON-RPC) (068-cmuser-user-resource)
- BCM JSON-RPC API (cookie-based auth) (070-category-optional-fields-tests)
- Terraform state (password as sensitive attribute) (082-bmc-password-drift)
- N/A (Terraform state management) (083-roles-uuid-computed)
- Go 1.24+ (Terraform provider), Python 3.x (investigation scripts) + terraform-plugin-framework v1.16.1, terraform-plugin-testing v1.13.3, requests (Python) (073-category-list-fields)
- N/A (BCM API is external system) (073-category-list-fields)

- Go 1.24.0 + terraform-plugin-framework v1.16.1, terraform-plugin-testing v1.13.3, terraform-plugin-log v0.10.0 (001-bcm-provider)

## Project Structure

```text
src/
tests/
```

## Commands

# Add commands for Go 1.24.0

## Code Style

Go 1.24.0: Follow standard conventions

## Recent Changes
- 073-category-list-fields: Added Go 1.24+ (Terraform provider), Python 3.x (investigation scripts) + terraform-plugin-framework v1.16.1, terraform-plugin-testing v1.13.3, requests (Python)
- 082-bmc-password-drift: Added Go 1.24.0 + terraform-plugin-framework v1.16.1, terraform-plugin-testing v1.13.3
- 083-roles-uuid-computed: Added Go 1.24.0 + terraform-plugin-framework v1.16.1, terraform-plugin-testing v1.13.3


<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
