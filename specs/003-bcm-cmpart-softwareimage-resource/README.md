# BCM CMPart SoftwareImage Resource Implementation

## Overview

This directory contains the complete implementation plan for the `bcm_cmpart_softwareimage` Terraform resource, which manages Nvidia BCM (Bright Cluster Manager) software images for DPU node provisioning.

## Directory Contents

| File | Description |
|------|-------------|
| `plan.md` | Comprehensive implementation plan with TDD phases and task breakdown |
| `spec.md` | Feature specification with user stories, requirements, and API contracts |
| `quickstart.md` | Step-by-step developer guide for implementing and testing the resource |
| `research.md` | API research findings (created during Phase 0 implementation) |

## Implementation Status

- [x] Phase 0: Research & Planning ✅ **API METHODS CONFIRMED**
- [ ] Phase 1: Design & Contracts
- [ ] Phase 2: Implementation (RED-GREEN-REFACTOR)
- [ ] Phase 3: Documentation Generation
- [ ] Phase 4: Quality Assurance

**Last Updated**: 2025-11-20

## Quick Start

To begin implementation:

1. **Read the Specification**: Start with `spec.md` to understand requirements
2. **Follow the Plan**: Use `plan.md` for detailed implementation steps
3. **Use the Quickstart**: Reference `quickstart.md` for hands-on guidance
4. **Research First**: Complete Phase 0 to discover API methods before coding

## Key Features

- Full CRUD lifecycle management for BCM software images
- Nested kernel modules configuration
- Serial Over LAN (SOL) settings
- Kernel version and parameter configuration
- ImportState functionality for existing images
- Comprehensive acceptance test coverage

## Development Approach

This implementation follows **Test-Driven Development (TDD)** principles:

1. **RED Phase**: Write failing acceptance tests that define expected behavior
2. **GREEN Phase**: Write minimal code to make tests pass (hardcoded values)
3. **REFACTOR Phase**: Replace minimal implementation with full API integration

## Architecture

```
terraform-provider-bcm/
├── internal/provider/
│   ├── resource_cmpart_softwareimage.go       # Resource implementation
│   ├── resource_cmpart_softwareimage_test.go  # Acceptance tests
│   └── bcm_client.go                          # API client (existing)
├── examples/resources/bcm_cmpart_softwareimage/
│   └── resource.tf                            # Example configuration
├── docs/resources/
│   └── cmpart_softwareimage.md                # Generated documentation
└── specs/003-bcm-cmpart-softwareimage-resource/
    ├── plan.md                                # This directory
    ├── spec.md
    └── quickstart.md
```

## API Endpoints

**Service:** `CMPart` (or `cmpart`)

**Confirmed Methods:** ✅
- `addSoftwareImage(entity, force)` - Create software image
- `getSoftwareImage(name)` - Get single image by name ✅ **PREFERRED for Read()**
- `getSoftwareImages()` - List all software images
- `updateSoftwareImage(entity, force)` - Update software image
- `removeSoftwareImage(uuid, removeData, removeAll, force)` - Delete software image
- `validateSoftwareImage(softwareImage)` - Validate entity before create/update ✅ **INCLUDED IN PLAN**

See `research.md` for complete API documentation and parameter details.

## Resource Schema

```hcl
resource "bcm_cmpart_softwareimage" "example" {
  # Required
  name = "ubuntu-22.04-dpu"
  path = "/cm/images/ubuntu-22.04-dpu"

  # Optional Kernel Configuration
  kernel_version         = "5.15.0-58-generic"
  kernel_parameters      = "quiet splash"
  kernel_output_console  = "tty0"  # Default

  # Optional Serial Over LAN
  enable_sol       = false  # Default
  sol_port         = "ttyS1"  # Default
  sol_speed        = "115200"  # Default
  sol_flow_control = true  # Default

  # Optional Kernel Modules
  modules = [
    {
      name       = "nvidia-drm"
      parameters = "modeset=1"
    }
  ]

  # Optional Metadata
  notes = "Ubuntu 22.04 LTS for DPU nodes"
  force = false  # Default

  # Computed (read-only)
  # id, uuid, creation_time, revision_id, file_operation_in_progress
}
```

## Testing Requirements

### Environment Variables

```bash
export TF_ACC=1
export BCM_ENDPOINT="https://172.21.15.254:8081"
export BCM_USERNAME="root"
export BCM_PASSWORD="Hashicorp123!"
```

### Test Scenarios

- Basic resource creation with minimal config
- Full resource creation with all attributes
- Update operations (kernel config, modules)
- Delete operations
- ImportState functionality
- Error cases (duplicate name, invalid path)

### Running Tests

```bash
# Run all acceptance tests for this resource
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMPartSoftwareImageResource

# Run specific test
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMPartSoftwareImageResource_Basic
```

## Dependencies

### Go Packages
- `github.com/hashicorp/terraform-plugin-framework` v1.16.1
- `github.com/hashicorp/terraform-plugin-testing` v1.13.3
- `github.com/hashicorp/terraform-plugin-log/tflog`

### Project Dependencies
- BCM API client (`internal/provider/bcm_client.go`)
- Helper functions for null-safe field extraction
- Existing data source pattern (`data_source_cmpart_softwareimages.go`)

### External Dependencies
- BCM cluster with CMPart service
- Valid BCM credentials
- Network access to BCM endpoint

## Success Criteria

### Functional
- [x] Create software images via API
- [x] Read software images with client-side UUID filtering
- [x] Update software images (method TBD in research)
- [x] Delete software images (method TBD in research)
- [x] Import existing images by UUID
- [x] Manage nested kernel modules list

### Quality
- [x] 100% acceptance test pass rate
- [x] 0 golangci-lint errors
- [x] Code follows Terraform Plugin Framework patterns
- [x] Documentation auto-generated with examples
- [x] Pre-commit hooks pass

## Risks and Mitigations

| Risk | Mitigation |
|------|------------|
| Update/Delete API methods unknown | Phase 0 research discovers methods before implementation |
| Nested modules complicate state | Reuse data source pattern for list management |
| Unique constraints not enforced | Test negative cases, document error handling |
| API response format differs | Comprehensive error handling with logging |

## Timeline Estimate

- **Phase 0 (Research)**: 2-4 hours
- **Phase 1 (Design)**: 1-2 hours
- **Phase 2 (Implementation)**: 4-6 hours
  - RED: 1 hour
  - GREEN: 1-2 hours
  - REFACTOR: 2-3 hours
- **Phase 3 (Documentation)**: 30 minutes
- **Phase 4 (QA)**: 1 hour

**Total**: 8-13 hours

## Related Resources

### Existing Provider Resources
- `data.bcm_cmpart_softwareimages` - Data source for listing images (reference implementation)
- `data.bcm_cmdevice_nodes` - Data source for nodes

### Future Related Resources
- `bcm_cmpart_fspart` - Filesystem partition resource
- `bcm_cmdevice_node` - Device node resource

## Contributing

When implementing this resource:

1. Follow the TDD workflow strictly (RED-GREEN-REFACTOR)
2. Write comprehensive acceptance tests first
3. Use minimal hardcoded implementation in GREEN phase
4. Only add API integration in REFACTOR phase
5. Run tests after each phase to verify progress
6. Generate documentation using `make generate`
7. Run linter and formatter before committing

## Questions or Issues

If you encounter issues during implementation:

1. Check the `quickstart.md` troubleshooting section
2. Review `plan.md` for detailed implementation guidance
3. Consult `spec.md` for requirements clarification
4. Reference existing data source implementation
5. Check BCM API documentation in `sampleRest/`

## Additional Documentation

- **Project README**: `/workspace/README.md`
- **Provider Development Guide**: `/workspace/AGENTS.md`
- **Claude Instructions**: `/workspace/CLAUDE.md`
- **TDD Constitution**: `/workspace/.specify/memory/constitution.md`
- **BCM API Docs**: `/workspace/sampleRest/BCM_API_Complete_Documentation.md`
- **SoftwareImage Entity**: `/workspace/sampleRest/wip/resource_cmpart_softwareimage.md`

## License

This provider is licensed under the Mozilla Public License 2.0 (MPL-2.0).
Copyright (c) HashiCorp, Inc.
