// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
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
	BCMResourceBase
}

// CMKubeClusterResourceModel describes the resource data model.
type CMKubeClusterResourceModel struct {
	// Identity fields
	ID   types.String `tfsdk:"id"`   // Computed, same as UUID
	UUID types.String `tfsdk:"uuid"` // Computed, BCM-assigned
	Name types.String `tfsdk:"name"` // Required

	// Node configuration
	MasterNodes types.List `tfsdk:"master_nodes"` // Required, list of UUIDs
	WorkerNodes types.List `tfsdk:"worker_nodes"` // Optional, list of UUIDs
	EtcdNodes   types.List `tfsdk:"etcd_nodes"`   // Optional, list of UUIDs for etcd cluster members

	// Network configuration
	ManagementNetwork types.String `tfsdk:"management_network"` // Optional, UUID
	OverlayNetwork    types.String `tfsdk:"overlay_network"`    // Optional, pod network overlay config
	DNSServers        types.List   `tfsdk:"dns_servers"`        // Optional, list of DNS server IPs

	// Kubernetes configuration
	Version   types.String `tfsdk:"version"`    // Optional, semver string
	CNIPlugin types.String `tfsdk:"cni_plugin"` // Optional, CNI plugin selection

	// Storage configuration
	StorageClasses types.String `tfsdk:"storage_classes"` // Optional, JSON-encoded storage class definitions

	// Load balancer configuration
	LoadBalancerMode types.String `tfsdk:"load_balancer_mode"` // Optional, load balancer strategy

	// Cluster addons
	Addons            types.String `tfsdk:"addons"`             // Optional, JSON-encoded addon configurations
	IngressController types.String `tfsdk:"ingress_controller"` // Optional, JSON-encoded ingress controller config

	// Operations
	Force types.Bool `tfsdk:"force"` // Optional, default false

	// Computed metadata
	CreationTime types.Int64 `tfsdk:"creation_time"` // Computed
	RevisionID   types.Int64 `tfsdk:"revision_id"`   // Computed
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
	r.ConfigureResource(req, resp)
}

// Schema defines the resource schema.
func (r *CMKubeClusterResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a BCM Kubernetes cluster.\n\n" +
			"Kubernetes clusters in BCM define the cluster topology (master and worker nodes), " +
			"networking configuration, and Kubernetes version for container orchestration workloads.",

		Attributes: map[string]schema.Attribute{
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
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Cluster name (alphanumeric, hyphens, underscores only)",
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-zA-Z0-9_-]+$`),
						"must contain only alphanumeric characters, hyphens, and underscores",
					),
				},
			},
			"master_nodes": schema.ListAttribute{
				ElementType:         types.StringType,
				Required:            true,
				MarkdownDescription: "List of master node UUIDs (minimum 1 required)",
			},
			"worker_nodes": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "List of worker node UUIDs",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"etcd_nodes": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "List of node UUIDs designated as etcd cluster members. NVIDIA recommends 3 nodes for production high availability. If not specified, etcd runs on master nodes.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"management_network": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Management network UUID for cluster management traffic",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"overlay_network": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Overlay network configuration for pod networking (UUID or configuration string)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"dns_servers": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "List of custom DNS server IPs for the cluster",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"version": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Kubernetes version (semver format, e.g., '1.28.0')",
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^\d+\.\d+\.\d+$`),
						"must be valid semver format (e.g., '1.28.0')",
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cni_plugin": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "CNI plugin selection (e.g., 'calico', 'flannel', 'weave')",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"storage_classes": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Storage class definitions (JSON-encoded array of storage class configurations)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"load_balancer_mode": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Load balancer strategy for the cluster",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"addons": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Cluster addons configuration (JSON-encoded array of addon definitions for monitoring, logging, etc.)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"ingress_controller": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Ingress controller configuration (JSON-encoded object)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"force": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Bypass validation warnings during operations (default: false)",
			},
			"creation_time": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Cluster creation timestamp",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"revision_id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "BCM revision ID for optimistic locking",
			},
		},
	}
}

// Create creates a new Kubernetes cluster.
func (r *CMKubeClusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.Client == nil {
		resp.Diagnostics.AddError(
			"Provider Not Configured",
			"The provider has not been configured. Please ensure the provider block is properly configured.",
		)
		return
	}

	var plan CMKubeClusterResourceModel

	// Read Terraform plan data
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate UUID for new cluster (BCM cmkube API requires client-generated UUID)
	clusterUUID := generateUUID()

	// Build cluster entity for BCM API
	entity, diags := buildClusterEntity(ctx, plan, clusterUUID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Pre-flight validation: Call validateKubeCluster before CREATE
	// Note: Service name is "cmkube" (lowercase) - exception to CamelCase pattern
	validationErrors, err := r.Client.ValidateEntity(ctx, "cmkube", "validateKubeCluster", entity, true)
	if err != nil {
		resp.Diagnostics.AddError(
			"Validation API Error",
			fmt.Sprintf("Could not validate cluster '%s': %s", plan.Name.ValueString(), err.Error()),
		)
		return
	}

	// Process validation results - halt if errors found
	if ProcessValidationErrors(validationErrors, &resp.Diagnostics) {
		return
	}

	// Call BCM API to create cluster
	tflog.Debug(ctx, "Creating Kubernetes cluster via BCM API", map[string]interface{}{
		"name": plan.Name.ValueString(),
	})

	body, err := r.Client.CallJSONRPC(ctx, "cmkube", "addKubeCluster", entity, plan.Force.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Kubernetes Cluster",
			fmt.Sprintf("Could not create cluster, unexpected error: %s", err.Error()),
		)
		return
	}

	// Parse response to check success
	var apiResponse map[string]interface{}
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		resp.Diagnostics.AddError(
			"Error Parsing Cluster Creation Response",
			fmt.Sprintf("Could not parse response: %s", err.Error()),
		)
		return
	}

	// Check for validation errors
	if success, ok := apiResponse["success"].(bool); ok && !success {
		resp.Diagnostics.AddError(
			"Cluster Creation Failed",
			fmt.Sprintf("BCM API returned success=false: %s", string(body)),
		)
		return
	}

	// Set UUID in state (we generated it, so use the generated value)
	plan.UUID = types.StringValue(clusterUUID)
	plan.ID = types.StringValue(clusterUUID)

	tflog.Info(ctx, "Created Kubernetes cluster", map[string]interface{}{
		"uuid": clusterUUID,
		"name": plan.Name.ValueString(),
	})

	// Read back full cluster state from BCM
	// (This populates computed fields like creation_time, revision_id)
	// CRITICAL: Preserve optional fields from plan before reading (BCM API may not return them)
	// Node lists (master_nodes, worker_nodes, etcd_nodes) are write-only fields
	planMasterNodes := plan.MasterNodes
	planWorkerNodes := plan.WorkerNodes
	planEtcdNodes := plan.EtcdNodes
	planManagementNetwork := plan.ManagementNetwork
	planOverlayNetwork := plan.OverlayNetwork
	planDNSServers := plan.DNSServers
	planVersion := plan.Version
	planCNIPlugin := plan.CNIPlugin
	planStorageClasses := plan.StorageClasses
	planLoadBalancerMode := plan.LoadBalancerMode
	planAddons := plan.Addons
	planIngressController := plan.IngressController

	readDiags := r.readCluster(ctx, &plan)
	resp.Diagnostics.Append(readDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// CRITICAL FIX: Restore write-only node list fields from plan
	// These fields are accepted by BCM API but NOT returned in getKubeCluster response
	if !planMasterNodes.IsUnknown() && !planMasterNodes.IsNull() {
		plan.MasterNodes = planMasterNodes
	}
	if !planWorkerNodes.IsUnknown() && !planWorkerNodes.IsNull() {
		plan.WorkerNodes = planWorkerNodes
	}
	if !planEtcdNodes.IsUnknown() && !planEtcdNodes.IsNull() {
		plan.EtcdNodes = planEtcdNodes
	}

	// CRITICAL FIX: Restore optional P3 fields from plan ONLY if they're known values
	// BCM API may not return these fields, but we want to keep the plan values in state
	// Never propagate Unknown values - they cause "invalid result object" errors
	if !planManagementNetwork.IsUnknown() && !planManagementNetwork.IsNull() {
		plan.ManagementNetwork = planManagementNetwork
	}
	if !planOverlayNetwork.IsUnknown() && !planOverlayNetwork.IsNull() {
		plan.OverlayNetwork = planOverlayNetwork
	}
	if !planDNSServers.IsUnknown() && !planDNSServers.IsNull() {
		plan.DNSServers = planDNSServers
	}
	if !planVersion.IsUnknown() && !planVersion.IsNull() {
		plan.Version = planVersion
	}
	if !planCNIPlugin.IsUnknown() && !planCNIPlugin.IsNull() {
		plan.CNIPlugin = planCNIPlugin
	}
	if !planStorageClasses.IsUnknown() && !planStorageClasses.IsNull() {
		plan.StorageClasses = planStorageClasses
	}
	if !planLoadBalancerMode.IsUnknown() && !planLoadBalancerMode.IsNull() {
		plan.LoadBalancerMode = planLoadBalancerMode
	}
	if !planAddons.IsUnknown() && !planAddons.IsNull() {
		plan.Addons = planAddons
	}
	if !planIngressController.IsUnknown() && !planIngressController.IsNull() {
		plan.IngressController = planIngressController
	}

	// Save state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read reads the current cluster state from BCM.
func (r *CMKubeClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.Client == nil {
		resp.Diagnostics.AddError(
			"Provider Not Configured",
			"The provider has not been configured. Please ensure the provider block is properly configured.",
		)
		return
	}

	var state CMKubeClusterResourceModel

	// Read current state
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// CRITICAL: Preserve optional fields from state before reading (BCM API may not return them)
	// Node lists (master_nodes, worker_nodes, etcd_nodes) are write-only fields
	stateMasterNodes := state.MasterNodes
	stateWorkerNodes := state.WorkerNodes
	stateEtcdNodes := state.EtcdNodes
	stateManagementNetwork := state.ManagementNetwork
	stateOverlayNetwork := state.OverlayNetwork
	stateDNSServers := state.DNSServers
	stateVersion := state.Version
	stateCNIPlugin := state.CNIPlugin
	stateStorageClasses := state.StorageClasses
	stateLoadBalancerMode := state.LoadBalancerMode
	stateAddons := state.Addons
	stateIngressController := state.IngressController

	// Read cluster from BCM API
	diags := r.readCluster(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// CRITICAL FIX: Restore write-only node list fields from prior state
	// These fields are accepted by BCM API but NOT returned in getKubeCluster response
	if state.MasterNodes.IsNull() && !stateMasterNodes.IsNull() && !stateMasterNodes.IsUnknown() {
		state.MasterNodes = stateMasterNodes
	}
	if state.WorkerNodes.IsNull() && !stateWorkerNodes.IsNull() && !stateWorkerNodes.IsUnknown() {
		state.WorkerNodes = stateWorkerNodes
	}
	if state.EtcdNodes.IsNull() && !stateEtcdNodes.IsNull() && !stateEtcdNodes.IsUnknown() {
		state.EtcdNodes = stateEtcdNodes
	}

	// CRITICAL FIX: Restore optional P3 fields from prior state ONLY if BCM returned null
	// This allows drift detection to work for fields BCM does return (like version, management_network)
	// while preserving state for P3 fields BCM may not return
	if state.ManagementNetwork.IsNull() && !stateManagementNetwork.IsNull() && !stateManagementNetwork.IsUnknown() {
		state.ManagementNetwork = stateManagementNetwork
	}
	if state.OverlayNetwork.IsNull() && !stateOverlayNetwork.IsNull() && !stateOverlayNetwork.IsUnknown() {
		state.OverlayNetwork = stateOverlayNetwork
	}
	if state.DNSServers.IsNull() && !stateDNSServers.IsNull() && !stateDNSServers.IsUnknown() {
		state.DNSServers = stateDNSServers
	}
	if state.Version.IsNull() && !stateVersion.IsNull() && !stateVersion.IsUnknown() {
		state.Version = stateVersion
	}
	if state.CNIPlugin.IsNull() && !stateCNIPlugin.IsNull() && !stateCNIPlugin.IsUnknown() {
		state.CNIPlugin = stateCNIPlugin
	}
	if state.StorageClasses.IsNull() && !stateStorageClasses.IsNull() && !stateStorageClasses.IsUnknown() {
		state.StorageClasses = stateStorageClasses
	}
	if state.LoadBalancerMode.IsNull() && !stateLoadBalancerMode.IsNull() && !stateLoadBalancerMode.IsUnknown() {
		state.LoadBalancerMode = stateLoadBalancerMode
	}
	if state.Addons.IsNull() && !stateAddons.IsNull() && !stateAddons.IsUnknown() {
		state.Addons = stateAddons
	}
	if state.IngressController.IsNull() && !stateIngressController.IsNull() && !stateIngressController.IsUnknown() {
		state.IngressController = stateIngressController
	}

	// Save updated state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// readCluster is a helper function to read cluster state from BCM API.
func (r *CMKubeClusterResource) readCluster(ctx context.Context, model *CMKubeClusterResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	clusterUUID := model.UUID.ValueString()

	tflog.Debug(ctx, "Reading Kubernetes cluster from BCM API", map[string]interface{}{
		"uuid": clusterUUID,
	})

	// Call BCM API with direct UUID lookup (args pattern)
	body, err := r.Client.CallJSONRPC(ctx, "cmkube", "getKubeCluster", clusterUUID)
	if err != nil {
		diags.AddError(
			"Error Reading Kubernetes Cluster",
			fmt.Sprintf("Could not read cluster UUID %s: %s", clusterUUID, err.Error()),
		)
		return diags
	}

	// Parse cluster data
	var clusterData map[string]interface{}
	if err := json.Unmarshal(body, &clusterData); err != nil {
		diags.AddError(
			"Error Parsing Cluster Data",
			fmt.Sprintf("Could not parse cluster response: %s", err.Error()),
		)
		return diags
	}

	// Map BCM API fields to Terraform model
	model.Name = getStringValue(clusterData, "name")

	// Master nodes and Worker nodes
	// NOTE: BCM cmkube API behavior for node lists:
	// - masterNodes and workerNodes are write-only fields (used during create/update)
	// - getKubeCluster does NOT return these fields in the response
	// - We preserve the plan values to maintain state consistency
	// - ImportState ignores these fields (see test ImportStateVerifyIgnore)
	if masterNodes, ok := clusterData["masterNodes"].([]interface{}); ok && len(masterNodes) > 0 {
		elements := make([]attr.Value, len(masterNodes))
		for i, node := range masterNodes {
			if nodeStr, ok := node.(string); ok {
				elements[i] = types.StringValue(nodeStr)
			}
		}
		model.MasterNodes, _ = types.ListValue(types.StringType, elements)
	} else {
		// If BCM doesn't return master nodes, preserve existing value or set empty list
		if model.MasterNodes.IsNull() || model.MasterNodes.IsUnknown() {
			model.MasterNodes, _ = types.ListValue(types.StringType, []attr.Value{})
		}
	}

	// Worker nodes
	if workerNodes, ok := clusterData["workerNodes"].([]interface{}); ok && len(workerNodes) > 0 {
		elements := make([]attr.Value, len(workerNodes))
		for i, node := range workerNodes {
			if nodeStr, ok := node.(string); ok {
				elements[i] = types.StringValue(nodeStr)
			}
		}
		model.WorkerNodes, _ = types.ListValue(types.StringType, elements)
	} else {
		// CRITICAL FIX: Preserve null vs empty list distinction
		// - If model currently has null/unknown, keep it null (don't convert to empty list)
		// - If model currently has empty list, preserve it as empty list
		// - This prevents "Provider produced inconsistent result after apply" errors
		if model.WorkerNodes.IsNull() || model.WorkerNodes.IsUnknown() {
			model.WorkerNodes = types.ListNull(types.StringType)
		}
		// else: preserve existing plan value (which might be an empty list)
	}

	// Etcd nodes (write-only field - same pattern as master_nodes/worker_nodes)
	// BCM cmkube API accepts etcdNodes during create/update but does NOT return them in getKubeCluster
	if model.EtcdNodes.IsNull() || model.EtcdNodes.IsUnknown() {
		model.EtcdNodes = types.ListNull(types.StringType)
	}
	// else: preserve existing state value (which was set via plan)

	// Network configuration (optional)
	model.ManagementNetwork = getStringValue(clusterData, "managementNetwork")
	model.OverlayNetwork = getStringValue(clusterData, "overlayNetwork")

	// DNS servers (optional list)
	if dnsServers, ok := clusterData["dnsServers"].([]interface{}); ok && len(dnsServers) > 0 {
		elements := make([]attr.Value, len(dnsServers))
		for i, server := range dnsServers {
			if serverStr, ok := server.(string); ok {
				elements[i] = types.StringValue(serverStr)
			}
		}
		model.DNSServers, _ = types.ListValue(types.StringType, elements)
	} else {
		if model.DNSServers.IsNull() || model.DNSServers.IsUnknown() {
			model.DNSServers = types.ListNull(types.StringType)
		}
	}

	// Kubernetes configuration (optional)
	model.Version = getStringValue(clusterData, "version")
	model.CNIPlugin = getStringValue(clusterData, "cniPlugin")

	// Storage configuration (optional, JSON-encoded)
	model.StorageClasses = getStringValue(clusterData, "storageClasses")

	// Load balancer configuration (optional)
	model.LoadBalancerMode = getStringValue(clusterData, "loadBalancerMode")

	// Cluster addons (optional, JSON-encoded)
	model.Addons = getStringValue(clusterData, "addons")

	// Ingress controller (optional, JSON-encoded)
	model.IngressController = getStringValue(clusterData, "ingressController")

	// Computed fields
	model.CreationTime = getInt64Value(clusterData, "creationTime")
	model.RevisionID = getInt64Value(clusterData, "revisionID")

	// Force is a client-side operation parameter, not stored in BCM
	// Preserve current value or default to false if null/unknown
	if model.Force.IsNull() || model.Force.IsUnknown() {
		model.Force = types.BoolValue(false)
	}

	return diags
}

// Update updates an existing Kubernetes cluster.
func (r *CMKubeClusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.Client == nil {
		resp.Diagnostics.AddError(
			"Provider Not Configured",
			"The provider has not been configured. Please ensure the provider block is properly configured.",
		)
		return
	}

	var plan CMKubeClusterResourceModel
	var state CMKubeClusterResourceModel

	// Read plan and current state
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build cluster entity with UUID for update
	entity, diags := buildClusterEntity(ctx, plan, state.UUID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Pre-flight validation: Call validateKubeCluster before UPDATE
	// Note: Service name is "cmkube" (lowercase) - exception to CamelCase pattern
	validationErrors, err := r.Client.ValidateEntity(ctx, "cmkube", "validateKubeCluster", entity, false)
	if err != nil {
		resp.Diagnostics.AddError(
			"Validation API Error",
			fmt.Sprintf("Could not validate cluster '%s': %s", plan.Name.ValueString(), err.Error()),
		)
		return
	}

	// Process validation results - halt if errors found
	if ProcessValidationErrors(validationErrors, &resp.Diagnostics) {
		return
	}

	// Call BCM API to update cluster
	tflog.Debug(ctx, "Updating Kubernetes cluster via BCM API", map[string]interface{}{
		"uuid": state.UUID.ValueString(),
		"name": plan.Name.ValueString(),
	})

	_, err = r.Client.CallJSONRPC(ctx, "cmkube", "updateKubeCluster", entity, plan.Force.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Kubernetes Cluster",
			fmt.Sprintf("Could not update cluster UUID %s: %s", state.UUID.ValueString(), err.Error()),
		)
		return
	}

	tflog.Info(ctx, "Updated Kubernetes cluster", map[string]interface{}{
		"uuid": state.UUID.ValueString(),
	})

	// Preserve UUID and ID from state
	plan.UUID = state.UUID
	plan.ID = state.ID

	// Read back updated state
	// CRITICAL: Preserve optional fields from plan before reading (BCM API may not return them)
	// Node lists (master_nodes, worker_nodes, etcd_nodes) are write-only fields
	planMasterNodes := plan.MasterNodes
	planWorkerNodes := plan.WorkerNodes
	planEtcdNodes := plan.EtcdNodes
	planManagementNetwork := plan.ManagementNetwork
	planOverlayNetwork := plan.OverlayNetwork
	planDNSServers := plan.DNSServers
	planVersion := plan.Version
	planCNIPlugin := plan.CNIPlugin
	planStorageClasses := plan.StorageClasses
	planLoadBalancerMode := plan.LoadBalancerMode
	planAddons := plan.Addons
	planIngressController := plan.IngressController

	readDiags := r.readCluster(ctx, &plan)
	resp.Diagnostics.Append(readDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// CRITICAL FIX: Restore write-only node list fields from plan
	// These fields are accepted by BCM API but NOT returned in getKubeCluster response
	if !planMasterNodes.IsUnknown() && !planMasterNodes.IsNull() {
		plan.MasterNodes = planMasterNodes
	}
	if !planWorkerNodes.IsUnknown() && !planWorkerNodes.IsNull() {
		plan.WorkerNodes = planWorkerNodes
	}
	if !planEtcdNodes.IsUnknown() && !planEtcdNodes.IsNull() {
		plan.EtcdNodes = planEtcdNodes
	}

	// CRITICAL FIX: Restore optional P3 fields from plan ONLY if they're known values
	// BCM API may not return these fields, but we want to keep the plan values in state
	// Never propagate Unknown values - they cause "invalid result object" errors
	if !planManagementNetwork.IsUnknown() && !planManagementNetwork.IsNull() {
		plan.ManagementNetwork = planManagementNetwork
	}
	if !planOverlayNetwork.IsUnknown() && !planOverlayNetwork.IsNull() {
		plan.OverlayNetwork = planOverlayNetwork
	}
	if !planDNSServers.IsUnknown() && !planDNSServers.IsNull() {
		plan.DNSServers = planDNSServers
	}
	if !planVersion.IsUnknown() && !planVersion.IsNull() {
		plan.Version = planVersion
	}
	if !planCNIPlugin.IsUnknown() && !planCNIPlugin.IsNull() {
		plan.CNIPlugin = planCNIPlugin
	}
	if !planStorageClasses.IsUnknown() && !planStorageClasses.IsNull() {
		plan.StorageClasses = planStorageClasses
	}
	if !planLoadBalancerMode.IsUnknown() && !planLoadBalancerMode.IsNull() {
		plan.LoadBalancerMode = planLoadBalancerMode
	}
	if !planAddons.IsUnknown() && !planAddons.IsNull() {
		plan.Addons = planAddons
	}
	if !planIngressController.IsUnknown() && !planIngressController.IsNull() {
		plan.IngressController = planIngressController
	}

	// Save state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes a Kubernetes cluster.
func (r *CMKubeClusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.Client == nil {
		resp.Diagnostics.AddError(
			"Provider Not Configured",
			"The provider has not been configured. Please ensure the provider block is properly configured.",
		)
		return
	}

	var state CMKubeClusterResourceModel

	// Read current state
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterUUID := state.UUID.ValueString()

	tflog.Debug(ctx, "Deleting Kubernetes cluster via BCM API", map[string]interface{}{
		"uuid": clusterUUID,
	})

	// Call BCM API to delete cluster
	_, err := r.Client.CallJSONRPC(ctx, "cmkube", "removeKubeCluster", clusterUUID, state.Force.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Kubernetes Cluster",
			fmt.Sprintf("Could not delete cluster UUID %s: %s", clusterUUID, err.Error()),
		)
		return
	}

	tflog.Info(ctx, "Deleted Kubernetes cluster", map[string]interface{}{
		"uuid": clusterUUID,
	})

	// State is automatically cleared by framework after successful Delete
}

// ImportState imports an existing cluster by UUID.
func (r *CMKubeClusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by ID (consistent with other resources - id and uuid are equivalent)
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	// Also set UUID to same value for consistency
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("uuid"), req.ID)...)
}

// buildClusterEntity constructs a BCM KubeCluster entity from Terraform model.
// NOTE: BCM cmkube API requires client-generated UUIDs (unlike other BCM services).
func buildClusterEntity(ctx context.Context, model CMKubeClusterResourceModel, uuid string) (map[string]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	entity := map[string]interface{}{
		"baseType":      "KubeCluster",
		"childType":     "",
		"modified":      true,
		"to_be_removed": false,
		"revision":      "",
	}

	// Add UUID if updating
	if uuid != "" {
		entity["uuid"] = uuid
	}

	// Required fields
	entity["name"] = model.Name.ValueString()

	// Master nodes (required)
	var masterNodes []string
	diags.Append(model.MasterNodes.ElementsAs(ctx, &masterNodes, false)...)
	entity["masterNodes"] = masterNodes

	// Worker nodes (optional)
	if !model.WorkerNodes.IsNull() && !model.WorkerNodes.IsUnknown() {
		var workerNodes []string
		diags.Append(model.WorkerNodes.ElementsAs(ctx, &workerNodes, false)...)
		entity["workerNodes"] = workerNodes
	} else {
		entity["workerNodes"] = []string{}
	}

	// Etcd nodes (optional) - for dedicated etcd cluster members
	if !model.EtcdNodes.IsNull() && !model.EtcdNodes.IsUnknown() {
		var etcdNodes []string
		diags.Append(model.EtcdNodes.ElementsAs(ctx, &etcdNodes, false)...)
		entity["etcdNodes"] = etcdNodes
	}

	// Network configuration (optional)
	if !model.ManagementNetwork.IsNull() && !model.ManagementNetwork.IsUnknown() {
		entity["managementNetwork"] = model.ManagementNetwork.ValueString()
	}

	if !model.OverlayNetwork.IsNull() && !model.OverlayNetwork.IsUnknown() {
		entity["overlayNetwork"] = model.OverlayNetwork.ValueString()
	}

	// DNS servers (optional list)
	if !model.DNSServers.IsNull() && !model.DNSServers.IsUnknown() {
		var dnsServers []string
		diags.Append(model.DNSServers.ElementsAs(ctx, &dnsServers, false)...)
		entity["dnsServers"] = dnsServers
	}

	// Kubernetes configuration (optional)
	if !model.Version.IsNull() && !model.Version.IsUnknown() {
		entity["version"] = model.Version.ValueString()
	}

	if !model.CNIPlugin.IsNull() && !model.CNIPlugin.IsUnknown() {
		entity["cniPlugin"] = model.CNIPlugin.ValueString()
	}

	// Storage configuration (optional, JSON-encoded)
	if !model.StorageClasses.IsNull() && !model.StorageClasses.IsUnknown() {
		// Store as JSON string that BCM API will parse
		entity["storageClasses"] = model.StorageClasses.ValueString()
	}

	// Load balancer configuration (optional)
	if !model.LoadBalancerMode.IsNull() && !model.LoadBalancerMode.IsUnknown() {
		entity["loadBalancerMode"] = model.LoadBalancerMode.ValueString()
	}

	// Cluster addons (optional, JSON-encoded)
	if !model.Addons.IsNull() && !model.Addons.IsUnknown() {
		entity["addons"] = model.Addons.ValueString()
	}

	// Ingress controller (optional, JSON-encoded)
	if !model.IngressController.IsNull() && !model.IngressController.IsUnknown() {
		entity["ingressController"] = model.IngressController.ValueString()
	}

	return entity, diags
}
