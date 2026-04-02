// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"strings"
)

// BuildDependencyError creates a formatted, actionable error message for dependency violations.
// The message includes:
// - Clear statement of the problem
// - List of dependent resources (max 10, truncated if more)
// - Resolution options specific to the resource type
//
// Parameters:
//   - resourceType: Type of resource being deleted (e.g., "Category", "Software Image")
//   - resourceName: Name of the resource being deleted
//   - dependentType: Type of dependent resources (e.g., "device", "category")
//   - identifiers: List of dependent resource identifiers
//
// Returns formatted multi-line error message suitable for Terraform diagnostics.
func BuildDependencyError(resourceType, resourceName, dependentType string, identifiers []ResourceIdentifier) string {
	count := len(identifiers)

	// Build dependent resource list (max 10, truncate if more)
	var dependentList strings.Builder
	maxShow := 10
	shown := minInt(count, maxShow)

	for i := 0; i < shown; i++ {
		fmt.Fprintf(&dependentList, "  - %s (uuid: %s)\n", identifiers[i].Name, identifiers[i].UUID)
	}

	if count > maxShow {
		fmt.Fprintf(&dependentList, "  ... (showing first %d of %d)\n", maxShow, count)
	}

	// Build resolution options based on resource type
	var resolutionOptions string
	switch resourceType {
	case "Category":
		resolutionOptions = `Resolution options:
  1. Reassign devices to another category before deleting
  2. Delete the dependent devices first
  3. Set 'force = true' to delete anyway (WARNING: will orphan device references)`

	case "Software Image":
		resolutionOptions = `Resolution options:
  1. Update categories to use a different software image before deleting
  2. Delete the dependent categories first (after removing their devices)
  3. Set 'force = true' to delete anyway (WARNING: will orphan category references)`

	default:
		resolutionOptions = fmt.Sprintf(`Resolution options:
  1. Remove dependencies on this %s before deleting
  2. Delete the dependent %s first
  3. Set 'force = true' to delete anyway (WARNING: will orphan %s references)`,
			resourceType, dependentType, dependentType)
	}

	// Format dependent type for grammatical correctness
	var dependentTypeDisplay string
	if count == 1 {
		// Singular form
		dependentTypeDisplay = dependentType
	} else {
		// Plural form with proper grammar
		if strings.HasSuffix(dependentType, "y") {
			// category -> categor(ies)
			dependentTypeDisplay = dependentType[:len(dependentType)-1] + "(ies)"
		} else {
			// device -> device(s)
			dependentTypeDisplay = dependentType + "(s)"
		}
	}

	// Build complete error message
	return fmt.Sprintf(
		"%s '%s' cannot be deleted because it has %d %s assigned.\n\n"+
			"Dependent %s:\n%s\n"+
			"%s",
		resourceType,
		resourceName,
		count,
		dependentTypeDisplay,
		dependentType,
		dependentList.String(),
		resolutionOptions,
	)
}

// BuildForceDeletionWarning creates a warning message for force deletion operations.
// This warning is logged when a resource is deleted with force=true, potentially
// creating orphaned references in dependent resources.
//
// Parameters:
//   - resourceType: Type of resource being force deleted (e.g., "Category", "Software Image")
//   - resourceName: Name of the resource being force deleted
//
// Returns formatted warning message suitable for Terraform diagnostics or logging.
func BuildForceDeletionWarning(resourceType, resourceName string) string {
	var impact string
	switch resourceType {
	case "Category":
		impact = `Potential Impact:
  - Devices assigned to this category will have invalid category references
  - Device provisioning may fail until devices are reassigned to a valid category`

	case "Software Image":
		impact = `Potential Impact:
  - Categories using this image will have invalid software image references
  - Device provisioning may fail for affected categories`

	default:
		impact = fmt.Sprintf(`Potential Impact:
  - Dependent resources may have invalid %s references
  - Operations may fail until dependencies are resolved`, resourceType)
	}

	return fmt.Sprintf(
		"%s '%s' is being deleted with force=true. This may create orphaned references in dependent resources.\n\n"+
			"%s\n\n"+
			"This operation cannot be undone.",
		resourceType,
		resourceName,
		impact,
	)
}

// minInt returns the minimum of two integers (Go 1.21+ has this in stdlib, but for compatibility).
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// =============================================================================
// CRUD Operation Error Messages
// =============================================================================

// BuildAPIError creates a standardized error message for API call failures.
// Use this for consistent error formatting across all resources.
//
// Parameters:
//   - operation: The operation that failed (e.g., "Create", "Read", "Update", "Delete")
//   - resourceType: Type of resource (e.g., "Category", "Device")
//   - resourceName: Name or identifier of the resource
//   - err: The underlying error
func BuildAPIError(operation, resourceType, resourceName string, err error) string {
	return fmt.Sprintf(
		"Failed to %s %s '%s': %s\n\n"+
			"Verify BCM endpoint, credentials, and network connectivity.",
		strings.ToLower(operation),
		resourceType,
		resourceName,
		err.Error(),
	)
}

// BuildParseError creates a standardized error message for JSON parsing failures.
//
// Parameters:
//   - operation: The operation context (e.g., "Create", "Read")
//   - resourceType: Type of resource
//   - err: The underlying parse error
func BuildParseError(operation, resourceType string, err error) string {
	return fmt.Sprintf(
		"Failed to parse %s response for %s operation: %s\n\n"+
			"This may indicate an incompatible BCM API version.",
		resourceType,
		strings.ToLower(operation),
		err.Error(),
	)
}

// BuildNotFoundError creates a standardized "not found" error message.
//
// Parameters:
//   - resourceType: Type of resource (e.g., "Category", "Device")
//   - identifier: Name or UUID of the resource
func BuildNotFoundError(resourceType, identifier string) string {
	return fmt.Sprintf(
		"%s '%s' not found in BCM. It may have been deleted externally or never existed.",
		resourceType,
		identifier,
	)
}

// BuildValidationAPIError creates an error message for validation API failures.
//
// Parameters:
//   - resourceType: Type of resource being validated
//   - resourceName: Name of the resource
//   - err: The underlying error
func BuildValidationAPIError(resourceType, resourceName string, err error) string {
	return fmt.Sprintf(
		"Could not validate %s '%s': %s\n\n"+
			"The validation API call failed. Check BCM connectivity and credentials.",
		resourceType,
		resourceName,
		err.Error(),
	)
}

// =============================================================================
// Error Title Constants - Standardized error titles
// =============================================================================

// ErrorTitle returns a standardized error title for CRUD operations.
// Format: "<ResourceType> <Operation> Failed".
func ErrorTitle(resourceType, operation string) string {
	return fmt.Sprintf("%s %s Failed", resourceType, operation)
}

// ErrorTitleNotFound returns a standardized "not found" error title.
func ErrorTitleNotFound(resourceType string) string {
	return fmt.Sprintf("%s Not Found", resourceType)
}

// ErrorTitleParse returns a standardized parse error title.
func ErrorTitleParse(resourceType string) string {
	return fmt.Sprintf("%s Parse Error", resourceType)
}

// ErrorTitleValidation returns a standardized validation error title.
func ErrorTitleValidation() string {
	return "Validation API Error"
}
