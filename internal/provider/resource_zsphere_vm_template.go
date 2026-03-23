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
	_ resource.Resource              = &vmTemplateResource{}
	_ resource.ResourceWithConfigure = &vmTemplateResource{}
)

type vmTemplateResource struct {
	client *client.ZSClient
}

type vmTemplateResourceModel struct {
	Uuid           types.String `tfsdk:"uuid"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	VmInstanceUuid types.String `tfsdk:"vm_instance_uuid"`
	ClusterUuid    types.String `tfsdk:"cluster_uuid"`
}

func (r *vmTemplateResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.ZSClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.ZSClient, got: %T. Please report this issue to the Provider developer. ", req.ProviderData),
		)
		return
	}

	r.client = client
}

func VmTemplateResource() resource.Resource {
	return &vmTemplateResource{}
}

func (r *vmTemplateResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vm_template"
}

func (r *vmTemplateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan vmTemplateResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.VmInstanceUuid.ValueString() != "" {
		tflog.Info(ctx, fmt.Sprintf("Creating VM template from VM instance: %s", plan.VmInstanceUuid.ValueString()))

		createParam := param.CreateTemplatedVmInstanceFromVmInstanceParam{
			Params: param.CreateTemplatedVmInstanceFromVmInstanceParamDetail{
				Name:        plan.Name.ValueString(),
				Description: plan.Description.ValueStringPointer(),
			},
		}

		result, err := r.client.CreateTemplatedVmInstanceFromVmInstance(ctx, plan.VmInstanceUuid.ValueString(), createParam)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to create VM template from VM instance",
				err.Error(),
			)
			return
		}

		queryParams := param.NewQueryParam()
		queryParams.AddQ("vmInstanceUuid=" + plan.VmInstanceUuid.ValueString())

		template, err := r.client.GetTemplatedVmInstance(ctx, result.TemplatedVmInstanceInventory.UUID)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to query created VM template",
				fmt.Sprintf("Template may have been created but could not be queried: %v", err),
			)
			return
		}

		plan.Uuid = types.StringValue(template.UUID)
		plan.Name = types.StringValue(template.Name)
		plan.ClusterUuid = types.StringValue(template.ZoneUuid)

		diags = resp.State.Set(ctx, plan)
		resp.Diagnostics.Append(diags...)
		return
	}

	if !plan.Uuid.IsNull() && plan.Uuid.ValueString() != "" {
		tflog.Info(ctx, fmt.Sprintf("Managing existing VM template with UUID: %s", plan.Uuid.ValueString()))

		template, err := r.client.GetTemplatedVmInstance(ctx, plan.Uuid.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to get existing VM template",
				err.Error(),
			)
			return
		}

		plan.Name = types.StringValue(template.Name)
		plan.ClusterUuid = types.StringValue(template.ZoneUuid)

		diags = resp.State.Set(ctx, plan)
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.AddError(
		"Missing required field",
		"Either 'vm_instance_uuid' (to create from VM instance) or 'uuid' (to manage existing template) must be specified.",
	)
}

func (r *vmTemplateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state vmTemplateResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	template, err := r.client.GetTemplatedVmInstance(ctx, state.Uuid.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to read VM template",
			err.Error(),
		)
		return
	}

	state.Uuid = types.StringValue(template.UUID)
	state.Name = types.StringValue(template.Name)
	state.ClusterUuid = types.StringValue(template.ZoneUuid)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *vmTemplateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state vmTemplateResourceModel

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
	plan.VmInstanceUuid = state.VmInstanceUuid
	plan.ClusterUuid = state.ClusterUuid

	updateParam := param.UpdateTemplatedVmInstanceParam{
		Params: param.UpdateTemplatedVmInstanceParamDetail{
			Name:        plan.Name.ValueString(),
			Description: plan.Description.ValueStringPointer(),
		},
	}

	tflog.Info(ctx, fmt.Sprintf("Updating VM template %s", plan.Uuid.ValueString()))
	_, err := r.client.UpdateTemplatedVmInstance(ctx, plan.Uuid.ValueString(), updateParam)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to update VM template",
			err.Error(),
		)
		return
	}

	template, err := r.client.GetTemplatedVmInstance(ctx, plan.Uuid.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to read updated VM template",
			err.Error(),
		)
		return
	}

	plan.Name = types.StringValue(template.Name)
	plan.ClusterUuid = types.StringValue(template.ZoneUuid)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *vmTemplateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state vmTemplateResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, fmt.Sprintf("Deleting VM template %s", state.Uuid.ValueString()))
	err := r.client.DeleteTemplatedVmInstance(ctx, state.Uuid.ValueString(), param.DeleteModeEnforcing)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to delete VM template",
			err.Error(),
		)
		return
	}
}

func (r *vmTemplateResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This resource allows you to create and manage VM templates in ZSphere. " +
			"A VM template is created by converting an existing VM instance.",
		Attributes: map[string]schema.Attribute{
			"uuid": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the VM template. Automatically generated by ZSphere.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the VM template.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The description of the VM template.",
			},
			"vm_instance_uuid": schema.StringAttribute{
				Optional:    true,
				Description: "The UUID of the VM instance to convert to a template. Changing this will recreate the resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"cluster_uuid": schema.StringAttribute{
				Computed:    true,
				Description: "The UUID of the cluster to which the VM template belongs.",
			},
		},
	}
}
