// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// =============================================================================
// BCMResourceBase Tests
// =============================================================================

func TestBCMResourceBase_ConfigureResource_NilProviderData(t *testing.T) {
	t.Parallel()

	base := &BCMResourceBase{}
	req := resource.ConfigureRequest{
		ProviderData: nil,
	}
	resp := &resource.ConfigureResponse{}

	base.ConfigureResource(req, resp)

	// Should not set client and should not add errors
	if base.Client != nil {
		t.Error("Client should be nil when ProviderData is nil")
	}
	if resp.Diagnostics.HasError() {
		t.Errorf("Should not have errors when ProviderData is nil: %v", resp.Diagnostics)
	}
}

func TestBCMResourceBase_ConfigureResource_ValidClient(t *testing.T) {
	t.Parallel()

	// Create a mock client
	mockClient := &BCMClient{
		Endpoint: "https://test.example.com",
	}

	base := &BCMResourceBase{}
	req := resource.ConfigureRequest{
		ProviderData: mockClient,
	}
	resp := &resource.ConfigureResponse{}

	base.ConfigureResource(req, resp)

	// Should set client and should not add errors
	if base.Client != mockClient {
		t.Errorf("Client should be set to mock client, got %v", base.Client)
	}
	if resp.Diagnostics.HasError() {
		t.Errorf("Should not have errors with valid BCMClient: %v", resp.Diagnostics)
	}
}

func TestBCMResourceBase_ConfigureResource_WrongType(t *testing.T) {
	t.Parallel()

	// Pass wrong type as ProviderData
	wrongType := "not a BCMClient"

	base := &BCMResourceBase{}
	req := resource.ConfigureRequest{
		ProviderData: wrongType,
	}
	resp := &resource.ConfigureResponse{}

	base.ConfigureResource(req, resp)

	// Should not set client and should add error
	if base.Client != nil {
		t.Error("Client should be nil when ProviderData is wrong type")
	}
	if !resp.Diagnostics.HasError() {
		t.Error("Should have error when ProviderData is wrong type")
	}
}

func TestBCMResourceBase_ConfigureResource_OtherStruct(t *testing.T) {
	t.Parallel()

	// Pass a different struct type
	type OtherClient struct {
		endpoint string
	}
	wrongStruct := &OtherClient{endpoint: "test"}

	base := &BCMResourceBase{}
	req := resource.ConfigureRequest{
		ProviderData: wrongStruct,
	}
	resp := &resource.ConfigureResponse{}

	base.ConfigureResource(req, resp)

	// Should not set client and should add error
	if base.Client != nil {
		t.Error("Client should be nil when ProviderData is wrong struct type")
	}
	if !resp.Diagnostics.HasError() {
		t.Error("Should have error when ProviderData is wrong struct type")
	}
}

// =============================================================================
// BCMDataSourceBase Tests
// =============================================================================

func TestBCMDataSourceBase_ConfigureDataSource_NilProviderData(t *testing.T) {
	t.Parallel()

	base := &BCMDataSourceBase{}
	req := datasource.ConfigureRequest{
		ProviderData: nil,
	}
	resp := &datasource.ConfigureResponse{}

	base.ConfigureDataSource(req, resp)

	// Should not set client and should not add errors
	if base.Client != nil {
		t.Error("Client should be nil when ProviderData is nil")
	}
	if resp.Diagnostics.HasError() {
		t.Errorf("Should not have errors when ProviderData is nil: %v", resp.Diagnostics)
	}
}

func TestBCMDataSourceBase_ConfigureDataSource_ValidClient(t *testing.T) {
	t.Parallel()

	// Create a mock client
	mockClient := &BCMClient{
		Endpoint: "https://test.example.com",
	}

	base := &BCMDataSourceBase{}
	req := datasource.ConfigureRequest{
		ProviderData: mockClient,
	}
	resp := &datasource.ConfigureResponse{}

	base.ConfigureDataSource(req, resp)

	// Should set client and should not add errors
	if base.Client != mockClient {
		t.Errorf("Client should be set to mock client, got %v", base.Client)
	}
	if resp.Diagnostics.HasError() {
		t.Errorf("Should not have errors with valid BCMClient: %v", resp.Diagnostics)
	}
}

func TestBCMDataSourceBase_ConfigureDataSource_WrongType(t *testing.T) {
	t.Parallel()

	// Pass wrong type as ProviderData
	wrongType := 12345

	base := &BCMDataSourceBase{}
	req := datasource.ConfigureRequest{
		ProviderData: wrongType,
	}
	resp := &datasource.ConfigureResponse{}

	base.ConfigureDataSource(req, resp)

	// Should not set client and should add error
	if base.Client != nil {
		t.Error("Client should be nil when ProviderData is wrong type")
	}
	if !resp.Diagnostics.HasError() {
		t.Error("Should have error when ProviderData is wrong type")
	}
}

func TestBCMDataSourceBase_ConfigureDataSource_OtherStruct(t *testing.T) {
	t.Parallel()

	// Pass a different struct type
	type OtherClient struct {
		endpoint string
	}
	wrongStruct := &OtherClient{endpoint: "test"}

	base := &BCMDataSourceBase{}
	req := datasource.ConfigureRequest{
		ProviderData: wrongStruct,
	}
	resp := &datasource.ConfigureResponse{}

	base.ConfigureDataSource(req, resp)

	// Should not set client and should add error
	if base.Client != nil {
		t.Error("Client should be nil when ProviderData is wrong struct type")
	}
	if !resp.Diagnostics.HasError() {
		t.Error("Should have error when ProviderData is wrong struct type")
	}
}

// =============================================================================
// Embedding Tests - verify the embedding pattern works correctly
// =============================================================================

// TestResourceEmbeddingPattern verifies that resources can properly embed BCMResourceBase
func TestResourceEmbeddingPattern(t *testing.T) {
	t.Parallel()

	// Create a type that embeds BCMResourceBase (simulating a real resource)
	type TestResource struct {
		BCMResourceBase
	}

	mockClient := &BCMClient{
		Endpoint: "https://test.example.com",
	}

	testRes := &TestResource{}
	req := resource.ConfigureRequest{
		ProviderData: mockClient,
	}
	resp := &resource.ConfigureResponse{}

	// Configure using the embedded method
	testRes.ConfigureResource(req, resp)

	// Verify we can access the client via the embedding
	if testRes.Client != mockClient {
		t.Error("Should be able to access Client via embedding")
	}
	if testRes.BCMResourceBase.Client != mockClient {
		t.Error("Should be able to access Client via explicit BCMResourceBase path")
	}
}

// TestDataSourceEmbeddingPattern verifies that data sources can properly embed BCMDataSourceBase
func TestDataSourceEmbeddingPattern(t *testing.T) {
	t.Parallel()

	// Create a type that embeds BCMDataSourceBase (simulating a real data source)
	type TestDataSource struct {
		BCMDataSourceBase
	}

	mockClient := &BCMClient{
		Endpoint: "https://test.example.com",
	}

	testDS := &TestDataSource{}
	req := datasource.ConfigureRequest{
		ProviderData: mockClient,
	}
	resp := &datasource.ConfigureResponse{}

	// Configure using the embedded method
	testDS.ConfigureDataSource(req, resp)

	// Verify we can access the client via the embedding
	if testDS.Client != mockClient {
		t.Error("Should be able to access Client via embedding")
	}
	if testDS.BCMDataSourceBase.Client != mockClient {
		t.Error("Should be able to access Client via explicit BCMDataSourceBase path")
	}
}
