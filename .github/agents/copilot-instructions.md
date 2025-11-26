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
- 070-category-optional-fields-tests: Added Go 1.24+ + terraform-plugin-framework v1.16.1, terraform-plugin-testing v1.13.3
- 068-cmuser-user-resource: Added Go 1.24.0 + terraform-plugin-framework v1.16.1, terraform-plugin-testing v1.13.3
- 039-device-roles-block: Added Go 1.24+ + terraform-plugin-framework v1.16.1, terraform-plugin-testing v1.13.3


<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
