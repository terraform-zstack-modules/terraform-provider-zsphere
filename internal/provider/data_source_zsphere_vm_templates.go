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
	_ datasource.DataSource              = &vmTemplateDataSource{}
	_ datasource.DataSourceWithConfigure = &vmTemplateDataSource{}
)

type vmTemplateDataSourceModel struct {
	Name        types.String      `tfsdk:"name"`
	NamePattern types.String      `tfsdk:"name_pattern"`
	VmTemplates []vmTemplateModel `tfsdk:"vm_templates"`
	Filter      []Filter          `tfsdk:"filter"`
}

type vmTemplateModel struct {
	Name        types.String `tfsdk:"name"`
	Uuid        types.String `tfsdk:"uuid"`
	ZoneUuid    types.String `tfsdk:"zone_uuid"`
	AccountUuid types.String `tfsdk:"account_uuid"`
}

func ZSphereVmTemplateDataSource() datasource.DataSource {
	return &vmTemplateDataSource{}
}

type vmTemplateDataSource struct {
	client *client.ZSClient
}

func (d *vmTemplateDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *vmTemplateDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vm_templates"
}

func (d *vmTemplateDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state vmTemplateDataSourceModel
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

	templates, err := d.client.QueryTemplatedVmInstance(ctx, &params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read VM Templates",
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

	filterTemplates, filterDiags := utils.FilterResource(ctx, templates, filters, "vmTemplate")
	resp.Diagnostics.Append(filterDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	for _, template := range filterTemplates {
		templateState := vmTemplateModel{
			Name:        types.StringValue(template.Name),
			Uuid:        types.StringValue(template.UUID),
			ZoneUuid:    types.StringValue(template.ZoneUuid),
			AccountUuid: types.StringValue(template.AccountUuid),
		}

		state.VmTemplates = append(state.VmTemplates, templateState)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (d *vmTemplateDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a list of VM templates and their associated attributes from the ZSphere environment.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Exact name for searching VM templates",
				Optional:    true,
			},
			"name_pattern": schema.StringAttribute{
				Description: "Pattern for fuzzy name search, similar to MySQL LIKE. Use % for multiple characters and _ for exactly one character.",
				Optional:    true,
			},
			"vm_templates": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"uuid": schema.StringAttribute{
							Computed:    true,
							Description: "The unique identifier (UUID) of the VM template.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the VM template.",
						},
						"zone_uuid": schema.StringAttribute{
							Computed:    true,
							Description: "The UUID of the zone in which the VM template is located.",
						},
						"account_uuid": schema.StringAttribute{
							Computed:    true,
							Description: "The UUID of the account that owns the VM template.",
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
