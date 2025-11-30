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
		dependentList.WriteString(fmt.Sprintf("  - %s (uuid: %s)\n", identifiers[i].Name, identifiers[i].UUID))
	}

	if count > maxShow {
		dependentList.WriteString(fmt.Sprintf("  ... (showing first %d of %d)\n", maxShow, count))
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
