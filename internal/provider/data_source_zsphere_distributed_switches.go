// Copyright (c) ZStack.io, Inc.

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terraform-zstack-modules/zsphere-sdk-go/pkg/client"
	"github.com/terraform-zstack-modules/zsphere-sdk-go/pkg/param"
)

var (
	_ datasource.DataSource              = &distributedSwitchDataSource{}
	_ datasource.DataSourceWithConfigure = &distributedSwitchDataSource{}
)

type distributedSwitchDataSource struct {
	client *client.ZSClient
}

type distributedSwitchDataSourceModel struct {
	Name              types.String                 `tfsdk:"name"`
	NamePattern       types.String                 `tfsdk:"name_pattern"`
	Filter            []Filter                     `tfsdk:"filter"`
	DistributedSwitches []distributedSwitchModel   `tfsdk:"distributed_switches"`
}

type distributedSwitchModel struct {
	Uuid                types.String `tfsdk:"uuid"`
	Name                types.String `tfsdk:"name"`
	Description         types.String `tfsdk:"description"`
	DatacenterUuid      types.String `tfsdk:"datacenter_uuid"`
	PhysicalInterface   types.String `tfsdk:"physical_interface"`
	VSwitchType         types.String `tfsdk:"vswitch_type"`
	Type                types.String `tfsdk:"type"`
	AttachedClusterUuids types.List  `tfsdk:"attached_cluster_uuids"`
}

func ZSphereDistributedSwitchDataSource() datasource.DataSource {
	return &distributedSwitchDataSource{}
}

func (d *distributedSwitchDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	cli, ok := req.ProviderData.(*client.ZSClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.ZSClient, got: %T. Please report this issue to the Provider developer. ", req.ProviderData),
		)
		return
	}
	d.client = cli
}

func (d *distributedSwitchDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_distributed_switches"
}

func (d *distributedSwitchDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state distributedSwitchDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := param.NewQueryParam()
	params.AddQ("type=VirtualSwitch")

	if !state.Name.IsNull() {
		params.AddQ("name=" + state.Name.ValueString())
	}
	if !state.NamePattern.IsNull() {
		params.AddQ("name~" + state.NamePattern.ValueString())
	}

	l2Networks, err := d.client.QueryL2Network(ctx, &params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Could not query distributed switches from ZSphere",
			"Error: "+err.Error(),
		)
		return
	}

	var distributedSwitches []distributedSwitchModel
	for _, network := range l2Networks {
		ds := distributedSwitchModel{
			Uuid:              types.StringValue(network.UUID),
			Name:              types.StringValue(network.Name),
			Description:       types.StringValue(network.Description),
			DatacenterUuid:    types.StringValue(network.ZoneUuid),
			PhysicalInterface: types.StringValue(network.PhysicalInterface),
			VSwitchType:       types.StringValue(network.VSwitchType),
			Type:              types.StringValue(network.Type),
		}

		if len(network.AttachedClusterUuids) > 0 {
			ds.AttachedClusterUuids, _ = types.ListValueFrom(ctx, types.StringType, network.AttachedClusterUuids)
		} else {
			ds.AttachedClusterUuids = types.ListNull(types.StringType)
		}

		distributedSwitches = append(distributedSwitches, ds)
	}

	state.DistributedSwitches = distributedSwitches

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (d *distributedSwitchDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Query distributed switches (L2VirtualSwitch) in ZSphere.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "Name of the distributed switch.",
			},
			"name_pattern": schema.StringAttribute{
				Optional:    true,
				Description: "Pattern to match distributed switch names.",
			},
			"filter": schema.ListNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name":   schema.StringAttribute{Required: true},
						"values": schema.ListAttribute{ElementType: types.StringType, Required: true},
						"regex":  schema.BoolAttribute{Optional: true},
					},
				},
			},
			"distributed_switches": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"uuid": schema.StringAttribute{
							Computed:    true,
							Description: "The UUID of the distributed switch.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Name of the distributed switch.",
						},
						"description": schema.StringAttribute{
							Computed:    true,
							Description: "Description of the distributed switch.",
						},
						"datacenter_uuid": schema.StringAttribute{
							Computed:    true,
							Description: "The UUID of the datacenter (zone) to which the distributed switch belongs.",
						},
						"physical_interface": schema.StringAttribute{
							Computed:    true,
							Description: "The physical interface used by the distributed switch.",
						},
						"vswitch_type": schema.StringAttribute{
							Computed:    true,
							Description: "The type of virtual switch.",
						},
						"type": schema.StringAttribute{
							Computed:    true,
							Description: "The type of the L2 network.",
						},
						"attached_cluster_uuids": schema.ListAttribute{
							Computed:    true,
							ElementType: types.StringType,
							Description: "List of cluster UUIDs attached to this distributed switch.",
						},
					},
				},
			},
		},
	}
}
