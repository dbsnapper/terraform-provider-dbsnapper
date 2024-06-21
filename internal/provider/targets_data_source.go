package provider

import (
	"context"
	"fmt"
	"terraform-provider-dbsnapper/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &TargetsDataSource{}
	_ datasource.DataSourceWithConfigure = &TargetsDataSource{}
)

// NewTargetsDataSource is a helper function to simplify the provider implementation.
func NewTargetsDataSource() datasource.DataSource {
	return &TargetsDataSource{}
}

// TargetsDataSource is the data source implementation.
type TargetsDataSource struct {
	client *client.DBSnapper
}

// TargetsDataSourceModel maps the data source schema data.
type TargetsDataSourceModel struct {
	Targets []TargetResourceModel `tfsdk:"targets"`
}

// Metadata returns the data source type name.
func (d *TargetsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_targets"
}

// Schema defines the schema for the data source.
// https://developer.hashicorp.com/terraform/plugin/framework/handling-data/attributes/single-nested
func (d *TargetsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Targets data source",

		Attributes: map[string]schema.Attribute{
			"targets": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The unique identifier for the target",
							Computed:    true,
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
							Computed:    true,
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
							Computed:    true,
							Attributes: map[string]schema.Attribute{
								"src_url": schema.StringAttribute{
									Description: "The source URL for the target snapshot",
									Computed:    true,
								},
								"dst_url": schema.StringAttribute{
									Description: "The destination URL for the target snapshot (will be overwritten)",
									Computed:    true,
								},
								"src_bytes": schema.Int64Attribute{
									Description: "The size of the source database in bytes",
									Computed:    true,
								},
							},
						},

						"sanitize": schema.SingleNestedAttribute{
							Description: "The sanitize configuration",
							Computed:    true,
							Attributes: map[string]schema.Attribute{
								"dst_url": schema.StringAttribute{
									Description: "The destination URL of the database used to sanitizea snapshot",
									Computed:    true,
								},
								"query": schema.StringAttribute{
									Description: "The query used to sanitize the snapshot",
									Computed:    true,
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
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *TargetsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state TargetsDataSourceModel

	targets := d.client.API.GetTargetsTF()

	// if err != nil {
	// 	resp.Diagnostics.AddError("Error fetching targets", err.Error())
	// 	return
	// }
	for _, target := range targets {
		idstr := target.ID.String()
		targetState := TargetResourceModel{

			ID:       types.StringValue(idstr),
			Name:     types.StringValue(target.Name),
			Status:   types.StringValue(target.Status),
			Messages: types.StringValue(target.Messages),
			Snapshot: &targetSnapshotModel{
				SrcURL: types.StringValue(target.Snapshot.SrcURL),
				DstURL: types.StringValue(target.Snapshot.DstURL),
				// SrcBytes: types.Int64Value(target.Snapshot.SrcBytes),
			},
			Sanitize: &targetSanitizeModel{
				DstURL: types.StringValue(target.Sanitize.DstURL),
				Query:  types.StringValue(target.Sanitize.Query),
			},
		}
		// Share
		if target.Share.SsoGroups != nil {
			l, diag := types.ListValueFrom(ctx, types.StringType, target.Share.SsoGroups)
			if diag.HasError() {
				resp.Diagnostics.AddError("Error reading SSOGroups", "")
				return

			}
			targetState.Share = new(targetShareModel)
			targetState.Share.SSOGroups = l
		}
		targetState.CreatedAt = types.StringValue(target.CreatedAt)
		targetState.UpdatedAt = types.StringValue(target.UpdatedAt)

		state.Targets = append(state.Targets, targetState)
	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source.
func (d *TargetsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.DBSnapper)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *DBSnapper, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}
