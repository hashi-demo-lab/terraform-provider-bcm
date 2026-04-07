// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// Package provider implements the bcm_cmkube_cluster resource for managing BCM KubeCluster entities.
// This resource enables Terraform to manage Kubernetes cluster definitions in BCM.
// Aligned with BCM CMKube API - FR-001 through FR-019.
package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &CMKubeClusterResource{}
	_ resource.ResourceWithImportState = &CMKubeClusterResource{}
)

// CMKubeClusterResource defines the resource implementation.
type CMKubeClusterResource struct {
	client *BCMClient
}

// NewCMKubeClusterResource creates a new resource instance.
func NewCMKubeClusterResource() resource.Resource {
	return &CMKubeClusterResource{}
}

// Metadata returns the resource type name.
func (r *CMKubeClusterResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cmkube_cluster"
}

// Configure adds the provider configured client to the resource.
func (r *CMKubeClusterResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if provider is not configured
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

	r.client = client
}

// Schema defines the resource schema - aligned with BCM KubeCluster API entity (FR-001 through FR-019).
func (r *CMKubeClusterResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a BCM Kubernetes cluster definition.\n\n" +
			"KubeCluster entities define the cluster configuration including network references, " +
			"API server settings, and application groups. Node membership is managed via " +
			"KubeletRole on device resources, not on the cluster itself.",

		Attributes: map[string]schema.Attribute{
			// Identifier attributes
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Cluster identifier (same as uuid)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"uuid": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "BCM-assigned cluster UUID",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			// Required attributes
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Cluster name (RFC 1123 DNS label: lowercase alphanumeric and hyphens, 1-63 characters).",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 63),
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`),
						"must contain only lowercase alphanumeric characters and hyphens, must start and end with alphanumeric",
					),
				},
			},

			// Network references (FR-001, FR-002, FR-003)
			"internal_network": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "UUID of the internal network for cluster node communication.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`),
						"must be a valid UUID",
					),
				},
			},
			"service_network": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "UUID of the service network for Kubernetes service IPs.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`),
						"must be a valid UUID",
					),
				},
			},
			"pod_network": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "UUID of the pod network for container IPs.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`),
						"must be a valid UUID",
					),
				},
			},

			// EtcdCluster reference (FR-004)
			"etcd_cluster": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "UUID of the EtcdCluster entity that backs this Kubernetes cluster.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`),
						"must be a valid UUID",
					),
				},
			},

			// Kubernetes configuration (FR-005, FR-006, FR-009)
			"version": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Kubernetes version (semver format, e.g., '1.28.0').",
				Default:             stringdefault.StaticString("1.28.0"),
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^\d+\.\d+\.\d+$`),
						"must be valid semver format (e.g., '1.28.0')",
					),
				},
			},
			"pod_network_node_mask": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Pod network node mask for CIDR allocation (e.g., '/24').",
				Default:             stringdefault.StaticString("/24"),
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^/\d{1,2}$`),
						"must be a valid CIDR mask (e.g., '/24')",
					),
				},
			},
			"kube_dns_ip": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Cluster DNS IP address. BCM sets a default value if not specified.",
				// No default - BCM is authoritative and may set server-side defaults
			},

			// API server configuration (FR-007, FR-008, FR-010)
			"kubernetes_api_server": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Kubernetes API server URL.",
				// No default - BCM is authoritative
			},
			"kubernetes_api_server_proxy_port": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Kubernetes API server proxy port. BCM defaults to 6444.",
				// No default - BCM is authoritative and sets 6444
				Validators: []validator.Int64{
					int64validator.Between(1, 65535),
				},
			},
			"trusted_domains": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "List of trusted domains for certificate SANs.",
				// No default - BCM is authoritative
			},

			// Ingress proxy configuration (FR-014)
			"ingress_proxy_enable": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Enable ingress proxy for external traffic routing.",
				// No default - BCM is authoritative
			},
			"ingress_proxy_listen_port": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Ingress proxy listen port. BCM may set a server-side default.",
				// No default - BCM is authoritative and may set 443
				Validators: []validator.Int64{
					int64validator.Between(0, 65535),
				},
			},
			"ingress_proxy_backend_port": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Ingress proxy backend port.",
				// No default - BCM is authoritative
				Validators: []validator.Int64{
					int64validator.Between(0, 65535),
				},
			},

			// Extensible options (FR-015)
			"options": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Extensible configuration options as JSON string.",
				Default:             stringdefault.StaticString("{}"),
			},

			// Computed metadata
			"creation_time": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Unix timestamp of when the cluster was created.",
			},
			"revision_id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Revision number for change tracking.",
			},
		},

		// Nested blocks (FR-011, FR-012, FR-013)
		Blocks: map[string]schema.Block{
			// AppGroups nested block (FR-011)
			"app_groups": schema.ListNestedBlock{
				MarkdownDescription: "Application groups for cluster addons. Each group contains applications (Kubernetes manifests) that can be enabled/disabled together.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Application group name.",
						},
						"enabled": schema.BoolAttribute{
							Optional:            true,
							Computed:            true,
							MarkdownDescription: "Whether the application group is enabled.",
							Default:             booldefault.StaticBool(true),
						},
					},
					Blocks: map[string]schema.Block{
						"applications": schema.ListNestedBlock{
							MarkdownDescription: "Applications within this group.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Required:            true,
										MarkdownDescription: "Application name.",
									},
									"enabled": schema.BoolAttribute{
										Optional:            true,
										Computed:            true,
										MarkdownDescription: "Whether the application is enabled.",
										Default:             booldefault.StaticBool(true),
									},
									"manifest": schema.StringAttribute{
										Optional:            true,
										Computed:            true,
										MarkdownDescription: "Kubernetes manifest YAML/JSON content.",
										Default:             stringdefault.StaticString(""),
									},
								},
							},
						},
					},
				},
			},

			// LabelSets nested block (FR-012)
			"label_sets": schema.ListNestedBlock{
				MarkdownDescription: "Label sets that can be applied to nodes, categories, or overlays.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Label set name.",
						},
						"labels": schema.MapAttribute{
							ElementType:         types.StringType,
							Optional:            true,
							Computed:            true,
							MarkdownDescription: "Map of label key-value pairs.",
						},
					},
				},
			},

			// Users nested block (FR-013)
			"users": schema.ListNestedBlock{
				MarkdownDescription: "Kubernetes users for kubeconfig management.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "User name.",
						},
						"groups": schema.ListAttribute{
							ElementType:         types.StringType,
							Optional:            true,
							Computed:            true,
							MarkdownDescription: "List of groups the user belongs to.",
						},
					},
				},
			},
		},
	}
}

// Create creates a new Kubernetes cluster (FR-018, FR-019).
func (r *CMKubeClusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data KubeClusterAlignedResourceModel

	// Read Terraform plan data into model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Nil client check
	if r.client == nil {
		resp.Diagnostics.AddError("Client Not Configured", "The BCM client is not configured. Please configure the provider.")
		return
	}

	// Generate UUID for new cluster (FR-019: BCM cmkube API requires client-generated UUID)
	clusterUUID := generateUUID()

	// Build entity
	entity, diags := r.buildEntity(ctx, &data, clusterUUID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Pre-flight validation (FR-018)
	validationErrors, err := r.client.ValidateEntity(ctx, "cmkube", "validateKubeCluster", entity, true)
	if err != nil {
		resp.Diagnostics.AddError("Validation API Failed", fmt.Sprintf("Failed to validate KubeCluster: %s", err))
		return
	}
	if ProcessValidationErrors(validationErrors, &resp.Diagnostics) {
		return
	}

	// Create via BCM API
	tflog.Debug(ctx, "Creating KubeCluster", map[string]interface{}{
		"name": data.Name.ValueString(),
		"uuid": clusterUUID,
	})

	body, err := r.client.CallJSONRPC(ctx, "cmkube", "addKubeCluster", entity, false)
	if err != nil {
		resp.Diagnostics.AddError("Create Failed", fmt.Sprintf("Failed to create KubeCluster: %s", err))
		return
	}

	// Parse response - BCM returns validation response: {"success": true/false, "validation": [...]}
	var validationResp struct {
		Success    bool                     `json:"success"`
		Validation []map[string]interface{} `json:"validation"`
	}

	if err := json.Unmarshal(body, &validationResp); err != nil {
		resp.Diagnostics.AddError(
			"Response Parse Error",
			fmt.Sprintf("Failed to parse KubeCluster creation response: %s", err),
		)
		return
	}

	// Check validation response
	if !validationResp.Success {
		var errorMsgs []string
		for _, v := range validationResp.Validation {
			if field, ok := v["field"].(string); ok {
				if msg, ok := v["message"].(string); ok {
					errorMsgs = append(errorMsgs, fmt.Sprintf("%s: %s", field, msg))
				}
			}
		}
		resp.Diagnostics.AddError(
			"KubeCluster Creation Failed",
			fmt.Sprintf("Failed to create KubeCluster '%s': validation errors: %v", data.Name.ValueString(), errorMsgs),
		)
		return
	}

	// Set UUID in model for read
	data.ID = types.StringValue(clusterUUID)
	data.UUID = types.StringValue(clusterUUID)

	tflog.Debug(ctx, "KubeCluster created successfully", map[string]interface{}{
		"name": data.Name.ValueString(),
		"uuid": clusterUUID,
	})

	// Preserve plan values for ALL optional fields that BCM may return different defaults for
	// This is critical to prevent "inconsistent result after apply" errors
	planOptions := data.Options
	planAppGroups := data.AppGroups
	planLabelSets := data.LabelSets
	planUsers := data.Users
	planKubeDnsIP := data.KubeDnsIP
	planKubernetesAPIServer := data.KubernetesAPIServer
	planKubernetesAPIServerProxyPort := data.KubernetesAPIServerProxyPort
	planIngressProxyEnable := data.IngressProxyEnable
	planIngressProxyListenPort := data.IngressProxyListenPort
	planIngressProxyBackendPort := data.IngressProxyBackendPort
	planVersion := data.Version
	planPodNetworkNodeMask := data.PodNetworkNodeMask
	planTrustedDomains := data.TrustedDomains

	// Read back created entity to populate all fields with eventual consistency handling
	maxRetries := 5
	var lastReadErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		readBody, err := r.client.CallJSONRPC(ctx, "cmkube", "getKubeCluster", clusterUUID)
		if err != nil {
			lastReadErr = err
			if attempt < maxRetries-1 {
				sleepDuration := time.Duration(1<<attempt) * time.Second
				tflog.Warn(ctx, "KubeCluster read after create failed, retrying", map[string]interface{}{
					"attempt":       attempt + 1,
					"sleep_seconds": sleepDuration.Seconds(),
					"error":         err.Error(),
				})
				time.Sleep(sleepDuration)
				continue
			}
			resp.Diagnostics.AddError(
				"Read After Create Failed",
				fmt.Sprintf("Failed to read KubeCluster after create: %s", lastReadErr),
			)
			return
		}

		// Parse the read response
		var responseData map[string]interface{}
		if err := json.Unmarshal(readBody, &responseData); err != nil {
			if attempt < maxRetries-1 {
				sleepDuration := time.Duration(1<<attempt) * time.Second
				tflog.Warn(ctx, "KubeCluster response parse failed, retrying", map[string]interface{}{
					"attempt":       attempt + 1,
					"sleep_seconds": sleepDuration.Seconds(),
				})
				time.Sleep(sleepDuration)
				continue
			}
			resp.Diagnostics.AddError(
				"Response Parse Failed",
				fmt.Sprintf("Failed to parse KubeCluster response: %s", err),
			)
			return
		}

		// Check if response has expected fields
		if responseData["name"] != nil && responseData["uuid"] != nil {
			diags := r.parseResponseIntoModel(ctx, responseData, &data)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			tflog.Debug(ctx, "Successfully read KubeCluster after create", map[string]interface{}{
				"uuid":    data.UUID.ValueString(),
				"name":    data.Name.ValueString(),
				"attempt": attempt + 1,
			})
			break
		}

		if attempt < maxRetries-1 {
			sleepDuration := time.Duration(1<<attempt) * time.Second
			tflog.Warn(ctx, "KubeCluster fields not populated, retrying", map[string]interface{}{
				"attempt":       attempt + 1,
				"sleep_seconds": sleepDuration.Seconds(),
			})
			time.Sleep(sleepDuration)
		} else {
			// Final attempt failed - error out instead of saving incomplete state
			resp.Diagnostics.AddError(
				"Read After Create Incomplete",
				"KubeCluster was created but read-back returned incomplete data after all retries",
			)
			return
		}
	}

	// Restore plan values ONLY for nested blocks that BCM may not return or returns differently
	// BCM does return scalar fields (kube_dns_ip, ingress ports, etc.) so we accept those values
	if !planOptions.IsNull() && !planOptions.IsUnknown() {
		data.Options = planOptions
	}
	if !planAppGroups.IsNull() && !planAppGroups.IsUnknown() {
		data.AppGroups = planAppGroups
	}
	if !planLabelSets.IsNull() && !planLabelSets.IsUnknown() {
		data.LabelSets = planLabelSets
	}
	if !planUsers.IsNull() && !planUsers.IsUnknown() {
		data.Users = planUsers
	}

	// Note: We do NOT preserve scalar optional fields (kube_dns_ip, ingress_proxy_listen_port, etc.)
	// because BCM returns actual values for these and we should accept the server values
	// This is the Terraform provider pattern for "eventually consistent" APIs where the server
	// applies defaults - we accept the server's authoritative state
	_ = planKubeDnsIP
	_ = planKubernetesAPIServer
	_ = planKubernetesAPIServerProxyPort
	_ = planIngressProxyEnable
	_ = planIngressProxyListenPort
	_ = planIngressProxyBackendPort
	_ = planVersion
	_ = planPodNetworkNodeMask
	_ = planTrustedDomains

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read reads the current cluster state from BCM.
func (r *CMKubeClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data KubeClusterAlignedResourceModel

	// Read Terraform state data into model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Nil client check
	if r.client == nil {
		resp.Diagnostics.AddError("Client Not Configured", "The BCM client is not configured. Please configure the provider.")
		return
	}

	// Preserve state values for fields BCM may not return
	stateOptions := data.Options
	stateAppGroups := data.AppGroups
	stateLabelSets := data.LabelSets
	stateUsers := data.Users

	// Get from BCM API using UUID
	identifier := data.UUID.ValueString()
	if identifier == "" {
		identifier = data.ID.ValueString()
	}

	tflog.Debug(ctx, "Reading KubeCluster", map[string]interface{}{
		"id": identifier,
	})

	body, err := r.client.CallJSONRPC(ctx, "cmkube", "getKubeCluster", identifier)
	if err != nil {
		// Check if resource no longer exists
		if containsAny(err.Error(), []string{"not found", "does not exist", "404", "null"}) {
			tflog.Info(ctx, "KubeCluster not found, removing from state", map[string]interface{}{
				"id": identifier,
			})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read Failed", fmt.Sprintf("Failed to read KubeCluster: %s", err))
		return
	}

	// Parse response
	var responseData map[string]interface{}
	if err := json.Unmarshal(body, &responseData); err != nil {
		resp.Diagnostics.AddError("Parse Failed", fmt.Sprintf("Failed to parse KubeCluster response: %s", err))
		return
	}

	// Check if response is empty (deleted)
	if len(responseData) == 0 {
		tflog.Info(ctx, "KubeCluster returned empty response, removing from state", map[string]interface{}{
			"id": identifier,
		})
		resp.State.RemoveResource(ctx)
		return
	}

	// Update model from response
	diags := r.parseResponseIntoModel(ctx, responseData, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Restore state values for fields BCM may not return
	if !stateOptions.IsNull() && !stateOptions.IsUnknown() && (data.Options.IsNull() || data.Options.ValueString() == "{}") {
		data.Options = stateOptions
	}
	if !stateAppGroups.IsNull() && !stateAppGroups.IsUnknown() && data.AppGroups.IsNull() {
		data.AppGroups = stateAppGroups
	}
	if !stateLabelSets.IsNull() && !stateLabelSets.IsUnknown() && data.LabelSets.IsNull() {
		data.LabelSets = stateLabelSets
	}
	if !stateUsers.IsNull() && !stateUsers.IsUnknown() && data.Users.IsNull() {
		data.Users = stateUsers
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates an existing Kubernetes cluster.
func (r *CMKubeClusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data KubeClusterAlignedResourceModel

	// Read Terraform plan data into model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Nil client check
	if r.client == nil {
		resp.Diagnostics.AddError("Client Not Configured", "The BCM client is not configured. Please configure the provider.")
		return
	}

	// Preserve plan values for fields BCM may not return
	planOptions := data.Options
	planAppGroups := data.AppGroups
	planLabelSets := data.LabelSets
	planUsers := data.Users

	// Get current state to preserve UUID
	var stateData KubeClusterAlignedResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &stateData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve UUID from state
	data.UUID = stateData.UUID
	data.ID = stateData.ID

	// Build entity with existing UUID
	entity, diags := r.buildEntity(ctx, &data, data.UUID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Pre-flight validation (FR-018)
	validationErrors, err := r.client.ValidateEntity(ctx, "cmkube", "validateKubeCluster", entity, false)
	if err != nil {
		resp.Diagnostics.AddError("Validation API Failed", fmt.Sprintf("Failed to validate KubeCluster: %s", err))
		return
	}
	if ProcessValidationErrors(validationErrors, &resp.Diagnostics) {
		return
	}

	tflog.Debug(ctx, "Updating KubeCluster", map[string]interface{}{
		"id":   data.UUID.ValueString(),
		"name": data.Name.ValueString(),
	})

	// Update via BCM API
	body, err := r.client.CallJSONRPC(ctx, "cmkube", "updateKubeCluster", entity, false)
	if err != nil {
		resp.Diagnostics.AddError("Update Failed", fmt.Sprintf("Failed to update KubeCluster: %s", err))
		return
	}

	// Parse response
	var validationResp struct {
		Success    bool                     `json:"success"`
		Validation []map[string]interface{} `json:"validation"`
	}

	if err := json.Unmarshal(body, &validationResp); err != nil {
		resp.Diagnostics.AddError(
			"Response Parse Error",
			fmt.Sprintf("Failed to parse KubeCluster update response: %s", err),
		)
		return
	}

	// Check validation response
	if !validationResp.Success {
		var errorMsgs []string
		for _, v := range validationResp.Validation {
			if field, ok := v["field"].(string); ok {
				if msg, ok := v["message"].(string); ok {
					errorMsgs = append(errorMsgs, fmt.Sprintf("%s: %s", field, msg))
				}
			}
		}
		resp.Diagnostics.AddError(
			"KubeCluster Update Failed",
			fmt.Sprintf("Failed to update KubeCluster '%s': validation errors: %v", data.Name.ValueString(), errorMsgs),
		)
		return
	}

	// Read back updated entity with eventual consistency handling
	maxRetries := 5
	var lastReadErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		readBody, err := r.client.CallJSONRPC(ctx, "cmkube", "getKubeCluster", data.UUID.ValueString())
		if err != nil {
			lastReadErr = err
			if attempt < maxRetries-1 {
				sleepDuration := time.Duration(1<<attempt) * time.Second
				tflog.Warn(ctx, "KubeCluster read after update failed, retrying", map[string]interface{}{
					"attempt":       attempt + 1,
					"sleep_seconds": sleepDuration.Seconds(),
					"error":         err.Error(),
				})
				time.Sleep(sleepDuration)
				continue
			}
			resp.Diagnostics.AddError(
				"Read After Update Failed",
				fmt.Sprintf("Failed to read KubeCluster after update: %s", lastReadErr),
			)
			return
		}

		var responseData map[string]interface{}
		if err := json.Unmarshal(readBody, &responseData); err != nil {
			if attempt < maxRetries-1 {
				sleepDuration := time.Duration(1<<attempt) * time.Second
				tflog.Warn(ctx, "KubeCluster response parse after update failed, retrying", map[string]interface{}{
					"attempt":       attempt + 1,
					"sleep_seconds": sleepDuration.Seconds(),
				})
				time.Sleep(sleepDuration)
				continue
			}
			resp.Diagnostics.AddError(
				"Response Parse Failed",
				fmt.Sprintf("Failed to parse KubeCluster response: %s", err),
			)
			return
		}

		// Parse response into model
		parseDiags := r.parseResponseIntoModel(ctx, responseData, &data)
		resp.Diagnostics.Append(parseDiags...)
		if resp.Diagnostics.HasError() {
			return
		}

		tflog.Debug(ctx, "Successfully read KubeCluster after update", map[string]interface{}{
			"uuid":    data.UUID.ValueString(),
			"name":    data.Name.ValueString(),
			"attempt": attempt + 1,
		})
		break
	}

	// Restore plan values for fields BCM may not return
	if !planOptions.IsNull() && !planOptions.IsUnknown() {
		data.Options = planOptions
	}
	if !planAppGroups.IsNull() && !planAppGroups.IsUnknown() {
		data.AppGroups = planAppGroups
	}
	if !planLabelSets.IsNull() && !planLabelSets.IsUnknown() {
		data.LabelSets = planLabelSets
	}
	if !planUsers.IsNull() && !planUsers.IsUnknown() {
		data.Users = planUsers
	}

	tflog.Debug(ctx, "Updated KubeCluster", map[string]interface{}{
		"id":   data.ID.ValueString(),
		"name": data.Name.ValueString(),
	})

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete removes a Kubernetes cluster.
func (r *CMKubeClusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data KubeClusterAlignedResourceModel

	// Read Terraform state data into model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Nil client check
	if r.client == nil {
		resp.Diagnostics.AddError("Client Not Configured", "The BCM client is not configured. Please configure the provider.")
		return
	}

	uuid := data.UUID.ValueString()
	if uuid == "" {
		uuid = data.ID.ValueString()
	}

	tflog.Debug(ctx, "Deleting KubeCluster", map[string]interface{}{
		"id":   uuid,
		"name": data.Name.ValueString(),
	})

	// Delete via BCM API
	_, err := r.client.CallJSONRPC(ctx, "cmkube", "removeKubeCluster", uuid, false)
	if err != nil {
		// Ignore "not found" errors during delete (idempotent)
		if !containsAny(err.Error(), []string{"not found", "does not exist", "404"}) {
			resp.Diagnostics.AddError("Delete Failed", fmt.Sprintf("Failed to delete KubeCluster: %s", err))
			return
		}
		tflog.Info(ctx, "KubeCluster already deleted", map[string]interface{}{
			"id": uuid,
		})
	}

	tflog.Debug(ctx, "Deleted KubeCluster", map[string]interface{}{
		"id": uuid,
	})
}

// ImportState imports an existing cluster by UUID (FR-017).
func (r *CMKubeClusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by UUID
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// =============================================================================
// Helper Methods
// =============================================================================

// buildEntity constructs a BCM KubeCluster entity from Terraform model.
func (r *CMKubeClusterResource) buildEntity(ctx context.Context, data *KubeClusterAlignedResourceModel, uuid string) (map[string]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	entity := map[string]interface{}{
		"baseType":      "KubeCluster",
		"childType":     "",
		"modified":      true,
		"to_be_removed": false,
		"revision":      "",
	}

	// Set or use provided UUID
	if uuid != "" {
		entity["uuid"] = uuid
	} else {
		entity["uuid"] = generateUUID()
	}

	// Required fields
	SetStringField(entity, "name", data.Name)

	// Network references (FR-001, FR-002, FR-003)
	SetStringField(entity, "internalNetwork", data.InternalNetwork)
	SetStringField(entity, "serviceNetwork", data.ServiceNetwork)
	SetStringField(entity, "podNetwork", data.PodNetwork)

	// EtcdCluster reference (FR-004)
	SetStringField(entity, "etcdCluster", data.EtcdCluster)

	// Kubernetes configuration (FR-005, FR-006, FR-009)
	SetStringField(entity, "version", data.Version)
	SetStringField(entity, "podNetworkNodeMask", data.PodNetworkNodeMask)
	if !data.KubeDnsIP.IsNull() && !data.KubeDnsIP.IsUnknown() && data.KubeDnsIP.ValueString() != "" {
		entity["kubeDnsIp"] = data.KubeDnsIP.ValueString()
	}

	// API server configuration (FR-007, FR-008, FR-010)
	if !data.KubernetesAPIServer.IsNull() && !data.KubernetesAPIServer.IsUnknown() && data.KubernetesAPIServer.ValueString() != "" {
		entity["kubernetesApiServer"] = data.KubernetesAPIServer.ValueString()
	}
	SetInt64Field(entity, "kubernetesApiServerProxyPort", data.KubernetesAPIServerProxyPort)

	// Trusted domains (FR-010)
	if !data.TrustedDomains.IsNull() && !data.TrustedDomains.IsUnknown() {
		var trustedDomains []string
		diags.Append(data.TrustedDomains.ElementsAs(ctx, &trustedDomains, false)...)
		entity["trustedDomains"] = trustedDomains
	} else {
		entity["trustedDomains"] = []string{}
	}

	// Ingress proxy configuration (FR-014)
	SetBoolField(entity, "ingressProxyEnable", data.IngressProxyEnable)
	if !data.IngressProxyListenPort.IsNull() && !data.IngressProxyListenPort.IsUnknown() && data.IngressProxyListenPort.ValueInt64() > 0 {
		entity["ingressProxyListenPort"] = data.IngressProxyListenPort.ValueInt64()
	}
	if !data.IngressProxyBackendPort.IsNull() && !data.IngressProxyBackendPort.IsUnknown() && data.IngressProxyBackendPort.ValueInt64() > 0 {
		entity["ingressProxyBackendPort"] = data.IngressProxyBackendPort.ValueInt64()
	}

	// Options (FR-015)
	if !data.Options.IsNull() && !data.Options.IsUnknown() && data.Options.ValueString() != "" && data.Options.ValueString() != "{}" {
		var options map[string]interface{}
		if err := json.Unmarshal([]byte(data.Options.ValueString()), &options); err == nil {
			entity["options"] = options
		} else {
			entity["options"] = map[string]interface{}{}
		}
	} else {
		entity["options"] = map[string]interface{}{}
	}

	// AppGroups (FR-011)
	if !data.AppGroups.IsNull() && !data.AppGroups.IsUnknown() {
		var appGroups []KubeAppGroupModel
		diags.Append(data.AppGroups.ElementsAs(ctx, &appGroups, false)...)
		if len(appGroups) > 0 {
			appGroupsData := make([]map[string]interface{}, len(appGroups))
			for i, ag := range appGroups {
				agData := map[string]interface{}{
					"baseType":  "KubeAppGroup",
					"childType": "",
				}
				SetStringField(agData, "name", ag.Name)
				SetBoolField(agData, "enabled", ag.Enabled)
				appGroupsData[i] = agData

				// Applications within the group
				if !ag.Applications.IsNull() && !ag.Applications.IsUnknown() {
					var apps []KubeAppModel
					diags.Append(ag.Applications.ElementsAs(ctx, &apps, false)...)
					if len(apps) > 0 {
						appsData := make([]map[string]interface{}, len(apps))
						for j, app := range apps {
							appData := map[string]interface{}{
								"baseType":  "KubeApp",
								"childType": "",
							}
							SetStringField(appData, "name", app.Name)
							SetBoolField(appData, "enabled", app.Enabled)
							SetStringField(appData, "manifest", app.Manifest)
							appsData[j] = appData
						}
						appGroupsData[i]["applications"] = appsData
					} else {
						appGroupsData[i]["applications"] = []map[string]interface{}{}
					}
				} else {
					appGroupsData[i]["applications"] = []map[string]interface{}{}
				}
			}
			entity["appGroups"] = appGroupsData
		}
	}

	// LabelSets (FR-012)
	if !data.LabelSets.IsNull() && !data.LabelSets.IsUnknown() {
		var labelSets []KubeLabelSetModel
		diags.Append(data.LabelSets.ElementsAs(ctx, &labelSets, false)...)
		if len(labelSets) > 0 {
			labelSetsData := make([]map[string]interface{}, len(labelSets))
			for i, ls := range labelSets {
				lsData := map[string]interface{}{
					"baseType":  "KubeLabelSet",
					"childType": "",
				}
				SetStringField(lsData, "name", ls.Name)
				labelSetsData[i] = lsData
				if !ls.Labels.IsNull() && !ls.Labels.IsUnknown() {
					var labels map[string]string
					diags.Append(ls.Labels.ElementsAs(ctx, &labels, false)...)
					labelSetsData[i]["labels"] = labels
				} else {
					labelSetsData[i]["labels"] = map[string]string{}
				}
			}
			entity["labelSets"] = labelSetsData
		}
	}

	// Users (FR-013)
	if !data.Users.IsNull() && !data.Users.IsUnknown() {
		var users []KubeUserModel
		diags.Append(data.Users.ElementsAs(ctx, &users, false)...)
		if len(users) > 0 {
			usersData := make([]map[string]interface{}, len(users))
			for i, u := range users {
				uData := map[string]interface{}{
					"baseType":  "KubeUser",
					"childType": "",
				}
				SetStringField(uData, "name", u.Name)
				usersData[i] = uData
				if !u.Groups.IsNull() && !u.Groups.IsUnknown() {
					var groups []string
					diags.Append(u.Groups.ElementsAs(ctx, &groups, false)...)
					usersData[i]["groups"] = groups
				} else {
					usersData[i]["groups"] = []string{}
				}
			}
			entity["users"] = usersData
		}
	}

	return entity, diags
}

// parseResponseIntoModel updates the Terraform model from BCM API response.
func (r *CMKubeClusterResource) parseResponseIntoModel(ctx context.Context, data map[string]interface{}, model *KubeClusterAlignedResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	// Parse identifiers
	model.UUID = getStringValue(data, "uuid")
	model.ID = model.UUID
	model.Name = getStringValue(data, "name")

	// Parse network references
	model.InternalNetwork = getStringValue(data, "internalNetwork")
	model.ServiceNetwork = getStringValue(data, "serviceNetwork")
	model.PodNetwork = getStringValue(data, "podNetwork")
	model.EtcdCluster = getStringValue(data, "etcdCluster")

	// Parse Kubernetes configuration
	model.Version = getStringValue(data, "version")
	model.PodNetworkNodeMask = getStringValue(data, "podNetworkNodeMask")
	model.KubeDnsIP = getStringValue(data, "kubeDnsIp")

	// Parse API server configuration
	model.KubernetesAPIServer = getStringValue(data, "kubernetesApiServer")
	model.KubernetesAPIServerProxyPort = getInt64Value(data, "kubernetesApiServerProxyPort")

	// Parse trusted domains - dynamically build slice to skip non-string values
	if trustedDomains, ok := data["trustedDomains"].([]interface{}); ok && len(trustedDomains) > 0 {
		elements := make([]attr.Value, 0, len(trustedDomains))
		for _, domain := range trustedDomains {
			if domainStr, ok := domain.(string); ok {
				elements = append(elements, types.StringValue(domainStr))
			}
		}
		model.TrustedDomains, _ = types.ListValue(types.StringType, elements)
	} else {
		model.TrustedDomains, _ = types.ListValue(types.StringType, []attr.Value{})
	}

	// Parse ingress proxy configuration
	model.IngressProxyEnable = getBoolValue(data, "ingressProxyEnable")
	model.IngressProxyListenPort = getInt64Value(data, "ingressProxyListenPort")
	model.IngressProxyBackendPort = getInt64Value(data, "ingressProxyBackendPort")

	// Parse options as JSON string
	if options, ok := data["options"]; ok && options != nil {
		if optMap, ok := options.(map[string]interface{}); ok {
			optBytes, err := json.Marshal(optMap)
			if err == nil {
				model.Options = types.StringValue(string(optBytes))
			} else {
				model.Options = types.StringValue("{}")
			}
		} else {
			model.Options = types.StringValue("{}")
		}
	} else {
		model.Options = types.StringValue("{}")
	}

	// Parse AppGroups (complex nested structure)
	if appGroupsData, ok := data["appGroups"].([]interface{}); ok && len(appGroupsData) > 0 {
		appGroupsList, appGroupsDiags := parseAppGroupsFromAPI(ctx, appGroupsData)
		diags.Append(appGroupsDiags...)
		model.AppGroups = appGroupsList
	} else {
		model.AppGroups = types.ListNull(types.ObjectType{AttrTypes: kubeAppGroupAttrTypes()})
	}

	// Parse LabelSets
	if labelSetsData, ok := data["labelSets"].([]interface{}); ok && len(labelSetsData) > 0 {
		labelSetsList, labelSetsDiags := parseLabelSetsFromAPI(ctx, labelSetsData)
		diags.Append(labelSetsDiags...)
		model.LabelSets = labelSetsList
	} else {
		model.LabelSets = types.ListNull(types.ObjectType{AttrTypes: kubeLabelSetAttrTypes()})
	}

	// Parse Users
	if usersData, ok := data["users"].([]interface{}); ok && len(usersData) > 0 {
		usersList, usersDiags := parseUsersFromAPI(ctx, usersData)
		diags.Append(usersDiags...)
		model.Users = usersList
	} else {
		model.Users = types.ListNull(types.ObjectType{AttrTypes: kubeUserAttrTypes()})
	}

	// Parse computed fields
	model.CreationTime = getInt64Value(data, "creationTime")
	model.RevisionID = getInt64Value(data, "revisionID")

	return diags
}

// =============================================================================
// Nested Type Helpers
// =============================================================================

// kubeAppGroupAttrTypes returns the attribute types for KubeAppGroup.
func kubeAppGroupAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":    types.StringType,
		"enabled": types.BoolType,
		"applications": types.ListType{
			ElemType: types.ObjectType{
				AttrTypes: kubeAppAttrTypes(),
			},
		},
	}
}

// kubeAppAttrTypes returns the attribute types for KubeApp.
func kubeAppAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":     types.StringType,
		"enabled":  types.BoolType,
		"manifest": types.StringType,
	}
}

// kubeLabelSetAttrTypes returns the attribute types for KubeLabelSet.
func kubeLabelSetAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name": types.StringType,
		"labels": types.MapType{
			ElemType: types.StringType,
		},
	}
}

// kubeUserAttrTypes returns the attribute types for KubeUser.
func kubeUserAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name": types.StringType,
		"groups": types.ListType{
			ElemType: types.StringType,
		},
	}
}

// parseAppGroupsFromAPI parses appGroups from BCM API response into Terraform types.
func parseAppGroupsFromAPI(_ context.Context, appGroupsData []interface{}) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics

	appGroupElements := make([]attr.Value, 0, len(appGroupsData))
	for _, agData := range appGroupsData {
		agMap, ok := agData.(map[string]interface{})
		if !ok {
			continue
		}

		// Parse applications within the group
		var applicationsValue types.List
		if appsData, ok := agMap["applications"].([]interface{}); ok && len(appsData) > 0 {
			appElements := make([]attr.Value, 0, len(appsData))
			for _, appData := range appsData {
				appMap, ok := appData.(map[string]interface{})
				if !ok {
					continue
				}
				appObj, appDiags := types.ObjectValue(kubeAppAttrTypes(), map[string]attr.Value{
					"name":     getStringValueForTF(appMap, "name"),
					"enabled":  getBoolValueForTF(appMap, "enabled"),
					"manifest": getStringValueForTF(appMap, "manifest"),
				})
				diags.Append(appDiags...)
				appElements = append(appElements, appObj)
			}
			var listDiags diag.Diagnostics
			applicationsValue, listDiags = types.ListValue(types.ObjectType{AttrTypes: kubeAppAttrTypes()}, appElements)
			diags.Append(listDiags...)
		} else {
			applicationsValue, _ = types.ListValue(types.ObjectType{AttrTypes: kubeAppAttrTypes()}, []attr.Value{})
		}

		agObj, agDiags := types.ObjectValue(kubeAppGroupAttrTypes(), map[string]attr.Value{
			"name":         getStringValueForTF(agMap, "name"),
			"enabled":      getBoolValueForTF(agMap, "enabled"),
			"applications": applicationsValue,
		})
		diags.Append(agDiags...)
		appGroupElements = append(appGroupElements, agObj)
	}

	result, listDiags := types.ListValue(types.ObjectType{AttrTypes: kubeAppGroupAttrTypes()}, appGroupElements)
	diags.Append(listDiags...)
	return result, diags
}

// parseLabelSetsFromAPI parses labelSets from BCM API response into Terraform types.
func parseLabelSetsFromAPI(_ context.Context, labelSetsData []interface{}) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics

	labelSetElements := make([]attr.Value, 0, len(labelSetsData))
	for _, lsData := range labelSetsData {
		lsMap, ok := lsData.(map[string]interface{})
		if !ok {
			continue
		}

		// Parse labels map
		var labelsValue types.Map
		if labelsData, ok := lsMap["labels"].(map[string]interface{}); ok && len(labelsData) > 0 {
			labelElements := make(map[string]attr.Value)
			for k, v := range labelsData {
				if vStr, ok := v.(string); ok {
					labelElements[k] = types.StringValue(vStr)
				}
			}
			var mapDiags diag.Diagnostics
			labelsValue, mapDiags = types.MapValue(types.StringType, labelElements)
			diags.Append(mapDiags...)
		} else {
			labelsValue, _ = types.MapValue(types.StringType, map[string]attr.Value{})
		}

		lsObj, lsDiags := types.ObjectValue(kubeLabelSetAttrTypes(), map[string]attr.Value{
			"name":   getStringValueForTF(lsMap, "name"),
			"labels": labelsValue,
		})
		diags.Append(lsDiags...)
		labelSetElements = append(labelSetElements, lsObj)
	}

	result, listDiags := types.ListValue(types.ObjectType{AttrTypes: kubeLabelSetAttrTypes()}, labelSetElements)
	diags.Append(listDiags...)
	return result, diags
}

// parseUsersFromAPI parses users from BCM API response into Terraform types.
func parseUsersFromAPI(_ context.Context, usersData []interface{}) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics

	userElements := make([]attr.Value, 0, len(usersData))
	for _, uData := range usersData {
		uMap, ok := uData.(map[string]interface{})
		if !ok {
			continue
		}

		// Parse groups list - use dynamic slice to skip non-string values safely
		var groupsValue types.List
		if groupsData, ok := uMap["groups"].([]interface{}); ok && len(groupsData) > 0 {
			groupElements := make([]attr.Value, 0, len(groupsData))
			for _, g := range groupsData {
				if gStr, ok := g.(string); ok {
					groupElements = append(groupElements, types.StringValue(gStr))
				}
			}
			var listDiags diag.Diagnostics
			groupsValue, listDiags = types.ListValue(types.StringType, groupElements)
			diags.Append(listDiags...)
		} else {
			groupsValue, _ = types.ListValue(types.StringType, []attr.Value{})
		}

		uObj, uDiags := types.ObjectValue(kubeUserAttrTypes(), map[string]attr.Value{
			"name":   getStringValueForTF(uMap, "name"),
			"groups": groupsValue,
		})
		diags.Append(uDiags...)
		userElements = append(userElements, uObj)
	}

	result, listDiags := types.ListValue(types.ObjectType{AttrTypes: kubeUserAttrTypes()}, userElements)
	diags.Append(listDiags...)
	return result, diags
}

// getStringValueForTF extracts a string value from a map and returns it as types.String.
func getStringValueForTF(data map[string]interface{}, key string) types.String {
	if v, ok := data[key]; ok && v != nil {
		if s, ok := v.(string); ok {
			return types.StringValue(s)
		}
	}
	return types.StringValue("")
}

// getBoolValueForTF extracts a bool value from a map and returns it as types.Bool.
func getBoolValueForTF(data map[string]interface{}, key string) types.Bool {
	if v, ok := data[key]; ok && v != nil {
		if b, ok := v.(bool); ok {
			return types.BoolValue(b)
		}
	}
	return types.BoolValue(false)
}
