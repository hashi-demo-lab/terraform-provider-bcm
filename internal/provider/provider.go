// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure BCMProvider satisfies various provider interfaces.
var _ provider.Provider = &BCMProvider{}
var _ provider.ProviderWithFunctions = &BCMProvider{}
var _ provider.ProviderWithEphemeralResources = &BCMProvider{}
var _ provider.ProviderWithActions = &BCMProvider{}

// BCMProvider defines the provider implementation.
type BCMProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// BCMProviderModel describes the provider data model.
type BCMProviderModel struct {
	Endpoint           types.String `tfsdk:"endpoint"`
	Username           types.String `tfsdk:"username"`
	Password           types.String `tfsdk:"password"`
	InsecureSkipVerify types.Bool   `tfsdk:"insecure_skip_verify"`
	Timeout            types.Int64  `tfsdk:"timeout"`
}

func (p *BCMProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "bcm"
	resp.Version = p.version
}

func (p *BCMProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `**PROTOTYPE PROVIDER** - This is a prototype Terraform provider for Nvidia BCM (Base Command Manager). It is recommended to fork this provider for your own use and development. This provider is not officially supported and may have breaking changes.`,
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "BCM JSON-RPC API endpoint (e.g., https://172.21.15.254:8081). Can also be set via BCM_ENDPOINT environment variable.",
				Optional:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "BCM username for authentication. Can also be set via BCM_USERNAME environment variable.",
				Optional:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "BCM password for authentication. Can also be set via BCM_PASSWORD environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"insecure_skip_verify": schema.BoolAttribute{
				MarkdownDescription: "Skip TLS certificate verification. WARNING: This makes connections susceptible to man-in-the-middle attacks. Only use for testing with self-signed certificates.",
				Optional:            true,
			},
			"timeout": schema.Int64Attribute{
				MarkdownDescription: "API timeout in seconds (default: 30)",
				Optional:            true,
			},
		},
	}
}

func (p *BCMProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data BCMProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read from environment variables if not set in configuration
	endpoint := data.Endpoint.ValueString()
	if endpoint == "" {
		endpoint = os.Getenv("BCM_ENDPOINT")
	}

	username := data.Username.ValueString()
	if username == "" {
		username = os.Getenv("BCM_USERNAME")
	}

	password := data.Password.ValueString()
	if password == "" {
		password = os.Getenv("BCM_PASSWORD")
	}

	// Validate required fields
	if endpoint == "" {
		resp.Diagnostics.AddError(
			"Missing BCM Endpoint",
			"The provider cannot create the BCM client as there is a missing or empty value for the BCM endpoint. "+
				"Set the endpoint value in the provider configuration or use the BCM_ENDPOINT environment variable.",
		)
	}

	if username == "" {
		resp.Diagnostics.AddError(
			"Missing BCM Username",
			"The provider cannot create the BCM client as there is a missing or empty value for the BCM username. "+
				"Set the username value in the provider configuration or use the BCM_USERNAME environment variable.",
		)
	}

	if password == "" {
		resp.Diagnostics.AddError(
			"Missing BCM Password",
			"The provider cannot create the BCM client as there is a missing or empty value for the BCM password. "+
				"Set the password value in the provider configuration or use the BCM_PASSWORD environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Set defaults for optional fields
	insecureSkipVerify := false
	if !data.InsecureSkipVerify.IsNull() {
		insecureSkipVerify = data.InsecureSkipVerify.ValueBool()
	}

	timeout := int64(30) // Default 30 seconds
	if !data.Timeout.IsNull() {
		timeout = data.Timeout.ValueInt64()
	}

	// Validate timeout value
	if timeout <= 0 {
		resp.Diagnostics.AddError(
			"Invalid Timeout Value",
			"Timeout must be a positive integer in seconds",
		)
		return
	}

	// Create BCM client with cookie-based authentication
	client, err := NewBCMClient(
		ctx,
		endpoint,
		username,
		password,
		insecureSkipVerify,
		int(timeout),
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create BCM Client",
			"An unexpected error occurred when creating the BCM client. "+
				"If the error is not clear, please check the BCM endpoint, credentials, and network connectivity.\n\n"+
				"BCM Client Error: "+err.Error(),
		)
		return
	}

	// Make client available to data sources, resources, actions, and ephemeral resources
	resp.DataSourceData = client
	resp.ResourceData = client
	resp.ActionData = client
	resp.EphemeralResourceData = client
}

func (p *BCMProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewCMPartSoftwareImageResource,
		NewCMDeviceCategoryResource,
		NewCMDeviceDeviceResource,
		NewCMNetNetworkResource,
		NewCMUserUserResource,
		NewCMKubeClusterResource,
	}
}

func (p *BCMProvider) EphemeralResources(ctx context.Context) []func() ephemeral.EphemeralResource {
	return []func() ephemeral.EphemeralResource{
		// No ephemeral resources in POV scope
	}
}

func (p *BCMProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewCMDeviceCategoriesDataSource,
		NewCMDeviceNodesDataSource,
		NewCMDeviceRolesDataSource,
		NewCMNetNetworksDataSource,
		NewCMPartEntityInfoDataSource,
		NewCMPartPartitionsDataSource,
		NewCMPartSoftwareImagesDataSource,
		NewCMUserUsersDataSource,
		NewCMKubeClustersDataSource,
	}
}

func (p *BCMProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{
		// No functions in POV scope
	}
}

func (p *BCMProvider) Actions(ctx context.Context) []func() action.Action {
	return []func() action.Action{
		NewCMDevicePowerAction,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &BCMProvider{
			version: version,
		}
	}
}
