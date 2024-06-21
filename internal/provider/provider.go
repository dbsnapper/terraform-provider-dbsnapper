// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"
	"terraform-provider-dbsnapper/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure dbSnapperProvider satisfies various provider interfaces.
var _ provider.Provider = &dbSnapperProvider{}

const baseURLProduction = "https://app.dbsnapper.com/api/v3"

// dbSnapperProvider defines the provider implementation.
type dbSnapperProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}
type dbSnapperProviderModel struct {
	AuthToken types.String `tfsdk:"authtoken"`
	BaseURL   types.String `tfsdk:"base_url"`
}

func (p *dbSnapperProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "dbsnapper"
	resp.Version = p.version
}

func (p *dbSnapperProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"authtoken": schema.StringAttribute{
				Description: "DBSnapper API Authtoken",
				Optional:    true,
				Sensitive:   true,
			},
			"base_url": schema.StringAttribute{
				Description: "DBSnapper API Base URL - for internal testing only",
				Optional:    true,
			},
		},
	}
}

func (p *dbSnapperProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config dbSnapperProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if config.AuthToken.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("authtoken"),
			"Unknown DBSnapper authtoken",
			"The provider cannot create the DBSnapper API client as there is an unknown configuration value for the DBSnapper API Authtoken."+
				" Either target apply the source of the value first, set the value statically in the configuration, or use the DBSNAPPER_AUTHTOKEN environment variable.",
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// Default to Env Vars / override with TF config value if set
	authtoken := os.Getenv("DBSNAPPER_AUTHTOKEN")
	baseURL := os.Getenv("DBSNAPPER_BASE_URL")

	if !config.AuthToken.IsNull() {
		authtoken = config.AuthToken.ValueString()
	}
	if !config.BaseURL.IsNull() {
		baseURL = config.BaseURL.ValueString()
	}

	// If expected configurations are missing, return errors

	// If the baseURL is not set, default to production
	if baseURL == "" {
		baseURL = baseURLProduction
	}

	if authtoken == "" {
		resp.Diagnostics.AddAttributeError(path.Root("host"),
			"Missing DBSnapper API AuthToken", "The provider cannot create the DBSnapper API client as there is a missing or empty value for the DBSnapper API authtoken. "+
				"Set the host value in the configuration or use the DBSNAPPER_AUTHTOKEN environment variable. "+
				"If either is already set, ensure the value is not empty.")
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// Create client with configuration values
	dbs := client.NewDBSnapper(authtoken, baseURL)
	if !dbs.IsReady {
		resp.Diagnostics.AddError("Failed to create DBSnapper API client", "API Not Ready")
		return
	}

	resp.DataSourceData = dbs
	resp.ResourceData = dbs

}

type Resourcer interface {
	GetResource() *resource.Resource
}

func (p *dbSnapperProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewTargetResource,
		NewStorageProviderResource,
	}
}

func (p *dbSnapperProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewTargetsDataSource,
	}
}

func (p *dbSnapperProvider) Functions(ctx context.Context) []func() function.Function {
	return nil
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &dbSnapperProvider{
			version: version,
		}
	}
}
