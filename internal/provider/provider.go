// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/common-fate/sdk/factoryconfig"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure DeploymentProvider satisfies various provider interfaces.
var _ provider.Provider = &DeploymentProvider{}
var _ provider.ProviderWithFunctions = &DeploymentProvider{}

// DeploymentProvider defines the provider implementation.
type DeploymentProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// DeploymentProviderModel describes the provider data model.
type DeploymentProviderModel struct {
	BaseURL        types.String `tfsdk:"base_url"`
	OIDCIssuer     types.String `tfsdk:"oidc_issuer"`
	LicenceKey     types.String `tfsdk:"licence_key"`
	DeploymentName types.String `tfsdk:"deployment_name"`
}

func (p *DeploymentProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "deploymeta"
	resp.Version = p.version
}

func (p *DeploymentProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"base_url": schema.StringAttribute{
				MarkdownDescription: "The Common Fate Factory base URL. Defaults to https://factory.commonfate.io.",
				Optional:            true,
			},
			"oidc_issuer": schema.StringAttribute{
				MarkdownDescription: "The Common Fate Factory OIDC issuer. Defaults to https://factory.commonfate.io.",
				Optional:            true,
			},
			"licence_key": schema.StringAttribute{
				MarkdownDescription: "The Common Fate licence key.",
				Required:            true,
			},
			"deployment_name": schema.StringAttribute{
				MarkdownDescription: "The Common Fate deployment name.",
				Required:            true,
			},
		},
	}
}

func (p *DeploymentProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data DeploymentProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	cfg, err := factoryconfig.Load(context.Background(), factoryconfig.Opts{
		LicenceKey:     data.LicenceKey.ValueString(),
		DeploymentName: data.DeploymentName.ValueString(),
		BaseURL:        data.BaseURL.ValueString(),
		OIDCIssuer:     data.OIDCIssuer.ValueString(),
	})

	if err != nil {
		resp.Diagnostics.AddError("Error loading Common Fate deployment configuration", err.Error())
		return
	}

	resp.DataSourceData = cfg
	resp.ResourceData = cfg
}

func (p *DeploymentProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewNameserversResource,
	}
}

func (p *DeploymentProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewDeploymentDataSource,
	}
}

func (p *DeploymentProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &DeploymentProvider{
			version: version,
		}
	}
}
