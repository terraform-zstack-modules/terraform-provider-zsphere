// Copyright (c) ZStack.io, Inc.

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/terraform-zstack-modules/zsphere-sdk-go/pkg/client"
	"github.com/terraform-zstack-modules/zsphere-sdk-go/pkg/param"
	"github.com/terraform-zstack-modules/zsphere-sdk-go/pkg/view"
)

var (
	_ resource.Resource                = &imageStorageResource{}
	_ resource.ResourceWithConfigure   = &imageStorageResource{}
	_ resource.ResourceWithImportState = &imageStorageResource{}
)

type imageStorageResource struct {
	client *client.ZSClient
}

type imageStorageResourceModel struct {
	Uuid              types.String `tfsdk:"uuid"`
	Name              types.String `tfsdk:"name"`
	Description       types.String `tfsdk:"description"`
	Type              types.String `tfsdk:"type"`
	Hostname          types.String `tfsdk:"hostname"`
	Url               types.String `tfsdk:"url"`
	SshPort           types.Int64  `tfsdk:"ssh_port"`
	Username          types.String `tfsdk:"username"`
	Password          types.String `tfsdk:"password"`
	ZoneUuid          types.String `tfsdk:"zone_uuid"`
	PoolName          types.String `tfsdk:"pool_name"`
	MonUrls           types.List   `tfsdk:"mon_urls"`
	ImportImages      types.Bool   `tfsdk:"import_images"`
	State             types.String `tfsdk:"state"`
	Status            types.String `tfsdk:"status"`
	TotalCapacity     types.Int64  `tfsdk:"total_capacity"`
	AvailableCapacity types.Int64  `tfsdk:"available_capacity"`
	AttachedZoneUuids types.List   `tfsdk:"attached_zone_uuids"`
}

func (r *imageStorageResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *imageStorageResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_image_storage"
}

func (r *imageStorageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("uuid"), req.ID)...)
}

func ImageStorageResource() resource.Resource {
	return &imageStorageResource{}
}

func (r *imageStorageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan imageStorageResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	backupStorageType := plan.Type.ValueString()
	var result interface{}
	var err error

	if backupStorageType == "Ceph" {
		result, err = r.createCephBackupStorage(ctx, plan)
	} else {
		result, err = r.createImageStoreBackupStorage(ctx, plan)
	}

	if err != nil {
		resp.Diagnostics.AddError(
			"fail to create Image Storage",
			fmt.Sprintf("fail to create Image Storage, err: %v", err),
		)
		return
	}

	r.syncFromResult(ctx, &plan, result, backupStorageType)

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *imageStorageResource) createImageStoreBackupStorage(ctx context.Context, plan imageStorageResourceModel) (interface{}, error) {
	sshPort := int(plan.SshPort.ValueInt64())
	url := plan.Url.ValueString()
	name := plan.Name.ValueString()
	description := plan.Description.ValueString()
	hostname := plan.Hostname.ValueString()
	username := plan.Username.ValueString()
	importImages := plan.ImportImages.ValueBool()

	addImageStoreBackupStorageParam := param.AddImageStoreBackupStorageParam{
		BaseParam: param.BaseParam{},
		Params: param.AddImageStoreBackupStorageParamDetail{
			Hostname:     hostname,
			Username:     username,
			Name:         name,
			Url:          url,
			SshPort:      &sshPort,
			ImportImages: &importImages,
		},
	}

	if !plan.Description.IsNull() && description != "" {
		addImageStoreBackupStorageParam.Params.Description = plan.Description.ValueStringPointer()
	}
	if !plan.Password.IsNull() && plan.Password.ValueString() != "" {
		addImageStoreBackupStorageParam.Params.Password = plan.Password.ValueStringPointer()
	}
	if !plan.ZoneUuid.IsNull() && plan.ZoneUuid.ValueString() != "" {
		addImageStoreBackupStorageParam.Params.ResourceUuid = plan.ZoneUuid.ValueStringPointer()
	}

	tflog.Info(ctx, "Creating ImageStore Backup Storage")
	result, err := r.client.AddImageStoreBackupStorage(ctx, addImageStoreBackupStorageParam)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *imageStorageResource) createCephBackupStorage(ctx context.Context, plan imageStorageResourceModel) (interface{}, error) {
	var monUrls []string
	plan.MonUrls.ElementsAs(ctx, &monUrls, false)

	name := plan.Name.ValueString()
	description := plan.Description.ValueString()
	poolName := plan.PoolName.ValueStringPointer()

	addCephBackupStorageParam := param.AddCephBackupStorageParam{
		BaseParam: param.BaseParam{},
		Params: param.AddCephBackupStorageParamDetail{
			MonUrls:      monUrls,
			Name:         name,
			PoolName:     poolName,
			ImportImages: func() *bool { v := false; return &v }(),
		},
	}

	if !plan.Description.IsNull() && description != "" {
		addCephBackupStorageParam.Params.Description = plan.Description.ValueStringPointer()
	}
	if !plan.ZoneUuid.IsNull() && plan.ZoneUuid.ValueString() != "" {
		addCephBackupStorageParam.Params.ResourceUuid = plan.ZoneUuid.ValueStringPointer()
	}

	tflog.Info(ctx, "Creating Ceph Backup Storage")
	result, err := r.client.AddCephBackupStorage(ctx, addCephBackupStorageParam)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *imageStorageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state imageStorageResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	backupStorageType := state.Type.ValueString()
	var result interface{}
	var err error

	if backupStorageType == "Ceph" {
		result, err = r.client.GetCephBackupStorage(ctx, state.Uuid.ValueString())
	} else if backupStorageType == "ImageStore" || backupStorageType == "" {
		result, err = r.client.GetImageStoreBackupStorage(ctx, state.Uuid.ValueString())
	} else {
		result, err = r.client.GetBackupStorage(ctx, state.Uuid.ValueString())
	}

	if err != nil {
		resp.Diagnostics.AddError(
			"fail to read Image Storage",
			fmt.Sprintf("fail to read Image Storage, err: %v", err),
		)
		return
	}

	r.syncFromResult(ctx, &state, result, backupStorageType)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *imageStorageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan imageStorageResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	backupStorageType := plan.Type.ValueString()

	if backupStorageType == "ImageStore" {
		sshPort := int(plan.SshPort.ValueInt64())
		updateParam := param.UpdateImageStoreBackupStorageParam{
			BaseParam: param.BaseParam{},
			Params: param.UpdateImageStoreBackupStorageParamDetail{
				Name:     plan.Name.ValueString(),
				Hostname: plan.Hostname.ValueStringPointer(),
				Username: plan.Username.ValueStringPointer(),
				SshPort:  &sshPort,
			},
		}
		if !plan.Description.IsNull() {
			updateParam.Params.Description = plan.Description.ValueStringPointer()
		}
		if !plan.Password.IsNull() && plan.Password.ValueString() != "" {
			updateParam.Params.Password = plan.Password.ValueStringPointer()
		}

		result, err := r.client.UpdateImageStoreBackupStorage(ctx, plan.Uuid.ValueString(), updateParam)
		if err != nil {
			resp.Diagnostics.AddError(
				"fail to update Image Storage",
				fmt.Sprintf("fail to update Image Storage, err: %v", err),
			)
			return
		}
		r.syncFromResult(ctx, &plan, result, backupStorageType)
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *imageStorageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state imageStorageResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteMode := param.DeleteModeEnforcing
	err := r.client.DeleteBackupStorage(ctx, state.Uuid.ValueString(), deleteMode)
	if err != nil {
		resp.Diagnostics.AddError(
			"fail to delete Image Storage",
			fmt.Sprintf("fail to delete Image Storage, err: %v", err),
		)
		return
	}
}

func (r *imageStorageResource) syncFromResult(ctx context.Context, plan *imageStorageResourceModel, result interface{}, backupStorageType string) {
	switch v := result.(type) {
	case *view.ImageStoreBackupStorageInventoryView:
		plan.Uuid = types.StringValue(v.UUID)
		plan.Name = types.StringValue(v.Name)
		if v.Description != "" {
			plan.Description = types.StringValue(v.Description)
		}
		if backupStorageType != "" {
			plan.Type = types.StringValue(backupStorageType)
		} else {
			plan.Type = types.StringValue(v.Type)
		}
		plan.Hostname = types.StringValue(v.Hostname)
		plan.Url = types.StringValue(v.Url)
		plan.SshPort = types.Int64Value(int64(v.SshPort))
		plan.Username = types.StringValue(v.Username)
		plan.State = types.StringValue(v.State)
		plan.Status = types.StringValue(v.Status)
		plan.TotalCapacity = types.Int64Value(v.TotalCapacity)
		plan.AvailableCapacity = types.Int64Value(v.AvailableCapacity)
		attachedZoneUuids, _ := types.ListValueFrom(ctx, types.StringType, v.AttachedZoneUuids)
		plan.AttachedZoneUuids = attachedZoneUuids
	case *view.BackupStorageInventoryView:
		plan.Uuid = types.StringValue(v.UUID)
		plan.Name = types.StringValue(v.Name)
		if v.Description != "" {
			plan.Description = types.StringValue(v.Description)
		}
		if backupStorageType != "" {
			plan.Type = types.StringValue(backupStorageType)
		} else {
			plan.Type = types.StringValue(v.Type)
		}
		plan.State = types.StringValue(v.State)
		plan.Status = types.StringValue(v.Status)
		plan.TotalCapacity = types.Int64Value(v.TotalCapacity)
		plan.AvailableCapacity = types.Int64Value(v.AvailableCapacity)
		attachedZoneUuids, _ := types.ListValueFrom(ctx, types.StringType, v.AttachedZoneUuids)
		plan.AttachedZoneUuids = attachedZoneUuids
	}
}

func (r *imageStorageResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Image Storage resource for managing backup storages.",
		Attributes: map[string]schema.Attribute{
			"uuid": schema.StringAttribute{
				Description: "UUID of the Image Storage",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the Image Storage",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the Image Storage",
				Optional:    true,
			},
			"type": schema.StringAttribute{
				Description: "Type of the Image Storage (ImageStore or Ceph)",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("ImageStore", "Ceph"),
				},
			},
			"hostname": schema.StringAttribute{
				Description: "Hostname for ImageStore type",
				Optional:    true,
			},
			"url": schema.StringAttribute{
				Description: "URL for ImageStore type",
				Optional:    true,
			},
			"ssh_port": schema.Int64Attribute{
				Description: "SSH port for ImageStore type",
				Optional:    true,
				Computed:    true,
				Validators: []validator.Int64{
					int64validator.Between(1, 65535),
				},
			},
			"username": schema.StringAttribute{
				Description: "Username for ImageStore type",
				Optional:    true,
			},
			"password": schema.StringAttribute{
				Description: "Password for ImageStore type",
				Optional:    true,
				Sensitive:   true,
			},
			"zone_uuid": schema.StringAttribute{
				Description: "Zone UUID to attach the Image Storage",
				Optional:    true,
			},
			"pool_name": schema.StringAttribute{
				Description: "Pool name for Ceph type",
				Optional:    true,
			},
			"mon_urls": schema.ListAttribute{
				Description: "MON URLs for Ceph type",
				Optional:    true,
				ElementType: types.StringType,
			},
			"import_images": schema.BoolAttribute{
				Description: "Import images for ImageStore type",
				Optional:    true,
				Computed:    true,
			},
			"state": schema.StringAttribute{
				Description: "State of the Image Storage",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "Status of the Image Storage",
				Computed:    true,
			},
			"total_capacity": schema.Int64Attribute{
				Description: "Total capacity of the Image Storage",
				Computed:    true,
			},
			"available_capacity": schema.Int64Attribute{
				Description: "Available capacity of the Image Storage",
				Computed:    true,
			},
			"attached_zone_uuids": schema.ListAttribute{
				Description: "Attached zone UUIDs",
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}
