// Copyright (c) ZStack.io, Inc.

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
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
	_ resource.Resource              = &clusterResource{}
	_ resource.ResourceWithConfigure = &clusterResource{}
)

type clusterResource struct {
	client *client.ZSClient
}

type clusterResourceModel struct {
	Uuid                     types.String `tfsdk:"uuid"`
	Name                     types.String `tfsdk:"name"`
	Description              types.String `tfsdk:"description"`
	State                    types.String `tfsdk:"state"`
	HypervisorType           types.String `tfsdk:"hypervisor_type"`
	Type                     types.String `tfsdk:"type"`
	DatacenterUuid           types.String `tfsdk:"datacenter_uuid"`
	Architecture             types.String `tfsdk:"architecture"`
	DisplayNetworkCidr       types.String `tfsdk:"display_network_cidr"`
	MigrateNetworkCidr       types.String `tfsdk:"migrate_network_cidr"`
	CheckCpuModel            types.String `tfsdk:"check_cpu_model"`
	CpuMode                  types.String `tfsdk:"cpu_mode"`
	NetworkHp                types.Bool   `tfsdk:"network_hp"`
	AutomationLevel          types.String `tfsdk:"automation_level"`
	CpuOverProvisioningRatio types.String `tfsdk:"cpu_over_provisioning_ratio"`
	VmHaLevel                types.String `tfsdk:"vm_ha_level"`
	DrsEnabled               types.Bool   `tfsdk:"drs_enabled"`
	DrsThresholdCpu          types.Int64  `tfsdk:"drs_threshold_cpu"`
	DrsThresholdMemory       types.Int64  `tfsdk:"drs_threshold_memory"`
	DrsThresholdDuration     types.Int64  `tfsdk:"drs_threshold_duration"`
}

func (r *clusterResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func ClusterResource() resource.Resource {
	return &clusterResource{}
}

func strPtr(s string) *string {
	return &s
}

func (r *clusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan clusterResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var description *string
	if plan.Description.IsNull() {
		description = nil
	} else {
		descriptionStr := plan.Description.ValueString()
		description = &descriptionStr
	}

	var clusterType *string
	if plan.Type.IsNull() {
		clusterType = nil
	} else {
		clusterTypeStr := plan.Type.ValueString()
		clusterType = &clusterTypeStr
	}

	var architecture *string
	if plan.Architecture.IsNull() {
		architecture = nil
	} else {
		architectureStr := plan.Architecture.ValueString()
		architecture = &architectureStr
	}

	hypervisorType := "KVM"
	if !plan.HypervisorType.IsNull() && plan.HypervisorType.ValueString() != "" {
		hypervisorType = plan.HypervisorType.ValueString()
	}

	systemTags := []string{}
	if !plan.DisplayNetworkCidr.IsNull() && plan.DisplayNetworkCidr.ValueString() != "" {
		systemTags = append(systemTags, fmt.Sprintf("display::network::cidr::%s", plan.DisplayNetworkCidr.ValueString()))
	}
	if !plan.MigrateNetworkCidr.IsNull() && plan.MigrateNetworkCidr.ValueString() != "" {
		systemTags = append(systemTags, fmt.Sprintf("cluster::migrate::network::cidr::%s", plan.MigrateNetworkCidr.ValueString()))
	}
	if !plan.CheckCpuModel.IsNull() && plan.CheckCpuModel.ValueString() != "" && plan.CheckCpuModel.ValueString() != "default" {
		systemTags = append(systemTags, fmt.Sprintf("check::cluster::cpu::model::%s", plan.CheckCpuModel.ValueString()))
	}
	if !plan.CpuOverProvisioningRatio.IsNull() && plan.CpuOverProvisioningRatio.ValueString() != "" {
		systemTags = append(systemTags, fmt.Sprintf("resourceConfig::kvm::cpu.overProvisioning.ratio::%s", plan.CpuOverProvisioningRatio.ValueString()))
	}
	if !plan.VmHaLevel.IsNull() && plan.VmHaLevel.ValueString() != "" {
		systemTags = append(systemTags, fmt.Sprintf("resourceConfig::ha::vm.ha.level::%s", plan.VmHaLevel.ValueString()))
	}
	if !plan.AutomationLevel.IsNull() && plan.AutomationLevel.ValueString() != "" {
		systemTags = append(systemTags, fmt.Sprintf("cluster::automationLevel::%s", plan.AutomationLevel.ValueString()))
	}

	clusterParam := param.CreateClusterParam{
		BaseParam: param.BaseParam{
			SystemTags: systemTags,
		},
		Params: param.CreateClusterParamDetail{
			ZoneUuid:       plan.DatacenterUuid.ValueString(),
			Name:           plan.Name.ValueString(),
			Description:    description,
			HypervisorType: hypervisorType,
			Type:           clusterType,
			Architecture:   architecture,
		},
	}

	cluster, err := r.client.CreateCluster(ctx, clusterParam)
	if err != nil {
		resp.Diagnostics.AddError(
			"Could not create cluster in ZSphere", "Error: "+err.Error(),
		)
		return
	}

	clusterUuid := cluster.UUID

	if !plan.CpuMode.IsNull() && plan.CpuMode.ValueString() != "" {
		cpuModeResourceConfig := param.UpdateResourceConfigsParam{
			Params: param.UpdateResourceConfigsParamDetail{
				ResourceConfigs: []param.UpdateResourceConfigs_ResourceConfigAOParam{
					{
						Category: strPtr("kvm"),
						Name:     "cpuMode",
						Value:    strPtr(plan.CpuMode.ValueString()),
					},
				},
			},
		}
		_, err := r.client.UpdateResourceConfigs(ctx, clusterUuid, cpuModeResourceConfig)
		if err != nil {
			tflog.Warn(ctx, fmt.Sprintf("Failed to update cpuMode for cluster %s: %s", clusterUuid, err.Error()))
		}
	}

	if !plan.NetworkHp.IsNull() && plan.NetworkHp.ValueBool() {
		networkHpResourceConfig := param.UpdateResourceConfigsParam{
			Params: param.UpdateResourceConfigsParamDetail{
				ResourceConfigs: []param.UpdateResourceConfigs_ResourceConfigAOParam{
					{
						Category: strPtr("kvm"),
						Name:     "ovsDbEnable",
						Value:    strPtr("true"),
					},
				},
			},
		}
		_, err := r.client.UpdateResourceConfigs(ctx, clusterUuid, networkHpResourceConfig)
		if err != nil {
			tflog.Warn(ctx, fmt.Sprintf("Failed to update networkHp for cluster %s: %s", clusterUuid, err.Error()))
		}
	}

	if !plan.DrsEnabled.IsNull() && plan.DrsEnabled.ValueBool() && plan.HypervisorType.ValueString() == "KVM" {
		drsThresholdCpu := int64(80)
		if !plan.DrsThresholdCpu.IsNull() {
			drsThresholdCpu = plan.DrsThresholdCpu.ValueInt64()
		}
		drsThresholdMemory := int64(80)
		if !plan.DrsThresholdMemory.IsNull() {
			drsThresholdMemory = plan.DrsThresholdMemory.ValueInt64()
		}
		drsThresholdDuration := 300
		if !plan.DrsThresholdDuration.IsNull() {
			drsThresholdDuration = int(plan.DrsThresholdDuration.ValueInt64())
		}

		cpuThreshold := fmt.Sprintf("%d", drsThresholdCpu)
		memThreshold := fmt.Sprintf("%d", drsThresholdMemory)

		drsParam := param.CreateClusterDRSParam{
			Params: param.CreateClusterDRSParamDetail{
				Name:            fmt.Sprintf("DRS-%s", clusterUuid),
				AutomationLevel: plan.AutomationLevel.ValueString(),
				Thresholds: []param.ThresholdParam{
					{
						ThresholdName:  strPtr("CPU"),
						ThresholdValue: &cpuThreshold,
						Operator:       strPtr("greater_than"),
					},
					{
						ThresholdName:  strPtr("Memory"),
						ThresholdValue: &memThreshold,
						Operator:       strPtr("greater_than"),
					},
				},
				ThresholdDuration: drsThresholdDuration,
			},
		}

		_, err := r.client.CreateClusterDRS(ctx, clusterUuid, drsParam)
		if err != nil {
			tflog.Warn(ctx, fmt.Sprintf("Failed to create DRS for cluster %s: %s", clusterUuid, err.Error()))
		}
	}

	plan.Uuid = types.StringValue(cluster.UUID)
	plan.Name = types.StringValue(cluster.Name)
	if cluster.Description != "" {
		plan.Description = types.StringValue(cluster.Description)
	} else {
		plan.Description = types.StringNull()
	}
	plan.State = types.StringValue(cluster.State)
	plan.HypervisorType = types.StringValue(cluster.HypervisorType)
	plan.Type = types.StringValue(cluster.Type)
	plan.DatacenterUuid = types.StringValue(cluster.ZoneUuid)
	plan.Architecture = types.StringValue(cluster.Architecture)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *clusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state clusterResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.Uuid.IsNull() || state.Uuid.ValueString() == "" {
		tflog.Warn(ctx, "cluster uuid is empty, so nothing to delete, skip it")
		return
	}

	uuid := state.Uuid.ValueString()

	tflog.Info(ctx, fmt.Sprintf("delete cluster %s", uuid))
	err := r.client.DeleteCluster(ctx, uuid, param.DeleteModeEnforcing)
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete cluster", err.Error())
		return
	}
}

func (r *clusterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster"
}

func (r *clusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state clusterResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	cluster, err := r.client.GetCluster(ctx, state.Uuid.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting ZStack cluster", "Could not read cluster: "+err.Error(),
		)
		return
	}

	state.Uuid = types.StringValue(cluster.UUID)
	state.Name = types.StringValue(cluster.Name)
	if cluster.Description != "" {
		state.Description = types.StringValue(cluster.Description)
	} else {
		state.Description = types.StringNull()
	}
	state.State = types.StringValue(cluster.State)
	state.HypervisorType = types.StringValue(cluster.HypervisorType)
	state.Type = types.StringValue(cluster.Type)
	state.DatacenterUuid = types.StringValue(cluster.ZoneUuid)
	state.Architecture = types.StringValue(cluster.Architecture)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *clusterResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This resource allows you to manage clusters in ZSphere. " +
			"A cluster is a logical grouping of hosts that share the same hypervisor type and storage. " +
			"It provides resource isolation and is the basic unit for resource management in ZSphere.",
		Attributes: map[string]schema.Attribute{
			"uuid": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the cluster. Automatically generated by ZSphere.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the cluster. This is a mandatory field.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "A description of the cluster, providing additional context or details.",
			},
			"state": schema.StringAttribute{
				Computed:    true,
				Description: "The state of the cluster, such as 'Enabled' or 'Disabled'.",
			},
			"hypervisor_type": schema.StringAttribute{
				Required:    true,
				Description: "The hypervisor type of the cluster. Valid values are: KVM, Simulator, baremetal, baremetal2, xdragon.",
				Validators: []validator.String{
					stringvalidator.OneOf("KVM", "Simulator", "baremetal", "baremetal2", "xdragon"),
				},
			},
			"type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The type of the cluster. Valid values are: zstack, baremetal, baremetal2.",
				Validators: []validator.String{
					stringvalidator.OneOf("zstack", "baremetal", "baremetal2"),
				},
			},
			"datacenter_uuid": schema.StringAttribute{
				Required:    true,
				Description: "The UUID of the datacenter (zone) to which the cluster belongs.",
			},
			"architecture": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The CPU architecture of the cluster. Valid values are: x86_64, aarch64, mips64el, loongarch64.",
				Validators: []validator.String{
					stringvalidator.OneOf("x86_64", "aarch64", "mips64el", "loongarch64"),
				},
			},
			"display_network_cidr": schema.StringAttribute{
				Optional:    true,
				Description: "Display network CIDR. Example: 192.168.1.0/24",
			},
			"migrate_network_cidr": schema.StringAttribute{
				Optional:    true,
				Description: "Migration network CIDR. Example: 192.168.2.0/24",
			},
			"check_cpu_model": schema.StringAttribute{
				Optional:    true,
				Description: "CPU model validation. Valid values: default, false, true",
			},
			"cpu_mode": schema.StringAttribute{
				Optional:    true,
				Description: "CPU mode. Valid values: hostPassthrough, virtio, kvm",
			},
			"network_hp": schema.BoolAttribute{
				Optional:    true,
				Description: "Enable network HP (ovsdpdk).",
			},
			"automation_level": schema.StringAttribute{
				Optional:    true,
				Description: "Automation level for DRS. Valid values: Manual, Automatic, closed",
			},
			"cpu_over_provisioning_ratio": schema.StringAttribute{
				Optional:    true,
				Description: "CPU over-provisioning ratio. Example: 4",
			},
			"vm_ha_level": schema.StringAttribute{
				Optional:    true,
				Description: "VM HA level. Valid values: NeverStop, BestEffort",
			},
			"drs_enabled": schema.BoolAttribute{
				Optional:    true,
				Description: "Enable Dynamic Resource Scheduling (DRS).",
			},
			"drs_threshold_cpu": schema.Int64Attribute{
				Optional:    true,
				Description: "DRS CPU threshold percentage. Example: 80",
			},
			"drs_threshold_memory": schema.Int64Attribute{
				Optional:    true,
				Description: "DRS memory threshold percentage. Example: 80",
			},
			"drs_threshold_duration": schema.Int64Attribute{
				Optional:    true,
				Description: "DRS threshold duration in seconds. Example: 300",
			},
		},
	}
}

func (r *clusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state clusterResourceModel

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
	plan.DatacenterUuid = state.DatacenterUuid
	plan.HypervisorType = state.HypervisorType
	plan.Architecture = state.Architecture

	var description *string
	if plan.Description.IsNull() {
		description = nil
	} else if plan.Description.ValueString() == "" {
		emptyStr := ""
		description = &emptyStr
	} else {
		descriptionStr := plan.Description.ValueString()
		description = &descriptionStr
	}

	updateParam := param.UpdateClusterParam{
		Params: param.UpdateClusterParamDetail{
			Name:        plan.Name.ValueString(),
			Description: description,
		},
	}

	cluster, err := r.client.UpdateCluster(ctx, plan.Uuid.ValueString(), updateParam)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating ZStack cluster", "Could not update cluster: "+err.Error(),
		)
		return
	}

	plan.Name = types.StringValue(cluster.Name)
	if cluster.Description != "" {
		plan.Description = types.StringValue(cluster.Description)
	} else {
		plan.Description = types.StringNull()
	}
	plan.State = types.StringValue(cluster.State)
	plan.Type = types.StringValue(cluster.Type)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
