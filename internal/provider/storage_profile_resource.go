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
	_ resource.Resource                = &storageProfileResource{}
	_ resource.ResourceWithConfigure   = &storageProfileResource{}
	_ resource.ResourceWithImportState = &storageProfileResource{}
)

func NewStorageProviderResource() resource.Resource {
	return &storageProfileResource{}
}

// storageProfileResource defines the resource implementation.
type storageProfileResource struct {
	client *client.DBSnapper
}

type StorageProfileResourceModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Provider  types.String `tfsdk:"sp_provider"`
	Region    types.String `tfsdk:"region"`
	AccountID types.String `tfsdk:"account_id"`
	AccessKey types.String `tfsdk:"access_key"`
	SecretKey types.String `tfsdk:"secret_key"`
	Bucket    types.String `tfsdk:"bucket"`
	Prefix    types.String `tfsdk:"prefix"`
	Status    types.String `tfsdk:"status"`

	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

func (r *storageProfileResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_storage_profile"
}

func (r *storageProfileResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Storage Profile resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier for the storage profile",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Description: "The time the storage profile was created",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "The time the storage profile was last updated",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the storage profile",
				Required:    true,
			},
			"sp_provider": schema.StringAttribute{
				Description: "The provider of the storage profile",
				Required:    true,
			},
			"region": schema.StringAttribute{
				Description: "The region of the storage profile",
				Optional:    true,
			},
			"account_id": schema.StringAttribute{
				Description: "The account ID of the storage provider - Required for Cloudflare",
				Optional:    true,
			},
			"access_key": schema.StringAttribute{
				Description: "The access key of the storage provider",
				Required:    true,
				Sensitive:   true,
			},
			"secret_key": schema.StringAttribute{
				Description: "The secret key of the storage provider",
				Required:    true,
				Sensitive:   true,
			},
			"bucket": schema.StringAttribute{
				Description: "The bucket of the storage provider",
				Required:    true,
			},
			"prefix": schema.StringAttribute{
				Description: "The prefix of the storage provider",
				Optional:    true,
			},
			"status": schema.StringAttribute{
				Description: "The status of the storage profile",
				Computed:    true,
			},
		},
	}
}

func (r *storageProfileResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates a new storage profile resource.
func (r *storageProfileResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan = new(StorageProfileResourceModel)

	// 1. Read Terraform PLAN into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read Terraform PLAN data into the model
	plan, err := TFToSPResourceModel(ctx, plan)

	if err != nil {
		resp.Diagnostics.AddError("Error reading Terraform plan", fmt.Sprintf("Unable to read Terraform plan, got error: %s", err))
		return
	}

	// Generate API Request Body
	spApiRequest, err := SPResourceModelToAPIRequest(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError("API Request Error", fmt.Sprintf("Unable to create API request body, got error: %s", err))
		return
	}

	// Call API to create storage profile
	targetResponse, err := r.client.API.CreateStorageProfile(spApiRequest)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create storage profile, got error: %s", err))
		return
	}

	// API response to Terraform mapping
	plan, err = APIResponseToSPResourceModel(ctx, targetResponse, plan)
	if err != nil {
		resp.Diagnostics.AddError("API Response Error", fmt.Sprintf("Unable to create storage profile, got error: %s", err))
		return
	}

	tflog.Info(ctx, "DBSnapper Provider: Create storage profile Resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

////////////////////////////// READ //////////////////////////////

// Read refreshes the Terraform state with the latest data.
func (r *storageProfileResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state = new(StorageProfileResourceModel)

	// 1. Read Terraform State into the model
	resp.Diagnostics.Append(req.State.Get(ctx, state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Call API to read the storage profile
	targetResponse, err := r.client.API.GetStorageProfile(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read storage profile, got error: %s", err))
		return
	}

	// API response to Terraform mapping
	state, err = APIResponseToSPResourceModel(ctx, targetResponse, state)
	if err != nil {
		resp.Diagnostics.AddError("API Response Error", fmt.Sprintf("Unable to map API response to Terraform state, got error: %s", err))
		return
	}

	tflog.Info(ctx, "DBSnapper Provider: Read storage profile Resource")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

////////////////////////////// UPDATE //////////////////////////////

// Update updates the storage profile resource.
func (r *storageProfileResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan = new(StorageProfileResourceModel)

	// Read Terraform PLAN data into the model.
	resp.Diagnostics.Append(req.Plan.Get(ctx, plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Convert Terraform PLAN data into the model
	plan, err := TFToSPResourceModel(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError("Error reading Terraform plan", fmt.Sprintf("Unable to read Terraform plan, got error: %s", err))
		return
	}

	// Generate API Request Body
	storageProfileApiRequest, err := SPResourceModelToAPIRequest(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError("API Request Error", fmt.Sprintf("Unable to create API request body, got error: %s", err))
		return
	}

	// Update storage profile via API
	storageProfileResponse, err := r.client.API.UpdateStorageProfile(plan.ID.ValueString(), storageProfileApiRequest)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update storage profile id: %s, got error: %s", plan.ID.ValueString(), err))
		return
	}

	// API response to Terraform mapping
	plan, err = APIResponseToSPResourceModel(ctx, storageProfileResponse, plan)
	if err != nil {
		resp.Diagnostics.AddError("API Response Error", fmt.Sprintf("Unable to update storage profile, got error: %s", err))
		return
	}

	tflog.Info(ctx, "DBSnapper Provider: Update storage profile Resource")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

////////////////////////////// DELETE //////////////////////////////

// Delete deletes the storage profile resource.
func (r *storageProfileResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data StorageProfileResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete storage profile via API
	err := r.client.API.DeleteStorageProfile(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete storage profile, got error: %s", err))
		return
	}
}

////////////////////////////// IMPORT STATE //////////////////////////////

// ImportState imports the storage profile resource state.
func (r *storageProfileResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
