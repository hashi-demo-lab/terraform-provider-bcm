// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
)

// ResourceIdentifier uniquely identifies a BCM resource in dependency checks.
type ResourceIdentifier struct {
	UUID string
	Name string
	Type string
}

// DependencyCheckResult encapsulates results of dependency validation.
type DependencyCheckResult struct {
	HasDependencies bool
	DependentCount  int
	DependentType   string
	Identifiers     []ResourceIdentifier
}

// CheckDevicesInCategory queries BCM for devices using the specified category.
// Returns dependency check result with list of dependent devices.
//
// This function queries all devices via CMDevice.getNodes and filters client-side
// by category UUID since BCM API does not support server-side filtering.
//
// Performance: Typical response < 1s, large clusters (1000+ devices) < 5s.
func CheckDevicesInCategory(ctx context.Context, client *BCMClient, categoryUUID string) (*DependencyCheckResult, error) {
	// Query all devices from BCM
	body, err := client.CallJSONRPC(ctx, "CMDevice", "getNodes")
	if err != nil {
		return nil, fmt.Errorf("failed to query devices: %w", err)
	}

	// Parse response as array of device objects
	var devices []map[string]interface{}
	if err := json.Unmarshal(body, &devices); err != nil {
		return nil, fmt.Errorf("failed to parse devices response: %w", err)
	}

	// Filter devices by category UUID (client-side filtering)
	var dependentDevices []ResourceIdentifier
	for _, device := range devices {
		// Extract category field (may be nil if device has no category)
		categoryField, ok := device["category"]
		if !ok || categoryField == nil {
			continue
		}

		// Check if category matches target UUID
		deviceCategoryUUID, ok := categoryField.(string)
		if !ok {
			continue
		}

		if deviceCategoryUUID == categoryUUID {
			// Extract device identifiers
			uuid, _ := device["uuid"].(string)
			hostname, _ := device["hostname"].(string)

			dependentDevices = append(dependentDevices, ResourceIdentifier{
				UUID: uuid,
				Name: hostname,
				Type: "device",
			})
		}
	}

	return &DependencyCheckResult{
		HasDependencies: len(dependentDevices) > 0,
		DependentCount:  len(dependentDevices),
		DependentType:   "devices",
		Identifiers:     dependentDevices,
	}, nil
}

// CheckCategoriesUsingImage queries BCM for categories using the specified software image.
// Returns dependency check result with list of dependent categories.
//
// This function queries all categories via CMDevice.getCategories and filters client-side
// by software image name (not UUID) since BCM API does not support server-side filtering.
//
// IMPORTANT: The category.softwareimage field contains the IMAGE NAME (string), not UUID.
//
// Performance: Typical response < 1s, large clusters (100+ categories) < 2s.
func CheckCategoriesUsingImage(ctx context.Context, client *BCMClient, imageName string) (*DependencyCheckResult, error) {
	// Query all categories from BCM
	body, err := client.CallJSONRPC(ctx, "CMDevice", "getCategories")
	if err != nil {
		return nil, fmt.Errorf("failed to query categories: %w", err)
	}

	// Parse response as array of category objects
	var categories []map[string]interface{}
	if err := json.Unmarshal(body, &categories); err != nil {
		return nil, fmt.Errorf("failed to parse categories response: %w", err)
	}

	// Filter categories by software image name (client-side filtering)
	var dependentCategories []ResourceIdentifier
	for _, category := range categories {
		// Extract softwareimage field (may be nil if category has no image)
		imageField, ok := category["softwareimage"]
		if !ok || imageField == nil {
			continue
		}

		// Check if software image matches target name
		categorySoftwareImage, ok := imageField.(string)
		if !ok {
			continue
		}

		if categorySoftwareImage == imageName {
			// Extract category identifiers
			uuid, _ := category["uuid"].(string)
			name, _ := category["name"].(string)

			dependentCategories = append(dependentCategories, ResourceIdentifier{
				UUID: uuid,
				Name: name,
				Type: "category",
			})
		}
	}

	return &DependencyCheckResult{
		HasDependencies: len(dependentCategories) > 0,
		DependentCount:  len(dependentCategories),
		DependentType:   "categories",
		Identifiers:     dependentCategories,
	}, nil
}
