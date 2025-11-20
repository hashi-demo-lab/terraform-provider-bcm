package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/hashi-demo-lab/terraform-provider-bcm/internal/provider"
)

func main() {
	endpoint := os.Getenv("BCM_ENDPOINT")
	username := os.Getenv("BCM_USERNAME")
	password := os.Getenv("BCM_PASSWORD")

	if endpoint == "" || username == "" || password == "" {
		log.Fatal("BCM_ENDPOINT, BCM_USERNAME, and BCM_PASSWORD must be set")
	}

	client, err := provider.NewBCMClient(context.Background(), endpoint, username, password, true, 30)
	if err != nil {
		log.Fatalf("Failed to create BCM client: %v", err)
	}

	// Get default-image UUID
	body, err := client.CallJSONRPC(context.Background(), "CMPart", "getSoftwareImage", "default-image")
	if err != nil {
		log.Fatalf("Failed to get default-image: %v", err)
	}

	var defaultImg map[string]interface{}
	if err := json.Unmarshal(body, &defaultImg); err != nil {
		log.Fatalf("Failed to parse default-image: %v", err)
	}

	defaultUUID := defaultImg["uuid"].(string)
	kernelVersion := defaultImg["kernelVersion"].(string)

	// Test different image names
	testCases := []struct {
		name string
		path string
	}{
		{"test-a", "/cm/images/test-a"},
		{"test-basic-image", "/cm/images/test-basic-image"},
		{"test-full-config", "/cm/images/test-full-config"},
	}

	for _, tc := range testCases {
		fmt.Printf("\nTesting creation of image: %s at path: %s\n", tc.name, tc.path)

		// First check if it exists and delete if needed
		body, err := client.CallJSONRPC(context.Background(), "CMPart", "getSoftwareImage", tc.name)
		if err == nil {
			var img map[string]interface{}
			if json.Unmarshal(body, &img) == nil {
				if uuid, ok := img["uuid"].(string); ok && uuid != "" {
					fmt.Printf("  Deleting existing image %s\n", tc.name)
					_, _ = client.CallJSONRPC(context.Background(), "CMPart", "removeSoftwareImage", uuid, false, false, false)
				}
			}
		}

		// Try to create the image
		entity := map[string]interface{}{
			"uuid":          uuid.New().String(),
			"baseType":      "SoftwareImage",
			"childType":     "",
			"modified":      true,
			"to_be_removed": false,
			"revision":      "",
			"name":          tc.name,
			"path":          tc.path,
			"kernelVersion": kernelVersion,
			"originalImage": defaultUUID,
			"modules":       []interface{}{},
		}

		_, err = client.CallJSONRPC(context.Background(), "CMPart", "addSoftwareImage", entity, false)
		if err != nil {
			fmt.Printf("  ❌ Failed: %v\n", err)
		} else {
			fmt.Printf("  ✅ Success!\n")
			// Clean up
			body, _ := client.CallJSONRPC(context.Background(), "CMPart", "getSoftwareImage", tc.name)
			var img map[string]interface{}
			if json.Unmarshal(body, &img) == nil {
				if uuid, ok := img["uuid"].(string); ok && uuid != "" {
					_, _ = client.CallJSONRPC(context.Background(), "CMPart", "removeSoftwareImage", uuid, false, false, false)
				}
			}
		}
	}
}