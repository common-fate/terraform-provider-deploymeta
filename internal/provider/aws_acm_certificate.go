package provider

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/common-fate/sdk/factory/service/deployment"
	"github.com/common-fate/sdk/factoryconfig"
	deploymentv1alpha1 "github.com/common-fate/sdk/gen/commonfate/factory/deployment/v1alpha1"
	"github.com/common-fate/sdk/gen/commonfate/factory/deployment/v1alpha1/deploymentv1alpha1connect"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &AWSACMCertificateResource{}
var _ resource.ResourceWithImportState = &AWSACMCertificateResource{}

func NewAWSACMCertificateResource() resource.Resource {
	return &AWSACMCertificateResource{}
}

// AWSACMCertificateResource defines the resource implementation.
type AWSACMCertificateResource struct {
	client deploymentv1alpha1connect.DeploymentServiceClient
}

type AWSACMCertificateResourceModel struct {
	ID                   types.String `tfsdk:"id"`
	ARN                  types.String `tfsdk:"arn"`
	DomainName           types.String `tfsdk:"domain_name"`
	ValidationCNameName  types.String `tfsdk:"validation_cname_name"`
	ValidationCNameValue types.String `tfsdk:"validation_cname_value"`
	Status               types.String `tfsdk:"status"`
}

func (r *AWSACMCertificateResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_aws_acm_certificate"
}

func (r *AWSACMCertificateResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Registers an AWS ACM certificate for a Common Fate deployment.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The certificate ID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"arn": schema.StringAttribute{
				MarkdownDescription: "The Amazon Resource Name (ARN) of the certificate",
				Required:            true,
			},
			"domain_name": schema.StringAttribute{
				MarkdownDescription: "The domain name for the certificate, for example: 'www.example.com'",
				Required:            true,
			},
			"validation_cname_name": schema.StringAttribute{
				MarkdownDescription: "The CNAME name used for domain validation",
				Required:            true,
			},
			"validation_cname_value": schema.StringAttribute{
				MarkdownDescription: "The CNAME value used for domain validation",
				Required:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The status of the certificate",
				Required:            true,
			},
		},
	}
}

func (r *AWSACMCertificateResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *AWSACMCertificateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data AWSACMCertificateResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.RegisterAWSACMCertificate(ctx, connect.NewRequest(&deploymentv1alpha1.RegisterAWSACMCertificateRequest{
		Arn:                  data.ARN.ValueString(),
		DomainName:           data.DomainName.ValueString(),
		ValidationCnameName:  data.ValidationCNameName.ValueString(),
		ValidationCnameValue: data.ValidationCNameValue.ValueString(),
		Status:               data.Status.ValueString(),
	}))
	if err != nil {
		resp.Diagnostics.AddError("Common Fate Deployment API error", fmt.Sprintf("Unable to register AWS ACM certificate for the deployment, got error: %s", err.Error()))
		return
	}

	tflog.Trace(ctx, "registered AWS ACM certificate")

	// Convert from the API data model to the Terraform data model
	// and set any unknown attribute values.
	data.ID = types.StringValue(res.Msg.Certificate.Id)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AWSACMCertificateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data AWSACMCertificateResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	apiRes, err := r.client.GetAWSACMCertificate(ctx, connect.NewRequest(&deploymentv1alpha1.GetAWSACMCertificateRequest{
		Id: data.ID.ValueString(),
	}))
	if connect.CodeOf(err) == connect.CodeNotFound {
		resp.State.RemoveResource(ctx)
		return
	} else if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read Common Fate DNS record, got error: %s", err))
		return
	}

	data.ID = types.StringValue(apiRes.Msg.Certificate.Id)
	data.ARN = types.StringValue(apiRes.Msg.Certificate.Arn)
	data.DomainName = types.StringValue(apiRes.Msg.Certificate.DomainName)
	data.ValidationCNameName = types.StringValue(apiRes.Msg.Certificate.ValidationCnameName)
	data.ValidationCNameValue = types.StringValue(apiRes.Msg.Certificate.ValidationCnameValue)
	data.Status = types.StringValue(apiRes.Msg.Certificate.Status)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
func (r *AWSACMCertificateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data AWSACMCertificateResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.UpdateAWSACMCertificate(ctx, connect.NewRequest(&deploymentv1alpha1.UpdateAWSACMCertificateRequest{
		Certificate: &deploymentv1alpha1.AWSACMCertificate{
			Id:                   data.ID.ValueString(),
			Arn:                  data.ARN.ValueString(),
			DomainName:           data.DomainName.ValueString(),
			ValidationCnameName:  data.ValidationCNameName.ValueString(),
			ValidationCnameValue: data.ValidationCNameValue.ValueString(),
			Status:               data.Status.ValueString(),
		},
	}))
	if err != nil {
		resp.Diagnostics.AddError("Common Fate Deployment API error", fmt.Sprintf("Unable to update AWS ACM certificate for the deployment, got error: %s", err.Error()))
		return
	}

	tflog.Trace(ctx, "updated AWS ACM certificate")

	// Convert from the API data model to the Terraform data model
	// and set any unknown attribute values.
	data.ID = types.StringValue(res.Msg.Certificate.Id)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AWSACMCertificateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data AWSACMCertificateResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	_, err := r.client.DeregisterAWSACMCertificate(ctx, connect.NewRequest(&deploymentv1alpha1.DeregisterAWSACMCertificateRequest{
		Id: data.ID.ValueString(),
	}))
	if connect.CodeOf(err) == connect.CodeNotFound {
		return
	}

	if err != nil {
		resp.Diagnostics.AddError("Common Fate Deployment API error", fmt.Sprintf("Unable to delete ACM certificate for the deployment, got error: %s", err.Error()))
		return
	}

	tflog.Trace(ctx, "deleted ACM cert")
}

func (r *AWSACMCertificateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
