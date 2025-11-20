#!/bin/bash
# Manual cleanup script for BCM test artifacts
# Usage: ./cleanup_test_images.sh

set -e

# Require environment variables
: ${BCM_ENDPOINT:?"BCM_ENDPOINT not set"}
: ${BCM_USERNAME:?"BCM_USERNAME not set"}
: ${BCM_PASSWORD:?"BCM_PASSWORD not set"}

echo "=== BCM Test Artifacts Cleanup ==="
echo "Endpoint: $BCM_ENDPOINT"
echo "Username: $BCM_USERNAME"
echo ""

# Create a simple Go cleanup script
cat > /tmp/cleanup_bcm.go << 'GOEOF'
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

func main() {
	endpoint := os.Getenv("BCM_ENDPOINT")
	username := os.Getenv("BCM_USERNAME")
	password := os.Getenv("BCM_PASSWORD")

	fmt.Println("Connecting to BCM...")
	
	// This would use the BCM client - for now just list what we'd clean
	fmt.Println("Would delete all images matching:")
	fmt.Println("  - test-basic-image-*")
	fmt.Println("  - test-full-config-*")
	fmt.Println("  - test-update-kernel-*")
	fmt.Println("  - test-update-modules-*")
	fmt.Println("  - test-update-sol-*")
	fmt.Println("  - test-with-modules-*")
	fmt.Println("")
	fmt.Println("To manually cleanup, use BCM web UI or API calls")
}
GOEOF

go run /tmp/cleanup_bcm.go
rm /tmp/cleanup_bcm.go

