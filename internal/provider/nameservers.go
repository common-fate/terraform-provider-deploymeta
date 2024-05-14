// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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
var _ resource.Resource = &NameserversResource{}

// var _ resource.ResourceWithImportState = &NameserversResource{}

func NewNameserversResource() resource.Resource {
	return &NameserversResource{}
}

// NameserversResource defines the resource implementation.
type NameserversResource struct {
	client deploymentv1alpha1connect.DeploymentServiceClient
}

// NameserversResourceModel describes the resource data model.
type NameserversResourceModel struct {
	Records types.Set `tfsdk:"ns_records"`
}

func (r *NameserversResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nameservers"
}

func (r *NameserversResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Registers DNS nameservers for the Common Fate deployment default deployment domain.",

		Attributes: map[string]schema.Attribute{
			"ns_records": schema.SetAttribute{
				MarkdownDescription: "The NS records to associate with the deployment.",
				Required:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (r *NameserversResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *NameserversResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data NameserversResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var records []string

	resp.Diagnostics.Append(data.Records.ElementsAs(ctx, &records, false)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.RegisterNameservers(ctx, connect.NewRequest(&deploymentv1alpha1.RegisterNameserversRequest{
		Nameservers: records,
	}))
	if err != nil {
		resp.Diagnostics.AddError("Common Fate Deployment API error", fmt.Sprintf("Unable to create register a DNS namespace for the deployment, got error: %s", err.Error()))
		return
	}

	tflog.Trace(ctx, "registered deployment nameservers")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NameserversResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NameserversResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	apiRes, err := r.client.GetDeployment(ctx, connect.NewRequest(&deploymentv1alpha1.GetDeploymentRequest{}))
	if connect.CodeOf(err) == connect.CodeNotFound {
		resp.State.RemoveResource(ctx)
		return
	} else if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read Common Fate deployment metadata, got error: %s", err))
		return
	}

	records, diags := types.SetValueFrom(ctx, types.StringType, apiRes.Msg.Deployment.Nameservers)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.Records = records

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NameserversResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data NameserversResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var records []string

	resp.Diagnostics.Append(data.Records.ElementsAs(ctx, &records, false)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.RegisterNameservers(ctx, connect.NewRequest(&deploymentv1alpha1.RegisterNameserversRequest{
		Nameservers: records,
	}))
	if err != nil {
		resp.Diagnostics.AddError("Common Fate Deployment API error", fmt.Sprintf("Unable to create register a DNS namespace for the deployment, got error: %s", err.Error()))
		return
	}

	tflog.Trace(ctx, "registered deployment nameservers")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NameserversResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// deleting is currently a no-op, the nameservers need to be deregistered by the Common Fate team.
}

// func (r *NameserversResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
// 	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
// }
