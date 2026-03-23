// Copyright (c) ZStack.io, Inc.

package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/terraform-zstack-modules/zsphere-sdk-go/pkg/client"
	"github.com/terraform-zstack-modules/zsphere-sdk-go/pkg/param"
	"github.com/terraform-zstack-modules/zsphere-sdk-go/pkg/view"
)

var (
	_ resource.Resource              = &distributedSwitchResource{}
	_ resource.ResourceWithConfigure = &distributedSwitchResource{}
)

type distributedSwitchResource struct {
	client *client.ZSClient
}

type HostParam struct {
	HostUuid          string `tfsdk:"host_uuid" json:"hostUuid"`
	PhysicalInterface string `tfsdk:"physical_interface" json:"physicalInterface"`
}

type ClusterAttach struct {
	ClusterUuid string      `tfsdk:"cluster_uuid"`
	HostParams  []HostParam `tfsdk:"host_params"`
}

type distributedSwitchResourceModel struct {
	Uuid                 types.String    `tfsdk:"uuid"`
	Name                 types.String    `tfsdk:"name"`
	Description          types.String    `tfsdk:"description"`
	DatacenterUuid       types.String    `tfsdk:"datacenter_uuid"`
	PhysicalInterface    types.String    `tfsdk:"physical_interface"`
	VSwitchType          types.String    `tfsdk:"vswitch_type"`
	Type                 types.String    `tfsdk:"type"`
	BondingName          types.String    `tfsdk:"bonding_name"`
	BondingMode          types.String    `tfsdk:"bonding_mode"`
	XmitHashPolicy       types.String    `tfsdk:"xmit_hash_policy"`
	ClusterUuids         types.List      `tfsdk:"cluster_uuids"`
	ClusterAttachments   []ClusterAttach `tfsdk:"cluster_attachments"`
	AttachedClusterUuids types.List      `tfsdk:"attached_cluster_uuids"`
}

func (r *distributedSwitchResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_distributed_switch"
}

func (r *distributedSwitchResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = cli
}

func DistributedSwitchResource() resource.Resource {
	return &distributedSwitchResource{}
}

func (r *distributedSwitchResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan distributedSwitchResourceModel
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

	vSwitchType := "LinuxBridge"
	if !plan.VSwitchType.IsNull() && plan.VSwitchType.ValueString() != "" {
		vSwitchType = plan.VSwitchType.ValueString()
	}

	clusterUuids := make([]string, 0)
	if !plan.ClusterUuids.IsNull() && !plan.ClusterUuids.IsUnknown() {
		diags = plan.ClusterUuids.ElementsAs(ctx, &clusterUuids, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	hostInterfaceList := make([]string, 0)
	if len(plan.ClusterAttachments) > 0 {
		for _, attach := range plan.ClusterAttachments {
			for _, hp := range attach.HostParams {
				hostInterfaceList = append(hostInterfaceList, hp.PhysicalInterface)
			}
		}
	}

	physicalInterface := ""
	systemTags := []string{}
	createBondPayloads := []map[string]interface{}{}

	bondingName := ""
	if !plan.BondingName.IsNull() && plan.BondingName.ValueString() != "" {
		bondingName = plan.BondingName.ValueString()
	}

	bondingMode := "balance-xor"
	if !plan.BondingMode.IsNull() && plan.BondingMode.ValueString() != "" {
		bondingMode = plan.BondingMode.ValueString()
	}

	xmitHashPolicy := ""
	if !plan.XmitHashPolicy.IsNull() && plan.XmitHashPolicy.ValueString() != "" {
		xmitHashPolicy = plan.XmitHashPolicy.ValueString()
	}

	if bondingName != "" {
		physicalInterface = bondingName

		if bondingMode == "802.3ad" && xmitHashPolicy != "" {
			systemTags = append(systemTags, fmt.Sprintf("uplink::bonding::%s::%s", bondingMode, xmitHashPolicy))
		} else if bondingMode == "802.3ad" {
			systemTags = append(systemTags, fmt.Sprintf("uplink::bonding::%s::", bondingMode))
		} else {
			systemTags = append(systemTags, fmt.Sprintf("uplink::bonding::%s::null", bondingMode))
		}

		if len(hostInterfaceList) > 0 {
			createBondPayloads = append(createBondPayloads, map[string]interface{}{
				"bondingName": bondingName,
				"slaveNames":  hostInterfaceList,
				"hostUuids":   []string{},
				"mode":        bondingMode,
				"xmitHashPolicy": func() string {
					if bondingMode == "802.3ad" && xmitHashPolicy != "" {
						return xmitHashPolicy
					}
					return ""
				}(),
			})
		}
	} else if len(hostInterfaceList) > 0 {
		physicalInterface = hostInterfaceList[0]
	} else if !plan.PhysicalInterface.IsNull() && plan.PhysicalInterface.ValueString() != "" {
		physicalInterface = plan.PhysicalInterface.ValueString()
	}

	createParam := struct {
		Name              string  `json:"name"`
		Description       *string `json:"description,omitempty"`
		ZoneUuid          string  `json:"zoneUuid"`
		PhysicalInterface string  `json:"physicalInterface,omitempty"`
		VSwitchType       string  `json:"vSwitchType"`
		IsDistributed     bool    `json:"isDistributed"`
	}{
		Name:              plan.Name.ValueString(),
		Description:       description,
		ZoneUuid:          plan.DatacenterUuid.ValueString(),
		PhysicalInterface: physicalInterface,
		VSwitchType:       vSwitchType,
		IsDistributed:     true,
	}

	tflog.Debug(ctx, "Creating distributed switch", map[string]interface{}{
		"name":     plan.Name.ValueString(),
		"zoneUuid": plan.DatacenterUuid.ValueString(),
	})

	var result view.L2NetworkInventoryView
	err := r.client.ZSHttpClient.Post(ctx, "v1/l2-networks/virtual-switch", map[string]interface{}{
		"params":     createParam,
		"systemTags": systemTags,
	}, &result)
	if err != nil {
		resp.Diagnostics.AddError(
			"Could not create distributed switch in ZSphere",
			"Error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Distributed switch created, now attaching to clusters", map[string]interface{}{
		"uuid":     result.UUID,
		"clusters": clusterUuids,
	})

	attachL2NetworkToClusterHostParams := make(map[string]string)
	if len(plan.ClusterAttachments) > 0 {
		for _, attach := range plan.ClusterAttachments {
			hostParamsJSON, err := json.Marshal(attach.HostParams)
			if err != nil {
				tflog.Warn(ctx, "Failed to marshal host params", map[string]interface{}{
					"error": err.Error(),
				})
				continue
			}
			attachL2NetworkToClusterHostParams[attach.ClusterUuid] = string(hostParamsJSON)
		}
	}

	for _, clusterUuid := range clusterUuids {
		hostParamsStr := attachL2NetworkToClusterHostParams[clusterUuid]

		attachParam := param.AttachL2NetworkToClusterParam{
			BaseParam: param.BaseParam{},
			Params: param.AttachL2NetworkToClusterParamDetail{
				HostParams: func() *string {
					if hostParamsStr != "" {
						return &hostParamsStr
					}
					return nil
				}(),
			},
		}

		tflog.Debug(ctx, "Attaching cluster to distributed switch", map[string]interface{}{
			"clusterUuid": clusterUuid,
			"hostParams":  hostParamsStr,
		})

		_, err = r.client.AttachL2NetworkToCluster(ctx, result.UUID, clusterUuid, attachParam)
		if err != nil {
			tflog.Warn(ctx, "Failed to attach cluster to distributed switch", map[string]interface{}{
				"clusterUuid": clusterUuid,
				"error":       err.Error(),
			})
		} else {
			tflog.Debug(ctx, "Successfully attached cluster", map[string]interface{}{
				"clusterUuid": clusterUuid,
			})
		}
	}

	plan.Uuid = types.StringValue(result.UUID)
	plan.Name = types.StringValue(result.Name)
	plan.Description = types.StringValue(result.Description)
	plan.DatacenterUuid = types.StringValue(result.ZoneUuid)
	if plan.PhysicalInterface.IsNull() || plan.PhysicalInterface.ValueString() == "" {
		plan.PhysicalInterface = types.StringValue(result.PhysicalInterface)
	}
	plan.VSwitchType = types.StringValue(result.VSwitchType)
	plan.Type = types.StringValue(result.Type)

	if len(result.AttachedClusterUuids) > 0 {
		plan.AttachedClusterUuids, _ = types.ListValueFrom(ctx, types.StringType, result.AttachedClusterUuids)
	} else {
		plan.AttachedClusterUuids = types.ListNull(types.StringType)
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *distributedSwitchResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state distributedSwitchResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result view.L2NetworkInventoryView
	err := r.client.ZSHttpClient.Get(ctx, "v1/l2-networks", state.Uuid.ValueString(), nil, &result)
	if err != nil {
		resp.Diagnostics.AddError(
			"Could not read distributed switch from ZSphere",
			"Error: "+err.Error(),
		)
		return
	}

	state.Name = types.StringValue(result.Name)
	state.Description = types.StringValue(result.Description)
	state.DatacenterUuid = types.StringValue(result.ZoneUuid)
	state.PhysicalInterface = types.StringValue(result.PhysicalInterface)
	state.VSwitchType = types.StringValue(result.VSwitchType)
	state.Type = types.StringValue(result.Type)

	if len(result.AttachedClusterUuids) > 0 {
		state.AttachedClusterUuids, _ = types.ListValueFrom(ctx, types.StringType, result.AttachedClusterUuids)
	} else {
		state.AttachedClusterUuids = types.ListNull(types.StringType)
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *distributedSwitchResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan distributedSwitchResourceModel
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

	updateParam := param.UpdateL2NetworkParam{
		BaseParam: param.BaseParam{},
		Params: param.UpdateL2NetworkParamDetail{
			Name:        plan.Name.ValueString(),
			Description: description,
		},
	}

	_, err := r.client.UpdateL2Network(ctx, plan.Uuid.ValueString(), updateParam)
	if err != nil {
		resp.Diagnostics.AddError(
			"Could not update distributed switch in ZSphere",
			"Error: "+err.Error(),
		)
		return
	}

	var state distributedSwitchResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.Name = plan.Name
	state.Description = plan.Description

	var result view.L2NetworkInventoryView
	err = r.client.ZSHttpClient.Get(ctx, "v1/l2-networks", state.Uuid.ValueString(), nil, &result)
	if err != nil {
		resp.Diagnostics.AddError(
			"Could not read distributed switch from ZSphere",
			"Error: "+err.Error(),
		)
		return
	}

	state.Name = types.StringValue(result.Name)
	state.Description = types.StringValue(result.Description)
	state.DatacenterUuid = types.StringValue(result.ZoneUuid)
	state.PhysicalInterface = types.StringValue(result.PhysicalInterface)
	state.VSwitchType = types.StringValue(result.VSwitchType)
	state.Type = types.StringValue(result.Type)

	if len(result.AttachedClusterUuids) > 0 {
		state.AttachedClusterUuids, _ = types.ListValueFrom(ctx, types.StringType, result.AttachedClusterUuids)
	} else {
		state.AttachedClusterUuids = types.ListNull(types.StringType)
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *distributedSwitchResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state distributedSwitchResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteL2Network(ctx, state.Uuid.ValueString(), param.DeleteModeEnforcing)
	if err != nil {
		resp.Diagnostics.AddError(
			"Could not delete distributed switch from ZSphere",
			"Error: "+err.Error(),
		)
		return
	}
}

func (r *distributedSwitchResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a distributed switch (L2VirtualSwitch) in ZSphere.",
		Attributes: map[string]schema.Attribute{
			"uuid": schema.StringAttribute{
				Computed:    true,
				Description: "The UUID of the distributed switch.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the distributed switch.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Description of the distributed switch.",
			},
			"datacenter_uuid": schema.StringAttribute{
				Required:    true,
				Description: "The UUID of the datacenter (zone) to which the distributed switch belongs.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"physical_interface": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The physical interface used by the distributed switch (used when not using bonding).",
			},
			"vswitch_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The type of virtual switch. Default is LinuxBridge.",
				Validators: []validator.String{
					stringvalidator.OneOf("LinuxBridge", "OvsDpdk", "MacVlan"),
				},
			},
			"type": schema.StringAttribute{
				Computed:    true,
				Description: "The type of the L2 network.",
			},
			"bonding_name": schema.StringAttribute{
				Optional:    true,
				Description: "The name of the bonding aggregate port (e.g., Uplink1).",
			},
			"bonding_mode": schema.StringAttribute{
				Optional:    true,
				Description: "The bonding mode. Valid values: 802.3ad, active-backup.",
				Validators: []validator.String{
					stringvalidator.OneOf("802.3ad", "active-backup"),
				},
			},
			"xmit_hash_policy": schema.StringAttribute{
				Optional:    true,
				Description: "The transmit hash policy for bonding mode 802.3ad. Valid values: layer2, layer2+3, layer3+4.",
				Validators: []validator.String{
					stringvalidator.OneOf("layer2", "layer2+3", "layer3+4"),
				},
			},
			"cluster_uuids": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "List of cluster UUIDs to attach the distributed switch to.",
			},
			"cluster_attachments": schema.ListNestedAttribute{
				Optional:    true,
				Description: "Cluster attachments with host parameters. Used to specify which host uses which physical interface for each cluster.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cluster_uuid": schema.StringAttribute{
							Required:    true,
							Description: "The UUID of the cluster to attach.",
						},
						"host_params": schema.ListNestedAttribute{
							Optional:    true,
							Description: "Host parameters for the attachment.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"host_uuid": schema.StringAttribute{
										Required:    true,
										Description: "The UUID of the host.",
									},
									"physical_interface": schema.StringAttribute{
										Required:    true,
										Description: "The physical interface name on the host.",
									},
								},
							},
						},
					},
				},
			},
			"attached_cluster_uuids": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "List of cluster UUIDs attached to this distributed switch.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}
