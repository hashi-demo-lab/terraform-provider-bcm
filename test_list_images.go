package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

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

	// Get all software images
	body, err := client.CallJSONRPC(context.Background(), "CMPart", "getSoftwareImages")
	if err != nil {
		log.Fatalf("Failed to get software images: %v", err)
	}

	var images []map[string]interface{}
	if err := json.Unmarshal(body, &images); err != nil {
		log.Fatalf("Failed to parse response: %v", err)
	}

	fmt.Printf("Found %d software images:\n\n", len(images))
	for _, img := range images {
		fmt.Printf("Name: %s\n", img["name"])
		fmt.Printf("  Path: %s\n", img["path"])
		fmt.Printf("  UUID: %s\n", img["uuid"])
		fmt.Printf("  Kernel Version: %s\n", img["kernelVersion"])
		fmt.Printf("  Notes: %s\n", img["notes"])
		fmt.Println()
	}
}