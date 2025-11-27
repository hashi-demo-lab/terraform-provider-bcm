package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

// TestValidateCategory_DirectAPICall tests the validateCategory BCM API method
// directly using our BCMClient to verify what validation errors are returned.
func TestValidateCategory_DirectAPICall(t *testing.T) {
	// Skip if not running acceptance tests
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set, skipping acceptance test")
	}

	ctx := context.Background()

	// Create BCM client
	client := createTestBCMClient(t)

	// Get default software image
	imagesBody, err := client.CallJSONRPC(ctx, "cmpart", "getSoftwareImages")
	if err != nil {
		t.Fatalf("Failed to get software images: %v", err)
	}

	var images []map[string]interface{}
	if err := json.Unmarshal(imagesBody, &images); err != nil {
		t.Fatalf("Failed to parse images: %v", err)
	}

	var defaultImageUUID string
	for _, img := range images {
		if name, ok := img["name"].(string); ok {
			if name == "default-image" {
				defaultImageUUID = img["uuid"].(string)
				break
			}
		}
	}
	if defaultImageUUID == "" && len(images) > 0 {
		defaultImageUUID = images[0]["uuid"].(string)
	}
	t.Logf("Using software image UUID: %s", defaultImageUUID)

	// Get default network
	networksBody, err := client.CallJSONRPC(ctx, "cmnet", "getNetworks")
	if err != nil {
		t.Fatalf("Failed to get networks: %v", err)
	}

	var networks []map[string]interface{}
	if err := json.Unmarshal(networksBody, &networks); err != nil {
		t.Fatalf("Failed to parse networks: %v", err)
	}

	var defaultNetworkUUID string
	if len(networks) > 0 {
		defaultNetworkUUID = networks[0]["uuid"].(string)
	}
	t.Logf("Using network UUID: %s", defaultNetworkUUID)

	// Test 1: Valid category entity
	t.Run("ValidCategory", func(t *testing.T) {
		entity := map[string]interface{}{
			"baseType":          "Category",
			"childType":         "",
			"name":              "tftest-validate-go-valid",
			"uuid":              "",
			"managementNetwork": defaultNetworkUUID,
			"softwareImageProxy": map[string]interface{}{
				"baseType":            "SoftwareImageProxy",
				"childType":           "",
				"parentSoftwareImage": defaultImageUUID,
				"revisionID":          -1,
			},
		}

		errors, err := client.ValidateEntity(ctx, "CMDevice", "validateCategory", entity, true)
		if err != nil {
			t.Fatalf("ValidateEntity failed: %v", err)
		}

		t.Logf("Validation errors (after Zero UUID filtering): %d", len(errors))
		for _, e := range errors {
			t.Logf("  [%s] %s: %s", e.Severity, e.Field, e.Message)
		}

		if len(errors) > 0 {
			t.Errorf("Expected no validation errors for valid category, got %d", len(errors))
		}
	})

	// Test 2: Empty name
	t.Run("EmptyName", func(t *testing.T) {
		entity := map[string]interface{}{
			"baseType":          "Category",
			"childType":         "",
			"name":              "",
			"uuid":              "",
			"managementNetwork": defaultNetworkUUID,
			"softwareImageProxy": map[string]interface{}{
				"baseType":            "SoftwareImageProxy",
				"childType":           "",
				"parentSoftwareImage": defaultImageUUID,
				"revisionID":          -1,
			},
		}

		errors, err := client.ValidateEntity(ctx, "CMDevice", "validateCategory", entity, true)
		if err != nil {
			t.Fatalf("ValidateEntity failed: %v", err)
		}

		t.Logf("Validation errors: %d", len(errors))
		for _, e := range errors {
			t.Logf("  [%s] %s: %s", e.Severity, e.Field, e.Message)
		}

		// Should have error for empty name
		foundNameError := false
		for _, e := range errors {
			if e.Field == "name" {
				foundNameError = true
				break
			}
		}
		if !foundNameError {
			t.Error("Expected validation error for empty name")
		}
	})

	// Test 3: Duplicate name (default already exists)
	t.Run("DuplicateName", func(t *testing.T) {
		entity := map[string]interface{}{
			"baseType":          "Category",
			"childType":         "",
			"name":              "default", // Already exists
			"uuid":              "",
			"managementNetwork": defaultNetworkUUID,
			"softwareImageProxy": map[string]interface{}{
				"baseType":            "SoftwareImageProxy",
				"childType":           "",
				"parentSoftwareImage": defaultImageUUID,
				"revisionID":          -1,
			},
		}

		errors, err := client.ValidateEntity(ctx, "CMDevice", "validateCategory", entity, true)
		if err != nil {
			t.Fatalf("ValidateEntity failed: %v", err)
		}

		t.Logf("Validation errors: %d", len(errors))
		for _, e := range errors {
			t.Logf("  [%s] %s: %s", e.Severity, e.Field, e.Message)
		}

		// Should have error for duplicate name
		foundDuplicateError := false
		for _, e := range errors {
			if e.Field == "name" && e.Message == "A category with that name already exists" {
				foundDuplicateError = true
				break
			}
		}
		if !foundDuplicateError {
			t.Error("Expected validation error for duplicate name")
		}
	})

	// Test 4: Missing parentSoftwareImage
	t.Run("MissingParentSoftwareImage", func(t *testing.T) {
		entity := map[string]interface{}{
			"baseType":          "Category",
			"childType":         "",
			"name":              "tftest-validate-go-missing-img",
			"uuid":              "",
			"managementNetwork": defaultNetworkUUID,
			// No softwareImageProxy - should fail validation
		}

		errors, err := client.ValidateEntity(ctx, "CMDevice", "validateCategory", entity, true)
		if err != nil {
			t.Fatalf("ValidateEntity failed: %v", err)
		}

		t.Logf("Validation errors: %d", len(errors))
		for _, e := range errors {
			t.Logf("  [%s] %s: %s", e.Severity, e.Field, e.Message)
		}

		// Should have error for missing parentSoftwareImage
		foundImageError := false
		for _, e := range errors {
			if e.Field == "parentSoftwareImage" {
				foundImageError = true
				break
			}
		}
		if !foundImageError {
			t.Error("Expected validation error for missing parentSoftwareImage")
		}
	})

	// Test 5: Validate existing category (with UUID)
	t.Run("ValidateExistingCategory", func(t *testing.T) {
		// Get the default category
		catBody, err := client.CallJSONRPC(ctx, "cmdevice", "getCategory", "default")
		if err != nil {
			t.Fatalf("Failed to get default category: %v", err)
		}

		var defaultCat map[string]interface{}
		if err := json.Unmarshal(catBody, &defaultCat); err != nil {
			t.Fatalf("Failed to parse category: %v", err)
		}

		t.Logf("Validating existing category: %s (uuid: %s)", defaultCat["name"], defaultCat["uuid"])

		// Validate with isCreate=false since it already exists
		errors, err := client.ValidateEntity(ctx, "CMDevice", "validateCategory", defaultCat, false)
		if err != nil {
			t.Fatalf("ValidateEntity failed: %v", err)
		}

		t.Logf("Validation errors: %d", len(errors))
		for _, e := range errors {
			t.Logf("  [%s] %s: %s", e.Severity, e.Field, e.Message)
		}

		if len(errors) > 0 {
			t.Errorf("Expected no validation errors for existing valid category, got %d", len(errors))
		}
	})

	fmt.Println("\n=== validateCategory Test Summary ===")
	fmt.Println("BCM validates: empty name, duplicate name, missing parentSoftwareImage")
	fmt.Println("BCM does NOT validate: invalid chars in name, invalid UUIDs, invalid URLs")
}
