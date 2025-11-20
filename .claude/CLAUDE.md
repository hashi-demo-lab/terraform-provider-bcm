# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Terraform Provider scaffold built on the Terraform Plugin Framework. It follows TDD (Test-Driven Development) principles with parallel execution patterns for resources, data sources, and ephemeral resources.

## Development Commands

### Building and Installing

- `make install` - Build and install the provider (default target includes fmt, lint, install, generate)
- `make build` - Build the provider without installing
- `go install` - Direct install command

### Testing

- `make test` - Run unit tests (120s timeout, parallel=10)
- `make testacc` - Run acceptance tests with `TF_ACC=1` (120m timeout)
- `TF_ACC=1 go test -v -timeout 120m ./internal/provider/` - Run acceptance tests for specific package

### Code Quality

- `make fmt` - Format Go code with `gofmt -s -w -e .`
- `make lint` - Run golangci-lint
- `pre-commit install` - Install pre-commit hooks (first time setup)
- `pre-commit run --all-files` - Run all pre-commit hooks manually

### Documentation

- `make generate` - Generate provider documentation (runs `cd tools; go generate ./...`)
- This generates:
  - Copyright headers via copywrite
  - Terraform documentation via tfplugindocs
  - Formats examples/ directory with `terraform fmt`

## Project Structure

- `internal/provider/` - Provider implementation (resources, data sources, ephemeral resources, functions)
- `examples/` - Example Terraform configurations for documentation
- `docs/` - Generated documentation (auto-generated, don't edit manually)
- `tools/tools.go` - Go generate directives for copyright headers and documentation
- `main.go` - Provider binary entry point

## TDD Workflow

This project follows RED-GREEN-REFACTOR cycles with parallel execution:

1. **RED**: Write failing acceptance tests first (`*_test.go` files)
2. **GREEN**: Write minimal CRUD implementation to pass tests
3. **REFACTOR**: Improve code quality while keeping tests green
4. **Document**: Run `make generate` to update documentation

## Key Dependencies

- **Framework**: `terraform-plugin-framework` (v1.16.1)
- **Testing**: `terraform-plugin-testing` (v1.13.3)
- **Documentation**: `terraform-plugin-docs` via tools/tools.go
- **Go Version**: 1.24+

## Primary Reference for TDD Patterns

See the root `./AGENTS.md` for comprehensive Terraform Provider TDD patterns, parallel execution strategies, and development best practices.

@/workspace/AGENTS.md

## Additional Component-Specific Guidance

Check for AGENTS.md files in subdirectories for module-specific implementation guides.

## Important Notes

- **Use subagents liberally**: Concurrent subagents can be used for performance and isolation with parallel tool calls in the same message
- **AskUserQuestion tool**: Use when clarification is needed (especially during speckit.clarify)
- **Pre-commit hooks**: Always run before commits to ensure code quality
- **Acceptance tests**: Set `TF_ACC=1` environment variable to enable acceptance tests (they create real resources)

## Updating Documentation

When you discover new implementation details, debugging insights, or architectural patterns:

- **Update existing AGENTS.md files** for component-specific guidance
- **Create new AGENTS.md files** in relevant directories for new areas
- **Add valuable insights** such as common pitfalls, debugging techniques, or implementation patterns
