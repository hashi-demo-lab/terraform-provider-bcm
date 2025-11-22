# Scripts Directory

This directory contains utility scripts for the Terraform BCM Provider.

## Documentation Generation

### generate-docs-docker.sh

Generates Terraform provider documentation using Docker containers for consistent, reproducible builds across different environments.

**Usage:**

```bash
# Standard usage
make generate-docker

# Or run directly
./scripts/generate-docs-docker.sh
```

**What it does:**

1. Formats Terraform examples using `hashicorp/terraform:1.13.5` container
2. Generates copyright headers using Go in `golang:1.24` container
3. Generates provider documentation using `tfplugindocs` in `golang:1.24` container

**Environment Variables:**

- `GOLANG_IMAGE` - Go Docker image to use (default: `golang:1.24`)
- `TERRAFORM_VERSION` - Terraform version for formatting (default: `1.13.5`)
- `PROVIDER_NAME` - Provider name (default: `bcm`)

**Examples:**

```bash
# Use a different Go version
GOLANG_IMAGE=golang:1.23 make generate-docker

# Use a different Terraform version
TERRAFORM_VERSION=1.14.0 make generate-docker

# Combine both
GOLANG_IMAGE=golang:1.23 TERRAFORM_VERSION=1.14.0 ./scripts/generate-docs-docker.sh
```

**Requirements:**

- Docker installed and running
- Sufficient permissions to run Docker commands
- Network access to pull Docker images

**Output:**

Generated documentation will be placed in:
- `docs/data-sources/` - Data source documentation
- `docs/resources/` - Resource documentation
- `docs/index.md` - Provider index page

## Comparison: Local vs Docker Generation

### Local Generation (`make generate`)

**Pros:**
- Faster (no container overhead)
- Uses local Go cache

**Cons:**
- Requires local Go installation (1.24+)
- Requires local Terraform installation
- Results may vary based on local environment

**Use when:**
- Developing locally with Go already installed
- Need quick iteration cycles
- Working in a controlled environment

### Docker Generation (`make generate-docker`)

**Pros:**
- Consistent across all environments
- No local Go/Terraform installation required
- Reproducible builds
- Works in CI/CD pipelines

**Cons:**
- Slightly slower (container startup overhead)
- Requires Docker installation

**Use when:**
- Working in diverse team environments
- CI/CD pipeline execution
- Ensuring consistent documentation across contributors
- Debugging documentation generation issues
