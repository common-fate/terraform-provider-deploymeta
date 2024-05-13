// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/common-fate/sdk/factory/service/monitoring"
	"github.com/common-fate/sdk/factoryconfig"
	monitoringv1alpha1 "github.com/common-fate/sdk/gen/commonfate/factory/monitoring/v1alpha1"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &MonitoringWriteTokenResource{}
var _ resource.ResourceWithImportState = &MonitoringWriteTokenResource{}

func NewMonitoringWriteTokenResource() resource.Resource {
	return &MonitoringWriteTokenResource{}
}

// MonitoringWriteTokenResource defines the resource implementation.
type MonitoringWriteTokenResource struct {
	client *monitoring.Client
}

// MonitoringWriteTokenResourceModel describes the resource data model.
type MonitoringWriteTokenResourceModel struct {
	ID    types.String `tfsdk:"id"`
	Token types.String `tfsdk:"token"`
}

func (r *MonitoringWriteTokenResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nameservers"
}

func (r *MonitoringWriteTokenResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A write token used to send events to Common Fate's centralised monitoring service.",

		Attributes: map[string]schema.Attribute{
			"token": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "The write token",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The token ID",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *MonitoringWriteTokenResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = monitoring.NewFromConfig(cfg)
}

func (r *MonitoringWriteTokenResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data MonitoringWriteTokenResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.Tokens().CreateWriteToken(ctx, connect.NewRequest(&monitoringv1alpha1.CreateWriteTokenRequest{}))
	if err != nil {
		resp.Diagnostics.AddError("Common Fate Deployment API error", fmt.Sprintf("Unable to create a monitoring write token for the deployment, got error: %s", err.Error()))
		return
	}

	// For the purposes of this example code, hardcoding a response value to
	// save into the Terraform state.
	data.Token = types.StringValue(res.Msg.WriteToken)
	data.ID = types.StringValue(res.Msg.Id)

	tflog.Trace(ctx, "created a monitoring token")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MonitoringWriteTokenResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MonitoringWriteTokenResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// // check the validity of the token
	// res, err := r.client.Tokens().GetWriteToken(ctx, connect.NewRequest(&monitoringv1alpha1.GetWriteTokenRequest{
	// 	Id: data.ID.ValueString(),
	// }))
	// if connect.CodeOf(err) == connect.CodeNotFound {
	// 	resp.State.RemoveResource(ctx)
	// 	return
	// }
	// if err != nil {
	// 	// don't block deployment on this, in case our API is unavailable.
	// 	resp.Diagnostics.AddWarning("Common Fate Deployment API error", fmt.Sprintf("Unable to validate the monitoring write token for the deployment, got error: %s", err.Error()))
	// 	return
	// }

	// if res.Msg.IsValid {
	// 	tflog.Debug(ctx, "token is no longer valid")
	// 	resp.State.RemoveResource(ctx)
	// 	return
	// }
}

func (r *MonitoringWriteTokenResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// resource is immutable once created, so this is a no-op.
}

func (r *MonitoringWriteTokenResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// var data MonitoringWriteTokenResourceModel

	// // Read Terraform prior state data into the model
	// resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	// if resp.Diagnostics.HasError() {
	// 	return
	// }

	// _, err := r.client.Tokens().GetWriteToken(ctx, connect.NewRequest(&monitoringv1alpha1.GetWriteTokenRequest{
	// 	Id: data.ID.ValueString(),
	// }))
	// if connect.CodeOf(err) == connect.CodeNotFound {
	// 	// not found, so it can be deleted.
	// 	return
	// }
	// if err != nil {
	// 	resp.Diagnostics.AddError("Common Fate Deployment API error", fmt.Sprintf("Unable to create a monitoring write token for the deployment, got error: %s", err.Error()))
	// 	return
	// }
}

func (r *MonitoringWriteTokenResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
