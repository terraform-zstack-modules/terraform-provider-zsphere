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
	"github.com/terraform-zstack-modules/zsphere-sdk-go/pkg/view"
)

var (
	_ resource.Resource              = &localPrimaryStorageResource{}
	_ resource.ResourceWithConfigure = &localPrimaryStorageResource{}
)

type localPrimaryStorageResource struct {
	client *client.ZSClient
}

type localPrimaryStorageResourceModel struct {
	Uuid                 types.String `tfsdk:"uuid"`
	Name                 types.String `tfsdk:"name"`
	Description          types.String `tfsdk:"description"`
	DatacenterUuid       types.String `tfsdk:"datacenter_uuid"`
	ClusterUuid          types.String `tfsdk:"cluster_uuid"`
	Url                  types.String `tfsdk:"url"`
	Type                 types.String `tfsdk:"type"`
	State                types.String `tfsdk:"state"`
	Status               types.String `tfsdk:"status"`
	MountPath            types.String `tfsdk:"mount_path"`
	TotalCapacity        types.Int64  `tfsdk:"total_capacity"`
	AvailableCapacity    types.Int64  `tfsdk:"available_capacity"`
	AttachedClusterUuids types.List   `tfsdk:"attached_cluster_uuids"`
}

func (r *localPrimaryStorageResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func LocalPrimaryStorageResource() resource.Resource {
	return &localPrimaryStorageResource{}
}

func (r *localPrimaryStorageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan localPrimaryStorageResourceModel
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

	localStorageParam := struct {
		Name        string  `json:"name"`
		Description *string `json:"description,omitempty"`
		ZoneUuid    string  `json:"zoneUuid"`
		ClusterUuid string  `json:"clusterUuid,omitempty"`
		Url         string  `json:"url"`
		Type        string  `json:"type,omitempty"`
	}{
		Name:        plan.Name.ValueString(),
		Description: description,
		ZoneUuid:    plan.DatacenterUuid.ValueString(),
		Url:         plan.Url.ValueString(),
	}

	if !plan.ClusterUuid.IsNull() && plan.ClusterUuid.ValueString() != "" {
		localStorageParam.ClusterUuid = plan.ClusterUuid.ValueString()
	}

	if !plan.Type.IsNull() && plan.Type.ValueString() != "" {
		localStorageParam.Type = plan.Type.ValueString()
	}

	var result view.PrimaryStorageInventoryView
	err := r.client.ZSHttpClient.Post(ctx, "v1/primary-storage/local-storage", map[string]interface{}{
		"params": localStorageParam,
	}, &result)
	if err != nil {
		resp.Diagnostics.AddError(
			"Could not create local primary storage in ZSphere", "Error: "+err.Error(),
		)
		return
	}

	plan.Uuid = types.StringValue(result.UUID)
	plan.Name = types.StringValue(result.Name)
	if result.Description != "" {
		plan.Description = types.StringValue(result.Description)
	} else {
		plan.Description = types.StringNull()
	}
	plan.DatacenterUuid = types.StringValue(result.ZoneUuid)
	plan.Url = types.StringValue(result.Url)
	plan.Type = types.StringValue(result.Type)
	plan.State = types.StringValue(result.State)
	plan.Status = types.StringValue(result.Status)
	plan.MountPath = types.StringValue(result.MountPath)
	plan.TotalCapacity = types.Int64Value(result.TotalCapacity)
	plan.AvailableCapacity = types.Int64Value(result.AvailableCapacity)

	attachedClusterUuids := make([]string, 0, len(result.AttachedClusterUuids))
	for _, uuid := range result.AttachedClusterUuids {
		attachedClusterUuids = append(attachedClusterUuids, uuid)
	}
	plan.AttachedClusterUuids, diags = types.ListValueFrom(ctx, types.StringType, attachedClusterUuids)
	resp.Diagnostics.Append(diags...)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *localPrimaryStorageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state localPrimaryStorageResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.Uuid.IsNull() || state.Uuid.ValueString() == "" {
		tflog.Warn(ctx, "local primary storage uuid is empty, so nothing to delete, skip it")
		return
	}

	uuid := state.Uuid.ValueString()

	tflog.Info(ctx, fmt.Sprintf("delete local primary storage %s", uuid))
	err := r.client.DeletePrimaryStorage(ctx, uuid, param.DeleteModeEnforcing)
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete local primary storage", err.Error())
		return
	}
}

func (r *localPrimaryStorageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_local_primary_storage"
}

func (r *localPrimaryStorageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state localPrimaryStorageResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	qparam := param.NewQueryParam()
	qparam.AddQ("uuid=" + state.Uuid.ValueString())
	primaryStorages, err := r.client.QueryPrimaryStorage(ctx, &qparam)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting ZStack local primary storage", "Could not read local primary storage: "+err.Error(),
		)
		return
	}

	if len(primaryStorages) == 0 {
		resp.Diagnostics.AddError(
			"Error getting ZStack local primary storage",
			"Could not find local primary storage: "+state.Uuid.ValueString(),
		)
		return
	}

	ps := primaryStorages[0]
	state.Uuid = types.StringValue(ps.UUID)
	state.Name = types.StringValue(ps.Name)
	if ps.Description != "" {
		state.Description = types.StringValue(ps.Description)
	} else {
		state.Description = types.StringNull()
	}
	state.DatacenterUuid = types.StringValue(ps.ZoneUuid)
	state.Url = types.StringValue(ps.Url)
	state.Type = types.StringValue(ps.Type)
	state.State = types.StringValue(ps.State)
	state.Status = types.StringValue(ps.Status)
	state.MountPath = types.StringValue(ps.MountPath)
	state.TotalCapacity = types.Int64Value(ps.TotalCapacity)
	state.AvailableCapacity = types.Int64Value(ps.AvailableCapacity)

	attachedClusterUuids := make([]string, 0, len(ps.AttachedClusterUuids))
	for _, uuid := range ps.AttachedClusterUuids {
		attachedClusterUuids = append(attachedClusterUuids, uuid)
	}
	state.AttachedClusterUuids, diags = types.ListValueFrom(ctx, types.StringType, attachedClusterUuids)
	resp.Diagnostics.Append(diags...)

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *localPrimaryStorageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan localPrimaryStorageResourceModel
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

	updateParam := param.UpdatePrimaryStorageParam{
		BaseParam: param.BaseParam{},
		Params: param.UpdatePrimaryStorageParamDetail{
			Name:        plan.Name.ValueString(),
			Description: description,
		},
	}

	_, err := r.client.UpdatePrimaryStorage(ctx, plan.Uuid.ValueString(), updateParam)
	if err != nil {
		resp.Diagnostics.AddError(
			"Could not update local primary storage in ZSphere", "Error: "+err.Error(),
		)
		return
	}

	qparam := param.NewQueryParam()
	qparam.AddQ("uuid=" + plan.Uuid.ValueString())
	primaryStorages, err := r.client.QueryPrimaryStorage(ctx, &qparam)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting ZStack local primary storage", "Could not read local primary storage: "+err.Error(),
		)
		return
	}

	if len(primaryStorages) == 0 {
		resp.Diagnostics.AddError(
			"Error getting ZStack local primary storage",
			"Could not find local primary storage: "+plan.Uuid.ValueString(),
		)
		return
	}

	ps := primaryStorages[0]
	plan.Name = types.StringValue(ps.Name)
	if ps.Description != "" {
		plan.Description = types.StringValue(ps.Description)
	} else {
		plan.Description = types.StringNull()
	}
	plan.DatacenterUuid = types.StringValue(ps.ZoneUuid)
	plan.Url = types.StringValue(ps.Url)
	plan.Type = types.StringValue(ps.Type)
	plan.State = types.StringValue(ps.State)
	plan.Status = types.StringValue(ps.Status)
	plan.MountPath = types.StringValue(ps.MountPath)
	plan.TotalCapacity = types.Int64Value(ps.TotalCapacity)
	plan.AvailableCapacity = types.Int64Value(ps.AvailableCapacity)

	attachedClusterUuids := make([]string, 0, len(ps.AttachedClusterUuids))
	for _, uuid := range ps.AttachedClusterUuids {
		attachedClusterUuids = append(attachedClusterUuids, uuid)
	}
	plan.AttachedClusterUuids, diags = types.ListValueFrom(ctx, types.StringType, attachedClusterUuids)
	resp.Diagnostics.Append(diags...)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *localPrimaryStorageResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Local Primary Storage Resource",
		Attributes: map[string]schema.Attribute{
			"uuid": schema.StringAttribute{
				MarkdownDescription: "The UUID of the local primary storage",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the local primary storage",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the local primary storage",
				Optional:            true,
				Computed:            true,
			},
			"datacenter_uuid": schema.StringAttribute{
				MarkdownDescription: "Datacenter UUID of the local primary storage",
				Required:            true,
			},
			"cluster_uuid": schema.StringAttribute{
				MarkdownDescription: "Cluster UUID to attach the local primary storage to",
				Optional:            true,
			},
			"url": schema.StringAttribute{
				MarkdownDescription: "Mount path of the local primary storage",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Type of the local primary storage",
				Computed:            true,
			},
			"state": schema.StringAttribute{
				MarkdownDescription: "State of the local primary storage",
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Status of the local primary storage",
				Computed:            true,
			},
			"mount_path": schema.StringAttribute{
				MarkdownDescription: "Mount path of the local primary storage",
				Computed:            true,
			},
			"total_capacity": schema.Int64Attribute{
				MarkdownDescription: "Total capacity of the local primary storage",
				Computed:            true,
			},
			"available_capacity": schema.Int64Attribute{
				MarkdownDescription: "Available capacity of the local primary storage",
				Computed:            true,
			},
			"attached_cluster_uuids": schema.ListAttribute{
				MarkdownDescription: "List of cluster UUIDs attached to the local primary storage",
				Computed:            true,
				ElementType:         types.StringType,
			},
		},
	}
}
