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
	_ resource.Resource              = &securityGroupResource{}
	_ resource.ResourceWithConfigure = &securityGroupResource{}
)

type securityGroupResource struct {
	client *client.ZSClient
}

type securityGroupResourceModel struct {
	Uuid                   types.String             `tfsdk:"uuid"`
	Name                   types.String             `tfsdk:"name"`
	Description            types.String             `tfsdk:"description"`
	State                  types.String             `tfsdk:"state"`
	IpVersion              types.Int64              `tfsdk:"ip_version"`
	L3NetworkUuids         types.List               `tfsdk:"l3_network_uuids"`
	VmNicUuids             types.List               `tfsdk:"vm_nic_uuids"`
	AttachedL3NetworkUuids types.List               `tfsdk:"attached_l3_network_uuids"`
	Rules                  []securityGroupRuleModel `tfsdk:"rules"`
}

type securityGroupRuleModel struct {
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

func (r *securityGroupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func SecurityGroupResource() resource.Resource {
	return &securityGroupResource{}
}

func (r *securityGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan securityGroupResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	description := stringPtrFromTfString(plan.Description)

	createSGParam := param.CreateSecurityGroupParam{
		BaseParam: param.BaseParam{},
		Params: param.CreateSecurityGroupParamDetail{
			Name:        plan.Name.ValueString(),
			Description: description,
		},
	}

	result, err := r.client.CreateSecurityGroup(ctx, createSGParam)
	if err != nil {
		resp.Diagnostics.AddError(
			"Could not create security group in ZSphere", "Error: "+err.Error(),
		)
		return
	}

	securityGroupUuid := result.UUID

	if len(plan.L3NetworkUuids.Elements()) > 0 {
		l3NetworkUuids := make([]string, 0, len(plan.L3NetworkUuids.Elements()))
		diags = plan.L3NetworkUuids.ElementsAs(ctx, &l3NetworkUuids, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		for _, l3NetworkUuid := range l3NetworkUuids {
			attachParam := param.AttachSecurityGroupToL3NetworkParam{
				BaseParam: param.BaseParam{},
				Params:    param.AttachSecurityGroupToL3NetworkParamDetail{},
			}
			_, err := r.client.AttachSecurityGroupToL3Network(ctx, securityGroupUuid, l3NetworkUuid, attachParam)
			if err != nil {
				resp.Diagnostics.AddError(
					"Could not attach security group to L3 network in ZSphere", "Error: "+err.Error(),
				)
				return
			}
		}
	}

	if len(plan.VmNicUuids.Elements()) > 0 {
		vmNicUuids := make([]string, 0, len(plan.VmNicUuids.Elements()))
		diags = plan.VmNicUuids.ElementsAs(ctx, &vmNicUuids, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		addNicParam := param.AddVmNicToSecurityGroupParam{
			BaseParam: param.BaseParam{},
			Params: param.AddVmNicToSecurityGroupParamDetail{
				VmNicUuids: vmNicUuids,
			},
		}
		_, err := r.client.AddVmNicToSecurityGroup(ctx, securityGroupUuid, addNicParam)
		if err != nil {
			resp.Diagnostics.AddError(
				"Could not add VM nics to security group in ZSphere", "Error: "+err.Error(),
			)
			return
		}
	}

	updatedResult, err := r.client.GetSecurityGroup(ctx, securityGroupUuid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Could not refresh security group after operations", "Error: "+err.Error(),
		)
		return
	}
	result = updatedResult

	if len(plan.Rules) > 0 {
		err := r.addSecurityGroupRules(ctx, securityGroupUuid, plan.Rules)
		if err != nil {
			resp.Diagnostics.AddError(
				"Could not add rules to security group in ZSphere", "Error: "+err.Error(),
			)
			return
		}

		updatedResult, err = r.client.GetSecurityGroup(ctx, securityGroupUuid)
		if err != nil {
			resp.Diagnostics.AddError(
				"Could not refresh security group after adding rules", "Error: "+err.Error(),
			)
			return
		}
		result = updatedResult

		if len(result.Rules) > 0 {
			plan.Rules = make([]securityGroupRuleModel, len(result.Rules))
			for i, rule := range result.Rules {
				priority := int64(0)
				if rule.Priority > 0 {
					priority = int64(rule.Priority)
				}
				state := rule.State
				if state == "" {
					state = "Enabled"
				}
				plan.Rules[i] = securityGroupRuleModel{
					Type:                    types.StringValue(rule.Type),
					Protocol:                types.StringValue(rule.Protocol),
					Priority:                types.Int64Value(priority),
					Action:                  types.StringValue(rule.Action),
					State:                   types.StringValue(state),
					AllowedCidr:             tfStringFromPtr(rule.AllowedCidr),
					RemoteSecurityGroupUuid: tfStringFromPtr(rule.RemoteSecurityGroupUuid),
					DstPortRange:            tfStringFromPtr(rule.DstPortRange),
					SrcIpRange:              tfStringFromPtr(rule.SrcIpRange),
					DstIpRange:              tfStringFromPtr(rule.DstIpRange),
					Description:             tfStringFromPtr(rule.Description),
				}
			}
		}
	}

	plan.Uuid = types.StringValue(result.UUID)
	plan.Name = types.StringValue(result.Name)
	plan.Description = tfStringFromPtr(result.Description)
	plan.State = types.StringValue(result.State)
	plan.IpVersion = types.Int64Value(int64(result.IpVersion))

	if len(result.AttachedL3NetworkUuids) > 0 {
		plan.AttachedL3NetworkUuids, _ = types.ListValueFrom(ctx, types.StringType, result.AttachedL3NetworkUuids)
	} else {
		plan.AttachedL3NetworkUuids = types.ListNull(types.StringType)
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *securityGroupResource) addSecurityGroupRules(ctx context.Context, securityGroupUuid string, rules []securityGroupRuleModel) error {
	sgRules := make([]param.AddSecurityGroupRule_SecurityGroupRuleAOParam, 0, len(rules))

	for _, rule := range rules {
		sgRule := param.AddSecurityGroupRule_SecurityGroupRuleAOParam{}

		if !rule.Type.IsNull() {
			sgRule.Type = rule.Type.ValueString()
		}
		if !rule.Protocol.IsNull() {
			protocol := rule.Protocol.ValueString()
			sgRule.Protocol = &protocol
		}
		if !rule.Action.IsNull() {
			action := rule.Action.ValueString()
			sgRule.Action = &action
		}
		if !rule.State.IsNull() {
			state := rule.State.ValueString()
			sgRule.State = &state
		}
		if !rule.AllowedCidr.IsNull() {
			allowedCidr := rule.AllowedCidr.ValueString()
			sgRule.AllowedCidr = &allowedCidr
		}
		if !rule.RemoteSecurityGroupUuid.IsNull() {
			remoteSG := rule.RemoteSecurityGroupUuid.ValueString()
			sgRule.RemoteSecurityGroupUuid = &remoteSG
		}
		if !rule.DstPortRange.IsNull() {
			dstPortRange := rule.DstPortRange.ValueString()
			sgRule.DstPortRange = &dstPortRange
		}
		if !rule.SrcIpRange.IsNull() {
			srcIpRange := rule.SrcIpRange.ValueString()
			sgRule.SrcIpRange = &srcIpRange
		}
		if !rule.DstIpRange.IsNull() {
			dstIpRange := rule.DstIpRange.ValueString()
			sgRule.DstIpRange = &dstIpRange
		}
		if !rule.Description.IsNull() {
			desc := rule.Description.ValueString()
			sgRule.Description = &desc
		}

		sgRules = append(sgRules, sgRule)
	}

	addRuleParam := param.AddSecurityGroupRuleParam{
		BaseParam: param.BaseParam{},
		Params: param.AddSecurityGroupRuleParamDetail{
			Rules: sgRules,
		},
	}

	_, err := r.client.AddSecurityGroupRule(ctx, securityGroupUuid, addRuleParam)
	return err
}

func (r *securityGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state securityGroupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.GetSecurityGroup(ctx, state.Uuid.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Could not read security group in ZSphere", "Error: "+err.Error(),
		)
		return
	}

	state.Uuid = types.StringValue(result.UUID)
	state.Name = types.StringValue(result.Name)
	state.Description = tfStringFromPtr(result.Description)
	state.State = types.StringValue(result.State)
	state.IpVersion = types.Int64Value(int64(result.IpVersion))

	if len(result.AttachedL3NetworkUuids) > 0 {
		state.AttachedL3NetworkUuids, _ = types.ListValueFrom(ctx, types.StringType, result.AttachedL3NetworkUuids)
	} else {
		state.AttachedL3NetworkUuids = types.ListNull(types.StringType)
	}

	if len(result.Rules) > 0 {
		state.Rules = make([]securityGroupRuleModel, len(result.Rules))
		for i, rule := range result.Rules {
			priority := int64(0)
			if rule.Priority > 0 {
				priority = int64(rule.Priority)
			}
			ruleState := rule.State
			if ruleState == "" {
				ruleState = "Enabled"
			}
			state.Rules[i] = securityGroupRuleModel{
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
	} else {
		state.Rules = []securityGroupRuleModel{}
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *securityGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan securityGroupResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	description := stringPtrFromTfString(plan.Description)

	updateParam := param.UpdateSecurityGroupParam{
		BaseParam: param.BaseParam{},
		Params: param.UpdateSecurityGroupParamDetail{
			Name:        plan.Name.ValueString(),
			Description: description,
		},
	}

	_, err := r.client.UpdateSecurityGroup(ctx, plan.Uuid.ValueString(), updateParam)
	if err != nil {
		resp.Diagnostics.AddError(
			"Could not update security group in ZSphere", "Error: "+err.Error(),
		)
		return
	}

	result, err := r.client.GetSecurityGroup(ctx, plan.Uuid.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Could not refresh security group after update", "Error: "+err.Error(),
		)
		return
	}

	plan.Uuid = types.StringValue(result.UUID)
	plan.Name = types.StringValue(result.Name)
	plan.Description = tfStringFromPtr(result.Description)
	plan.State = types.StringValue(result.State)
	plan.IpVersion = types.Int64Value(int64(result.IpVersion))

	if len(result.AttachedL3NetworkUuids) > 0 {
		plan.AttachedL3NetworkUuids, _ = types.ListValueFrom(ctx, types.StringType, result.AttachedL3NetworkUuids)
	} else {
		plan.AttachedL3NetworkUuids = types.ListNull(types.StringType)
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *securityGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state securityGroupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteSecurityGroup(ctx, state.Uuid.ValueString(), param.DeleteModeEnforcing)
	if err != nil {
		resp.Diagnostics.AddError(
			"Could not delete security group in ZSphere", "Error: "+err.Error(),
		)
		return
	}
}

func (r *securityGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_group"
}

func (r *securityGroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Resource definition for ZSphere Security Group.",
		Attributes: map[string]schema.Attribute{
			"uuid": schema.StringAttribute{
				Computed:    true,
				Description: "UUID of the security group.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the security group.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
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
			"l3_network_uuids": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "UUIDs of L3 networks to attach the security group to.",
			},
			"vm_nic_uuids": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "UUIDs of VM nics to bind to the security group.",
			},
			"attached_l3_network_uuids": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "UUIDs of L3 networks that the security group is attached to.",
			},
		},
		Blocks: map[string]schema.Block{
			"rules": schema.ListNestedBlock{
				Description: "Security group rules.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Optional:    true,
							Description: "Rule type: Ingress or Egress.",
							Validators: []validator.String{
								stringvalidator.OneOf("Ingress", "Egress"),
							},
						},
						"protocol": schema.StringAttribute{
							Optional:    true,
							Description: "Protocol: TCP, UDP, ICMP, or ALL.",
							Validators: []validator.String{
								stringvalidator.OneOf("TCP", "UDP", "ICMP", "ALL"),
							},
						},
						"priority": schema.Int64Attribute{
							Optional:    true,
							Computed:    true,
							Description: "Rule priority.",
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.UseStateForUnknown(),
							},
							Validators: []validator.Int64{
								int64validator.AtLeast(1),
							},
						},
						"action": schema.StringAttribute{
							Optional:    true,
							Description: "Rule action: ACCEPT or DROP.",
							Validators: []validator.String{
								stringvalidator.OneOf("ACCEPT", "DROP"),
							},
						},
						"state": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Rule state: Enabled or Disabled.",
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"allowed_cidr": schema.StringAttribute{
							Optional:    true,
							Description: "Allowed CIDR block.",
						},
						"remote_security_group_uuid": schema.StringAttribute{
							Optional:    true,
							Description: "Remote security group UUID.",
						},
						"dst_port_range": schema.StringAttribute{
							Optional:    true,
							Description: "Destination port range (e.g., '22' or '80-443').",
						},
						"src_ip_range": schema.StringAttribute{
							Optional:    true,
							Description: "Source IP range.",
						},
						"dst_ip_range": schema.StringAttribute{
							Optional:    true,
							Description: "Destination IP range.",
						},
						"description": schema.StringAttribute{
							Optional:    true,
							Description: "Description of the rule.",
						},
					},
				},
			},
		},
	}
}
