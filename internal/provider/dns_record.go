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
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &DNSRecordResource{}
var _ resource.ResourceWithImportState = &DNSRecordResource{}

func NewDNSRecordResource() resource.Resource {
	return &DNSRecordResource{}
}

// DNSRecordResource defines the resource implementation.
type DNSRecordResource struct {
	client deploymentv1alpha1connect.DeploymentServiceClient
}

// DNSRecordResourceModel describes the resource data model.
type DNSRecordResourceModel struct {
	ID       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Type     types.String `tfsdk:"type"`
	ZoneName types.String `tfsdk:"zone_name"`
	Values   types.Set    `tfsdk:"values"`
}

func (r *DNSRecordResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_record"
}

func (r *DNSRecordResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Registers DNS records for a Common Fate deployment.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The DNS record ID",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The DNS record name",
				Required:            true,
			},
			"zone_name": schema.StringAttribute{
				MarkdownDescription: "The DNS zone name",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The DNS record type. Must be one of ['TXT', 'CNAME']",
				Required:            true,
			},
			"values": schema.SetAttribute{
				MarkdownDescription: "The DNS record values",
				Required:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (r *DNSRecordResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DNSRecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DNSRecordResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var values []string

	resp.Diagnostics.Append(data.Values.ElementsAs(ctx, &values, false)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var rrType deploymentv1alpha1.DNSRecordType

	switch data.Type.ValueString() {
	case "TXT":
		rrType = deploymentv1alpha1.DNSRecordType_DNS_RECORD_TYPE_TXT
	case "CNAME":
		rrType = deploymentv1alpha1.DNSRecordType_DNS_RECORD_TYPE_CNAME
	default:
		resp.Diagnostics.AddError("Invalid DNS record type", fmt.Sprintf("the DNS record type '%s' is invalid. Valid values are ['TXT', 'CNAME']", data.Type.ValueString()))
		return
	}

	res, err := r.client.CreateDNSRecord(ctx, connect.NewRequest(&deploymentv1alpha1.CreateDNSRecordRequest{
		Name:        data.Name.ValueString(),
		DnsZoneName: data.ZoneName.ValueString(),
		Type:        rrType,
		Values:      values,
	}))
	if err != nil {
		resp.Diagnostics.AddError("Common Fate Deployment API error", fmt.Sprintf("Unable to create a DNS record for the deployment, got error: %s", err.Error()))
		return
	}

	tflog.Trace(ctx, "registered DNS record")

	// Convert from the API data model to the Terraform data model
	// and set any unknown attribute values.
	data.ID = types.StringValue(res.Msg.Created.Id)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSRecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DNSRecordResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	apiRes, err := r.client.GetDNSRecord(ctx, connect.NewRequest(&deploymentv1alpha1.GetDNSRecordRequest{
		Id: data.ID.ValueString(),
	}))
	if connect.CodeOf(err) == connect.CodeNotFound {
		resp.State.RemoveResource(ctx)
		return
	} else if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read Common Fate DNS record, got error: %s", err))
		return
	}

	data.ID = types.StringValue(apiRes.Msg.Record.Id)

	values, diags := types.SetValueFrom(ctx, types.StringType, apiRes.Msg.Record.Values)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.Values = values

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSRecordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DNSRecordResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var values []string

	resp.Diagnostics.Append(data.Values.ElementsAs(ctx, &values, false)...)

	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.UpdateDNSRecord(ctx, connect.NewRequest(&deploymentv1alpha1.UpdateDNSRecordRequest{
		Id:     data.ID.ValueString(),
		Values: values,
	}))
	if err != nil {
		resp.Diagnostics.AddError("Common Fate Deployment API error", fmt.Sprintf("Unable to update a DNS record for the deployment, got error: %s", err.Error()))
		return
	}

	tflog.Trace(ctx, "registered DNS record")

	// Convert from the API data model to the Terraform data model
	// and set any unknown attribute values.
	data.ID = types.StringValue(res.Msg.Updated.Id)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSRecordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DNSRecordResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var values []string

	resp.Diagnostics.Append(data.Values.ElementsAs(ctx, &values, false)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.DeleteDNSRecord(ctx, connect.NewRequest(&deploymentv1alpha1.DeleteDNSRecordRequest{
		Id: data.ID.ValueString(),
	}))
	if err != nil {
		resp.Diagnostics.AddError("Common Fate Deployment API error", fmt.Sprintf("Unable to delete DNS record for the deployment, got error: %s", err.Error()))
		return
	}

	tflog.Trace(ctx, "deleted DNS record")
}

func (r *DNSRecordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
