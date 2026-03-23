// Copyright (c) ZStack.io, Inc.

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terraform-zstack-modules/zsphere-sdk-go/pkg/client"
	"github.com/terraform-zstack-modules/zsphere-sdk-go/pkg/param"
)

var (
	_ resource.Resource              = &portGroupResource{}
	_ resource.ResourceWithConfigure = &portGroupResource{}
)

type portGroupResource struct {
	client *client.ZSClient
}

type portGroupResourceModel struct {
	Uuid           types.String            `tfsdk:"uuid"`
	Name           types.String            `tfsdk:"name"`
	Description    types.String            `tfsdk:"description"`
	DatacenterUuid types.String            `tfsdk:"datacenter_uuid"`
	VSwitchUuid    types.String            `tfsdk:"vswitch_uuid"`
	Vlan           types.Int64             `tfsdk:"vlan"`
	VlanMode       types.String            `tfsdk:"vlan_mode"`
	Type           types.String            `tfsdk:"type"`
	Category       types.String            `tfsdk:"category"`
	State          types.String            `tfsdk:"state"`
	DnsDomain      types.String            `tfsdk:"dns_domain"`
	IpVersion      types.Int64             `tfsdk:"ip_version"`
	EnableIPAM     types.Bool              `tfsdk:"enable_ipam"`
	DhcpService    types.Bool              `tfsdk:"dhcp_service"`
	DhcpIp         types.String            `tfsdk:"dhcp_ip"`
	Dns            types.List              `tfsdk:"dns"`
	IpRange        []portGroupIpRangeModel `tfsdk:"ip_range"`
}

type portGroupIpRangeModel struct {
	Name               types.String `tfsdk:"name"`
	Uuid               types.String `tfsdk:"uuid"`
	IsByCidr           types.Bool   `tfsdk:"is_by_cidr"`
	StartIp            types.String `tfsdk:"start_ip"`
	EndIp              types.String `tfsdk:"end_ip"`
	Netmask            types.String `tfsdk:"netmask"`
	PrefixLen          types.Int64  `tfsdk:"prefix_len"`
	Gateway            types.String `tfsdk:"gateway"`
	NetworkCidr        types.String `tfsdk:"network_cidr"`
	IpVersion          types.Int64  `tfsdk:"ip_version"`
	AddressMode        types.String `tfsdk:"address_mode"`
	IpAllocateStrategy types.String `tfsdk:"ip_allocate_strategy"`
}

func (r *portGroupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func PortGroupResource() resource.Resource {
	return &portGroupResource{}
}

func (r *portGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan portGroupResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	description := stringPtrFromTfString(plan.Description)
	vlanMode := stringPtrFromTfString(plan.VlanMode)
	resourceType := stringPtrFromTfString(plan.Type)
	category := stringPtrFromTfString(plan.Category)
	dnsDomain := stringPtrFromTfString(plan.DnsDomain)
	ipVersion := intPtrFromTfInt64(plan.IpVersion)
	enableIPAM := boolPtrFromTfBool(plan.EnableIPAM)

	systemTags := []string{}

	var dhcpService bool
	if !plan.DhcpService.IsNull() {
		dhcpService = plan.DhcpService.ValueBool()
	}

	if !plan.DhcpIp.IsNull() && plan.DhcpIp.ValueString() != "" {
		dhcpIpStr := plan.DhcpIp.ValueString()
		systemTags = append(systemTags, fmt.Sprintf("flatNetwork::DhcpServer::%s::ipUuid::null", dhcpIpStr))
	}

	portGroupParam := param.CreatePortGroupParam{
		BaseParam: param.BaseParam{
			SystemTags: systemTags,
		},
		Params: param.CreatePortGroupParamDetail{
			VSwitchUuid:   plan.VSwitchUuid.ValueString(),
			VlanMode:      vlanMode,
			Vlan:          int(plan.Vlan.ValueInt64()),
			Name:          plan.Name.ValueString(),
			Description:   description,
			Type:          resourceType,
			L2NetworkUuid: nil,
			Category:      category,
			IpVersion:     ipVersion,
			System:        false,
			DnsDomain:     dnsDomain,
			EnableIPAM:    enableIPAM,
		},
	}

	result, err := r.client.CreatePortGroup(ctx, portGroupParam)
	if err != nil {
		resp.Diagnostics.AddError(
			"Could not create port group in ZSphere", "Error: "+err.Error(),
		)
		return
	}

	plan.Uuid = types.StringValue(result.UUID)
	plan.Name = types.StringValue(result.Name)
	plan.Description = tfStringFromPtr(result.Description)
	plan.DatacenterUuid = types.StringValue(result.ZoneUuid)
	plan.VSwitchUuid = types.StringValue(result.VSwitchUuid)
	plan.Vlan = types.Int64Value(int64(result.VlanId))
	plan.VlanMode = tfStringFromPtr(result.VlanMode)
	plan.Type = types.StringValue(result.Type)
	plan.State = types.StringValue(result.State)
	plan.DnsDomain = tfStringFromPtr(result.DnsDomain)
	plan.IpVersion = types.Int64Value(int64(result.IpVersion))
	plan.EnableIPAM = types.BoolValue(result.EnableIPAM)

	plan.DhcpService = types.BoolValue(dhcpService)

	if len(result.Dns) > 0 {
		dnsValues := make([]string, len(result.Dns))
		copy(dnsValues, result.Dns)
		plan.Dns, _ = types.ListValueFrom(ctx, types.StringType, dnsValues)
	}

	if !plan.EnableIPAM.IsNull() && plan.EnableIPAM.ValueBool() && len(plan.IpRange) > 0 {
		for _, ipRangePlan := range plan.IpRange {
			err := r.addIpRangeToPortGroup(ctx, result.UUID, ipRangePlan)
			if err != nil {
				resp.Diagnostics.AddError(
					"Could not add IP range to port group in ZSphere", "Error: "+err.Error(),
				)
				return
			}
		}
	}

	updatedResult, err := r.client.GetPortGroup(ctx, result.UUID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Could not refresh port group after adding IP range", "Error: "+err.Error(),
		)
		return
	}
	result = updatedResult

	if len(result.IpRanges) > 0 {
		for i, ipRange := range result.IpRanges {
			if i < len(plan.IpRange) {
				plan.IpRange[i].Uuid = types.StringValue(ipRange.UUID)
				plan.IpRange[i].Name = types.StringValue(ipRange.Name)
				plan.IpRange[i].StartIp = types.StringValue(ipRange.StartIp)
				plan.IpRange[i].EndIp = types.StringValue(ipRange.EndIp)
				plan.IpRange[i].Netmask = types.StringValue(ipRange.Netmask)
				plan.IpRange[i].Gateway = types.StringValue(ipRange.Gateway)
				plan.IpRange[i].IpVersion = types.Int64Value(int64(ipRange.IpVersion))
				if plan.IpRange[i].AddressMode.IsUnknown() {
					plan.IpRange[i].AddressMode = types.StringNull()
				}
				if plan.IpRange[i].IsByCidr.IsUnknown() {
					plan.IpRange[i].IsByCidr = types.BoolNull()
				}
				if plan.IpRange[i].PrefixLen.IsUnknown() {
					plan.IpRange[i].PrefixLen = types.Int64Null()
				}
				if plan.IpRange[i].IpAllocateStrategy.IsUnknown() {
					plan.IpRange[i].IpAllocateStrategy = types.StringNull()
				}
			}
		}
	}

	if len(result.Dns) > 0 {
		dnsValues := make([]string, len(result.Dns))
		copy(dnsValues, result.Dns)
		plan.Dns, _ = types.ListValueFrom(ctx, types.StringType, dnsValues)
	} else if plan.Dns.IsUnknown() {
		emptyDns, _ := types.ListValueFrom(ctx, types.StringType, []string{})
		plan.Dns = emptyDns
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *portGroupResource) addIpRangeToPortGroup(ctx context.Context, l3NetworkUuid string, ipRangePlan portGroupIpRangeModel) error {
	var ipRangeParam param.AddIpRangeParam
	var systemTags []string

	if !ipRangePlan.IpAllocateStrategy.IsNull() && ipRangePlan.IpAllocateStrategy.ValueString() != "" {
		strategy := ipRangePlan.IpAllocateStrategy.ValueString()
		systemTags = append(systemTags, fmt.Sprintf("resourceConfig::l3Network::ipAllocateStrategy::%sStrategy", strategy))
	}

	ipRangeParam = param.AddIpRangeParam{
		BaseParam: param.BaseParam{
			SystemTags: systemTags,
		},
		Params: param.AddIpRangeParamDetail{
			Name:    ipRangePlan.Name.ValueString(),
			StartIp: ipRangePlan.StartIp.ValueString(),
			EndIp:   ipRangePlan.EndIp.ValueString(),
			Netmask: ipRangePlan.Netmask.ValueString(),
		},
	}

	if !ipRangePlan.Gateway.IsNull() && ipRangePlan.Gateway.ValueString() != "" {
		gateway := ipRangePlan.Gateway.ValueString()
		ipRangeParam.Params.Gateway = &gateway
	}

	_, err := r.client.AddIpRange(ctx, l3NetworkUuid, ipRangeParam)
	return err
}

func (r *portGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state portGroupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.GetPortGroup(ctx, state.Uuid.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Could not read port group in ZSphere", "Error: "+err.Error(),
		)
		return
	}

	state.Uuid = types.StringValue(result.UUID)
	state.Name = types.StringValue(result.Name)
	state.Description = tfStringFromPtr(result.Description)
	state.DatacenterUuid = types.StringValue(result.ZoneUuid)
	state.VSwitchUuid = types.StringValue(result.VSwitchUuid)
	state.Vlan = types.Int64Value(int64(result.VlanId))
	state.VlanMode = tfStringFromPtr(result.VlanMode)
	state.Type = types.StringValue(result.Type)
	state.Category = types.StringValue(result.Category)
	state.State = types.StringValue(result.State)
	state.DnsDomain = tfStringFromPtr(result.DnsDomain)
	state.IpVersion = types.Int64Value(int64(result.IpVersion))
	state.EnableIPAM = types.BoolValue(result.EnableIPAM)

	if len(result.Dns) > 0 {
		state.Dns, _ = types.ListValueFrom(ctx, types.StringType, result.Dns)
	} else {
		state.Dns = types.ListNull(types.StringType)
	}

	if len(result.IpRanges) > 0 {
		state.IpRange = make([]portGroupIpRangeModel, len(result.IpRanges))
		for i, ipRange := range result.IpRanges {
			state.IpRange[i] = portGroupIpRangeModel{
				Uuid:      types.StringValue(ipRange.UUID),
				Name:      types.StringValue(ipRange.Name),
				StartIp:   types.StringValue(ipRange.StartIp),
				EndIp:     types.StringValue(ipRange.EndIp),
				Netmask:   types.StringValue(ipRange.Netmask),
				Gateway:   types.StringValue(ipRange.Gateway),
				IpVersion: types.Int64Value(int64(ipRange.IpVersion)),
			}
		}
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *portGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan portGroupResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	description := stringPtrFromTfString(plan.Description)
	dnsDomain := stringPtrFromTfString(plan.DnsDomain)
	category := stringPtrFromTfString(plan.Category)

	updateParam := param.UpdatePortGroupParam{
		BaseParam: param.BaseParam{},
		Params: param.UpdatePortGroupParamDetail{
			Name:        plan.Name.ValueString(),
			Description: description,
			DnsDomain:   dnsDomain,
			Category:    category,
		},
	}

	result, err := r.client.UpdatePortGroup(ctx, plan.Uuid.ValueString(), updateParam)
	if err != nil {
		resp.Diagnostics.AddError(
			"Could not update port group in ZSphere", "Error: "+err.Error(),
		)
		return
	}

	plan.Uuid = types.StringValue(result.UUID)
	plan.Name = types.StringValue(result.Name)
	plan.Description = tfStringFromPtr(result.Description)
	plan.DatacenterUuid = types.StringValue(result.ZoneUuid)
	plan.VSwitchUuid = types.StringValue(result.VSwitchUuid)
	plan.Vlan = types.Int64Value(int64(result.VlanId))
	plan.VlanMode = tfStringFromPtr(result.VlanMode)
	plan.Type = types.StringValue(result.Type)
	plan.State = types.StringValue(result.State)
	plan.DnsDomain = tfStringFromPtr(result.DnsDomain)
	plan.IpVersion = types.Int64Value(int64(result.IpVersion))
	plan.EnableIPAM = types.BoolValue(result.EnableIPAM)

	if len(result.Dns) > 0 {
		dnsValues := make([]string, len(result.Dns))
		copy(dnsValues, result.Dns)
		plan.Dns, _ = types.ListValueFrom(ctx, types.StringType, dnsValues)
	} else if plan.Dns.IsUnknown() {
		emptyDns, _ := types.ListValueFrom(ctx, types.StringType, []string{})
		plan.Dns = emptyDns
	}

	if len(result.IpRanges) > 0 {
		for i, ipRange := range result.IpRanges {
			if i == len(plan.IpRange) {
				plan.IpRange = append(plan.IpRange, portGroupIpRangeModel{})
			}
			plan.IpRange[i].Uuid = types.StringValue(ipRange.UUID)
			plan.IpRange[i].Name = types.StringValue(ipRange.Name)
			plan.IpRange[i].StartIp = types.StringValue(ipRange.StartIp)
			plan.IpRange[i].EndIp = types.StringValue(ipRange.EndIp)
			plan.IpRange[i].Netmask = types.StringValue(ipRange.Netmask)
			plan.IpRange[i].Gateway = types.StringValue(ipRange.Gateway)
			plan.IpRange[i].IpVersion = types.Int64Value(int64(ipRange.IpVersion))
			if plan.IpRange[i].AddressMode.IsUnknown() {
				plan.IpRange[i].AddressMode = types.StringNull()
			}
			if plan.IpRange[i].IsByCidr.IsUnknown() {
				plan.IpRange[i].IsByCidr = types.BoolNull()
			}
			if plan.IpRange[i].PrefixLen.IsUnknown() {
				plan.IpRange[i].PrefixLen = types.Int64Null()
			}
			if plan.IpRange[i].IpAllocateStrategy.IsUnknown() {
				plan.IpRange[i].IpAllocateStrategy = types.StringNull()
			}
		}
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *portGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state portGroupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeletePortGroup(ctx, state.Uuid.ValueString(), param.DeleteModeEnforcing)
	if err != nil {
		resp.Diagnostics.AddError(
			"Could not delete port group in ZSphere", "Error: "+err.Error(),
		)
		return
	}
}

func (r *portGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_port_group"
}

func stringPtrFromTfString(s types.String) *string {
	if s.IsNull() || s.ValueString() == "" {
		return nil
	}
	str := s.ValueString()
	return &str
}

func tfStringFromPtr(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}

func intPtrFromTfInt64(i types.Int64) *int {
	if i.IsNull() {
		return nil
	}
	val := int(i.ValueInt64())
	return &val
}

func boolPtrFromTfBool(b types.Bool) *bool {
	if b.IsNull() {
		return nil
	}
	val := b.ValueBool()
	return &val
}

func (r *portGroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Resource definition for ZSphere Port Group.",
		Attributes: map[string]schema.Attribute{
			"uuid": schema.StringAttribute{
				Computed:    true,
				Description: "UUID of the port group.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the port group.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Description of the port group.",
			},
			"datacenter_uuid": schema.StringAttribute{
				Computed:    true,
				Description: "UUID of the datacenter (zone) that this port group belongs to.",
			},
			"vswitch_uuid": schema.StringAttribute{
				Required:    true,
				Description: "UUID of the vSwitch that this port group is attached to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"vlan": schema.Int64Attribute{
				Required:    true,
				Description: "VLAN ID for the port group.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"vlan_mode": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "VLAN mode for the port group.",
				Validators: []validator.String{
					stringvalidator.OneOf("", "ACCESS", "NONE", "PVLAN", "TRUNK"),
				},
			},
			"type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Type of the port group.",
			},
			"category": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Category of the port group.",
			},
			"state": schema.StringAttribute{
				Computed:    true,
				Description: "State of the port group.",
			},
			"dns_domain": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "DNS domain for the port group.",
			},
			"ip_version": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "IP version (4 or 6).",
				Validators: []validator.Int64{
					int64validator.OneOf(4, 6),
				},
			},
			"enable_ipam": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether IPAM is enabled for the port group.",
			},
			"dhcp_service": schema.BoolAttribute{
				Optional:    true,
				Description: "Whether DHCP service is enabled for the port group.",
			},
			"dhcp_ip": schema.StringAttribute{
				Optional:    true,
				Description: "DHCP service IP address.",
			},
			"dns": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Description: "DNS servers for the port group.",
			},
		},
		Blocks: map[string]schema.Block{
			"ip_range": schema.ListNestedBlock{
				Description: "IP range configuration for the port group.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:    true,
							Description: "Name of the IP range.",
						},
						"uuid": schema.StringAttribute{
							Computed:    true,
							Description: "UUID of the IP range.",
						},
						"is_by_cidr": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Whether to use CIDR notation instead of IP range.",
						},
						"start_ip": schema.StringAttribute{
							Optional:    true,
							Description: "Start IP address of the range (IP range mode).",
						},
						"end_ip": schema.StringAttribute{
							Optional:    true,
							Description: "End IP address of the range (IP range mode).",
						},
						"netmask": schema.StringAttribute{
							Optional:    true,
							Description: "Netmask (IP range mode).",
						},
						"prefix_len": schema.Int64Attribute{
							Optional:    true,
							Computed:    true,
							Description: "Prefix length for IPv6 CIDR.",
							Validators: []validator.Int64{
								int64validator.Between(64, 126),
							},
						},
						"gateway": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Gateway address.",
						},
						"network_cidr": schema.StringAttribute{
							Optional:    true,
							Description: "Network CIDR (CIDR mode).",
						},
						"ip_version": schema.Int64Attribute{
							Optional:    true,
							Computed:    true,
							Description: "IP version (4 or 6).",
						},
						"address_mode": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "IPv6 address mode: Stateful-DHCP, Stateless-DHCP, or SLAAC.",
							Validators: []validator.String{
								stringvalidator.OneOf("", "Stateful-DHCP", "Stateless-DHCP", "SLAAC"),
							},
						},
						"ip_allocate_strategy": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "IP allocation strategy: RandomIpAllocator, FirstAvailableIpAllocator, AscDelayRecycleIpAllocator.",
							Validators: []validator.String{
								stringvalidator.OneOf("", "RandomIpAllocator", "FirstAvailableIpAllocator", "AscDelayRecycleIpAllocator"),
							},
						},
					},
				},
			},
		},
	}
}
