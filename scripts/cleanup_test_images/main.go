// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// cleanup_test_images.go - Manual cleanup tool for BCM test artifacts
// Usage: BCM_ENDPOINT="https://..." BCM_USERNAME="root" BCM_PASSWORD="..." go run cleanup_test_images.go

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/hashi-demo-lab/terraform-provider-bcm/internal/provider"
)

func main() {
	endpoint := os.Getenv("BCM_ENDPOINT")
	username := os.Getenv("BCM_USERNAME")
	password := os.Getenv("BCM_PASSWORD")

	if endpoint == "" || username == "" || password == "" {
		fmt.Println("Error: BCM_ENDPOINT, BCM_USERNAME, and BCM_PASSWORD must be set")
		os.Exit(1)
	}

	fmt.Println("=== BCM Test Artifacts Cleanup ===")
	fmt.Printf("Endpoint: %s\n", endpoint)
	fmt.Printf("Username: %s\n\n", username)

	// Create BCM client
	fmt.Println("Connecting to BCM...")
	client, err := provider.NewBCMClient(context.Background(), endpoint, username, password, true, 30)
	if err != nil {
		fmt.Printf("Error: Failed to connect to BCM: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Connected successfully")

	// Get all software images
	fmt.Println("Fetching all software images...")
	body, err := client.CallJSONRPC(context.Background(), "CMPart", "getSoftwareImages")
	if err != nil {
		fmt.Printf("Error: Failed to get software images: %v\n", err)
		os.Exit(1)
	}

	var images []map[string]interface{}
	if err := json.Unmarshal(body, &images); err != nil {
		fmt.Printf("Error: Failed to parse response: %v\n", err)
		os.Exit(1)
	}

	// Test image prefixes to clean
	testPrefixes := []string{
		"test-basic-image",
		"test-full-config",
		"test-update-kernel",
		"test-update-modules",
		"test-update-sol",
		"test-with-modules",
	}

	// Find test images
	var testImages []map[string]interface{}
	for _, img := range images {
		name, ok := img["name"].(string)
		if !ok {
			continue
		}
		for _, prefix := range testPrefixes {
			if strings.HasPrefix(name, prefix) {
				testImages = append(testImages, img)
				break
			}
		}
	}

	if len(testImages) == 0 {
		fmt.Println("✓ No test images found to clean up")
		return
	}

	fmt.Printf("Found %d test images to delete:\n\n", len(testImages))
	for _, img := range testImages {
		name, ok := img["name"].(string)
		if !ok {
			continue
		}
		uuid, _ := img["uuid"].(string)
		fmt.Printf("  - %s (UUID: %s)\n", name, uuid)
	}
	fmt.Println()

	// Confirm deletion
	fmt.Print("Delete these images? (yes/no): ")
	var response string
	_, err = fmt.Scanln(&response)
	if err != nil {
		fmt.Printf("Error reading input: %v\n", err)
		return
	}

	if strings.ToLower(response) != "yes" {
		fmt.Println("Aborted")
		return
	}

	// Delete images
	fmt.Println("\nDeleting test images...")
	deleted := 0
	failed := 0

	for _, img := range testImages {
		name, ok := img["name"].(string)
		if !ok {
			fmt.Printf("  ✗ unknown - invalid name field\n")
			failed++
			continue
		}
		uuid, uuidOk := img["uuid"].(string)
		if !uuidOk || uuid == "" {
			fmt.Printf("  ✗ %s - no UUID\n", name)
			failed++
			continue
		}

		_, err := client.CallJSONRPC(context.Background(), "CMPart", "removeSoftwareImage", uuid, false, false, false)
		if err != nil {
			fmt.Printf("  ✗ %s - %v\n", name, err)
			failed++
		} else {
			fmt.Printf("  ✓ %s\n", name)
			deleted++
		}
	}

	fmt.Printf("\n=== Summary ===\n")
	fmt.Printf("Deleted: %d\n", deleted)
	fmt.Printf("Failed:  %d\n", failed)
}
