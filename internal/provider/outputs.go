package provider

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/common-fate/sdk/factory/service/deployment"
	"github.com/common-fate/sdk/factoryconfig"
	deploymentv1alpha1 "github.com/common-fate/sdk/gen/commonfate/factory/deployment/v1alpha1"
	"github.com/common-fate/sdk/gen/commonfate/factory/deployment/v1alpha1/deploymentv1alpha1connect"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &TerraformOutputResource{}

func NewTerraformOutputResource() resource.Resource {
	return &TerraformOutputResource{}
}

// TerraformOutputResource defines the resource implementation.
type TerraformOutputResource struct {
	client deploymentv1alpha1connect.DeploymentServiceClient
}

// TerraformOutputResourceModel describes the resource data model.
type TerraformOutputResourceModel struct {
	SAMLSSOACSURL               types.String `tfsdk:"saml_sso_acs_url"`
	SAMLSSOEntityID             types.String `tfsdk:"saml_sso_entity_id"`
	CognitoUserPoolID           types.String `tfsdk:"cognito_user_pool_id"`
	DNSCNAMERecordForAppDomain  types.String `tfsdk:"dns_cname_record_for_app_domain"`
	DNSCNAMERecordForAuthDomain types.String `tfsdk:"dns_cname_record_for_auth_domain"`
	WebClientID                 types.String `tfsdk:"web_client_id"`
	CLIClientID                 types.String `tfsdk:"cli_client_id"`
	TerraformClientID           types.String `tfsdk:"terraform_client_id"`
	ReadOnlyClientID            types.String `tfsdk:"read_only_client_id"`
	ProvisionerClientID         types.String `tfsdk:"provisioner_client_id"`
	VPCID                       types.String `tfsdk:"vpc_id"`
}

func (r *TerraformOutputResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_terraform_output"
}

func (r *TerraformOutputResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Registers Terraform outputs for a Common Fate deployment.",

		Attributes: map[string]schema.Attribute{
			"saml_sso_acs_url": schema.StringAttribute{
				MarkdownDescription: "The SAML SSO ACS URL",
				Required:            true,
			},
			"saml_sso_entity_id": schema.StringAttribute{
				MarkdownDescription: "The SAML SSO Entity ID",
				Required:            true,
			},
			"cognito_user_pool_id": schema.StringAttribute{
				MarkdownDescription: "The Cognito user pool ID",
				Required:            true,
			},
			"dns_cname_record_for_app_domain": schema.StringAttribute{
				MarkdownDescription: "The DNS CNAME record for the app domain",
				Required:            true,
			},
			"dns_cname_record_for_auth_domain": schema.StringAttribute{
				MarkdownDescription: "The DNS CNAME record for the auth domain",
				Required:            true,
			},
			"web_client_id": schema.StringAttribute{
				MarkdownDescription: "The web console client ID",
				Required:            true,
			},
			"cli_client_id": schema.StringAttribute{
				MarkdownDescription: "The CLI client ID",
				Required:            true,
			},
			"terraform_client_id": schema.StringAttribute{
				MarkdownDescription: "The Terraform client ID",
				Required:            true,
			},
			"read_only_client_id": schema.StringAttribute{
				MarkdownDescription: "The Read-Only client ID",
				Required:            true,
			},
			"provisioner_client_id": schema.StringAttribute{
				MarkdownDescription: "The Provisioner client ID",
				Required:            true,
			},
			"vpc_id": schema.StringAttribute{
				MarkdownDescription: "The VPC ID",
				Required:            true,
			},
		},
	}
}

func (r *TerraformOutputResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	cfg, ok := req.ProviderData.(*factoryconfig.Context)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *factoryconfig.Context, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = deployment.NewFromConfig(cfg)
}

func (r *TerraformOutputResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data TerraformOutputResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.SetTerraformOutput(ctx, connect.NewRequest(&deploymentv1alpha1.SetTerraformOutputRequest{
		Output: &deploymentv1alpha1.TerraformOutput{
			SamlSsoAcsUrl:               data.SAMLSSOACSURL.ValueString(),
			SamlSsoEntityId:             data.SAMLSSOEntityID.ValueString(),
			CognitoUserPoolId:           data.CognitoUserPoolID.ValueString(),
			DnsCnameRecordForAppDomain:  data.DNSCNAMERecordForAppDomain.ValueString(),
			DnsCnameRecordForAuthDomain: data.DNSCNAMERecordForAuthDomain.ValueString(),
			WebClientId:                 data.WebClientID.ValueString(),
			CliClientId:                 data.CLIClientID.ValueString(),
			TerraformClientId:           data.TerraformClientID.ValueString(),
			ReadOnlyClientId:            data.ReadOnlyClientID.ValueString(),
			ProvisionerClientId:         data.ProvisionerClientID.ValueString(),
			VpcId:                       data.VPCID.ValueString(),
		},
	}))
	if err != nil {
		resp.Diagnostics.AddError("Common Fate Deployment API error", fmt.Sprintf("Unable to set Terraform outputs for the deployment, got error: %s", err.Error()))
		return
	}

	tflog.Trace(ctx, "set Terraform outputs")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TerraformOutputResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data TerraformOutputResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	apiRes, err := r.client.GetTerraformOutput(ctx, connect.NewRequest(&deploymentv1alpha1.GetTerraformOutputRequest{}))
	if connect.CodeOf(err) == connect.CodeNotFound {
		resp.State.RemoveResource(ctx)
		return
	} else if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read Common Fate DNS record, got error: %s", err))
		return
	}

	data.SAMLSSOACSURL = types.StringValue(apiRes.Msg.Output.SamlSsoAcsUrl)
	data.SAMLSSOEntityID = types.StringValue(apiRes.Msg.Output.SamlSsoEntityId)
	data.CognitoUserPoolID = types.StringValue(apiRes.Msg.Output.CognitoUserPoolId)
	data.DNSCNAMERecordForAppDomain = types.StringValue(apiRes.Msg.Output.DnsCnameRecordForAppDomain)
	data.DNSCNAMERecordForAuthDomain = types.StringValue(apiRes.Msg.Output.DnsCnameRecordForAuthDomain)
	data.WebClientID = types.StringValue(apiRes.Msg.Output.WebClientId)
	data.CLIClientID = types.StringValue(apiRes.Msg.Output.CliClientId)
	data.TerraformClientID = types.StringValue(apiRes.Msg.Output.TerraformClientId)
	data.ReadOnlyClientID = types.StringValue(apiRes.Msg.Output.ReadOnlyClientId)
	data.ProvisionerClientID = types.StringValue(apiRes.Msg.Output.ProvisionerClientId)
	data.VPCID = types.StringValue(apiRes.Msg.Output.VpcId)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TerraformOutputResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data TerraformOutputResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.SetTerraformOutput(ctx, connect.NewRequest(&deploymentv1alpha1.SetTerraformOutputRequest{
		Output: &deploymentv1alpha1.TerraformOutput{
			SamlSsoAcsUrl:               data.SAMLSSOACSURL.ValueString(),
			SamlSsoEntityId:             data.SAMLSSOEntityID.ValueString(),
			CognitoUserPoolId:           data.CognitoUserPoolID.ValueString(),
			DnsCnameRecordForAppDomain:  data.DNSCNAMERecordForAppDomain.ValueString(),
			DnsCnameRecordForAuthDomain: data.DNSCNAMERecordForAuthDomain.ValueString(),
			WebClientId:                 data.WebClientID.ValueString(),
			CliClientId:                 data.CLIClientID.ValueString(),
			TerraformClientId:           data.TerraformClientID.ValueString(),
			ReadOnlyClientId:            data.ReadOnlyClientID.ValueString(),
			ProvisionerClientId:         data.ProvisionerClientID.ValueString(),
			VpcId:                       data.VPCID.ValueString(),
		},
	}))
	if err != nil {
		resp.Diagnostics.AddError("Common Fate Deployment API error", fmt.Sprintf("Unable to set Terraform outputs for the deployment, got error: %s", err.Error()))
		return
	}

	tflog.Trace(ctx, "set Terraform outputs")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TerraformOutputResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// no-op at the moment.
}
