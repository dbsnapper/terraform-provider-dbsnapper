// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"terraform-provider-dbsnapper/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &targetResource{}
	_ resource.ResourceWithConfigure   = &targetResource{}
	_ resource.ResourceWithImportState = &targetResource{}
)

func NewTargetResource() resource.Resource {
	return &targetResource{}
}

// targetResource defines the resource implementation.
type targetResource struct {
	client *client.DBSnapper
}

type TargetResourceModel struct {
	ID        types.String         `tfsdk:"id"`
	Name      types.String         `tfsdk:"name"`
	Status    types.String         `tfsdk:"status"`
	Messages  types.String         `tfsdk:"messages"`
	Snapshot  *targetSnapshotModel `tfsdk:"snapshot"`
	Sanitize  *targetSanitizeModel `tfsdk:"sanitize"`
	Share     *targetShareModel    `tfsdk:"share"`
	CreatedAt types.String         `tfsdk:"created_at"`
	UpdatedAt types.String         `tfsdk:"updated_at"`
}

// targetSnapshotModel maps snapshot data.
type targetSnapshotModel struct {
	SrcURL         types.String               `tfsdk:"src_url"`
	DstURL         types.String               `tfsdk:"dst_url"`
	SrcBytes       types.Int64                `tfsdk:"src_bytes"`
	StorageProfile *targetStorageProfileModel `tfsdk:"storage_profile"`
}

// targetSanitizeModel maps sanitization data.
type targetSanitizeModel struct {
	DstURL         types.String               `tfsdk:"dst_url"`
	Query          types.String               `tfsdk:"query"`
	StorageProfile *targetStorageProfileModel `tfsdk:"storage_profile"`
}

// targetShareModel maps share data.
type targetShareModel struct {
	SSOGroups types.List `tfsdk:"sso_groups"`
}

type targetStorageProfileModel struct {
	ID types.String `tfsdk:"id"`
}

func (r *targetResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_target"
}

func (r *targetResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Target resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier for the target",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Description: "The time the target was created",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "The time the target was last updated",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the target",
				Required:    true,
			},
			"status": schema.StringAttribute{
				Description: "The status of the target - determined by agent",
				Computed:    true,
			},
			"messages": schema.StringAttribute{
				Description: "The error messages from the target - determined by agent",
				Computed:    true,
			},

			"snapshot": schema.SingleNestedAttribute{
				Description: "The snapshot configuration",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"src_url": schema.StringAttribute{
						Description: "The source URL for the target snapshot",
						Required:    true,
					},
					"dst_url": schema.StringAttribute{
						Description: "The destination URL for the target snapshot (will be overwritten)",
						Optional:    true,
					},
					"src_bytes": schema.Int64Attribute{
						Description: "The size of the source database in bytes",
						Computed:    true,
					},
					"storage_profile": schema.SingleNestedAttribute{
						Description: "Storage provider configuration for Snapshots",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"id": schema.StringAttribute{
								Description: "The unique identifier for the storage profile",
								Optional:    true,
							},
						},
					},
				},
			},

			"sanitize": schema.SingleNestedAttribute{
				Description: "The sanitize configuration",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"dst_url": schema.StringAttribute{
						Description: "The destination URL of the database used to sanitizea snapshot",
						Optional:    true,
					},
					"query": schema.StringAttribute{
						Description: "The query used to sanitize the snapshot",
						Optional:    true,
					},
					"storage_profile": schema.SingleNestedAttribute{
						Description: "Storage provider configuration for Sanitized Snapshots",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"id": schema.StringAttribute{
								Description: "The unique identifier for the storage profile",
								Optional:    true,
							},
						},
					},
				},
			},

			"share": schema.SingleNestedAttribute{
				Description: "The share configuration",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"sso_groups": schema.ListAttribute{
						Description: "The SSO groups that have access to the target snapshot",
						Optional:    true,
						ElementType: types.StringType,
					},
				},
			},
		},
	}
}

func (r *targetResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.DBSnapper)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *DBSnapper, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

////////////////////////////// CREATE //////////////////////////////

// Create creates a new target resource.
func (r *targetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan = new(TargetResourceModel)

	// 1. Read Terraform PLAN into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read Terraform PLAN data into the model
	plan, err := TFToResourceModel(ctx, plan)

	if err != nil {
		resp.Diagnostics.AddError("Error reading Terraform plan", fmt.Sprintf("Unable to read Terraform plan, got error: %s", err))
		return
	}

	// Generate API Request Body
	targetApiRequest, err := ResourceModelToAPIRequest(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError("API Request Error", fmt.Sprintf("Unable to create API request body, got error: %s", err))
		return
	}

	// Call API to create target
	targetResponse, err := r.client.API.CreateTarget(targetApiRequest)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create target, got error: %s", err))
		return
	}

	// API response to Terraform mapping
	plan, err = APIResponseToResourceModel(ctx, targetResponse, plan)
	if err != nil {
		resp.Diagnostics.AddError("API Response Error", fmt.Sprintf("Unable to create target, got error: %s", err))
		return
	}

	tflog.Info(ctx, "DBSnapper Provider: Create Target Resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

////////////////////////////// READ //////////////////////////////

// Read refreshes the Terraform state with the latest data.
func (r *targetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state = new(TargetResourceModel)

	// Read Terraform prior STATE data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	state, err := TFToResourceModel(ctx, state)
	if err != nil {
		resp.Diagnostics.AddError("Error reading Terraform state", fmt.Sprintf("Unable to read Terraform state, got error: %s", err))
		return
	}

	// Call API to get refreshed target data
	targetResponse, err := r.client.API.GetTarget(state.ID.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read target, got error: %s\n Could not read Target ID: %s\n", err, state.ID.ValueString()))

		return
	}

	// API response to Terraform mapping
	state, err = APIResponseToResourceModel(ctx, targetResponse, state)
	if err != nil {
		resp.Diagnostics.AddError("API Response Error", fmt.Sprintf("Unable to create target, got error: %s", err))
		return
	}

	tflog.Info(ctx, "DBSnapper Provider: Read Target Resource")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

////////////////////////////// UPDATE //////////////////////////////

// Update updates the target resource.
func (r *targetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan = new(TargetResourceModel)

	// Read Terraform PLAN data into the model.
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Convert  Terraform PLAN data into the model
	plan, err := TFToResourceModel(ctx, plan)

	if err != nil {
		resp.Diagnostics.AddError("Error reading Terraform plan", fmt.Sprintf("Unable to read Terraform plan, got error: %s", err))
		return
	}

	// Generate API Request Body
	targetApiRequest, err := ResourceModelToAPIRequest(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError("API Request Error", fmt.Sprintf("Unable to create API request body, got error: %s", err))
		return
	}

	// Update target via API
	targetResponse, err := r.client.API.UpdateTarget(plan.ID.ValueString(), targetApiRequest)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Errorf("Unable to update target id: %s, got error: %w", targetResponse.ID.String(), err).Error())
		return
	}

	// API response to Terraform mapping
	plan, err = APIResponseToResourceModel(ctx, targetResponse, plan)
	if err != nil {
		resp.Diagnostics.AddError("API Response Error", fmt.Sprintf("Unable to update target, got error: %s", err))
		return
	}

	tflog.Info(ctx, "DBSnapper Provider: Update Target Resource")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

////////////////////////////// DELETE //////////////////////////////

// Delete deletes the target resource.
func (r *targetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data TargetResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete target via API
	err := r.client.API.DeleteTarget(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete target, got error: %s", err))
		return
	}
}

////////////////////////////// IMPORT STATE //////////////////////////////

// ImportState imports the target resource state.
func (r *targetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
