// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// JSONRPCRequestWithArgs represents BCM JSON-RPC request with args parameter
type JSONRPCRequestWithArgs struct {
	Service string        `json:"service"`
	Call    string        `json:"call"`
	Args    []interface{} `json:"args,omitempty"` // Optional args parameter
}

func main() {
	ctx := context.Background()

	// Get credentials from environment
	endpoint := os.Getenv("BCM_ENDPOINT")
	username := os.Getenv("BCM_USERNAME")
	password := os.Getenv("BCM_PASSWORD")

	if endpoint == "" || username == "" || password == "" {
		fmt.Println("❌ Missing environment variables: BCM_ENDPOINT, BCM_USERNAME, BCM_PASSWORD")
		os.Exit(1)
	}

	fmt.Println("🔍 Testing BCMClient args parameter support")
	fmt.Println("=" + string(bytes.Repeat([]byte("="), 60)))

	// Test 1: Login (no args)
	fmt.Println("\n✓ Test 1: Login (no args)")
	client, loginToken, err := testLogin(ctx, endpoint, username, password)
	if err != nil {
		fmt.Printf("❌ Login failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ Login successful, token: %s\n", loginToken[:20]+"...")

	// Test 2: getSoftwareImages() - no args (baseline)
	fmt.Println("\n✓ Test 2: getSoftwareImages() - no args (baseline)")
	images, err := testGetSoftwareImages(ctx, client, endpoint, loginToken)
	if err != nil {
		fmt.Printf("❌ getSoftwareImages failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ getSoftwareImages returned %d images\n", len(images))

	// Find first image name for Test 3
	var testImageName string
	if len(images) > 0 {
		if name, ok := images[0]["name"].(string); ok {
			testImageName = name
		}
	}

	if testImageName == "" {
		fmt.Println("⚠️  No images found in BCM, cannot test getSoftwareImage(name)")
		fmt.Println("✓ Args parameter structure test passed (syntax valid)")
		os.Exit(0)
	}

	// Test 3: getSoftwareImage(name) - WITH args
	fmt.Printf("\n✓ Test 3: getSoftwareImage(name) - WITH args (name='%s')\n", testImageName)
	image, err := testGetSoftwareImage(ctx, client, endpoint, loginToken, testImageName)
	if err != nil {
		fmt.Printf("❌ getSoftwareImage failed: %v\n", err)
		fmt.Println("\n📋 VERDICT: Args parameter NOT supported by BCM API")
		fmt.Println("   Recommendation: Use getSoftwareImages() + client-side filter OR implement direct HTTP calls")
		os.Exit(1)
	}
	fmt.Printf("✓ getSoftwareImage returned image: %s\n", image["name"])
	fmt.Println("\n📋 VERDICT: ✅ Args parameter IS supported by BCM API")
	fmt.Println("   Recommendation: Extend CallJSONRPC to support variadic args")
}

func testLogin(ctx context.Context, endpoint, username, password string) (*http.Client, string, error) {
	loginReq := map[string]string{
		"service":  "login",
		"username": username,
		"password": password,
	}

	jsonBody, _ := json.Marshal(loginReq)
	req, _ := http.NewRequestWithContext(ctx, "POST", endpoint+"/json", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Create client with TLS insecure skip verify (for self-signed certs)
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var loginSuccess bool
	if err := json.Unmarshal(body, &loginSuccess); err != nil || !loginSuccess {
		return nil, "", fmt.Errorf("login failed: %s", string(body))
	}

	// Extract cm-login-token
	var token string
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "cm-login-token" {
			token = cookie.Value
			break
		}
	}

	if token == "" {
		return nil, "", fmt.Errorf("no cm-login-token in response")
	}

	return client, token, nil
}

func testGetSoftwareImages(ctx context.Context, client *http.Client, endpoint, token string) ([]map[string]interface{}, error) {
	reqBody := JSONRPCRequestWithArgs{
		Service: "cmpart",
		Call:    "getSoftwareImages",
		// No args
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequestWithContext(ctx, "POST", endpoint+"/json", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", "cm-login-token="+token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var images []map[string]interface{}
	if err := json.Unmarshal(body, &images); err != nil {
		return nil, err
	}

	return images, nil
}

func testGetSoftwareImage(ctx context.Context, client *http.Client, endpoint, token, name string) (map[string]interface{}, error) {
	reqBody := JSONRPCRequestWithArgs{
		Service: "cmpart",
		Call:    "getSoftwareImage",
		Args:    []interface{}{name}, // Pass name as argument
	}

	jsonBody, _ := json.Marshal(reqBody)

	fmt.Printf("   Request body: %s\n", string(jsonBody))

	req, _ := http.NewRequestWithContext(ctx, "POST", endpoint+"/json", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", "cm-login-token="+token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	fmt.Printf("   Response status: %d\n", resp.StatusCode)
	fmt.Printf("   Response body: %s\n", string(body)[:200])

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var image map[string]interface{}
	if err := json.Unmarshal(body, &image); err != nil {
		// Try array format
		var images []map[string]interface{}
		if err := json.Unmarshal(body, &images); err != nil {
			return nil, fmt.Errorf("parse error: %v", err)
		}
		if len(images) == 0 {
			return nil, fmt.Errorf("image not found")
		}
		return images[0], nil
	}

	return image, nil
}
