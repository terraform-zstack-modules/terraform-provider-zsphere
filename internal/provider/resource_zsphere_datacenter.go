// Copyright (c) ZStack.io, Inc.

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/terraform-zstack-modules/zsphere-sdk-go/pkg/client"
	"github.com/terraform-zstack-modules/zsphere-sdk-go/pkg/param"
)

var (
	_ resource.Resource              = &datacenterResource{}
	_ resource.ResourceWithConfigure = &datacenterResource{}
)

type datacenterResource struct {
	client *client.ZSClient
}

type datacenterResourceModel struct {
	Uuid        types.String `tfsdk:"uuid"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	State       types.String `tfsdk:"state"`
	Type        types.String `tfsdk:"type"`
	IsDefault   types.Bool   `tfsdk:"is_default"`
}

func (r *datacenterResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.ZSClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.ZSClient, got: %T. Please report this issue to the Provider developer. ", req.ProviderData),
		)
		return
	}

	r.client = client
}

func DatacenterResource() resource.Resource {
	return &datacenterResource{}
}

func (r *datacenterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan datacenterResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var description *string
	if plan.Description.IsNull() {
		description = nil
	} else {
		descriptionStr := plan.Description.ValueString()
		description = &descriptionStr
	}

	var isDefault *bool
	if plan.IsDefault.IsNull() || plan.IsDefault.IsUnknown() {
		isDefault = nil
	} else {
		isDefaultVal := plan.IsDefault.ValueBool()
		isDefault = &isDefaultVal
	}

	zoneParam := param.CreateZoneParam{
		Params: param.CreateZoneParamDetail{
			Name:        plan.Name.ValueString(),
			Description: description,
			IsDefault:   isDefault,
		},
	}

	zone, err := r.client.CreateZone(ctx, zoneParam)
	if err != nil {
		resp.Diagnostics.AddError(
			"Could not create datacenter in ZSphere", "Error: "+err.Error(),
		)
		return
	}

	plan.Uuid = types.StringValue(zone.UUID)
	plan.Name = types.StringValue(zone.Name)
	plan.Description = types.StringValue(zone.Description)
	plan.State = types.StringValue(zone.State)
	plan.Type = types.StringValue(zone.Type)
	plan.IsDefault = types.BoolValue(zone.IsDefault)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *datacenterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state datacenterResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.Uuid.IsNull() || state.Uuid.ValueString() == "" {
		tflog.Warn(ctx, "datacenter uuid is empty, so nothing to delete, skip it")
		return
	}

	uuid := state.Uuid.ValueString()

	tflog.Info(ctx, fmt.Sprintf("delete datacenter %s", uuid))
	err := r.client.DeleteZone(ctx, uuid, param.DeleteModeEnforcing)
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete datacenter", err.Error())
		return
	}
}

func (r *datacenterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter"
}

func (r *datacenterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state datacenterResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone, err := r.client.GetZone(ctx, state.Uuid.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting ZStack datacenter", "Could not read datacenter: "+err.Error(),
		)
		return
	}

	state.Uuid = types.StringValue(zone.UUID)
	state.Name = types.StringValue(zone.Name)
	state.Description = types.StringValue(zone.Description)
	state.State = types.StringValue(zone.State)
	state.Type = types.StringValue(zone.Type)
	state.IsDefault = types.BoolValue(zone.IsDefault)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *datacenterResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This resource allows you to manage datacenters (zones) in ZSphere. " +
			"A datacenter (zone) is a logical container that groups compute resources like clusters and hosts. " +
			"It provides resource isolation and can be used to organize infrastructure for different environments or tenants.",
		Attributes: map[string]schema.Attribute{
			"uuid": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the datacenter. Automatically generated by ZSphere.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the datacenter. This is a mandatory field.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "A description of the datacenter, providing additional context or details.",
			},
			"state": schema.StringAttribute{
				Computed:    true,
				Description: "The state of the datacenter, such as 'Enabled' or 'Disabled'.",
			},
			"type": schema.StringAttribute{
				Computed:    true,
				Description: "The type of the datacenter. Currently only 'zstack' is supported.",
			},
			"is_default": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Indicates if this is the default datacenter. Only one datacenter can be the default.",
			},
		},
	}
}

func (r *datacenterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state datacenterResourceModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.Uuid = state.Uuid

	var description *string
	if plan.Description.IsNull() {
		description = nil
	} else if plan.Description.ValueString() == "" {
		emptyStr := ""
		description = &emptyStr
	} else {
		descriptionStr := plan.Description.ValueString()
		description = &descriptionStr
	}

	var isDefault *bool
	if plan.IsDefault.IsNull() || plan.IsDefault.IsUnknown() {
		isDefault = nil
	} else {
		isDefaultVal := plan.IsDefault.ValueBool()
		isDefault = &isDefaultVal
	}

	updateParam := param.UpdateZoneParam{
		Params: param.UpdateZoneParamDetail{
			Name:        plan.Name.ValueString(),
			Description: description,
			IsDefault:   isDefault,
		},
	}

	zone, err := r.client.UpdateZone(ctx, plan.Uuid.ValueString(), updateParam)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating ZStack datacenter", "Could not update datacenter: "+err.Error(),
		)
		return
	}

	plan.Name = types.StringValue(zone.Name)
	plan.Description = types.StringValue(zone.Description)
	plan.State = types.StringValue(zone.State)
	plan.Type = types.StringValue(zone.Type)
	plan.IsDefault = types.BoolValue(zone.IsDefault)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
