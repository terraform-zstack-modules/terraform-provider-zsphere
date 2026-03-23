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
	_ resource.Resource              = &tagResource{}
	_ resource.ResourceWithConfigure = &tagResource{}
)

type tagResource struct {
	client *client.ZSClient
}

type tagResourceModel struct {
	Uuid        types.String `tfsdk:"uuid"`
	Name        types.String `tfsdk:"name"`
	Value       types.String `tfsdk:"value"`
	Description types.String `tfsdk:"description"`
	Color       types.String `tfsdk:"color"`
	Type        types.String `tfsdk:"type"`
}

func (r *tagResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	cli, ok := req.ProviderData.(*client.ZSClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.ZSClient, got: %T. Please report this issue to the Provider developer. ", req.ProviderData),
		)
		return
	}

	r.client = cli
}

func TagResource() resource.Resource {
	return &tagResource{}
}

func (r *tagResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan tagResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var description *string
	if !plan.Description.IsNull() && plan.Description.ValueString() != "" {
		descStr := plan.Description.ValueString()
		description = &descStr
	}

	var color *string
	if !plan.Color.IsNull() && plan.Color.ValueString() != "" {
		colorStr := plan.Color.ValueString()
		color = &colorStr
	}

	tagParam := param.CreateTagParam{
		Params: param.CreateTagParamDetail{
			Name:        plan.Name.ValueString(),
			Value:       plan.Value.ValueString(),
			Description: description,
			Color:       color,
		},
	}

	result, err := r.client.CreateTag(ctx, tagParam)
	if err != nil {
		resp.Diagnostics.AddError(
			"Could not create tag in ZSphere", "Error: "+err.Error(),
		)
		return
	}

	plan.Uuid = types.StringValue(result.UUID)
	plan.Name = types.StringValue(result.Name)
	plan.Value = types.StringValue(result.Value)
	plan.Description = tfStringFromPtr(result.Description)
	plan.Color = tfStringFromPtr(result.Color)
	plan.Type = tfStringFromPtr(result.Type)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *tagResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state tagResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.Uuid.IsNull() || state.Uuid.ValueString() == "" {
		tflog.Warn(ctx, "tag uuid is empty, so nothing to delete, skip it")
		return
	}

	uuid := state.Uuid.ValueString()

	tflog.Info(ctx, fmt.Sprintf("delete tag %s", uuid))
	err := r.client.DeleteTag(ctx, uuid, param.DeleteModeEnforcing)
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete tag", err.Error())
		return
	}
}

func (r *tagResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tag"
}

func (r *tagResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state tagResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tag, err := r.client.GetTag(ctx, state.Uuid.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting ZStack tag", "Could not read tag: "+err.Error(),
		)
		return
	}

	state.Uuid = types.StringValue(tag.UUID)
	state.Name = types.StringValue(tag.Name)
	state.Value = types.StringValue(tag.Value)
	state.Description = tfStringFromPtr(tag.Description)
	state.Color = tfStringFromPtr(tag.Color)
	state.Type = tfStringFromPtr(tag.Type)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *tagResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This resource allows you to manage tags in ZSphere. " +
			"A tag is a label that can be attached to resources for organization and filtering. " +
			"Tags have a name, value, color, and optional description.",
		Attributes: map[string]schema.Attribute{
			"uuid": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the tag. Automatically generated by ZSphere.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the tag. This is a mandatory field.",
			},
			"value": schema.StringAttribute{
				Required:    true,
				Description: "The value of the tag. This is a mandatory field.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "A description of the tag, providing additional context or details.",
			},
			"color": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The color of the tag in hex format (e.g., #3185FC for blue).",
			},
			"type": schema.StringAttribute{
				Computed:    true,
				Description: "The type of the tag.",
			},
		},
	}
}

func (r *tagResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state tagResourceModel

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
		descStr := plan.Description.ValueString()
		description = &descStr
	}

	var color *string
	if plan.Color.IsNull() {
		color = nil
	} else if plan.Color.ValueString() == "" {
		emptyStr := ""
		color = &emptyStr
	} else {
		colorStr := plan.Color.ValueString()
		color = &colorStr
	}

	updateParam := param.UpdateTagParam{
		Params: param.UpdateTagParamDetail{
			Name:        plan.Name.ValueString(),
			Description: description,
			Color:       color,
		},
	}

	// currently, cannot update simple tag pattern if we change the tag value.
	// if !plan.Value.IsNull() && plan.Value.ValueString() != "" {
	// 	valueStr := plan.Value.ValueString()
	// 	updateParam.Params.Value = &valueStr
	// }

	tag, err := r.client.UpdateTag(ctx, plan.Uuid.ValueString(), updateParam)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating ZStack tag", "Could not update tag: "+err.Error(),
		)
		return
	}

	plan.Name = types.StringValue(tag.Name)
	plan.Value = types.StringValue(tag.Value)
	plan.Description = tfStringFromPtr(tag.Description)
	plan.Color = tfStringFromPtr(tag.Color)
	plan.Type = tfStringFromPtr(tag.Type)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
