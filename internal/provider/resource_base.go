// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// Package provider contains base types for BCM Terraform provider resources and data sources.
// These base types reduce code duplication by providing common functionality like Configure methods.
package provider

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// =============================================================================
// Base Types for Resources and Data Sources
// =============================================================================

// BCMResourceBase provides common functionality for all BCM resources.
// Embed this type in resource structs to inherit the Configure method and client access.
//
// Example:
//
//	type CMPartSoftwareImageResource struct {
//	    BCMResourceBase
//	}
//
//	func (r *CMPartSoftwareImageResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
//	    r.ConfigureResource(req, resp)
//	}
//
//	func (r *CMPartSoftwareImageResource) Create(...) {
//	    // Access client via embedded field
//	    r.Client.CallJSONRPC(...)
//	}
type BCMResourceBase struct {
	// Client is the BCM API client. Access directly from embedding structs.
	Client *BCMClient
}

// ConfigureResource handles the common Configure logic for resources.
// Call this from your resource's Configure method to set up the client.
//
// Example:
//
//	func (r *MyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
//	    r.ConfigureResource(req, resp)
//	}
func (b *BCMResourceBase) ConfigureResource(req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*BCMClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *BCMClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	b.Client = client
}

// =============================================================================
// Base Types for Data Sources
// =============================================================================

// BCMDataSourceBase provides common functionality for all BCM data sources.
// Embed this type in data source structs to inherit the Configure method and client access.
//
// Example:
//
//	type CMPartSoftwareImagesDataSource struct {
//	    BCMDataSourceBase
//	}
//
//	func (d *CMPartSoftwareImagesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
//	    d.ConfigureDataSource(req, resp)
//	}
//
//	func (d *CMPartSoftwareImagesDataSource) Read(...) {
//	    // Access client via embedded field
//	    d.Client.CallJSONRPC(...)
//	}
type BCMDataSourceBase struct {
	// Client is the BCM API client. Access directly from embedding structs.
	Client *BCMClient
}

// ConfigureDataSource handles the common Configure logic for data sources.
// Call this from your data source's Configure method to set up the client.
//
// Example:
//
//	func (d *MyDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
//	    d.ConfigureDataSource(req, resp)
//	}
func (b *BCMDataSourceBase) ConfigureDataSource(req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*BCMClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *BCMClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	b.Client = client
}
