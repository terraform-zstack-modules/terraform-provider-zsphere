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
	_ datasource.DataSource              = &securityGroupDataSource{}
	_ datasource.DataSourceWithConfigure = &securityGroupDataSource{}
)

type securityGroupDataSource struct {
	client *client.ZSClient
}

type securityGroupDataSourceModel struct {
	Name           types.String                       `tfsdk:"name"`
	NamePattern    types.String                       `tfsdk:"name_pattern"`
	Filter         []Filter                           `tfsdk:"filter"`
	SecurityGroups []securityGroupDataSourceModelType `tfsdk:"security_groups"`
}

type securityGroupDataSourceModelType struct {
	Uuid                   types.String                       `tfsdk:"uuid"`
	Name                   types.String                       `tfsdk:"name"`
	Description            types.String                       `tfsdk:"description"`
	State                  types.String                       `tfsdk:"state"`
	IpVersion              types.Int64                        `tfsdk:"ip_version"`
	AttachedL3NetworkUuids types.List                         `tfsdk:"attached_l3_network_uuids"`
	VmNicUuids             types.List                         `tfsdk:"vm_nic_uuids"`
	Rules                  []securityGroupRuleDataSourceModel `tfsdk:"rules"`
}

type securityGroupRuleDataSourceModel struct {
	Type                    types.String `tfsdk:"type"`
	Protocol                types.String `tfsdk:"protocol"`
	Priority                types.Int64  `tfsdk:"priority"`
	Action                  types.String `tfsdk:"action"`
	State                   types.String `tfsdk:"state"`
	AllowedCidr             types.String `tfsdk:"allowed_cidr"`
	RemoteSecurityGroupUuid types.String `tfsdk:"remote_security_group_uuid"`
	DstPortRange            types.String `tfsdk:"dst_port_range"`
	SrcIpRange              types.String `tfsdk:"src_ip_range"`
	DstIpRange              types.String `tfsdk:"dst_ip_range"`
	Description             types.String `tfsdk:"description"`
}

func SecurityGroupDataSource() datasource.DataSource {
	return &securityGroupDataSource{}
}

func (d *securityGroupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *securityGroupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_groups"
}

func (d *securityGroupDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a list of security groups and their associated attributes from the ZSphere environment.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Exact name for searching security groups.",
				Optional:    true,
			},
			"name_pattern": schema.StringAttribute{
				Description: "Pattern for fuzzy name search, similar to MySQL LIKE. Use % for multiple characters and _ for exactly one character.",
				Optional:    true,
			},
			"security_groups": schema.ListNestedAttribute{
				Description: "List of security groups matching the specified filters.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"uuid": schema.StringAttribute{
							Computed:    true,
							Description: "UUID of the security group.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Name of the security group.",
						},
						"description": schema.StringAttribute{
							Computed:    true,
							Description: "Description of the security group.",
						},
						"state": schema.StringAttribute{
							Computed:    true,
							Description: "State of the security group.",
						},
						"ip_version": schema.Int64Attribute{
							Computed:    true,
							Description: "IP version (4 or 6).",
						},
						"attached_l3_network_uuids": schema.ListAttribute{
							ElementType: types.StringType,
							Computed:    true,
							Description: "UUIDs of L3 networks that the security group is attached to.",
						},
						"vm_nic_uuids": schema.ListAttribute{
							ElementType: types.StringType,
							Computed:    true,
							Description: "UUIDs of VM nics bound to the security group.",
						},
					},
				},
			},
		},
		Blocks: map[string]schema.Block{
			"filter": schema.ListNestedBlock{
				Description: "Filter resources based on any field in the schema. For example, to filter by state, use `name = \"state\"` and `values = [\"Enabled\"]`.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "Name of the field to filter by (e.g., state, name).",
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
			"rules": schema.ListNestedBlock{
				Description: "Security group rules.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Computed:    true,
							Description: "Rule type: Ingress or Egress.",
						},
						"protocol": schema.StringAttribute{
							Computed:    true,
							Description: "Protocol: TCP, UDP, ICMP, or ALL.",
						},
						"priority": schema.Int64Attribute{
							Computed:    true,
							Description: "Rule priority.",
						},
						"action": schema.StringAttribute{
							Computed:    true,
							Description: "Rule action: ACCEPT or DROP.",
						},
						"state": schema.StringAttribute{
							Computed:    true,
							Description: "Rule state: Enabled or Disabled.",
						},
						"allowed_cidr": schema.StringAttribute{
							Computed:    true,
							Description: "Allowed CIDR block.",
						},
						"remote_security_group_uuid": schema.StringAttribute{
							Computed:    true,
							Description: "Remote security group UUID.",
						},
						"dst_port_range": schema.StringAttribute{
							Computed:    true,
							Description: "Destination port range.",
						},
						"src_ip_range": schema.StringAttribute{
							Computed:    true,
							Description: "Source IP range.",
						},
						"dst_ip_range": schema.StringAttribute{
							Computed:    true,
							Description: "Destination IP range.",
						},
						"description": schema.StringAttribute{
							Computed:    true,
							Description: "Description of the rule.",
						},
					},
				},
			},
		},
	}
}

func (d *securityGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state securityGroupDataSourceModel
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

	securityGroups, err := d.client.QuerySecurityGroup(ctx, &params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read ZSphere Security Groups",
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

	filterSecurityGroups, filterDiags := utils.FilterResource(ctx, securityGroups, filters, "security_group")
	resp.Diagnostics.Append(filterDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	for _, sg := range filterSecurityGroups {
		sgState := securityGroupDataSourceModelType{
			Uuid:        types.StringValue(sg.UUID),
			Name:        types.StringValue(sg.Name),
			Description: tfStringFromPtr(sg.Description),
			State:       types.StringValue(sg.State),
			IpVersion:   types.Int64Value(int64(sg.IpVersion)),
		}

		if len(sg.AttachedL3NetworkUuids) > 0 {
			sgState.AttachedL3NetworkUuids, _ = types.ListValueFrom(ctx, types.StringType, sg.AttachedL3NetworkUuids)
		} else {
			sgState.AttachedL3NetworkUuids = types.ListNull(types.StringType)
		}

		vmNicUuids, err := d.queryVmNicsForSecurityGroup(ctx, sg.UUID)
		if err == nil && len(vmNicUuids) > 0 {
			sgState.VmNicUuids, _ = types.ListValueFrom(ctx, types.StringType, vmNicUuids)
		} else {
			sgState.VmNicUuids = types.ListNull(types.StringType)
		}

		if len(sg.Rules) > 0 {
			sgState.Rules = make([]securityGroupRuleDataSourceModel, len(sg.Rules))
			for i, rule := range sg.Rules {
				priority := int64(0)
				if rule.Priority > 0 {
					priority = int64(rule.Priority)
				}
				ruleState := rule.State
				if ruleState == "" {
					ruleState = "Enabled"
				}
				sgState.Rules[i] = securityGroupRuleDataSourceModel{
					Type:                    types.StringValue(rule.Type),
					Protocol:                types.StringValue(rule.Protocol),
					Priority:                types.Int64Value(priority),
					Action:                  types.StringValue(rule.Action),
					State:                   types.StringValue(ruleState),
					AllowedCidr:             tfStringFromPtr(rule.AllowedCidr),
					RemoteSecurityGroupUuid: tfStringFromPtr(rule.RemoteSecurityGroupUuid),
					DstPortRange:            tfStringFromPtr(rule.DstPortRange),
					SrcIpRange:              tfStringFromPtr(rule.SrcIpRange),
					DstIpRange:              tfStringFromPtr(rule.DstIpRange),
					Description:             tfStringFromPtr(rule.Description),
				}
			}
		}

		state.SecurityGroups = append(state.SecurityGroups, sgState)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (d *securityGroupDataSource) queryVmNicsForSecurityGroup(ctx context.Context, securityGroupUuid string) ([]string, error) {
	if securityGroupUuid == "" {
		return nil, fmt.Errorf("securityGroupUuid cannot be empty")
	}

	params := &param.QueryParam{}
	params.AddQ("securityGroupUuid=" + securityGroupUuid)

	vmNicRefs, err := d.client.QueryVmNicInSecurityGroup(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to query VM NICs for security group %s: %w", securityGroupUuid, err)
	}

	vmNicUuids := make([]string, 0, len(vmNicRefs))
	for _, ref := range vmNicRefs {
		if ref.VmNicUuid != "" {
			vmNicUuids = append(vmNicUuids, ref.VmNicUuid)
		}
	}

	return vmNicUuids, nil
}
