// Copyright (c) ZStack.io, Inc.

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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
)

var (
	_ resource.Resource                = &imageResource{}
	_ resource.ResourceWithConfigure   = &imageResource{}
	_ resource.ResourceWithImportState = &imageResource{}
)

type imageResource struct {
	client *client.ZSClient
}

type imageBackupStorageRefModel struct {
	BackupStorageUuid types.String `tfsdk:"backup_storage_uuid"`
	InstallPath       types.String `tfsdk:"install_path"`
}

type imageResourceModel struct {
	Uuid               types.String `tfsdk:"uuid"`
	Name               types.String `tfsdk:"name"`
	Description        types.String `tfsdk:"description"`
	Url                types.String `tfsdk:"url"`
	MediaType          types.String `tfsdk:"media_type"`
	GuestOsType        types.String `tfsdk:"guest_os_type"`
	System             types.String `tfsdk:"system"`
	Platform           types.String `tfsdk:"platform"`
	Format             types.String `tfsdk:"format"`
	BackupStorageUuids types.List   `tfsdk:"image_storage_uuids"`
	Architecture       types.String `tfsdk:"architecture"`
	Virtio             types.Bool   `tfsdk:"virtio"`
	BootMode           types.String `tfsdk:"boot_mode"`
	Expunge            types.Bool   `tfsdk:"expunge"`
	State              types.String `tfsdk:"state"`
	Status             types.String `tfsdk:"status"`
	Size               types.Int64  `tfsdk:"size"`
	ActualSize         types.Int64  `tfsdk:"actual_size"`
	Md5Sum             types.String `tfsdk:"md5_sum"`
	Type               types.String `tfsdk:"type"`
	BackupStorageRefs  types.List   `tfsdk:"backup_storage_refs"`
	Enable             types.Bool   `tfsdk:"enable"`
}

func (r *imageResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *imageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("uuid"), req.ID)...)
}

func ImageResource() resource.Resource {
	return &imageResource{}
}

func (r *imageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var imagePlan imageResourceModel
	diags := req.Plan.Get(ctx, &imagePlan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var backupStorageUuids []string
	if imagePlan.BackupStorageUuids.IsNull() {
		storage, err := r.client.QueryBackupStorage(ctx, &param.QueryParam{})
		if err != nil {
			resp.Diagnostics.AddError(
				"fail to get Image storage",
				fmt.Sprintf("fail to get backup storage, err: %v", err),
			)
			return
		}
		backupStorageUuids = []string{storage[0].UUID}
	} else {
		imagePlan.BackupStorageUuids.ElementsAs(ctx, &backupStorageUuids, false)
	}

	var systemTags []string

	if imagePlan.BootMode.IsNull() || imagePlan.BootMode.ValueString() == "" {
		if imagePlan.Architecture.ValueString() == "aarch64" {
			systemTags = append(systemTags, "bootMode::UEFI")
		} else {
			systemTags = append(systemTags, "bootMode::Legacy")
		}
	} else {
		bootMode := imagePlan.BootMode.ValueString()

		switch bootMode {
		case "UEFI":
			systemTags = append(systemTags, "bootMode::UEFI")
		case "Legacy":
			systemTags = append(systemTags, "bootMode::Legacy")
		case "UEFI_WITH_CSM":
			systemTags = append(systemTags, "bootMode::UEFI_WITH_CSM")
		default:
			resp.Diagnostics.AddError(
				"invalid boot mode",
				fmt.Sprintf("invalid boot mode: %s", bootMode),
			)
			return
		}
	}

	if imagePlan.Description.IsNull() {
		imagePlan.Description = types.StringValue("")
	}
	if imagePlan.GuestOsType.IsNull() {
		imagePlan.GuestOsType = types.StringValue("Linux")
	}
	if imagePlan.Platform.IsNull() {
		imagePlan.Platform = types.StringValue("Linux")
	}

	tflog.Info(ctx, "Configuring ZStack client")
	imageParam := param.AddImageParam{
		BaseParam: param.BaseParam{
			SystemTags: systemTags,
		},
		Params: param.AddImageParamDetail{
			Name:               imagePlan.Name.ValueString(),
			Description:        imagePlan.Description.ValueStringPointer(),
			Url:                imagePlan.Url.ValueString(),
			MediaType:          imagePlan.MediaType.ValueStringPointer(),
			GuestOsType:        imagePlan.GuestOsType.ValueStringPointer(),
			System:             false,
			Format:             imagePlan.Format.ValueStringPointer(),
			Platform:           imagePlan.Platform.ValueStringPointer(),
			BackupStorageUuids: backupStorageUuids,
			ResourceUuid:       nil,
			Architecture:       imagePlan.Architecture.ValueStringPointer(),
			Virtio:             imagePlan.Virtio.ValueBoolPointer(),
		},
	}

	ctx = tflog.SetField(ctx, "url", imagePlan.Url)
	image, err := r.client.AddImage(ctx, imageParam)
	if err != nil {
		resp.Diagnostics.AddError(
			"Could not Add image to ZSphere Image storage", "Error "+err.Error(),
		)
		return
	}

	imagePlan.Uuid = types.StringValue(image.UUID)
	imagePlan.Name = types.StringValue(image.Name)
	imagePlan.Description = types.StringValue(image.Description)
	imagePlan.Url = types.StringValue(image.Url)
	imagePlan.MediaType = types.StringValue(image.MediaType)
	imagePlan.GuestOsType = types.StringValue(image.GuestOsType)
	imagePlan.System = types.StringValue(fmt.Sprintf("%v", image.System))
	imagePlan.Platform = types.StringValue(image.Platform)
	imagePlan.Format = types.StringValue(image.Format)
	imagePlan.Architecture = types.StringValue(string(image.Architecture))
	imagePlan.Virtio = types.BoolValue(image.Virtio)
	imagePlan.State = types.StringValue(image.State)
	imagePlan.Status = types.StringValue(image.Status)
	imagePlan.Size = types.Int64Value(image.Size)
	imagePlan.ActualSize = types.Int64Value(image.ActualSize)
	imagePlan.Md5Sum = types.StringValue(image.Md5Sum)
	imagePlan.Type = types.StringValue(image.Type)

	var backupStorageRefs []imageBackupStorageRefModel
	for _, ref := range image.BackupStorageRefs {
		backupStorageRefs = append(backupStorageRefs, imageBackupStorageRefModel{
			BackupStorageUuid: types.StringValue(ref.BackupStorageUuid),
			InstallPath:       types.StringValue(ref.InstallPath),
		})
	}
	backupStorageRefsList, diags := types.ListValueFrom(ctx, types.ObjectType{}.WithAttributeTypes(map[string]attr.Type{
		"backup_storage_uuid": types.StringType,
		"install_path":        types.StringType,
	}), backupStorageRefs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	imagePlan.BackupStorageRefs = backupStorageRefsList

	enableImage := false
	if imagePlan.Enable.IsNull() || imagePlan.Enable.IsUnknown() {
		enableImage = true
	} else {
		enableImage = imagePlan.Enable.ValueBool()
	}

	if enableImage && image.State == "Disabled" {
		stateChangeParam := param.ChangeImageStateParam{
			Params: param.ChangeImageStateParamDetail{
				StateEvent: "enable",
			},
		}
		result, err := r.client.ChangeImageState(ctx, imagePlan.Uuid.ValueString(), stateChangeParam)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error enabling image", "Could not enable image: "+err.Error(),
			)
			return
		}
		image.State = result.State
		imagePlan.State = types.StringValue(result.State)
	}

	imagePlan.Enable = types.BoolValue(image.State == "Enabled")

	ctx = tflog.SetField(ctx, "url", image.Url)
	diags = resp.State.Set(ctx, imagePlan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *imageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state imageResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	expunge := false
	if !state.Expunge.IsNull() && !state.Expunge.IsUnknown() {
		expunge = state.Expunge.ValueBool()
	}

	if state.Uuid == types.StringValue("") {
		tflog.Warn(ctx, "image uuid is empty, so nothing to delete, skip it")
		return
	}

	uuid := state.Uuid.ValueString()

	tflog.Info(ctx, fmt.Sprintf("delete image %s", uuid))
	err := r.client.DeleteImage(ctx, uuid, param.DeleteModeEnforcing)
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete image", err.Error())
		return
	}

	if expunge {
		var backupStorageUuids []string
		state.BackupStorageUuids.ElementsAs(ctx, &backupStorageUuids, false)
		expungeParam := param.ExpungeImageParam{
			Params: param.ExpungeImageParamDetail{
				Uuid:               uuid,
				BackupStorageUuids: backupStorageUuids,
			},
		}
		state.BackupStorageUuids.ElementsAs(ctx, &expungeParam.Params.BackupStorageUuids, false)
		tflog.Info(ctx, fmt.Sprintf("expunge image %s", uuid))
		err := r.client.ExpungeImage(ctx, uuid, expungeParam)
		if err != nil {
			resp.Diagnostics.AddError("Failed to expunge image", err.Error())
			return
		}
	}
}

func (r *imageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_image"
}

func (r *imageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state imageResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	image, err := r.client.GetImage(ctx, state.Uuid.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting ZStack Image uuid", "Could not read image uuid"+err.Error(),
		)
		return
	}

	state.Uuid = types.StringValue(image.UUID)
	state.Name = types.StringValue(image.Name)
	state.Url = types.StringValue(image.Url)
	state.Description = types.StringValue(image.Description)
	state.MediaType = types.StringValue(image.MediaType)
	state.GuestOsType = types.StringValue(image.GuestOsType)
	state.System = types.StringValue(fmt.Sprintf("%v", image.System))
	state.Platform = types.StringValue(image.Platform)
	state.Format = types.StringValue(image.Format)
	state.Architecture = types.StringValue(string(image.Architecture))
	state.Virtio = types.BoolValue(image.Virtio)
	state.State = types.StringValue(image.State)
	state.Status = types.StringValue(image.Status)
	state.Size = types.Int64Value(image.Size)
	state.ActualSize = types.Int64Value(image.ActualSize)
	state.Md5Sum = types.StringValue(image.Md5Sum)
	state.Type = types.StringValue(image.Type)

	var backupStorageRefs []imageBackupStorageRefModel
	for _, ref := range image.BackupStorageRefs {
		backupStorageRefs = append(backupStorageRefs, imageBackupStorageRefModel{
			BackupStorageUuid: types.StringValue(ref.BackupStorageUuid),
			InstallPath:       types.StringValue(ref.InstallPath),
		})
	}
	backupStorageRefsList, diags := types.ListValueFrom(ctx, types.ObjectType{}.WithAttributeTypes(map[string]attr.Type{
		"backup_storage_uuid": types.StringType,
		"install_path":        types.StringType,
	}), backupStorageRefs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.BackupStorageRefs = backupStorageRefsList

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *imageResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This resource allows you to manage images in ZSphere. " +
			"An image represents a virtual machine image format qcow2, raw, vmdk or an ISO file that can be used to create or boot virtual machines. " +
			"You can define the image's properties, such as its URL, format, architecture, and backup storage locations.",
		Attributes: map[string]schema.Attribute{
			"uuid": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the image. Automatically generated by ZSphere.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the image. This is a mandatory field.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "A description of the image, providing additional context or details.",
			},
			"url": schema.StringAttribute{
				Required:    true,
				Description: "The URL where the image is located. This can be a file path or an HTTP link. Changing this will force recreation of the resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"media_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The type of media for the image. Examples include 'ISO' or 'RootVolumeTemplate' or DataVolumeTemplate.",
				Validators: []validator.String{
					stringvalidator.OneOf("ISO", "RootVolumeTemplate", "DataVolumeTemplate"),
				},
			},
			"guest_os_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The guest operating system type that the image is optimized for.",
			},
			"system": schema.StringAttribute{
				Computed:    true,
				Description: "Indicates if the image is a system image. Set automatically by ZStack.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"platform": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The platform that the image is intended for, such as 'Linux', 'Windows', or others.",
				Validators: []validator.String{
					stringvalidator.OneOf("Linux", "Windows", "Other", "Paravirtualization", "WindowsVirtio"),
				},
			},
			"format": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The format of the image file, such as 'qcow2', 'raw', or 'vmdk'.",
				Validators: []validator.String{
					stringvalidator.OneOf("qcow2", "iso", "raw", "vmdk"),
				},
			},
			"image_storage_uuids": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    true,
				Description: "A list of UUIDs for the image storages where the image is stored.",
			},
			"architecture": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The architecture of the image, such as 'x86_64' or 'aarch64'.",
				Validators: []validator.String{
					stringvalidator.OneOf("x86_64", "aarch64", "mips64el", "loongarch64"),
				},
			},
			"virtio": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Indicates if the VirtIO drivers are required for the image.",
			},
			"expunge": schema.BoolAttribute{
				Optional:    true,
				Description: "Indicates if the image should be expunged (permanently deleted) after deletion. If true, the image will be expunged instead of just deleted.",
			},
			"boot_mode": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The boot mode supported by the image, such as 'Legacy', 'UEFI', or 'UEFI_WITH_CSM'.",
				Validators: []validator.String{
					stringvalidator.OneOf("Legacy", "UEFI", "UEFI_WITH_CSM"),
				},
			},
			"state": schema.StringAttribute{
				Computed:    true,
				Description: "The state of the image, such as 'Enabled' or 'Disabled'.",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "The status of the image, such as 'Ready' or 'NotReady'.",
			},
			"size": schema.Int64Attribute{
				Computed:    true,
				Description: "The size of the image in bytes.",
			},
			"actual_size": schema.Int64Attribute{
				Computed:    true,
				Description: "The actual size of the image in bytes after compression.",
			},
			"md5_sum": schema.StringAttribute{
				Computed:    true,
				Description: "The MD5 checksum of the image.",
			},
			"type": schema.StringAttribute{
				Computed:    true,
				Description: "The type of the image, such as 'zstack' or 'iso'.",
			},
			"backup_storage_refs": schema.ListNestedAttribute{
				Computed:    true,
				Description: "References to the backup storages where this image is stored.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"backup_storage_uuid": schema.StringAttribute{
							Computed:    true,
							Description: "The UUID of the backup storage.",
						},
						"install_path": schema.StringAttribute{
							Computed:    true,
							Description: "The install path of the image on the backup storage.",
						},
					},
				},
			},
			"enable": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Enable or disable the image. When set to true, the image will be enabled.",
			},
		},
	}
}

func (r *imageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state imageResourceModel

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
	plan.Url = state.Url
	plan.BackupStorageUuids = state.BackupStorageUuids
	plan.System = state.System
	plan.State = state.State
	plan.Status = state.Status
	plan.Size = state.Size
	plan.ActualSize = state.ActualSize
	plan.Md5Sum = state.Md5Sum
	plan.Type = state.Type
	plan.BackupStorageRefs = state.BackupStorageRefs

	if !state.Enable.IsNull() && !plan.Enable.IsUnknown() {
		desiredState := plan.Enable.ValueBool()
		currentState := state.State.ValueString()
		if (desiredState && currentState != "Enabled") || (!desiredState && currentState != "Disabled") {
			stateEvent := "enable"
			if !desiredState {
				stateEvent = "disable"
			}
			stateChangeParam := param.ChangeImageStateParam{
				Params: param.ChangeImageStateParamDetail{
					StateEvent: stateEvent,
				},
			}
			_, err := r.client.ChangeImageState(ctx, plan.Uuid.ValueString(), stateChangeParam)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error changing image state", "Could not change image state: "+err.Error(),
				)
				return
			}
		}
	}

	updateParam := param.UpdateImageParam{
		Params: param.UpdateImageParamDetail{
			Name:         plan.Name.ValueString(),
			Description:  plan.Description.ValueStringPointer(),
			GuestOsType:  plan.GuestOsType.ValueStringPointer(),
			MediaType:    plan.MediaType.ValueStringPointer(),
			Format:       plan.Format.ValueStringPointer(),
			Platform:     plan.Platform.ValueStringPointer(),
			Architecture: plan.Architecture.ValueStringPointer(),
			Virtio:       plan.Virtio.ValueBoolPointer(),
		},
	}

	image, err := r.client.UpdateImage(ctx, plan.Uuid.ValueString(), updateParam)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating ZStack Image", "Could not update image: "+err.Error(),
		)
		return
	}

	plan.Name = types.StringValue(image.Name)
	plan.Description = types.StringValue(image.Description)
	plan.MediaType = types.StringValue(image.MediaType)
	plan.GuestOsType = types.StringValue(image.GuestOsType)
	plan.Platform = types.StringValue(image.Platform)
	plan.Format = types.StringValue(image.Format)
	plan.Architecture = types.StringValue(string(image.Architecture))
	plan.Virtio = types.BoolValue(image.Virtio)
	plan.State = types.StringValue(image.State)
	plan.Status = types.StringValue(image.Status)
	plan.Size = types.Int64Value(image.Size)
	plan.ActualSize = types.Int64Value(image.ActualSize)
	plan.Md5Sum = types.StringValue(image.Md5Sum)
	plan.Type = types.StringValue(image.Type)

	if plan.Enable.IsNull() || plan.Enable.IsUnknown() {
		plan.Enable = types.BoolValue(image.State == "Enabled")
	} else {
		plan.Enable = types.BoolValue(image.State == "Enabled")
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
