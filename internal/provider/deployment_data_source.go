package provider

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/common-fate/sdk/factory/service/deployment"
	"github.com/common-fate/sdk/factoryconfig"
	deploymentv1alpha1 "github.com/common-fate/sdk/gen/commonfate/factory/deployment/v1alpha1"
	"github.com/common-fate/sdk/gen/commonfate/factory/deployment/v1alpha1/deploymentv1alpha1connect"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &DeploymentDataSource{}

func NewDeploymentDataSource() datasource.DataSource {
	return &DeploymentDataSource{}
}

// DeploymentDataSource defines the data source implementation.
type DeploymentDataSource struct {
	client deploymentv1alpha1connect.DeploymentServiceClient
}

// DeploymentDataSourceModel describes the data source data model.
type DeploymentDataSourceModel struct {
	DNSZoneName      types.String `tfsdk:"dns_zone_name"`
	DefaultSubdomain types.String `tfsdk:"default_subdomain"`
	Id               types.String `tfsdk:"id"`
}

func (d *DeploymentDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_deployment"
}

func (d *DeploymentDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Metadata about the current Common Fate deployment.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The deployment ID",
				Computed:            true,
			},
			"dns_zone_name": schema.StringAttribute{
				MarkdownDescription: "The default DNS zone name associated with the deployment",
				Computed:            true,
			},
			"default_subdomain": schema.StringAttribute{
				MarkdownDescription: "The default DNS subdomain associated with the deployment",
				Computed:            true,
			},
		},
	}
}

func (d *DeploymentDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	cfg, ok := req.ProviderData.(*factoryconfig.Context)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *factoryconfig.Context, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = deployment.NewFromConfig(cfg)
}

func (d *DeploymentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DeploymentDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	apiRes, err := d.client.GetDeployment(ctx, connect.NewRequest(&deploymentv1alpha1.GetDeploymentRequest{}))
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read Common Fate deployment metadata, got error: %s", err))
		return
	}

	data.Id = types.StringValue(apiRes.Msg.Deployment.Id)
	data.DNSZoneName = types.StringValue(apiRes.Msg.Deployment.DnsZoneName)
	data.DefaultSubdomain = types.StringValue(apiRes.Msg.Deployment.DefaultSubdomain)

	tflog.Trace(ctx, "read deployment metadata")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
