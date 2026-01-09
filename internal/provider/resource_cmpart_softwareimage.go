// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"time"

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
	_ resource.Resource                = &CMPartSoftwareImageResource{}
	_ resource.ResourceWithImportState = &CMPartSoftwareImageResource{}
)

// CMPartSoftwareImageResource defines the resource implementation.
type CMPartSoftwareImageResource struct {
	BCMResourceBase
}

// CMPartSoftwareImageResourceModel describes the resource data model.
type CMPartSoftwareImageResourceModel struct {
	// Identity fields
	ID   types.String `tfsdk:"id"`
	UUID types.String `tfsdk:"uuid"`
	Name types.String `tfsdk:"name"`
	Path types.String `tfsdk:"path"`

	// Kernel configuration
	KernelVersion       types.String `tfsdk:"kernel_version"`
	KernelParameters    types.String `tfsdk:"kernel_parameters"`
	KernelOutputConsole types.String `tfsdk:"kernel_output_console"`

	// Serial Over LAN (SOL) configuration
	EnableSOL      types.Bool   `tfsdk:"enable_sol"`
	SOLPort        types.String `tfsdk:"sol_port"`
	SOLSpeed       types.String `tfsdk:"sol_speed"`
	SOLFlowControl types.Bool   `tfsdk:"sol_flow_control"`

	// Kernel modules (using types.List to handle Unknown values during planning)
	Modules types.List `tfsdk:"modules"`

	// Filesystem partitions (optional references)
	FSPart     types.String `tfsdk:"fspart"`
	BootFSPart types.String `tfsdk:"bootfspart"`

	// Metadata fields (computed)
	Notes                   types.String `tfsdk:"notes"`
	CreationTime            types.Int64  `tfsdk:"creation_time"`
	RevisionID              types.Int64  `tfsdk:"revision_id"`
	FileOperationInProgress types.Bool   `tfsdk:"file_operation_in_progress"`
	OriginalImage           types.String `tfsdk:"original_image"`
	ParentSoftwareImage     types.String `tfsdk:"parent_software_image"`

	// Force parameter
	Force types.Bool `tfsdk:"force"` // Optional, default: false
}

// KernelModuleResourceModel describes a kernel module nested object.
type KernelModuleResourceModel struct {
	Name       types.String `tfsdk:"name"`
	Parameters types.String `tfsdk:"parameters"`
}

// NewCMPartSoftwareImageResource creates a new resource instance.
func NewCMPartSoftwareImageResource() resource.Resource {
	return &CMPartSoftwareImageResource{}
}

// Metadata returns the resource type name.
func (r *CMPartSoftwareImageResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cmpart_softwareimage"
}

// Schema defines the resource schema.
func (r *CMPartSoftwareImageResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a BCM software image (OS image for DPU node provisioning).\n\n" +
			"Software images define the operating system kernel, kernel parameters, modules, and boot configuration " +
			"used to provision compute nodes in the BCM cluster.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Resource identifier (same as UUID)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"uuid": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier assigned by BCM",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Software image name (must be unique)",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
			},
			"path": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Image path in BCM filesystem (e.g., `/cm/images/ubuntu-22.04`). Must be unique.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^(/[-+_.a-zA-Z0-9]+)+/?(@\d+)?$`),
						"path must match format: /cm/images/name",
					),
				},
			},
			"kernel_version": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Kernel version string (e.g., `5.15.0-58-generic`). When cloning an image, this value is inherited from the source image and becomes known after the clone completes.",
			},
			"kernel_parameters": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Kernel command-line parameters (e.g., `quiet splash`)",
			},
			"kernel_output_console": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("tty0"),
				MarkdownDescription: "Kernel output console device. Defaults to `tty0`.",
			},
			"enable_sol": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Enable Serial Over LAN for remote console access. Defaults to `false`.",
			},
			"sol_port": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("ttyS1"),
				MarkdownDescription: "SOL serial port device. Defaults to `ttyS1`.",
			},
			"sol_speed": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("115200"),
				MarkdownDescription: "SOL baud rate. Valid values: `115200`, `57600`, `38400`, `19200`, `9600`, `4800`, `2400`, `1200`. Defaults to `115200`.",
				Validators: []validator.String{
					stringvalidator.OneOf("115200", "57600", "38400", "19200", "9600", "4800", "2400", "1200"),
				},
			},
			"sol_flow_control": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Enable SOL hardware flow control. Defaults to `true`.",
			},
			"fspart": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Filesystem partition UUID reference (auto-generated when cloning)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"bootfspart": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Boot filesystem partition UUID reference (auto-generated when cloning)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"notes": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "User notes or description for the software image",
			},
			"creation_time": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Unix timestamp of image creation (seconds since epoch)",
			},
			"revision_id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Image revision number",
			},
			"file_operation_in_progress": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Indicates if a file operation is currently in progress for this image",
			},
			"original_image": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "UUID of the original image to clone from. When set, BCM will copy the filesystem from the specified image. This is only used during resource creation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"parent_software_image": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID of the parent image if this is a revision",
			},
			"modules": schema.ListNestedAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "List of kernel modules to load at boot",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Kernel module name (e.g., `nvidia-drm`, `e1000e`)",
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
							},
						},
						"parameters": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "Module parameters (e.g., `modeset=1`)",
						},
					},
				},
			},
			"force": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Force deletion even if categories reference this software image. **WARNING**: Force deletion may create orphaned references in the BCM database. Use with caution.",
			},
		},
	}
}

// Configure stores the BCM client in the resource instance.
func (r *CMPartSoftwareImageResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.ConfigureResource(req, resp)
}

// Create implements the resource Create operation (REFACTOR phase - API integration).
func (r *CMPartSoftwareImageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.Client == nil {
		resp.Diagnostics.AddError(
			"Provider Not Configured",
			"The provider has not been configured. Please ensure the provider block is properly configured.",
		)
		return
	}

	var plan CMPartSoftwareImageResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build API entity from plan
	// UUID will be generated automatically by buildAPIEntity for new resources
	entity := r.buildAPIEntity(&plan, "")

	// Pre-flight validation: Call validateSoftwareImage before CREATE
	validationErrors, err := r.Client.ValidateEntity(ctx, "CMPart", "validateSoftwareImage", entity, true)
	if err != nil {
		resp.Diagnostics.AddError(
			"Validation API Error",
			fmt.Sprintf("Could not validate software image '%s': %s", plan.Name.ValueString(), err.Error()),
		)
		return
	}

	// Process validation results - halt if errors found
	if ProcessValidationErrors(validationErrors, &resp.Diagnostics) {
		return
	}

	// Log the complete request entity for debugging
	entityJSON, _ := json.MarshalIndent(entity, "", "  ")
	tflog.Debug(ctx, "Creating software image via BCM API with entity", map[string]interface{}{
		"name":   plan.Name.ValueString(),
		"path":   plan.Path.ValueString(),
		"entity": string(entityJSON),
	})

	createBody, err := r.Client.CallJSONRPC(ctx, "CMPart", "addSoftwareImage", entity, false)
	if err != nil {
		resp.Diagnostics.AddError(
			"Software Image Creation Failed",
			fmt.Sprintf("Failed to create software image '%s': %s", plan.Name.ValueString(), err.Error()),
		)
		return
	}

	// Parse UUID from response (may be string, object with uuid, or complex response with updated_entity)
	var createdUUID string
	var uuidStr string
	if err := json.Unmarshal(createBody, &uuidStr); err == nil {
		createdUUID = uuidStr
	} else {
		var objResp map[string]interface{}
		if err := json.Unmarshal(createBody, &objResp); err == nil {
			// Check for direct uuid field
			if uuid, ok := objResp["uuid"].(string); ok {
				createdUUID = uuid
			} else if updatedEntity, ok := objResp["updated_entity"].(map[string]interface{}); ok {
				// Check for uuid in updated_entity (clone response format)
				if uuid, ok := updatedEntity["uuid"].(string); ok {
					createdUUID = uuid
				}
			}
		}
	}

	if createdUUID == "" {
		resp.Diagnostics.AddError(
			"UUID Extraction Failed",
			fmt.Sprintf("Failed to extract UUID from create response: %s", string(createBody)),
		)
		return
	}

	// Set ID and UUID
	plan.ID = types.StringValue(createdUUID)
	plan.UUID = types.StringValue(createdUUID)

	// STEP 2.5: Wait for clone operation to complete (eventual consistency handling)
	// BCM's addSoftwareImage initiates clone asynchronously - kernel files aren't immediately available
	// Poll fileOperationInProgress field until clone completes
	if !plan.OriginalImage.IsNull() {
		tflog.Debug(ctx, "Waiting for clone operation to complete", map[string]interface{}{
			"uuid": createdUUID,
		})

		maxRetries := 6 // 1s + 2s + 4s + 8s + 16s = 31s total wait time
		var lastErr error
		cloneComplete := false

		for attempt := 0; attempt < maxRetries; attempt++ {
			if attempt > 0 {
				// Exponential backoff: 1s, 2s, 4s, 8s, 16s
				waitDuration := time.Duration(1<<uint(attempt-1)) * time.Second
				tflog.Debug(ctx, "Waiting before retry", map[string]interface{}{
					"attempt":      attempt + 1,
					"wait_seconds": waitDuration.Seconds(),
				})
				time.Sleep(waitDuration)
			}

			// Check if clone is complete by reading fileOperationInProgress
			// NOTE: We use getSoftwareImages() (list all) instead of getSoftwareImage(uuid) because
			// the BCM API's getSoftwareImage method does not reliably return the fileOperationInProgress
			// field during active clone operations. The list endpoint provides consistent status updates.
			statusBody, err := r.Client.CallJSONRPC(ctx, "CMPart", "getSoftwareImages")
			if err != nil {
				lastErr = err
				tflog.Warn(ctx, "Failed to check clone status", map[string]interface{}{
					"attempt": attempt + 1,
					"error":   err.Error(),
				})
				continue
			}

			var images []map[string]interface{}
			if err := json.Unmarshal(statusBody, &images); err != nil {
				lastErr = err
				continue
			}

			// Find our newly created image
			for _, img := range images {
				if imgUUID, ok := img["uuid"].(string); ok && imgUUID == createdUUID {
					// Check fileOperationInProgress field
					if fileOpInProgress, ok := img["fileOperationInProgress"].(bool); ok && !fileOpInProgress {
						tflog.Info(ctx, "Clone operation completed", map[string]interface{}{
							"uuid":    createdUUID,
							"attempt": attempt + 1,
						})
						cloneComplete = true
						break
					} else {
						tflog.Debug(ctx, "Clone still in progress", map[string]interface{}{
							"uuid":                    createdUUID,
							"fileOperationInProgress": fileOpInProgress,
							"attempt":                 attempt + 1,
						})
					}
					break
				}
			}

			if cloneComplete {
				break
			}
		}

		if !cloneComplete {
			tflog.Warn(ctx, "Clone operation may still be in progress after max retries", map[string]interface{}{
				"uuid":       createdUUID,
				"maxRetries": maxRetries,
				"lastErr":    lastErr,
				"proceeding": "Will attempt to read image anyway",
			})
		}
	}

	// STEP 3: Read back created resource to populate computed fields
	// Preserve original_image from plan before reading (BCM API resets it after clone)
	planOriginalImage := plan.OriginalImage
	r.readSoftwareImage(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		// Attempt to clean up orphaned resource using correct API method
		if !plan.ID.IsNull() && !plan.ID.IsUnknown() {
			tflog.Warn(ctx, "Attempting to clean up orphaned software image after read failure", map[string]interface{}{
				"uuid": plan.ID.ValueString(),
			})
			// Use removeSoftwareImage (singular) with UUID, matching Delete method signature
			_, cleanupErr := r.Client.CallJSONRPC(ctx, "CMPart", "removeSoftwareImage", plan.ID.ValueString(), false, false, false)
			if cleanupErr != nil {
				tflog.Error(ctx, "Failed to clean up orphaned software image", map[string]interface{}{
					"uuid":  plan.ID.ValueString(),
					"error": cleanupErr.Error(),
				})
			}
		}
		return
	}

	// CRITICAL FIX: Preserve plan's original_image ONLY if it's a known value
	// Never propagate Unknown values to state - they cause "invalid result object" errors
	// BCM API returns zero UUID after cloning, but we want to keep the original value in state
	if !planOriginalImage.IsUnknown() {
		plan.OriginalImage = planOriginalImage
	}
	// If planOriginalImage was Unknown, keep the value from readSoftwareImage (null or actual value)

	tflog.Info(ctx, "Created software image resource", map[string]interface{}{
		"name": plan.Name.ValueString(),
		"uuid": createdUUID,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read implements the resource Read operation (REFACTOR phase - API integration).
func (r *CMPartSoftwareImageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.Client == nil {
		resp.Diagnostics.AddError(
			"Provider Not Configured",
			"The provider has not been configured. Please ensure the provider block is properly configured.",
		)
		return
	}

	var state CMPartSoftwareImageResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve original_image from state before reading (BCM API resets it after clone)
	stateOriginalImage := state.OriginalImage

	// Fetch current state from BCM API
	r.readSoftwareImage(ctx, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// CRITICAL FIX: Restore original_image from prior state ONLY if it's a known value
	// Never propagate Unknown values - they cause "invalid result object" errors
	// BCM API returns zero UUID after cloning, but we want to keep the original value in state
	if !stateOriginalImage.IsUnknown() {
		state.OriginalImage = stateOriginalImage
	}
	// If stateOriginalImage was Unknown, keep the value from readSoftwareImage (null or actual value)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update implements the resource Update operation (REFACTOR phase - API integration).
func (r *CMPartSoftwareImageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.Client == nil {
		resp.Diagnostics.AddError(
			"Provider Not Configured",
			"The provider has not been configured. Please ensure the provider block is properly configured.",
		)
		return
	}

	var plan CMPartSoftwareImageResourceModel
	var state CMPartSoftwareImageResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build updated entity from plan (include UUID for update)
	entity := r.buildAPIEntity(&plan, plan.UUID.ValueString())

	// Pre-flight validation: Call validateSoftwareImage before UPDATE
	validationErrors, err := r.Client.ValidateEntity(ctx, "CMPart", "validateSoftwareImage", entity, false)
	if err != nil {
		resp.Diagnostics.AddError(
			"Validation API Error",
			fmt.Sprintf("Could not validate software image '%s': %s", plan.Name.ValueString(), err.Error()),
		)
		return
	}

	// Process validation results - halt if errors found
	if ProcessValidationErrors(validationErrors, &resp.Diagnostics) {
		return
	}

	// Log the complete update entity for debugging
	entityJSON, _ := json.MarshalIndent(entity, "", "  ")
	tflog.Debug(ctx, "Updating software image via BCM API with entity", map[string]interface{}{
		"name":   plan.Name.ValueString(),
		"uuid":   plan.UUID.ValueString(),
		"entity": string(entityJSON),
	})

	_, err = r.Client.CallJSONRPC(ctx, "CMPart", "updateSoftwareImage", entity, false)
	if err != nil {
		resp.Diagnostics.AddError(
			"Software Image Update Failed",
			fmt.Sprintf("Failed to update software image '%s': %s", plan.Name.ValueString(), err.Error()),
		)
		return
	}

	// STEP 3: Read back updated resource to verify changes
	// Preserve original_image: use plan value if set, otherwise use state value
	// This handles both cases: user updating original_image, or keeping previous value
	planOriginalImage := plan.OriginalImage
	stateOriginalImage := state.OriginalImage
	r.readSoftwareImage(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// CRITICAL FIX: Restore original_image from plan or state, but NEVER use Unknown values
	// Unknown values cause "invalid result object after apply" errors
	// BCM API returns zero UUID after cloning, but we want to preserve the original value
	if !planOriginalImage.IsNull() && !planOriginalImage.IsUnknown() {
		plan.OriginalImage = planOriginalImage
	} else if !stateOriginalImage.IsNull() && !stateOriginalImage.IsUnknown() {
		plan.OriginalImage = stateOriginalImage
	}
	// If both are null/unknown, keep the value from readSoftwareImage (null or actual value)

	tflog.Info(ctx, "Updated software image resource", map[string]interface{}{
		"name": plan.Name.ValueString(),
		"uuid": plan.UUID.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete implements the resource Delete operation (REFACTOR phase - API integration).
func (r *CMPartSoftwareImageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.Client == nil {
		resp.Diagnostics.AddError(
			"Provider Not Configured",
			"The provider has not been configured. Please ensure the provider block is properly configured.",
		)
		return
	}

	var state CMPartSoftwareImageResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get force parameter from state (default to false)
	force := false
	if !state.Force.IsNull() {
		force = state.Force.ValueBool()
	}

	tflog.Debug(ctx, "Deleting software image via BCM API", map[string]interface{}{
		"name":  state.Name.ValueString(),
		"uuid":  state.UUID.ValueString(),
		"force": force,
	})

	// PROACTIVE DEPENDENCY CHECK: Check for dependent categories before deletion (unless force=true)
	if !force {
		tflog.Debug(ctx, "Performing dependency check for software image deletion", map[string]interface{}{
			"image_name": state.Name.ValueString(),
			"image_uuid": state.UUID.ValueString(),
		})

		result, err := CheckCategoriesUsingImage(ctx, r.Client, state.Name.ValueString())
		if err != nil {
			// Dependency check failed - report error so resource is NOT removed from state
			// User can retry or use force=true to skip dependency check
			resp.Diagnostics.AddError(
				"Dependency Check Failed",
				fmt.Sprintf(
					"Unable to verify dependencies for software image '%s': %s\n\n"+
						"Please retry the deletion. If the issue persists, you can set 'force = true' "+
						"to skip the dependency check, but this may create orphaned references.",
					state.Name.ValueString(),
					err.Error(),
				),
			)
			return
		}

		if result.HasDependencies {
			// Dependencies exist - block deletion with detailed error message
			tflog.Info(ctx, "Software image deletion blocked due to dependencies", map[string]interface{}{
				"image_name":      state.Name.ValueString(),
				"image_uuid":      state.UUID.ValueString(),
				"dependent_count": result.DependentCount,
				"dependent_type":  result.DependentType,
			})

			resp.Diagnostics.AddError(
				"Software Image In Use - Cannot Delete",
				BuildDependencyError(
					"Software Image",
					state.Name.ValueString(),
					"category",
					result.Identifiers,
				),
			)
			return
		}

		tflog.Debug(ctx, "No dependencies found - proceeding with deletion", map[string]interface{}{
			"image_name": state.Name.ValueString(),
			"image_uuid": state.UUID.ValueString(),
		})
	} else {
		// Force deletion - log warning about potential orphaned references
		tflog.Warn(ctx, BuildForceDeletionWarning("Software Image", state.Name.ValueString()), map[string]interface{}{
			"image_name": state.Name.ValueString(),
			"image_uuid": state.UUID.ValueString(),
			"force":      true,
		})
	}

	// Call removeSoftwareImage with UUID, removeData=false, removeAll=false, force parameter
	_, err := r.Client.CallJSONRPC(ctx, "CMPart", "removeSoftwareImage", state.UUID.ValueString(), false, false, force)
	if err != nil {
		// Enhanced error handling - check if already deleted (idempotent)
		errStr := err.Error()

		// Check if image was already deleted externally (idempotent delete)
		if containsAny(errStr, []string{"not found", "does not exist", "No such"}) {
			tflog.Info(ctx, "Software image already deleted (idempotent)", map[string]interface{}{
				"name": state.Name.ValueString(),
				"uuid": state.UUID.ValueString(),
			})
			return
		}

		// Other deletion errors
		resp.Diagnostics.AddError(
			"Software Image Deletion Failed",
			fmt.Sprintf("Failed to delete software image '%s': %s", state.Name.ValueString(), err.Error()),
		)
		return
	}

	tflog.Info(ctx, "Deleted software image resource", map[string]interface{}{
		"name":  state.Name.ValueString(),
		"uuid":  state.UUID.ValueString(),
		"force": force,
	})
}

// ImportState implements resource import using UUID.
func (r *CMPartSoftwareImageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Helper Methods

// readSoftwareImage fetches software image from API and populates model.
func (r *CMPartSoftwareImageResource) readSoftwareImage(ctx context.Context, model *CMPartSoftwareImageResourceModel, diags *diag.Diagnostics) {
	// Determine which identifier to use for lookup
	// During import, name may be empty but ID is set (to UUID)
	// During normal operations, name is populated
	var lookupName string
	if !model.Name.IsNull() && model.Name.ValueString() != "" {
		lookupName = model.Name.ValueString()
	} else if !model.ID.IsNull() && model.ID.ValueString() != "" {
		// Import case: ID is set (UUID), need to look up by ID
		// First, get all images and find by UUID
		allImagesBody, err := r.Client.CallJSONRPC(ctx, "CMPart", "getSoftwareImages")
		if err != nil {
			diags.AddError(
				"Software Image Read Failed",
				fmt.Sprintf("Failed to list software images during import: %s", err.Error()),
			)
			return
		}

		var imagesList []map[string]interface{}
		if err := json.Unmarshal(allImagesBody, &imagesList); err != nil {
			diags.AddError(
				"Response Parse Error",
				fmt.Sprintf("Failed to parse software images list: %s", err.Error()),
			)
			return
		}

		// Find image with matching UUID
		targetUUID := model.ID.ValueString()
		for _, img := range imagesList {
			if uuid, ok := img["uuid"].(string); ok && uuid == targetUUID {
				if name, ok := img["name"].(string); ok {
					lookupName = name
					break
				}
			}
		}

		if lookupName == "" {
			diags.AddError(
				"Software Image Not Found",
				fmt.Sprintf("Software image with ID '%s' not found in BCM", targetUUID),
			)
			return
		}
	} else {
		diags.AddError(
			"Invalid State",
			"Cannot read software image: both name and ID are empty",
		)
		return
	}

	tflog.Debug(ctx, "Reading software image via BCM API", map[string]interface{}{
		"name": lookupName,
	})

	// Use efficient getSoftwareImage(name) API for direct lookup
	body, err := r.Client.CallJSONRPC(ctx, "CMPart", "getSoftwareImage", lookupName)
	if err != nil {
		diags.AddError(
			"Software Image Read Failed",
			fmt.Sprintf("Failed to read software image '%s': %s", lookupName, err.Error()),
		)
		return
	}

	// Parse response as single software image entity
	var imageData map[string]interface{}
	if err := json.Unmarshal(body, &imageData); err != nil {
		diags.AddError(
			"Response Parse Error",
			fmt.Sprintf("Failed to parse software image response: %s", err.Error()),
		)
		return
	}

	// Check if image not found (null or empty response)
	if len(imageData) == 0 {
		diags.AddError(
			"Software Image Not Found",
			fmt.Sprintf("Software image '%s' not found in BCM", model.Name.ValueString()),
		)
		return
	}

	// Map API fields to model
	model.UUID = getStringValue(imageData, "uuid")
	model.Name = getStringValue(imageData, "name")
	model.Path = getStringValue(imageData, "path")

	// Kernel configuration
	model.KernelVersion = getStringValue(imageData, "kernelVersion")
	model.KernelParameters = getStringValue(imageData, "kernelParameters")
	model.KernelOutputConsole = getStringValue(imageData, "kernelOutputConsole")

	// SOL configuration
	model.EnableSOL = getBoolValue(imageData, "enableSOL")
	model.SOLPort = getStringValue(imageData, "SOLPort")
	model.SOLSpeed = getStringValue(imageData, "SOLSpeed")
	model.SOLFlowControl = getBoolValue(imageData, "SOLFlowControl")

	// Partitions
	model.FSPart = getStringValue(imageData, "fspart")
	model.BootFSPart = getStringValue(imageData, "bootfspart")

	// Metadata
	model.Notes = getStringValue(imageData, "notes")
	model.CreationTime = getInt64Value(imageData, "creationTime")
	model.RevisionID = getInt64Value(imageData, "revisionID")
	model.FileOperationInProgress = getBoolValue(imageData, "fileOperationInProgress")

	// Handle original_image: BCM resets this to zero UUID after cloning completes
	// For consistency, we don't expose the zero UUID to users - just set it to null
	// IMPORTANT: Always set to a known value (never preserve Unknown values from plan)
	originalImageFromAPI := getStringValue(imageData, "originalImage")
	zeroUUID := "00000000-0000-0000-0000-000000000000"
	if !originalImageFromAPI.IsNull() && originalImageFromAPI.ValueString() != "" && originalImageFromAPI.ValueString() != zeroUUID {
		model.OriginalImage = originalImageFromAPI
	} else {
		// Set to null if API returns zero UUID, empty, or null
		// This ensures the value is always known after apply (never Unknown)
		model.OriginalImage = types.StringNull()
	}

	model.ParentSoftwareImage = getStringValue(imageData, "parentSoftwareImage")

	// Parse modules list - ALWAYS set to a known value (never leave as Unknown)
	moduleType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"name":       types.StringType,
			"parameters": types.StringType,
		},
	}

	if modulesRaw, ok := imageData["modules"]; ok {
		if modulesList, ok := modulesRaw.([]interface{}); ok {
			modules := make([]KernelModuleResourceModel, 0, len(modulesList))
			for _, modRaw := range modulesList {
				if modData, ok := modRaw.(map[string]interface{}); ok {
					module := KernelModuleResourceModel{
						Name:       getStringValue(modData, "name"),
						Parameters: getStringValue(modData, "parameters"),
					}
					modules = append(modules, module)
				}
			}
			// Convert to types.List
			if len(modules) > 0 {
				moduleObjects := make([]attr.Value, 0, len(modules))
				for _, mod := range modules {
					obj, _ := types.ObjectValue(moduleType.AttrTypes, map[string]attr.Value{
						"name":       mod.Name,
						"parameters": mod.Parameters,
					})
					moduleObjects = append(moduleObjects, obj)
				}
				listVal, _ := types.ListValue(moduleType, moduleObjects)
				model.Modules = listVal
			} else {
				model.Modules, _ = types.ListValue(moduleType, []attr.Value{})
			}
		} else {
			// modules exists but is not a list - set to empty list
			model.Modules, _ = types.ListValue(moduleType, []attr.Value{})
		}
	} else {
		// modules key doesn't exist in API response - set to empty list to ensure known value
		// This prevents "Value Conversion Error" when modules is Unknown in plan
		model.Modules, _ = types.ListValue(moduleType, []attr.Value{})
	}

	// Ensure ID is set for import
	if model.ID.IsNull() && !model.UUID.IsNull() {
		model.ID = model.UUID
	}

	tflog.Trace(ctx, "Successfully read software image", map[string]interface{}{
		"name": model.Name.ValueString(),
		"uuid": model.UUID.ValueString(),
	})
}

// buildAPIEntity constructs BCM API entity from Terraform model
// If uuid is provided, this is an update operation, otherwise it's a create.
func (r *CMPartSoftwareImageResource) buildAPIEntity(model *CMPartSoftwareImageResourceModel, uuid string) map[string]interface{} {
	entity := map[string]interface{}{
		"baseType":      "SoftwareImage",
		"childType":     "",
		"modified":      true,
		"to_be_removed": false,
		"revision":      "",
	}

	// BCM REQUIRES a UUID for software images (like categories and networks)
	// For create: generate new UUID, for update: use existing UUID
	if uuid != "" {
		entity["uuid"] = uuid
	} else {
		// Generate UUID for new software image
		entity["uuid"] = generateUUID()
	}

	// Identity fields
	if !model.Name.IsNull() {
		entity["name"] = model.Name.ValueString()
	}
	if !model.Path.IsNull() {
		entity["path"] = model.Path.ValueString()
	}

	// Kernel configuration
	if !model.KernelVersion.IsNull() {
		entity["kernelVersion"] = model.KernelVersion.ValueString()
	}
	if !model.KernelParameters.IsNull() {
		entity["kernelParameters"] = model.KernelParameters.ValueString()
	}
	if !model.KernelOutputConsole.IsNull() {
		entity["kernelOutputConsole"] = model.KernelOutputConsole.ValueString()
	}

	// SOL configuration
	if !model.EnableSOL.IsNull() {
		entity["enableSOL"] = model.EnableSOL.ValueBool()
	}
	if !model.SOLPort.IsNull() {
		entity["SOLPort"] = model.SOLPort.ValueString()
	}
	if !model.SOLSpeed.IsNull() {
		entity["SOLSpeed"] = model.SOLSpeed.ValueString()
	}
	if !model.SOLFlowControl.IsNull() {
		entity["SOLFlowControl"] = model.SOLFlowControl.ValueBool()
	}

	// Note: fspart and bootfspart are computed-only (auto-generated by BCM)
	// They are not included in API requests

	// Notes
	if !model.Notes.IsNull() {
		entity["notes"] = model.Notes.ValueString()
	}

	// Original image (for cloning) - primarily used during CREATE
	// BCM API requires this field to be present with a valid UUID format
	// If not set or empty, use zero UUID; otherwise use the provided value
	if !model.OriginalImage.IsNull() && model.OriginalImage.ValueString() != "" {
		entity["originalImage"] = model.OriginalImage.ValueString()
	} else {
		// Use zero UUID when originalImage is null or empty
		// This satisfies BCM's validation requirement for UUID format
		entity["originalImage"] = "00000000-0000-0000-0000-000000000000"
	}

	// Build modules list from types.List
	if !model.Modules.IsNull() && !model.Modules.IsUnknown() {
		var moduleModels []KernelModuleResourceModel
		diags := model.Modules.ElementsAs(context.Background(), &moduleModels, false)
		if diags.HasError() {
			// Fallback to empty list if conversion fails
			entity["modules"] = []interface{}{}
		} else {
			modules := make([]map[string]interface{}, 0, len(moduleModels))
			for _, mod := range moduleModels {
				module := map[string]interface{}{
					"baseType":      "KernelModule",
					"childType":     "",
					"modified":      true,
					"to_be_removed": false,
					"revision":      "",
				}
				if !mod.Name.IsNull() {
					module["name"] = mod.Name.ValueString()
				}
				if !mod.Parameters.IsNull() {
					module["parameters"] = mod.Parameters.ValueString()
				} else {
					module["parameters"] = ""
				}
				modules = append(modules, module)
			}
			entity["modules"] = modules
		}
	} else {
		entity["modules"] = []interface{}{}
	}

	return entity
}

// Helper functions:
// - getStringValue, getBoolValue, getInt64Value are defined in data_source_cmpart_softwareimages.go
// - containsAny is defined in resource_cmdevice_category.go (shared helper)
