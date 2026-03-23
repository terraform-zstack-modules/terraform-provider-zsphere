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
	_ datasource.DataSource              = &nfsPrimaryStorageDataSource{}
	_ datasource.DataSourceWithConfigure = &nfsPrimaryStorageDataSource{}
)

func ZSphereNfsPrimaryStorageDataSource() datasource.DataSource {
	return &nfsPrimaryStorageDataSource{}
}

type nfsStorage struct {
	Name                      types.String `tfsdk:"name"`
	Uuid                      types.String `tfsdk:"uuid"`
	State                     types.String `tfsdk:"state"`
	Status                    types.String `tfsdk:"status"`
	TotalCapacity             types.Int64  `tfsdk:"total_capacity"`
	AvailableCapacity         types.Int64  `tfsdk:"available_capacity"`
	TotalPhysicalCapacity     types.Int64  `tfsdk:"total_physical_capacity"`
	AvailablePhysicalCapacity types.Int64  `tfsdk:"available_physical_capacity"`
	SystemUsedCapacity        types.Int64  `tfsdk:"system_used_capacity"`
	Type                      types.String `tfsdk:"type"`
	DatacenterUuid            types.String `tfsdk:"datacenter_uuid"`
	Url                       types.String `tfsdk:"url"`
	MountPath                 types.String `tfsdk:"mount_path"`
}

type nfsPrimaryStorageDataSourceModel struct {
	Name           types.String     `tfsdk:"name"`
	NamePattern    types.String     `tfsdk:"name_pattern"`
	Filter         []Filter         `tfsdk:"filter"`
	NfsStorages []nfsStorage      `tfsdk:"nfs_storages"`
}

type nfsPrimaryStorageDataSource struct {
	client *client.ZSClient
}

func (d *nfsPrimaryStorageDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
	d.client = client
}

func (d *nfsPrimaryStorageDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nfs_primary_storages"
}

func (d *nfsPrimaryStorageDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state nfsPrimaryStorageDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := param.NewQueryParam()
	params.AddQ("type=NFS")

	if !state.Name.IsNull() {
		params.AddQ("name=" + state.Name.ValueString())
	} else if !state.NamePattern.IsNull() {
		params.AddQ("name~=" + state.NamePattern.ValueString())
	}

	nfsStorages, err := d.client.QueryPrimaryStorage(ctx, &params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read ZStack NFS Primary Storages",
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

	filterNfsStorage, filterDiags := utils.FilterResource(ctx, nfsStorages, filters, "nfs_primary_storage")
	resp.Diagnostics.Append(filterDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	for _, storage := range filterNfsStorage {
		nfsStorageState := nfsStorage{
			TotalCapacity:             types.Int64Value(storage.TotalCapacity),
			State:                     types.StringValue(storage.State),
			Status:                    types.StringValue(storage.Status),
			Uuid:                      types.StringValue(storage.UUID),
			AvailableCapacity:         types.Int64Value(storage.AvailableCapacity),
			Name:                      types.StringValue(storage.Name),
			TotalPhysicalCapacity:     types.Int64Value(storage.TotalPhysicalCapacity),
			AvailablePhysicalCapacity: types.Int64Value(storage.AvailablePhysicalCapacity),
			SystemUsedCapacity:        types.Int64Value(storage.SystemUsedCapacity),
			Type:                      types.StringValue(storage.Type),
			DatacenterUuid:            types.StringValue(storage.ZoneUuid),
			Url:                       types.StringValue(storage.Url),
			MountPath:                 types.StringValue(storage.MountPath),
		}

		state.NfsStorages = append(state.NfsStorages, nfsStorageState)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

func (d *nfsPrimaryStorageDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "List all NFS primary storages, or query NFS primary storages by exact name match, or query NFS primary storages by name pattern fuzzy match.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Exact name for searching NFS primary storage.",
				Optional:    true,
			},
			"name_pattern": schema.StringAttribute{
				Description: "Pattern for fuzzy name search, similar to MySQL LIKE. Use % for multiple characters and _ for exactly one character.",
				Optional:    true,
			},
			"nfs_storages": schema.ListNestedAttribute{
				Description: "List of NFS primary storage entries",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "Name of the NFS primary storage",
							Computed:    true,
						},

						"uuid": schema.StringAttribute{
							Description: "UUID identifier of the NFS primary storage",
							Computed:    true,
						},
						"state": schema.StringAttribute{
							Description: "State of the NFS primary storage (Enabled or Disabled)",
							Computed:    true,
						},
						"status": schema.StringAttribute{
							Description: "Readiness status of the NFS primary storage",
							Computed:    true,
						},
						"total_capacity": schema.Int64Attribute{
							Description: "Total capacity of the NFS primary storage in bytes",
							Computed:    true,
						},
						"available_capacity": schema.Int64Attribute{
							Description: "Available capacity of the NFS primary storage in bytes",
							Computed:    true,
						},
						"total_physical_capacity": schema.Int64Attribute{
							Description: "Total physical capacity of the NFS primary storage in bytes",
							Computed:    true,
						},
						"available_physical_capacity": schema.Int64Attribute{
							Description: "Available physical capacity of the NFS primary storage in bytes",
							Computed:    true,
						},
						"system_used_capacity": schema.Int64Attribute{
							Description: "System used capacity of the NFS primary storage in bytes",
							Computed:    true,
						},
						"type": schema.StringAttribute{
							Description: "Type of the NFS primary storage",
							Computed:    true,
						},
						"datacenter_uuid": schema.StringAttribute{
							Description: "Datacenter UUID of the NFS primary storage",
							Computed:    true,
						},
						"url": schema.StringAttribute{
							Description: "URL of the NFS primary storage",
							Computed:    true,
						},
						"mount_path": schema.StringAttribute{
							Description: "Mount path of the NFS primary storage",
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
