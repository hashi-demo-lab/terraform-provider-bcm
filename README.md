# WORK IN PROGRESS

# Terraform Provider for Nvidia BCM

Terraform provider for managing Nvidia BCM (Base Comand Manager) infrastructure.

## Features

- **Data Sources**: Query BCM resources (nodes, categories, software images, networks)
- **Resources**: Manage BCM software images and device categories
- **Environment Variable Auth**: Seamless authentication via `BCM_ENDPOINT`, `BCM_USERNAME`, `BCM_PASSWORD`
- **100% Tested Examples**: All documentation examples validated with automated testing

## Quick Start

### Installation

```bash
# Clone repository
git clone https://github.com/hashi-demo-lab/terraform-provider-bcm.git
cd terraform-provider-bcm

# Build and install provider
make install
```

### Authentication

Set environment variables for BCM authentication:

```bash
export BCM_ENDPOINT="https://bcm.example.com:8081"
export BCM_USERNAME="admin"
export BCM_PASSWORD="your-password"
```

### Basic Usage

```hcl
terraform {
  required_providers {
    bcm = {
      source  = "hashicorp/bcm"
      version = "~> 0.1"
    }
  }
}

provider "bcm" {
  insecure_skip_verify = true  # Only for self-signed certificates
}

# Query all software images
data "bcm_cmpart_softwareimages" "all" {}

output "images" {
  value = data.bcm_cmpart_softwareimages.all.images
}
```

## Available Resources and Data Sources

### Data Sources

- `bcm_cmdevice_nodes` - Query compute nodes
- `bcm_cmdevice_categories` - Query device categories
- `bcm_cmpart_softwareimages` - Query software images
- `bcm_cmnet_networks` - Query network configurations

### Resources

- `bcm_cmpart_softwareimage` - Manage software images (create, clone, update, delete)
- `bcm_cmdevice_category` - Manage device categories

## Development

### Prerequisites

- Go 1.24+
- Terraform 1.5+
- Access to BCM cluster

### Building

```bash
# Format code
make fmt

# Run linters
make lint

# Build provider
make build

# Install locally
make install
```

### Testing

**Unit Tests:**
```bash
make test
```

**Acceptance Tests:**
```bash
export TF_ACC=1
export BCM_ENDPOINT="https://172.21.15.254:8081"
export BCM_USERNAME="root"
export BCM_PASSWORD="Hashicorp123!"

make testacc
```

**Example Testing:**
```bash
# Test all examples (data sources + resources)
BCM_ENDPOINT="..." BCM_USERNAME="..." BCM_PASSWORD="..." \
  SKIP_BUILD=true \
  ./scripts/test-examples.sh

# Test data sources only (parallel execution)
./scripts/test-examples.sh --data-sources-only

# Test resources only (sequential execution)
./scripts/test-examples.sh --resources-only

# Verbose mode (show full terraform output)
./scripts/test-examples.sh --verbose
```

**Expected Results:**
```
Examples tested: 17
  Passed: 17 ✓
  Failed: 0

Data sources: 10 examples (10s parallel execution)
Resources: 7 examples (26s sequential execution)
Total time: 40s
```

### Code Generation

```bash
# Generate documentation from examples
make generate
```

Documentation is auto-generated from:
- Schema definitions in `internal/provider/`
- Examples in `examples/data-sources/` and `examples/resources/`
- Generated docs output to `docs/`

## Project Structure

```
terraform-provider-bcm/
├── internal/provider/          # Provider implementation
│   ├── provider.go            # Provider configuration
│   ├── bcm_client.go          # BCM JSON-RPC API client
│   ├── data_source_*.go       # Data source implementations
│   ├── resource_*.go          # Resource implementations
│   └── *_test.go              # Acceptance tests
├── examples/                   # Terraform example configurations
│   ├── provider/              # Provider configuration examples
│   ├── data-sources/          # Data source usage examples
│   └── resources/             # Resource usage examples
├── docs/                       # Generated documentation (auto-generated)
├── scripts/                    # Test and utility scripts
│   └── test-examples.sh       # Automated example testing
├── specs/                      # Feature specifications
├── tools/                      # Code generation tools
└── main.go                     # Provider entry point
```

## Documentation

- **Examples**: See `examples/` directory for working Terraform configurations
- **API Patterns**: See `CLAUDE.md` for provider development patterns
- **Testing Guide**: See `AGENTS.md` for comprehensive testing infrastructure documentation
- **Generated Docs**: Run `make generate` to create documentation in `docs/`

## Architecture

### BCM API Client

The provider uses BCM's JSON-RPC API:

- **Authentication**: Cookie-based (`cm-login-token`)
- **Endpoint**: `https://[bcm-host]:8081/json`
- **Protocol**: JSON-RPC over HTTPS
- **Services**: CMDevice, CMPart, CMNet, etc.

### Environment Variable Support

Provider reads authentication from environment variables following HashiCorp best practices:

```go
// provider.go configuration
endpoint := data.Endpoint.ValueString()
if endpoint == "" {
    endpoint = os.Getenv("BCM_ENDPOINT")  // Fallback to env var
}
```

This pattern matches AWS, Azure, and GCP providers.

## Testing Infrastructure

### Automated Example Testing

All documentation examples are automatically tested:

- **Data sources**: Parallel execution (4 concurrent, ~10s)
- **Resources**: Sequential execution (~26s)
- **Validation**: `terraform init` → `validate` → `plan`
- **Integration**: `test-citest` runs full `apply` + `destroy` cycle

### Running Example Tests

```bash
# Set credentials
export BCM_ENDPOINT="https://172.21.15.254:8081"
export BCM_USERNAME="root"
export BCM_PASSWORD="Hashicorp123!"

# Run all tests
./scripts/test-examples.sh

# Skip provider build (faster)
SKIP_BUILD=true ./scripts/test-examples.sh
```

## Contributing

1. Follow TDD workflow: Write tests first, then implementation
2. Run `make fmt` and `make lint` before committing
3. Add examples in `examples/` for new resources/data sources
4. Run `make generate` to update documentation
5. Test examples with `./scripts/test-examples.sh`

## License

Mozilla Public License 2.0 (MPL-2.0)

## Support

For issues and feature requests, please use the GitHub issue tracker.
