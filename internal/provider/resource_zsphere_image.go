// Copyright (c) ZStack.io, Inc.

package provider

import (
	"context"
	"fmt"

	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/terraform-zstack-modules/zsphere-sdk-go/pkg/client"
	"github.com/terraform-zstack-modules/zsphere-sdk-go/pkg/param"
)

var (
	_ resource.Resource              = &imageResource{}
	_ resource.ResourceWithConfigure = &imageResource{}
)

type imageResource struct {
	client *client.ZSClient
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
}

// Configure implements resource.ResourceWithConfigure.
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

func ImageResource() resource.Resource {
	return &imageResource{}
}

// Create implements resource.Resource.
func (r *imageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var imagePlan imageResourceModel
	diags := req.Plan.Get(ctx, &imagePlan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var backupStorageUuids []string
	if imagePlan.BackupStorageUuids.IsNull() {
		storage, err := r.client.QueryBackupStorage(param.QueryParam{})
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
		// if boot mode not set, use uefi in aarch64 and legacy in x86_64
		if imagePlan.Architecture.ValueString() == "aarch64" {
			systemTags = append(systemTags, param.SystemTagBootModeUEFI)
		} else {
			systemTags = append(systemTags, param.SystemTagBootModeLegacy)
		}
	} else {
		bootMode := strings.ToLower(imagePlan.BootMode.ValueString())

		switch bootMode {
		case "uefi":
			systemTags = append(systemTags, param.SystemTagBootModeUEFI)
		case "legacy":
			systemTags = append(systemTags, param.SystemTagBootModeLegacy)
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
		Params: param.AddImageDetailParam{
			Name:               imagePlan.Name.ValueString(),
			Description:        imagePlan.Description.ValueString(),
			Url:                imagePlan.Url.ValueString(),
			MediaType:          param.MediaType(imagePlan.MediaType.ValueString()), // param.RootVolumeTemplate,
			GuestOsType:        imagePlan.GuestOsType.ValueString(),
			System:             false,
			Format:             param.ImageFormat(imagePlan.Format.ValueString()), // param.Qcow2,
			Platform:           imagePlan.Platform.ValueString(),
			BackupStorageUuids: backupStorageUuids,
			//Type:               imagePlan.Type.ValueString(),
			ResourceUuid: "",
			Architecture: param.Architecture(imagePlan.Architecture.ValueString()),
			Virtio:       imagePlan.Virtio.ValueBool(),
		},
	}

	ctx = tflog.SetField(ctx, "url", imagePlan.Url)
	image, err := r.client.AddImage(imageParam)
	if err != nil {
		resp.Diagnostics.AddError(
			"Could not Add image to ZSphere Image storage"+image.Name, "Error "+err.Error(),
		)
		return
	}

	imagePlan.Uuid = types.StringValue(image.UUID)
	imagePlan.Name = types.StringValue(image.Name)
	imagePlan.Description = types.StringValue(image.Description)
	imagePlan.Url = types.StringValue(image.Url)
	imagePlan.GuestOsType = types.StringValue(image.GuestOsType)
	imagePlan.System = types.StringValue(image.System)
	imagePlan.Platform = types.StringValue(image.Platform)

	ctx = tflog.SetField(ctx, "url", image.Url)
	diags = resp.State.Set(ctx, imagePlan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete implements resource.Resource.
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

	err := r.client.DeleteImage(state.Uuid.ValueString(), param.DeleteModeEnforcing)

	if err != nil {
		resp.Diagnostics.AddError("fail to delete image", ""+err.Error())
		return
	}

	if expunge {
		tflog.Info(ctx, fmt.Sprintf("expunge image %s", state.Uuid.ValueString()))

		err = r.client.ExpungeImage(state.Uuid.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to expunge image", "Error: "+err.Error(),
			)
			return
		}
	}
}

// Metadata implements resource.Resource.
func (r *imageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_image"
}

// Read implements resource.Resource.
func (r *imageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state imageResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	image, err := r.client.GetImage(state.Uuid.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting ZStack Image uuid", "Could not read image uuid"+err.Error(),
		)
		return
	}

	state.Uuid = types.StringValue(image.UUID)
	state.Name = types.StringValue(image.Name)
	state.Url = types.StringValue(image.Url)

	if !state.Description.IsNull() {
		state.Description = types.StringValue(image.Description)
	}
	if !state.GuestOsType.IsNull() {
		state.GuestOsType = types.StringValue(image.GuestOsType)
	}
	if !state.Platform.IsNull() {
		state.Platform = types.StringValue(image.Platform)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Schema implements resource.Resource.
func (r *imageResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This resource allows you to manage images in ZSphere. " +
			"An image represents a virtual machine image format qcow2, raw, vmdk or an ISO file that can be used to create or boot virtual machines. " +
			"You can define the image's properties, such as its URL, format, architecture, and backup storage locations.",
		Attributes: map[string]schema.Attribute{
			"uuid": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the image. Automatically generated by ZSphere.",
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
				Description: "The URL where the image is located. This can be a file path or an HTTP link.",
			},
			"media_type": schema.StringAttribute{
				Optional:    true,
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
			},
			"platform": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The platform that the image is intended for, such as 'Linux', 'Windows', or others.",
				Validators: []validator.String{
					stringvalidator.OneOf("Linux", "Windows", "Other"),
				},
			},
			"format": schema.StringAttribute{
				Required:    true,
				Description: "The format of the image file, such as 'qcow2', 'raw', or 'vmdk'.",
				Validators: []validator.String{
					stringvalidator.OneOf("qcow2", "iso", "raw", "vmdk"),
				},
			},
			"image_storage_uuids": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "A list of UUIDs for the image storages where the image is stored.",
			},
			"architecture": schema.StringAttribute{
				Optional:    true,
				Description: "The architecture of the image, such as 'x86_64' or 'aarch64'.",
				Validators: []validator.String{
					stringvalidator.OneOf("x86_64", "aarch64", "mips64el", "loongarch64"),
				},
			},
			"virtio": schema.BoolAttribute{
				Optional:    true,
				Description: "Indicates if the VirtIO drivers are required for the image.",
			},
			"expunge": schema.BoolAttribute{
				Optional:    true,
				Description: "Indicates if the image should be expunged after deletion.",
			},
			"boot_mode": schema.StringAttribute{
				Optional:    true,
				Description: "The boot mode supported by the image, such as 'Legacy' or 'UEFI'.",
				Validators: []validator.String{
					stringvalidator.OneOf("Legacy", "UEFI"),
				},
			},
		},
	}
}

func (r *imageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

}
