// Copyright (c) ZStack.io, Inc.

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/terraform-zstack-modules/zsphere-sdk-go/pkg/client"
	"github.com/terraform-zstack-modules/zsphere-sdk-go/pkg/param"
	"github.com/terraform-zstack-modules/zsphere-sdk-go/pkg/view"
)

var (
	_ resource.Resource              = &hostResource{}
	_ resource.ResourceWithConfigure = &hostResource{}
)

type hostResource struct {
	client *client.ZSClient
}

type hostResourceModel struct {
	Uuid           types.String `tfsdk:"uuid"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	ManagementIp   types.String `tfsdk:"management_ip"`
	ClusterUuid    types.String `tfsdk:"cluster_uuid"`
	ZoneUuid       types.String `tfsdk:"zone_uuid"`
	State          types.String `tfsdk:"state"`
	Status         types.String `tfsdk:"status"`
	HypervisorType types.String `tfsdk:"hypervisor_type"`
	Architecture   types.String `tfsdk:"architecture"`
	Username       types.String `tfsdk:"username"`
	Password       types.String `tfsdk:"password"`
	SshPort        types.Int64  `tfsdk:"ssh_port"`
	Type           types.String `tfsdk:"type"`
	Iommu          types.Bool   `tfsdk:"iommu"`
	Ept            types.Bool   `tfsdk:"ept"`
}

func (r *hostResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.ZSClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.ZSClient, got: %T. Please report this issue to the Provider developer. ", req.ProviderData),
		)
		return
	}

	r.client = client
}

func HostResource() resource.Resource {
	return &hostResource{}
}

func (r *hostResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan hostResourceModel
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

	var sshPort *int
	if plan.SshPort.IsNull() {
		sshPort = nil
	} else {
		sshPortInt := int(plan.SshPort.ValueInt64())
		sshPort = &sshPortInt
	}

	systemTags := []string{}
	if !plan.Iommu.IsNull() && plan.Iommu.ValueBool() {
		systemTags = append(systemTags, "iommuState::Enabled")
	}
	if !plan.Ept.IsNull() && !plan.Ept.ValueBool() {
		systemTags = append(systemTags, "pageTableExtensionDisabled")
	}

	hostParam := struct {
		Name         string   `json:"name"`
		Description  *string  `json:"description,omitempty"`
		ManagementIp string   `json:"managementIp"`
		ClusterUuid  string   `json:"clusterUuid"`
		Username     string   `json:"username"`
		Password     string   `json:"password"`
		SshPort      *int     `json:"sshPort,omitempty"`
		SystemTags   []string `json:"systemTags,omitempty"`
	}{
		Name:         plan.Name.ValueString(),
		Description:  description,
		ManagementIp: plan.ManagementIp.ValueString(),
		ClusterUuid:  plan.ClusterUuid.ValueString(),
		Username:     plan.Username.ValueString(),
		Password:     plan.Password.ValueString(),
		SshPort:      sshPort,
		SystemTags:   systemTags,
	}

	var result view.HostInventoryView
	err := r.client.ZSHttpClient.Post(ctx, "v1/hosts/kvm", map[string]interface{}{
		"params": hostParam,
	}, &result)
	if err != nil {
		resp.Diagnostics.AddError(
			"Could not create host in ZSphere", "Error: "+err.Error(),
		)
		return
	}

	plan.Uuid = types.StringValue(result.UUID)
	plan.Name = types.StringValue(result.Name)
	if result.Description != "" {
		plan.Description = types.StringValue(result.Description)
	} else {
		plan.Description = types.StringNull()
	}
	plan.ManagementIp = types.StringValue(result.ManagementIp)
	plan.ClusterUuid = types.StringValue(result.ClusterUuid)
	plan.ZoneUuid = types.StringValue(result.ZoneUuid)
	plan.State = types.StringValue(result.State)
	plan.Status = types.StringValue(result.Status)
	plan.HypervisorType = types.StringValue(result.HypervisorType)
	plan.Architecture = types.StringValue(result.Architecture)
	plan.Type = types.StringValue(result.HypervisorType)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *hostResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state hostResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.Uuid.IsNull() || state.Uuid.ValueString() == "" {
		tflog.Warn(ctx, "host uuid is empty, so nothing to delete, skip it")
		return
	}

	uuid := state.Uuid.ValueString()

	tflog.Info(ctx, fmt.Sprintf("delete host %s", uuid))
	err := r.client.DeleteHost(ctx, uuid, param.DeleteModeEnforcing)
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete host", err.Error())
		return
	}
}

func (r *hostResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_host"
}

func (r *hostResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state hostResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	host, err := r.client.GetHost(ctx, state.Uuid.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting ZStack host", "Could not read host: "+err.Error(),
		)
		return
	}

	state.Uuid = types.StringValue(host.UUID)
	state.Name = types.StringValue(host.Name)
	if host.Description != "" {
		state.Description = types.StringValue(host.Description)
	} else {
		state.Description = types.StringNull()
	}
	state.ManagementIp = types.StringValue(host.ManagementIp)
	state.ClusterUuid = types.StringValue(host.ClusterUuid)
	state.ZoneUuid = types.StringValue(host.ZoneUuid)
	state.State = types.StringValue(host.State)
	state.Status = types.StringValue(host.Status)
	state.HypervisorType = types.StringValue(host.HypervisorType)
	state.Architecture = types.StringValue(host.Architecture)
	state.Type = types.StringValue(host.HypervisorType)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *hostResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This resource allows you to manage hosts in ZSphere. " +
			"A host is a physical or virtual machine that provides compute resources for running virtual machines. " +
			"It is a fundamental component of the ZSphere cloud infrastructure.",
		Attributes: map[string]schema.Attribute{
			"uuid": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the host. Automatically generated by ZSphere.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the host. This is a mandatory field.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "A description of the host, providing additional context or details.",
			},
			"management_ip": schema.StringAttribute{
				Required:    true,
				Description: "The management IP address of the host.",
			},
			"cluster_uuid": schema.StringAttribute{
				Required:    true,
				Description: "The UUID of the cluster to which the host belongs.",
			},
			"zone_uuid": schema.StringAttribute{
				Computed:    true,
				Description: "The UUID of the zone to which the host belongs.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"state": schema.StringAttribute{
				Computed:    true,
				Description: "The state of the host, such as 'Enabled' or 'Disabled'.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "The operational status of the host, such as 'Connected' or 'Disconnected'.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"hypervisor_type": schema.StringAttribute{
				Computed:    true,
				Description: "The hypervisor type of the host. Valid values are: KVM, Simulator, baremetal, baremetal2, xdragon.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"architecture": schema.StringAttribute{
				Computed:    true,
				Description: "The CPU architecture of the host. Valid values are: x86_64, aarch64, mips64el, loongarch64.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"username": schema.StringAttribute{
				Required:    true,
				Description: "The username for SSH authentication to the host.",
				Sensitive:   true,
			},
			"password": schema.StringAttribute{
				Required:    true,
				Description: "The password for SSH authentication to the host.",
				Sensitive:   true,
			},
			"ssh_port": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "The SSH port for connecting to the host. Default is 22.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"iommu": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Enable scanning host IOMMU settings. This is required for GPU passthrough, vGPU virtualization, and SR-IOV features. Default is false.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"ept": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Enable Intel EPT hardware-assisted virtualization. Default is true. Note: Only applicable for Intel CPUs.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"type": schema.StringAttribute{
				Computed:    true,
				Description: "The type of the host. Same as hypervisor_type.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *hostResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state hostResourceModel

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
	plan.ClusterUuid = state.ClusterUuid
	plan.ZoneUuid = state.ZoneUuid
	plan.HypervisorType = state.HypervisorType
	plan.Architecture = state.Architecture
	plan.ManagementIp = state.ManagementIp
	plan.Username = state.Username
	plan.Password = state.Password
	plan.SshPort = state.SshPort
	plan.Iommu = state.Iommu
	plan.Ept = state.Ept

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

	updateParam := param.UpdateHostParam{
		Params: param.UpdateHostParamDetail{
			Name:        plan.Name.ValueString(),
			Description: description,
		},
	}

	host, err := r.client.UpdateHost(ctx, plan.Uuid.ValueString(), updateParam)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating ZStack host", "Could not update host: "+err.Error(),
		)
		return
	}

	plan.Name = types.StringValue(host.Name)
	if host.Description != "" {
		plan.Description = types.StringValue(host.Description)
	} else {
		plan.Description = types.StringNull()
	}
	plan.State = types.StringValue(host.State)
	plan.Status = types.StringValue(host.Status)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
