// Copyright (c) ZStack.io, Inc.

package provider

import (
	"context"
	"fmt"
	"terraform-provider-zsphere/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terraform-zstack-modules/zsphere-sdk-go/pkg/client"
	"github.com/terraform-zstack-modules/zsphere-sdk-go/pkg/param"
)

var (
	_ datasource.DataSource              = &imageDataSource{}
	_ datasource.DataSourceWithConfigure = &imageDataSource{}
)

type imageDataSource struct {
	client *client.ZSClient
}

type imagesModel struct {
	Name         types.String `tfsdk:"name"`
	State        types.String `tfsdk:"state"`
	Status       types.String `tfsdk:"status"`
	Uuid         types.String `tfsdk:"uuid"`
	Format       types.String `tfsdk:"format"`
	Platform     types.String `tfsdk:"platform"`
	Architecture types.String `tfsdk:"architecture"`
}

type imagesDataSourceModel struct {
	Name        types.String  `tfsdk:"name"`
	NamePattern types.String  `tfsdk:"name_pattern"`
	Images      []imagesModel `tfsdk:"images"`
	Filter      []Filter      `tfsdk:"filter"`
}

func ZSphereImageDataSource() datasource.DataSource {
	return &imageDataSource{}
}

// Configure implements datasource.DataSourceWithConfigure.
func (d *imageDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.ZSClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.ZSClient, got: %T. Please report this issue to the ZSphere Provider developer. ", req.ProviderData),
		)
		return
	}

	d.client = client
}

// Metadata implements datasource.DataSource.
func (d *imageDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_images"
}

// Read implements datasource.DataSource.
func (d *imageDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	var state imagesDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := param.NewQueryParam()

	if !state.Name.IsNull() {
		params.AddQ("name=" + state.Name.ValueString())
	} else if !state.NamePattern.IsNull() {
		params.AddQ("name~=" + state.NamePattern.ValueString())
	}

	images, err := d.client.QueryImage(params)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read ZStack Images",
			err.Error(),
		)
		return
	}

	filters := make(map[string][]string)
	for _, filter := range state.Filter {
		values := make([]string, 0, len(filter.Values.Elements()))
		diags := filter.Values.ElementsAs(ctx, &values, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		filters[filter.Name.ValueString()] = values
	}

	filterImages, filterDiags := utils.FilterResource(ctx, images, filters, "image")
	resp.Diagnostics.Append(filterDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	for _, image := range filterImages {
		imageState := imagesModel{
			Name:         types.StringValue(image.Name),
			State:        types.StringValue(image.State),
			Status:       types.StringValue(image.Status),
			Uuid:         types.StringValue(image.UUID),
			Format:       types.StringValue(image.Format),
			Platform:     types.StringValue(image.Platform),
			Architecture: types.StringValue(string(image.Architecture)),
		}
		state.Images = append(state.Images, imageState)
	}

	diags = resp.State.Set(ctx, state)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

// Schema implements datasource.DataSource.
func (d *imageDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a list of images and their associated attributes from the ZSphere environment.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Exact name for searching images",
				Optional:    true,
			},
			"name_pattern": schema.StringAttribute{
				Description: "Pattern for fuzzy name search, similar to MySQL LIKE. Use % for multiple characters and _ for exactly one character.",
				Optional:    true,
			},
			"images": schema.ListNestedAttribute{
				Description: "List of Images",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "Name of the image",
							Computed:    true,
						},

						"uuid": schema.StringAttribute{
							Description: "UUID identifier of the image",
							Computed:    true,
						},
						"state": schema.StringAttribute{
							Description: "State of the image, indicating if it is Enabled or Disabled",
							Computed:    true,
						},
						"status": schema.StringAttribute{
							Description: "Readiness status of the image (e.g., Ready or Not Ready)",
							Computed:    true,
						},
						"format": schema.StringAttribute{
							Description: "Format of the image, such as qcow2, iso, vmdk, or raw",
							Computed:    true,
						},
						"platform": schema.StringAttribute{
							Description: "Platform of the image, such as Linux, Windows, or Other",
							Computed:    true,
						},
						"architecture": schema.StringAttribute{
							Description: "CPU architecture of the image, such as x86_64, aarch64, mips64, or longarch64",
							Computed:    true,
						},
					},
				},
			},
		},
		Blocks: map[string]schema.Block{
			"filter": schema.ListNestedBlock{
				Description: "Filter resources based on any field in the schema. For example, to filter by status, use `name = \"status\"` and `values = [\"Ready\"]`.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "Name of the field to filter by (e.g., status, state).",
							Required:    true,
						},
						"values": schema.SetAttribute{
							Description: "Values to filter by. Multiple values will be treated as an OR condition.",
							Required:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
		},
	}
}
