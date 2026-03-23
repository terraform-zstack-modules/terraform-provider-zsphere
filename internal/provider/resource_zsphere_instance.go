// Copyright (c) ZStack.io, Inc.

package provider

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"terraform-provider-zsphere/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/terraform-zstack-modules/zsphere-sdk-go/pkg/client"
	"github.com/terraform-zstack-modules/zsphere-sdk-go/pkg/param"
	"github.com/terraform-zstack-modules/zsphere-sdk-go/pkg/view"
	"golang.org/x/sync/errgroup"
)

type vmResource struct {
	client *client.ZSClient
}

var (
	_ resource.Resource              = &vmResource{}
	_ resource.ResourceWithConfigure = &vmResource{}
)

var networkModelAttrTypes = map[string]attr.Type{
	"uuid":    types.StringType,
	"ip":      types.StringType,
	"netmask": types.StringType,
	"gateway": types.StringType,
}

type diskModel struct {
	Size               types.Int64  `tfsdk:"size"`
	PrimaryStorageUuid types.String `tfsdk:"primary_storage_uuid"`
	Name               types.String `tfsdk:"name"`
	Boot               types.Bool   `tfsdk:"boot"`
	BusType            types.String `tfsdk:"bus_type"`
}

type diskAOModel struct {
	Size               types.Int64  `tfsdk:"size"`
	PrimaryStorageUuid types.String `tfsdk:"primary_storage_uuid"`
	Name               types.String `tfsdk:"name"`
	Boot               types.Bool   `tfsdk:"boot"`
	BusType            types.String `tfsdk:"bus_type"`
	SourceType         types.String `tfsdk:"source_type"`
	SourceUuid         types.String `tfsdk:"source_uuid"`
	SystemTags         types.List   `tfsdk:"system_tags"`
}

type vmNicConfigModel struct {
	L3NetworkUuid     types.String `tfsdk:"l3_network_uuid"`
	DriverType        types.String `tfsdk:"driver_type"`
	MultiQueueNum     types.String `tfsdk:"multi_queue_num"`
	State             types.String `tfsdk:"state"`
	StaticIp          types.String `tfsdk:"static_ip"`
	StaticIpv6        types.String `tfsdk:"static_ipv6"`
	Ipv4Netmask       types.String `tfsdk:"ipv4_netmask"`
	Ipv6Prefix        types.Int64  `tfsdk:"ipv6_prefix"`
	Ipv4Gateway       types.String `tfsdk:"ipv4_gateway"`
	Ipv6Gateway       types.String `tfsdk:"ipv6_gateway"`
	InboundBandwidth  types.Int64  `tfsdk:"inbound_bandwidth"`
	OutboundBandwidth types.Int64  `tfsdk:"outbound_bandwidth"`
	CustomMac         types.String `tfsdk:"custom_mac"`
}

type cdromModel struct {
	CdRom   types.String `tfsdk:"cdrom"`
	IsoUuid types.String `tfsdk:"iso_uuid"`
}

type usbConfigModel struct {
	UsbDeviceUuid types.String `tfsdk:"usb_device_uuid"`
	AttachType    types.String `tfsdk:"attach_type"`
}

type vgpuDeviceModel struct {
	Uuid types.String `tfsdk:"uuid"`
	Type types.String `tfsdk:"type"`
}

type cpuBindItemModel struct {
	VCPU     types.String `tfsdk:"vcpu"`
	PCPUList types.List   `tfsdk:"pcpu_list"`
}

type vmInstanceDataSourceModel struct {
	Uuid              types.String `tfsdk:"uuid"`
	Name              types.String `tfsdk:"name"`
	ImageUuid         types.String `tfsdk:"image_uuid"`
	TemplateUuid      types.String `tfsdk:"template_uuid"`
	NetworkInterfaces types.List   `tfsdk:"network_interfaces"`
	RootDisk          types.Object `tfsdk:"root_disk"`
	DataDisks         types.List   `tfsdk:"data_disks"`
	ZoneUuid          types.String `tfsdk:"datacenter_uuid"`
	ClusterUuid       types.String `tfsdk:"cluster_uuid"`
	HostUuid          types.String `tfsdk:"host_uuid"`
	Description       types.String `tfsdk:"description"`
	Platform          types.String `tfsdk:"platform"`
	GuestOsType       types.String `tfsdk:"guest_os_type"`
	Strategy          types.String `tfsdk:"strategy"`
	MemorySize        types.Int64  `tfsdk:"memory_size"`
	CPUNum            types.Int64  `tfsdk:"cpu_num"`
	NeverStop         types.Bool   `tfsdk:"never_stop"`
	UserData          types.String `tfsdk:"user_data"`
	VMNics            types.List   `tfsdk:"vm_nics"`
	Expunge           types.Bool   `tfsdk:"expunge"`
	BootMode          types.String `tfsdk:"boot_mode"`
	Hostname          types.String `tfsdk:"hostname"`
	Netmask           types.String `tfsdk:"netmask"`
	Gateway           types.String `tfsdk:"gateway"`
	CpuMode           types.String `tfsdk:"cpu_mode"`
	VmMachineType     types.String `tfsdk:"vm_machine_type"`
	Architecture      types.String `tfsdk:"architecture"`
	// Template-specific fields
	VmNicParams         types.String  `tfsdk:"vm_nic_params"`
	VmNicConfig         types.List    `tfsdk:"vm_nic_config"`
	DiskAOs             types.List    `tfsdk:"disk_aos"`
	CdromList           types.List    `tfsdk:"cdrom_list"`
	VmUSBConfig         types.List    `tfsdk:"vm_usb_config"`
	VgpuDevice          types.Object  `tfsdk:"vgpu_device"`
	GpuDeviceUuidList   types.List    `tfsdk:"gpu_device_uuid_list"`
	VmGroupUuid         types.String  `tfsdk:"vm_group_uuid"`
	CpuQuota            types.Float64 `tfsdk:"cpu_quota"`
	SockedNum           types.Int64   `tfsdk:"socked_num"`
	VnumaEnabled        types.Bool    `tfsdk:"vnuma_enabled"`
	CpuBindListByVCpu   types.List    `tfsdk:"cpu_bind_list_by_vcpu"`
	CpuBindType         types.String  `tfsdk:"cpu_bind_type"`
	TotalGPUMemory      types.Float64 `tfsdk:"total_gpu_memory"`
	MemoryResourceLevel types.String  `tfsdk:"memory_resource_level"`
	CpuResourceLevel    types.String  `tfsdk:"cpu_resource_level"`
	UsbRedirect         types.Bool    `tfsdk:"usb_redirect"`
	GpuType             types.String  `tfsdk:"gpu_type"`
	SoundCard           types.String  `tfsdk:"sound_card"`
	MotherboardType     types.String  `tfsdk:"motherboard_type"`
	Group               types.String  `tfsdk:"group"`
	// Post-creation configuration
	BootOrders           types.List   `tfsdk:"boot_orders"`
	VdiMonitorNumber     types.Int64  `tfsdk:"vdi_monitor_number"`
	EmulatorPinning      types.String `tfsdk:"emulator_pinning"`
	ClockTrack           types.String `tfsdk:"clock_track"`
	ClockSyncAfterResume types.Bool   `tfsdk:"clock_sync_after_resume"`
	ClockSyncInterval    types.Int64  `tfsdk:"clock_sync_interval"`
	// Console and access configuration
	ConsolePassword types.String `tfsdk:"console_password"`
	ConsoleMode     types.String `tfsdk:"console_mode"`
	SshKey          types.String `tfsdk:"ssh_key"`
	RootPassword    types.String `tfsdk:"root_password"`
	RootUsername    types.String `tfsdk:"root_username"`
	// Advanced configuration
	BiosTimeSync          types.Bool   `tfsdk:"bios_time_sync"`
	SpiceStreamingMode    types.String `tfsdk:"spice_streaming_mode"`
	FaultStrategy         types.String `tfsdk:"fault_strategy"`
	EmulateHyperV         types.Bool   `tfsdk:"emulate_hyperv"`
	VmPortOff             types.Bool   `tfsdk:"vm_port_off"`
	BootMenuSplashTimeout types.String `tfsdk:"boot_menu_splash_timeout"`
	CpuHideKVMMark        types.Bool   `tfsdk:"cpu_hide_kvm_mark"`
	MigrateAutoConverge   types.Bool   `tfsdk:"migrate_auto_converge"`
	HotPlugEnabled        types.Bool   `tfsdk:"hot_plug_enabled"`
	VmCpuidVendor         types.String `tfsdk:"vm_cpuid_vendor"`
	AffinityGroupUuid     types.String `tfsdk:"affinity_group_uuid"`
	HaStickStrategy       types.Bool   `tfsdk:"ha_stick_strategy"`
	AutoReleaseGpuDevice  types.Bool   `tfsdk:"auto_release_gpu_device"`
	Se                    types.Bool   `tfsdk:"se"`
}

type NicsModel struct {
	Uuid    types.String `tfsdk:"uuid"`
	Ip      types.String `tfsdk:"ip"`
	Netmask types.String `tfsdk:"netmask"`
	Gateway types.String `tfsdk:"gateway"`
}

type NetworkInterfaceModel struct {
	L3NetworkUuid types.String `tfsdk:"port_group_uuid"`
	DefaultL3     types.Bool   `tfsdk:"default_l3"`
	StaticIp      types.String `tfsdk:"static_ip"`
}

func InstanceResource() resource.Resource {
	return &vmResource{}
}

// Configure implements resource.ResourceWithConfigure.
func (r *vmResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.ZSClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.ZSClient, got: %T. Please report this issue to the Provider developer.", req.ProviderData),
		)

		return
	}

	r.client = client
}

// Metadata implements resource.Resource.
func (r *vmResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_instance"
}

// Schema implements resource.Resource.
func (r *vmResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This resource allows you to manage virtual machine (VM) instances in ZStack. " +
			"A VM instance represents a virtualized compute resource that can be created, updated, and deleted. " +
			"You can define the VM's properties, such as its name, image, network configuration, disks, and GPU devices.",
		Attributes: map[string]schema.Attribute{
			"uuid": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the VM instance.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the VM instance.",
			},
			"network_interfaces": schema.ListNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Defines network interfaces attached to the VM. Each NIC corresponds to an L3 network, and optionally configures a static IP.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"port_group_uuid": schema.StringAttribute{
							Required:    true,
							Description: "The UUID of the L3 network for this NIC.",
						},
						"default_l3": schema.BoolAttribute{
							Required:    true,
							Description: "Whether this NIC is the default route NIC.",
						},
						"static_ip": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Static IP address to assign. The format will be converted to system tag `staticIp::<l3_uuid>::<ip>`.",
						},
					},
				},
			},
			"vm_nics": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"uuid": schema.StringAttribute{
							Computed:    true,
							Description: "The UUID of the network.",
						},
						"ip": schema.StringAttribute{
							Computed:    true,
							Description: "The IP address assigned to the network.",
						},
						"netmask": schema.StringAttribute{
							Computed:    true,
							Description: "The netmask of the network.",
						},
						"gateway": schema.StringAttribute{
							Computed:    true,
							Description: "The gateway of the network.",
						},
					},
				},
				Computed:    true,
				Description: "The network interfaces assigned to the VM instance.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"platform": schema.StringAttribute{
				Description: "Platform of the image, such as Linux, Windows, or Other",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("Linux", "Windows", "Other", "Paravirtualization", "WindowsVirtio"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"guest_os_type": schema.StringAttribute{
				Description: "GuestOsType of the image, such as Linux, Windows, or Other",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"image_uuid": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The UUID of the image used to create the VM instance. Either image_uuid or template_uuid must be specified.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"template_uuid": schema.StringAttribute{
				Optional:    true,
				Description: "The UUID of the templated VM instance to create the VM from. Either image_uuid or template_uuid must be specified.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"root_disk": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"size": schema.Int64Attribute{
						Optional:    true,
						Computed:    true,
						Description: "The size of the root disk in bytes.",
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
					"primary_storage_uuid": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "The UUID of the primary storage for the root disk.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},

					"name": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "The name of the root disk.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"boot": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Whether this disk is the boot disk.",
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
					},
					"bus_type": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Bus type of the disk, such as virtio, ide, virtio-scsi, scsi.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						Validators: []validator.String{
							stringvalidator.OneOf("virtio", "ide", "virtio-scsi", "scsi"),
						},
					},
				},
				Optional:    true,
				Computed:    true,
				Description: "The configuration for the root disk of the VM instance. When creating from template, this is inherited from the template.",
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
			},
			"data_disks": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"size": schema.Int64Attribute{
							Optional:    true,
							Computed:    true,
							Description: "The size of the data disk in bytes.",
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.UseStateForUnknown(),
							},
						},
						"primary_storage_uuid": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "The UUID of the primary storage for the data disk.",
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},

						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the data disk. Auto-generated as {vm-name}-1, {vm-name}-2, etc., matching the frontend naming convention.",
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"boot": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Whether this disk is the boot disk.",
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.UseStateForUnknown(),
							},
						},
						"bus_type": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Bus type of the disk, such as virtio, ide, virtio-scsi, scsi.",
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
							Validators: []validator.String{
								stringvalidator.OneOf("virtio", "ide", "virtio-scsi", "scsi"),
							},
						},
					},
				},
				Optional:    true,
				Computed:    true,
				Description: "The configuration for additional data disks. When creating from template, these are inherited from the template.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"datacenter_uuid": schema.StringAttribute{
				Optional:    true,
				Description: "The UUID of the zone where the VM instance is deployed.",
			},
			"cluster_uuid": schema.StringAttribute{
				Optional:    true,
				Description: "The UUID of the cluster where the VM instance is deployed.",
			},
			"host_uuid": schema.StringAttribute{
				Optional:    true,
				Description: "The UUID of the host where the VM instance is running.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "A description of the VM instance.",
			},
			"memory_size": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "The memory size allocated to the VM instance in megabytes (MB). When used together with `cpu_num`, the `instance_offering_uuid` is not required.",
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"cpu_num": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "The number of CPUs allocated to the VM instance.  When used together with `memory_size`, the `instance_offering_uuid` is not required.",
				Validators: []validator.Int64{
					int64validator.Between(1, 1024),
				},
			},
			"strategy": schema.StringAttribute{
				Optional:    true,
				Description: "The deployment strategy for the VM instance.",
				Validators: []validator.String{
					stringvalidator.OneOf("InstantStart", "JustCreate", "CreateStopped"),
				},
			},
			"user_data": schema.StringAttribute{
				Optional:    true,
				Description: "User data injected into the VM instance at boot time.",
			},
			"never_stop": schema.BoolAttribute{
				Optional:    true,
				Description: "Whether the VM instance should never stop automatically.",
			},
			"expunge": schema.BoolAttribute{
				Optional:    true,
				Description: "Indicates if the instance should be expunged after deletion.",
			},
			// Provider enhancement fields
			"boot_mode": schema.StringAttribute{
				Optional:    true,
				Description: "Boot mode for the VM instance. Valid values are 'Legacy' and 'UEFI'.",
				Validators: []validator.String{
					stringvalidator.OneOf("Legacy", "UEFI"),
				},
			},
			"hostname": schema.StringAttribute{
				Optional:    true,
				Description: "Hostname to be set for the VM instance after creation (requires VMtools).",
			},
			"netmask": schema.StringAttribute{
				Optional:    true,
				Description: "Netmask for the VM's network interface (used with static IP).",
			},
			"gateway": schema.StringAttribute{
				Optional:    true,
				Description: "Gateway for the VM's network interface (used with static IP).",
			},
			"cpu_mode": schema.StringAttribute{
				Optional:    true,
				Description: "CPU mode for the VM. Valid values are 'host-model', 'host-passthrough', or 'custom'.",
				Validators: []validator.String{
					stringvalidator.OneOf("host-model", "host-passthrough", "custom"),
				},
			},
			"vm_machine_type": schema.StringAttribute{
				Optional:    true,
				Description: "VM machine type. For UEFI boot mode, 'q35' is recommended.",
			},
			"architecture": schema.StringAttribute{
				Optional:    true,
				Description: "Architecture of the VM instance. Default is 'x86_64'.",
				Validators: []validator.String{
					stringvalidator.OneOf("x86_64", "aarch64"),
				},
			},
			// Template-specific fields
			"vm_nic_params": schema.StringAttribute{
				Optional:    true,
				Description: "JSON string of VM NIC parameters for template creation. Used for advanced NIC configuration.",
			},
			"vm_nic_config": schema.ListNestedAttribute{
				Optional:    true,
				Description: "List of VM NIC configurations for template creation.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"l3_network_uuid": schema.StringAttribute{
							Required:    true,
							Description: "UUID of the L3 network.",
						},
						"driver_type": schema.StringAttribute{
							Optional:    true,
							Description: "Driver type for the NIC (e.g., virtio, e1000).",
						},
						"multi_queue_num": schema.StringAttribute{
							Optional:    true,
							Description: "Number of multi-queues for the NIC.",
						},
						"state": schema.StringAttribute{
							Optional:    true,
							Description: "State of the NIC (enable/disable).",
						},
						"static_ip": schema.StringAttribute{
							Optional:    true,
							Description: "Static IPv4 address.",
						},
						"static_ipv6": schema.StringAttribute{
							Optional:    true,
							Description: "Static IPv6 address.",
						},
						"ipv4_netmask": schema.StringAttribute{
							Optional:    true,
							Description: "IPv4 netmask.",
						},
						"ipv6_prefix": schema.Int64Attribute{
							Optional:    true,
							Description: "IPv6 prefix length.",
						},
						"ipv4_gateway": schema.StringAttribute{
							Optional:    true,
							Description: "IPv4 gateway.",
						},
						"ipv6_gateway": schema.StringAttribute{
							Optional:    true,
							Description: "IPv6 gateway.",
						},
						"inbound_bandwidth": schema.Int64Attribute{
							Optional:    true,
							Description: "Inbound bandwidth limit.",
						},
						"outbound_bandwidth": schema.Int64Attribute{
							Optional:    true,
							Description: "Outbound bandwidth limit.",
						},
						"custom_mac": schema.StringAttribute{
							Optional:    true,
							Description: "Custom MAC address.",
						},
					},
				},
			},
			"disk_aos": schema.ListNestedAttribute{
				Optional:    true,
				Description: "List of disk AO configurations for template creation.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"size": schema.Int64Attribute{
							Optional:    true,
							Description: "Size of the disk in bytes.",
						},
						"primary_storage_uuid": schema.StringAttribute{
							Optional:    true,
							Description: "UUID of the primary storage.",
						},
						"name": schema.StringAttribute{
							Optional:    true,
							Description: "Name of the disk.",
						},
						"boot": schema.BoolAttribute{
							Optional:    true,
							Description: "Whether this is the boot disk.",
						},
						"bus_type": schema.StringAttribute{
							Optional:    true,
							Description: "Bus type of the disk (virtio, ide, virtio-scsi, scsi).",
						},
						"source_type": schema.StringAttribute{
							Optional:    true,
							Description: "Source type of the disk (e.g., TemplatedVmInstanceVO, VolumeVO).",
						},
						"source_uuid": schema.StringAttribute{
							Optional:    true,
							Description: "Source UUID of the disk.",
						},
						"system_tags": schema.ListAttribute{
							Optional:    true,
							ElementType: types.StringType,
							Description: "System tags for the disk.",
						},
					},
				},
			},
			"cdrom_list": schema.ListNestedAttribute{
				Optional:    true,
				Description: "List of CD-ROM configurations.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cdrom": schema.StringAttribute{
							Required:    true,
							Description: "CD-ROM identifier.",
						},
						"iso_uuid": schema.StringAttribute{
							Optional:    true,
							Description: "UUID of the ISO image (Empty for empty CD-ROM).",
						},
					},
				},
			},
			"vm_usb_config": schema.ListNestedAttribute{
				Optional:    true,
				Description: "List of USB device configurations.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"usb_device_uuid": schema.StringAttribute{
							Required:    true,
							Description: "UUID of the USB device.",
						},
						"attach_type": schema.StringAttribute{
							Optional:    true,
							Description: "Attachment type for the USB device. Valid values: PassThrough (direct connection, requires USB device on same host), Redirect (forwarding, allows USB device on any host in same datacenter).",
							Validators: []validator.String{
								stringvalidator.OneOf("PassThrough", "Redirect"),
							},
						},
					},
				},
			},
			"vgpu_device": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "vGPU device configuration.",
				Attributes: map[string]schema.Attribute{
					"uuid": schema.StringAttribute{
						Required:    true,
						Description: "UUID of the vGPU device.",
					},
					"type": schema.StringAttribute{
						Required:    true,
						Description: "Type of the vGPU device (e.g., MdevDevice).",
					},
				},
			},
			"gpu_device_uuid_list": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "List of GPU device UUIDs.",
			},
			"vm_group_uuid": schema.StringAttribute{
				Optional:    true,
				Description: "UUID of the VM scheduling group.",
			},
			"cpu_quota": schema.Float64Attribute{
				Optional:    true,
				Description: "CPU quota for the VM.",
			},
			"socked_num": schema.Int64Attribute{
				Optional:    true,
				Description: "Number of CPU sockets.",
			},
			"vnuma_enabled": schema.BoolAttribute{
				Optional:    true,
				Description: "Whether vNUMA is enabled.",
			},
			"cpu_bind_list_by_vcpu": schema.ListNestedAttribute{
				Optional:    true,
				Description: "List of CPU binding configurations per vCPU.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"vcpu": schema.StringAttribute{
							Required:    true,
							Description: "vCPU identifier.",
						},
						"pcpu_list": schema.ListAttribute{
							Required:    true,
							ElementType: types.StringType,
							Description: "List of physical CPUs to bind.",
						},
					},
				},
			},
			"cpu_bind_type": schema.StringAttribute{
				Optional:    true,
				Description: "CPU binding type.",
			},
			"total_gpu_memory": schema.Float64Attribute{
				Optional:    true,
				Description: "Total GPU memory in MB.",
			},
			"memory_resource_level": schema.StringAttribute{
				Optional:    true,
				Description: "Memory resource level (Normal, High).",
			},
			"cpu_resource_level": schema.StringAttribute{
				Optional:    true,
				Description: "CPU resource level (Normal, CpuHigh).",
			},
			"usb_redirect": schema.BoolAttribute{
				Optional:    true,
				Description: "Whether USB redirect is enabled.",
			},
			"gpu_type": schema.StringAttribute{
				Optional:    true,
				Description: "Type of GPU (e.g., qxl, vga).",
			},
			"sound_card": schema.StringAttribute{
				Optional:    true,
				Description: "Type of sound card (e.g., ich6, ac97).",
			},
			"motherboard_type": schema.StringAttribute{
				Optional:    true,
				Description: "Type of motherboard (e.g., q35, pc).",
			},
			"group": schema.StringAttribute{
				Optional:    true,
				Description: "VM group identifier.",
			},
			// Post-creation configuration
			"boot_orders": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Boot order for the VM (e.g., CdRom, HardDisk, Network).",
			},
			"vdi_monitor_number": schema.Int64Attribute{
				Optional:    true,
				Description: "Number of VDI monitors.",
			},
			"emulator_pinning": schema.StringAttribute{
				Optional:    true,
				Description: "Emulator pinning configuration.",
			},
			"clock_track": schema.StringAttribute{
				Optional:    true,
				Description: "Clock track mode for the VM.",
			},
			"clock_sync_after_resume": schema.BoolAttribute{
				Optional:    true,
				Description: "Whether to sync clock after VM resume.",
			},
			"clock_sync_interval": schema.Int64Attribute{
				Optional:    true,
				Description: "Clock sync interval in seconds.",
			},
			"console_password": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Password for VM console access.",
			},
			"console_mode": schema.StringAttribute{
				Optional:    true,
				Description: "Console mode for the VM (vnc, spice).",
				Validators: []validator.String{
					stringvalidator.OneOf("vnc", "spice"),
				},
			},
			"ssh_key": schema.StringAttribute{
				Optional:    true,
				Description: "SSH public key for VM access.",
			},
			"root_password": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Root password for the VM (used with user_data).",
			},
			"root_username": schema.StringAttribute{
				Optional:    true,
				Description: "Root username for the VM (default is root for Linux, Administrator for Windows).",
			},
			"bios_time_sync": schema.BoolAttribute{
				Optional:    true,
				Description: "Whether to sync VM time with BIOS time.",
			},
			"spice_streaming_mode": schema.StringAttribute{
				Optional:    true,
				Description: "Spice streaming mode for the VM.",
			},
			"fault_strategy": schema.StringAttribute{
				Optional:    true,
				Description: "Fault handling strategy for the VM.",
			},
			"emulate_hyperv": schema.BoolAttribute{
				Optional:    true,
				Description: "Whether to emulate Hyper-V for the VM.",
			},
			"vm_port_off": schema.BoolAttribute{
				Optional:    true,
				Description: "Whether to turn off VM ports.",
			},
			"boot_menu_splash_timeout": schema.StringAttribute{
				Optional:    true,
				Description: "Timeout for boot menu splash screen.",
			},
			"cpu_hide_kvm_mark": schema.BoolAttribute{
				Optional:    true,
				Description: "Whether to hide KVM mark from VM CPU.",
			},
			"migrate_auto_converge": schema.BoolAttribute{
				Optional:    true,
				Description: "Whether to enable auto-converge for VM migration.",
			},
			"hot_plug_enabled": schema.BoolAttribute{
				Optional:    true,
				Description: "Whether hot-plug is enabled for PCI devices.",
			},
			"vm_cpuid_vendor": schema.StringAttribute{
				Optional:    true,
				Description: "Custom CPUID vendor string for the VM.",
			},
			"affinity_group_uuid": schema.StringAttribute{
				Optional:    true,
				Description: "UUID of the affinity group for the VM.",
			},
			"ha_stick_strategy": schema.BoolAttribute{
				Optional:    true,
				Description: "Whether to enable HA stick strategy.",
			},
			"auto_release_gpu_device": schema.BoolAttribute{
				Optional:    true,
				Description: "Whether to auto-release GPU device when VM stops.",
			},
			"se": schema.BoolAttribute{
				Optional:    true,
				Description: "Whether security element is enabled.",
			},
		},
	}
}

// Create implements resource.Resource.
func (r *vmResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan vmInstanceDataSourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that exactly one of image_uuid or template_uuid is specified
	if err := r.validateCreationSource(ctx, plan); err != nil {
		resp.Diagnostics.AddError("Validation Error", err.Error())
		return
	}

	// Route to appropriate creation method
	var instance *view.VmInstanceInventoryView
	var createNics []NetworkInterfaceModel
	var err error

	isTemplate := !plan.TemplateUuid.IsNull() && plan.TemplateUuid.ValueString() != ""

	if isTemplate {
		// Create from template
		instance, createNics, err = r.createFromTemplate(ctx, plan, resp)
	} else {
		// Create from image
		instance, createNics, err = r.createFromImage(ctx, plan, resp)
	}

	if err != nil {
		// Error already handled in respective methods
		return
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Common post-creation processing
	// fromImage: true for image creation, false for template creation
	r.handlePostCreation(ctx, plan, instance, createNics, !isTemplate, resp)

	// Execute post-creation configuration tasks based on creation method
	if isTemplate {
		// Template-based creation uses different post-creation tasks
		r.executePostCreationTasksForTemplate(ctx, plan, instance, resp)
	} else {
		// Image-based creation uses full post-creation tasks
		r.executePostCreationTasks(ctx, plan, instance, resp)
	}
}

// validateCreationSource validates that exactly one of image_uuid or template_uuid is specified
// and validates parameters specific to each creation method
func (r *vmResource) validateCreationSource(ctx context.Context, plan vmInstanceDataSourceModel) error {
	hasImage := !plan.ImageUuid.IsNull() && plan.ImageUuid.ValueString() != ""
	hasTemplate := !plan.TemplateUuid.IsNull() && plan.TemplateUuid.ValueString() != ""

	if !hasImage && !hasTemplate {
		return fmt.Errorf("either image_uuid or template_uuid must be specified")
	}
	if hasImage && hasTemplate {
		return fmt.Errorf("image_uuid and template_uuid are mutually exclusive, specify only one")
	}

	// Validate parameters specific to each creation method
	if hasImage {
		// Image creation: validate that template-specific parameters are not set
		// Parameters exclusive to template creation:
		if !plan.VmNicParams.IsNull() && plan.VmNicParams.ValueString() != "" {
			return fmt.Errorf("vm_nic_params is not supported when creating from image, use network_interfaces instead")
		}
		if !plan.DiskAOs.IsNull() && len(plan.DiskAOs.Elements()) > 0 {
			return fmt.Errorf("disk_aos is not supported when creating from image, use root_disk and data_disks instead")
		}
		// vm_nic_config is also template-specific
		if !plan.VmNicConfig.IsNull() && len(plan.VmNicConfig.Elements()) > 0 {
			return fmt.Errorf("vm_nic_config is not supported when creating from image, use network_interfaces instead")
		}
	} else if hasTemplate {
		// Template creation: validate that image-specific parameters are not set
		// Note: root_disk and data_disks are Computed fields, they will be filled from template
		// We only validate if user explicitly provides them (check config instead of plan)
		if !plan.RootDisk.IsUnknown() && !plan.RootDisk.IsNull() {
			// Check if user provided root_disk in config by checking if size is set
			var rootDisk diskModel
			if diags := plan.RootDisk.As(ctx, &rootDisk, basetypes.ObjectAsOptions{}); !diags.HasError() {
				if !rootDisk.Size.IsNull() && rootDisk.Size.ValueInt64() > 0 {
					return fmt.Errorf("root_disk is not supported when creating from template, the root disk is inherited from the template. Use disk_aos to add additional disks")
				}
			}
		}
		if !plan.DataDisks.IsUnknown() && !plan.DataDisks.IsNull() && len(plan.DataDisks.Elements()) > 0 {
			return fmt.Errorf("data_disks is not supported when creating from template, data disks are inherited from the template. Use disk_aos to add additional disks")
		}
		if !plan.Platform.IsNull() && plan.Platform.ValueString() != "" {
			return fmt.Errorf("platform is not supported when creating from template, the platform is inherited from the template")
		}
		if !plan.GuestOsType.IsNull() && plan.GuestOsType.ValueString() != "" {
			return fmt.Errorf("guest_os_type is not supported when creating from template, the guest OS type is inherited from the template")
		}
		// Image-specific post-creation parameters not applicable to template
		if !plan.VdiMonitorNumber.IsNull() && plan.VdiMonitorNumber.ValueInt64() > 0 {
			return fmt.Errorf("vdi_monitor_number is not supported when creating from template")
		}
		if !plan.BootOrders.IsNull() && len(plan.BootOrders.Elements()) > 0 {
			return fmt.Errorf("boot_orders is not supported when creating from template")
		}
		if !plan.ClockTrack.IsNull() && plan.ClockTrack.ValueString() != "" {
			return fmt.Errorf("clock_track is not supported when creating from template")
		}
		// Clock sync parameters are also image-specific
		if !plan.ClockSyncAfterResume.IsNull() && plan.ClockSyncAfterResume.ValueBool() {
			return fmt.Errorf("clock_sync_after_resume is not supported when creating from template")
		}
		if !plan.ClockSyncInterval.IsNull() && plan.ClockSyncInterval.ValueInt64() > 0 {
			return fmt.Errorf("clock_sync_interval is not supported when creating from template")
		}
		// BIOS time sync is image-specific
		if !plan.BiosTimeSync.IsNull() && plan.BiosTimeSync.ValueBool() {
			return fmt.Errorf("bios_time_sync is not supported when creating from template")
		}
		// Spice streaming mode is image-specific
		if !plan.SpiceStreamingMode.IsNull() && plan.SpiceStreamingMode.ValueString() != "" {
			return fmt.Errorf("spice_streaming_mode is not supported when creating from template")
		}
		// Fault strategy is image-specific
		if !plan.FaultStrategy.IsNull() && plan.FaultStrategy.ValueString() != "" {
			return fmt.Errorf("fault_strategy is not supported when creating from template")
		}
		// Boot menu splash timeout is image-specific
		if !plan.BootMenuSplashTimeout.IsNull() && plan.BootMenuSplashTimeout.ValueString() != "" {
			return fmt.Errorf("boot_menu_splash_timeout is not supported when creating from template")
		}
	}

	return nil
}

// createFromImage creates a VM instance from an image
func (r *vmResource) createFromImage(ctx context.Context, plan vmInstanceDataSourceModel, resp *resource.CreateResponse) (*view.VmInstanceInventoryView, []NetworkInterfaceModel, error) {
	var rootDiskPlan diskModel
	var dataDisksPlan []diskModel

	hostUuid := ""
	clusterUuid := ""
	zoneUuid := ""

	// Build DiskAOs for all disks (including root disk)
	var diskAOs []param.DiskAOParam
	// Check if any disk uses virtio bus type
	var virtio *bool
	// SET ROOT DISK
	if !plan.RootDisk.IsNull() {
		diags := plan.RootDisk.As(ctx, &rootDiskPlan, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return nil, nil, fmt.Errorf("failed to parse root_disk")
		}

		if err := isDiskParamValid(ctx, r, rootDiskPlan); err != nil {
			resp.Diagnostics.AddError("Params Error", fmt.Sprintf("invalid root_disk param: %v", err))
			return nil, nil, err
		}

		// Build root disk DiskAO
		rootDiskAO := param.DiskAOParam{
			Size: rootDiskPlan.Size.ValueInt64(),
			Boot: true,
		}

		// Set platform, guestOsType, architecture from plan (not from disk plan)
		if !plan.Platform.IsNull() && plan.Platform.ValueString() != "" {
			rootDiskAO.Platform = plan.Platform.ValueStringPointer()
		}
		if !plan.GuestOsType.IsNull() && plan.GuestOsType.ValueString() != "" {
			rootDiskAO.GuestOsType = plan.GuestOsType.ValueStringPointer()
		}
		if !plan.Architecture.IsNull() && plan.Architecture.ValueString() != "" {
			rootDiskAO.Architecture = plan.Architecture.ValueStringPointer()
		}

		// Set name
		if !rootDiskPlan.Name.IsNull() && rootDiskPlan.Name.ValueString() != "" {
			rootDiskAO.Name = rootDiskPlan.Name.ValueString()
		}

		// Set primary storage
		if !rootDiskPlan.PrimaryStorageUuid.IsNull() && rootDiskPlan.PrimaryStorageUuid.ValueString() != "" {
			rootDiskAO.PrimaryStorageUuid = rootDiskPlan.PrimaryStorageUuid.ValueStringPointer()
		}

		// Build system tags for root disk
		var rootDiskSystemTags []string
		if !rootDiskPlan.BusType.IsNull() && rootDiskPlan.BusType.ValueString() != "" {
			switch rootDiskPlan.BusType.ValueString() {
			case "virtio-scsi":
				rootDiskSystemTags = append(rootDiskSystemTags, "capability::virtio-scsi")
			case "scsi":
				rootDiskSystemTags = append(rootDiskSystemTags, "capability::scsi")
			case "virtio":
				virtio = utils.BoolPointer(true)
			}
		}
		if len(rootDiskSystemTags) > 0 {
			rootDiskAO.SystemTags = rootDiskSystemTags
		}

		diskAOs = append(diskAOs, rootDiskAO)
	}

	// SET DATA DISK
	if !plan.DataDisks.IsNull() {
		plan.DataDisks.ElementsAs(ctx, &dataDisksPlan, false)

		for i := range dataDisksPlan {
			if dataDisksPlan[i].Size.IsNull() {
				resp.Diagnostics.AddError("Params Error", "data_disk size cannot be null")
				return nil, nil, fmt.Errorf("data_disk size cannot be null")
			}

			// Auto-generate disk name: {vm-name}-{index}, matching frontend behavior
			// Frontend uses format: {vm-name}-1, {vm-name}-2, etc.
			diskName := fmt.Sprintf("%s-%d", plan.Name.ValueString(), i+1)

			// Build data disk DiskAO
			dataDiskAO := param.DiskAOParam{
				Size: dataDisksPlan[i].Size.ValueInt64(),
				Name: diskName,
			}

			if !dataDisksPlan[i].PrimaryStorageUuid.IsNull() && dataDisksPlan[i].PrimaryStorageUuid.ValueString() != "" {
				dataDiskAO.PrimaryStorageUuid = dataDisksPlan[i].PrimaryStorageUuid.ValueStringPointer()
			}

			// Build system tags for data disk
			var dataDiskSystemTags []string

			// Handle bus type system tags
			if !dataDisksPlan[i].BusType.IsNull() && dataDisksPlan[i].BusType.ValueString() != "" {
				switch dataDisksPlan[i].BusType.ValueString() {
				case "virtio-scsi":
					dataDiskSystemTags = append(dataDiskSystemTags, "capability::virtio-scsi")
				case "scsi":
					dataDiskSystemTags = append(dataDiskSystemTags, "capability::scsi")
				case "virtio":
					virtio = utils.BoolPointer(true)
				}
			}

			if len(dataDiskSystemTags) > 0 {
				dataDiskAO.SystemTags = dataDiskSystemTags
			}

			diskAOs = append(diskAOs, dataDiskAO)
		}
	}
	// Build network configuration
	l3NetworkUuids, defaultL3Uuid, createNics, err := r.buildNetworkConfig(ctx, plan, resp)
	if err != nil {
		return nil, nil, err
	}

	// Validate image
	if err := r.validateImage(ctx, plan.ImageUuid.ValueString(), resp); err != nil {
		return nil, nil, err
	}

	// SET HOST UUID
	if !plan.HostUuid.IsNull() && plan.HostUuid.ValueString() != "" {
		hostUuid = plan.HostUuid.ValueString()
	}

	// SET CLUSTER UUID
	if !plan.ClusterUuid.IsNull() && plan.ClusterUuid.ValueString() != "" {
		clusterUuid = plan.ClusterUuid.ValueString()
	}

	// SET ZONE UUID
	if !plan.ZoneUuid.IsNull() && plan.ZoneUuid.ValueString() != "" {
		zoneUuid = plan.ZoneUuid.ValueString()
	}

	// Build system tags with all parameters
	systemTags := r.buildSystemTags(ctx, plan)

	// Validate strategy
	if !plan.Strategy.IsNull() {
		strategyValue := plan.Strategy.ValueString()
		if strategyValue != "InstantStart" && strategyValue != "CreateStopped" {
			resp.Diagnostics.AddError("Params Error", fmt.Sprintf("strategy %s is invalid, valid values are InstantStart or CreateStopped", strategyValue))
			return nil, nil, fmt.Errorf("invalid strategy")
		}
	}

	// Build creation parameters
	memorySize := utils.MBToBytes(plan.MemorySize.ValueInt64())
	cpuNum := int(plan.CPUNum.ValueInt64())

	// Build rootVolumeSystemTags for root disk QoS and other settings
	rootVolumeSystemTags := r.buildRootVolumeSystemTags(ctx, plan)

	createVmInstanceParam := param.CreateVmInstanceParam{
		BaseParam: param.BaseParam{
			SystemTags: systemTags,
			UserTags:   nil,
			RequestIp:  "",
		},
		Params: param.CreateVmInstanceParamDetail{
			Name:                 plan.Name.ValueString(),
			ImageUuid:            plan.ImageUuid.ValueStringPointer(),
			L3NetworkUuids:       l3NetworkUuids,
			ZoneUuid:             utils.StringPointer(zoneUuid),
			ClusterUuid:          utils.StringPointer(clusterUuid),
			HostUuid:             utils.StringPointer(hostUuid),
			Description:          plan.Description.ValueStringPointer(),
			DefaultL3NetworkUuid: utils.StringPointer(defaultL3Uuid),
			Strategy:             plan.Strategy.ValueStringPointer(),
			MemorySize:           &memorySize,
			CpuNum:               &cpuNum,
			Virtio:               virtio,
			DiskAOs:              diskAOs,
			RootVolumeSystemTags: rootVolumeSystemTags,
		},
	}

	instance, err := r.client.CreateVmInstance(ctx, createVmInstanceParam)
	if err != nil {
		resp.Diagnostics.AddError("Create VmInstance Error", fmt.Sprintf("failed to create VM instance: %v", err))
		return nil, nil, err
	}

	return instance, createNics, nil
}

// createFromTemplate creates a VM instance from a templated VM instance
func (r *vmResource) createFromTemplate(ctx context.Context, plan vmInstanceDataSourceModel, resp *resource.CreateResponse) (*view.VmInstanceInventoryView, []NetworkInterfaceModel, error) {
	templateUuid := plan.TemplateUuid.ValueString()

	// Query template VM to get inherited resources (disks and nics)
	templateVm, err := r.client.GetVmInstance(ctx, templateUuid)
	if err != nil {
		resp.Diagnostics.AddError("Get Template VM Error", fmt.Sprintf("failed to get template VM: %v", err))
		return nil, nil, err
	}

	// Build network configuration (merge template nics with user-specified nics)
	l3NetworkUuids, defaultL3Uuid, createNics, err := r.buildNetworkConfigFromTemplate(ctx, plan, templateVm, resp)
	if err != nil {
		return nil, nil, err
	}

	// Build system tags with all parameters
	systemTags := r.buildSystemTags(ctx, plan)

	// Validate strategy
	if !plan.Strategy.IsNull() {
		strategyValue := plan.Strategy.ValueString()
		if strategyValue != "InstantStart" && strategyValue != "CreateStopped" && strategyValue != "JustCreate" {
			resp.Diagnostics.AddError("Params Error", fmt.Sprintf("strategy %s is invalid, valid values are InstantStart, CreateStopped, or JustCreate", strategyValue))
			return nil, nil, fmt.Errorf("invalid strategy")
		}
	}

	// Build creation parameters
	memorySize := utils.MBToBytes(plan.MemorySize.ValueInt64())
	cpuNum := int(plan.CPUNum.ValueInt64())

	// Set zone/cluster/host UUIDs from plan
	zoneUuid := ""
	if !plan.ZoneUuid.IsNull() && plan.ZoneUuid.ValueString() != "" {
		zoneUuid = plan.ZoneUuid.ValueString()
	}
	clusterUuid := ""
	if !plan.ClusterUuid.IsNull() && plan.ClusterUuid.ValueString() != "" {
		clusterUuid = plan.ClusterUuid.ValueString()
	}
	hostUuid := ""
	if !plan.HostUuid.IsNull() && plan.HostUuid.ValueString() != "" {
		hostUuid = plan.HostUuid.ValueString()
	}
	vmNicParams := ""
	// Handle vm_nic_params if provided
	if !plan.VmNicParams.IsNull() && plan.VmNicParams.ValueString() != "" {
		vmNicParams = plan.VmNicParams.ValueString()
	}

	createParam := param.CreateVmInstanceFromTemplatedVmInstanceParam{
		BaseParam: param.BaseParam{
			SystemTags: systemTags,
			UserTags:   nil,
			RequestIp:  "",
		},
		Params: param.CreateVmInstanceFromTemplatedVmInstanceParamDetail{
			Names:                []string{plan.Name.ValueString()},
			Strategy:             plan.Strategy.ValueStringPointer(),
			Description:          plan.Description.ValueStringPointer(),
			CpuNum:               &cpuNum,
			MemorySize:           &memorySize,
			L3NetworkUuids:       l3NetworkUuids,
			DefaultL3NetworkUuid: utils.StringPointer(defaultL3Uuid),
			ZoneUuid:             utils.StringPointer(zoneUuid),
			ClusterUuid:          utils.StringPointer(clusterUuid),
			HostUuid:             utils.StringPointer(hostUuid),
			VmNicParams:          utils.StringPointer(vmNicParams),
		},
	}

	// Build diskAOs from template VM volumes (inherited disks)
	diskAOs := r.buildTemplateDiskAOs(templateVm)

	// Add user-specified disk_aos if provided
	if !plan.DiskAOs.IsNull() && len(plan.DiskAOs.Elements()) > 0 {
		userDiskAOs, err := r.buildDiskAOs(ctx, plan, resp)
		if err != nil {
			return nil, nil, err
		}
		diskAOs = append(diskAOs, userDiskAOs...)
	}

	createParam.Params.DiskAOs = diskAOs

	result, err := r.client.CreateVmInstanceFromTemplatedVmInstance(ctx, templateUuid, createParam)
	if err != nil {
		resp.Diagnostics.AddError("Create VmInstance From Template Error", fmt.Sprintf("failed to create VM instance from template: %v", err))
		return nil, nil, err
	}

	if !result.Success || len(result.Result.Inventories) == 0 {
		resp.Diagnostics.AddError("Create VmInstance From Template Error", "failed to create VM instance from template: empty result")
		return nil, nil, fmt.Errorf("empty result")
	}

	// Get the created VM instance details
	instance, err := r.client.GetVmInstance(ctx, result.Result.Inventories[0].Inventory.UUID)
	if err != nil {
		resp.Diagnostics.AddError("Get VmInstance Error", fmt.Sprintf("failed to get created VM instance: %v", err))
		return nil, nil, err
	}

	return instance, createNics, nil
}

// buildDiskAOs builds disk AO parameters from plan
func (r *vmResource) buildDiskAOs(ctx context.Context, plan vmInstanceDataSourceModel, resp *resource.CreateResponse) ([]param.DiskAOParam, error) {
	var diskModels []diskAOModel
	diags := plan.DiskAOs.ElementsAs(ctx, &diskModels, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return nil, fmt.Errorf("failed to parse disk_aos")
	}

	var diskAOs []param.DiskAOParam
	for i, disk := range diskModels {
		diskAO := param.DiskAOParam{
			Boot: i == 0, // First disk is boot disk
		}

		if !disk.Size.IsNull() {
			diskAO.Size = disk.Size.ValueInt64()
		}
		if !disk.PrimaryStorageUuid.IsNull() && disk.PrimaryStorageUuid.ValueString() != "" {
			diskAO.PrimaryStorageUuid = disk.PrimaryStorageUuid.ValueStringPointer()
		}
		if !disk.Name.IsNull() && disk.Name.ValueString() != "" {
			diskAO.Name = disk.Name.ValueString()
		}
		if !disk.SourceType.IsNull() && disk.SourceType.ValueString() != "" {
			sourceType := disk.SourceType.ValueString()
			diskAO.SourceType = &sourceType
		}
		if !disk.SourceUuid.IsNull() && disk.SourceUuid.ValueString() != "" {
			sourceUuid := disk.SourceUuid.ValueString()
			diskAO.SourceUuid = &sourceUuid
		}

		// Handle system tags
		var systemTags []string
		if !disk.BusType.IsNull() && disk.BusType.ValueString() != "" {
			switch disk.BusType.ValueString() {
			case "virtio-scsi":
				systemTags = append(systemTags, "capability::virtio-scsi")
			case "scsi":
				systemTags = append(systemTags, "capability::scsi")
			}
		}
		if !disk.SystemTags.IsNull() && len(disk.SystemTags.Elements()) > 0 {
			var tags []string
			disk.SystemTags.ElementsAs(ctx, &tags, false)
			systemTags = append(systemTags, tags...)
		}
		if len(systemTags) > 0 {
			diskAO.SystemTags = systemTags
		}

		diskAOs = append(diskAOs, diskAO)
	}

	return diskAOs, nil
}

// buildRootVolumeSystemTags builds system tags for root volume (disk) from plan
// This handles QoS, cacheMode, aio, and other disk-related settings for the root disk
func (r *vmResource) buildRootVolumeSystemTags(ctx context.Context, plan vmInstanceDataSourceModel) []string {
	var rootVolumeSystemTags []string

	// Note: The root disk settings from plan (like cacheMode, aio, etc.) would need to be passed
	// Currently, the root disk settings are handled in diskAOs as SystemTags
	// This function can be extended to include additional root volume specific system tags

	return rootVolumeSystemTags
}

// buildTemplateDiskAOs builds disk AO parameters from template VM volumes
func (r *vmResource) buildTemplateDiskAOs(templateVm *view.VmInstanceInventoryView) []param.DiskAOParam {
	var diskAOs []param.DiskAOParam

	for i, volume := range templateVm.AllVolumes {
		diskAO := param.DiskAOParam{
			Size:       volume.Size,
			Name:       fmt.Sprintf("-%d", i+1), // Name as "-1", "-2", etc.
			SourceType: utils.StringPointer("TemplatedVmInstanceVO"),
			SourceUuid: utils.StringPointer(volume.UUID),
			Boot:       volume.Type == "Root",
		}

		diskAOs = append(diskAOs, diskAO)
	}

	return diskAOs
}

// buildVmNicParams builds vmNicParams JSON string from VmNicConfig
// This includes multiQueueNum for each NIC (matching frontend behavior)
func (r *vmResource) buildVmNicParams(ctx context.Context, plan vmInstanceDataSourceModel) (string, error) {
	if plan.VmNicConfig.IsNull() || len(plan.VmNicConfig.Elements()) == 0 {
		return "", nil
	}

	var nicConfigs []vmNicConfigModel
	diags := plan.VmNicConfig.ElementsAs(ctx, &nicConfigs, false)
	if diags.HasError() {
		return "", fmt.Errorf("failed to parse vm_nic_config")
	}

	type vmNicParam struct {
		L3NetworkUuid     string `json:"l3NetworkUuid"`
		DriverType        string `json:"driverType,omitempty"`
		MultiQueueNum     string `json:"multiQueueNum,omitempty"`
		State             string `json:"state,omitempty"`
		InboundBandwidth  int64  `json:"inboundBandwidth,omitempty"`
		OutboundBandwidth int64  `json:"outboundBandwidth,omitempty"`
	}

	var nicParams []vmNicParam
	for _, nic := range nicConfigs {
		np := vmNicParam{
			L3NetworkUuid: nic.L3NetworkUuid.ValueString(),
		}

		if !nic.DriverType.IsNull() && nic.DriverType.ValueString() != "" {
			np.DriverType = nic.DriverType.ValueString()
		}

		// MultiQueueNum - use user-provided value or calculate from CPU count
		if !nic.MultiQueueNum.IsNull() && nic.MultiQueueNum.ValueString() != "" {
			np.MultiQueueNum = nic.MultiQueueNum.ValueString()
		} else if !plan.CPUNum.IsNull() {
			cpuNum := int(plan.CPUNum.ValueInt64())
			multiQueueNum := cpuNum
			if multiQueueNum > 12 {
				multiQueueNum = 12
			}
			np.MultiQueueNum = fmt.Sprintf("%d", multiQueueNum)
		}

		if !nic.State.IsNull() && nic.State.ValueString() != "" {
			np.State = nic.State.ValueString()
		}

		if !nic.InboundBandwidth.IsNull() {
			np.InboundBandwidth = nic.InboundBandwidth.ValueInt64()
		}

		if !nic.OutboundBandwidth.IsNull() {
			np.OutboundBandwidth = nic.OutboundBandwidth.ValueInt64()
		}

		nicParams = append(nicParams, np)
	}

	jsonBytes, err := json.Marshal(nicParams)
	if err != nil {
		return "", fmt.Errorf("failed to marshal vm_nic_params: %v", err)
	}

	return string(jsonBytes), nil
}

// buildSystemTags builds system tags from plan
func (r *vmResource) buildSystemTags(ctx context.Context, plan vmInstanceDataSourceModel) []string {
	systemTags := []string{}

	// UserData - base64 encode to handle special characters
	if !plan.UserData.IsNull() && plan.UserData.ValueString() != "" {
		encodedUserData := base64.StdEncoding.EncodeToString([]byte(plan.UserData.ValueString()))
		systemTags = append(systemTags, fmt.Sprintf("userdata::%s", encodedUserData))
	}

	// Hostname
	if !plan.Hostname.IsNull() && plan.Hostname.ValueString() != "" {
		systemTags = append(systemTags, fmt.Sprintf("hostname::%s", plan.Hostname.ValueString()))
	}

	// CPU cores (sockedNum)
	if !plan.SockedNum.IsNull() && plan.SockedNum.ValueInt64() > 0 {
		systemTags = append(systemTags, fmt.Sprintf("cpuCores::%d", plan.SockedNum.ValueInt64()))
	}

	// CPU quota
	if !plan.CpuQuota.IsNull() && plan.CpuQuota.ValueFloat64() > 0 {
		quota := int(plan.CpuQuota.ValueFloat64() * 10000)
		systemTags = append(systemTags, fmt.Sprintf("resourceConfig::kvm::vm.cpu.quota::%d", quota))
	}

	// Group/Directory
	if !plan.Group.IsNull() && plan.Group.ValueString() != "" && plan.Group.ValueString() != "-2" && plan.Group.ValueString() != "-1" {
		systemTags = append(systemTags, fmt.Sprintf("directoryUuid::%s", plan.Group.ValueString()))
	}

	// HA (NeverStop)
	if !plan.NeverStop.IsNull() && plan.NeverStop.ValueBool() {
		systemTags = append(systemTags, "ha::NeverStop")
	}

	// Total GPU memory
	if !plan.TotalGPUMemory.IsNull() && plan.TotalGPUMemory.ValueFloat64() > 0 {
		memory := int(plan.TotalGPUMemory.ValueFloat64() * 1024)
		systemTags = append(systemTags, fmt.Sprintf("qxlMemory::0::%d::0", memory))
	}

	// VM Priority (CPU and Memory resource levels)
	cpuLevel := "Normal"
	memLevel := "Normal"
	if !plan.CpuResourceLevel.IsNull() && plan.CpuResourceLevel.ValueString() != "" {
		cpuLevel = plan.CpuResourceLevel.ValueString()
	}
	if !plan.MemoryResourceLevel.IsNull() && plan.MemoryResourceLevel.ValueString() != "" {
		memLevel = plan.MemoryResourceLevel.ValueString()
	}
	if cpuLevel == "Normal" && memLevel == "Normal" {
		systemTags = append(systemTags, "vmPriority::Normal")
	} else if cpuLevel == "Normal" && memLevel == "High" {
		systemTags = append(systemTags, "vmPriority::MemoryHigh")
	} else if cpuLevel == "CpuHigh" && memLevel == "Normal" {
		systemTags = append(systemTags, "vmPriority::CpuHigh")
	} else if cpuLevel == "CpuHigh" && memLevel == "High" {
		systemTags = append(systemTags, "vmPriority::High")
	}

	// VM Group UUID (Scheduling group)
	if !plan.VmGroupUuid.IsNull() && plan.VmGroupUuid.ValueString() != "" {
		systemTags = append(systemTags, fmt.Sprintf("vmSchedulingRuleGroupUuid::%s", plan.VmGroupUuid.ValueString()))
	}

	// NUMA (hotPlug)
	if !plan.VnumaEnabled.IsNull() {
		if plan.VnumaEnabled.ValueBool() {
			systemTags = append(systemTags, "resourceConfig::vm::numa::true")
		} else {
			systemTags = append(systemTags, "resourceConfig::vm::numa::false")
		}
	}

	// GPU Type
	if !plan.GpuType.IsNull() && plan.GpuType.ValueString() != "" {
		systemTags = append(systemTags, fmt.Sprintf("resourceConfig::vm::videoType::%s", plan.GpuType.ValueString()))
	}

	// Sound Card
	if !plan.SoundCard.IsNull() && plan.SoundCard.ValueString() != "" {
		systemTags = append(systemTags, fmt.Sprintf("resourceConfig::vm::soundType::%s", plan.SoundCard.ValueString()))
	}

	// Motherboard Type (vmMachineType)
	if !plan.MotherboardType.IsNull() && plan.MotherboardType.ValueString() == "q35" {
		systemTags = append(systemTags, "vmMachineType::q35")
	}

	// CPU Mode
	if !plan.CpuMode.IsNull() && plan.CpuMode.ValueString() != "" {
		systemTags = append(systemTags, fmt.Sprintf("resourceConfig::kvm::vm.cpuMode::%s", plan.CpuMode.ValueString()))
	}

	// Boot Mode
	if !plan.BootMode.IsNull() && plan.BootMode.ValueString() != "" {
		systemTags = append(systemTags, fmt.Sprintf("bootMode::%s", plan.BootMode.ValueString()))
	}

	// USB Redirect
	if !plan.UsbRedirect.IsNull() && plan.UsbRedirect.ValueBool() {
		systemTags = append(systemTags, fmt.Sprintf("usbRedirect::%t", plan.UsbRedirect.ValueBool()))
	}

	// GPU Devices (vgpuDevice or gpuDeviceUuidList)
	if !plan.VgpuDevice.IsNull() {
		var vgpu vgpuDeviceModel
		diags := plan.VgpuDevice.As(ctx, &vgpu, basetypes.ObjectAsOptions{})
		if diags.HasError() == false && !vgpu.Uuid.IsNull() {
			deviceType := "pciDevice"
			if !vgpu.Type.IsNull() && vgpu.Type.ValueString() == "MdevDevice" {
				deviceType = "mdevDevice"
			}
			systemTags = append(systemTags, fmt.Sprintf("%s::%s", deviceType, vgpu.Uuid.ValueString()))
		}
	}

	if !plan.GpuDeviceUuidList.IsNull() && len(plan.GpuDeviceUuidList.Elements()) > 0 {
		var gpuUuids []string
		plan.GpuDeviceUuidList.ElementsAs(ctx, &gpuUuids, false)
		for _, uuid := range gpuUuids {
			systemTags = append(systemTags, fmt.Sprintf("pciDevice::%s", uuid))
		}
	}

	// CD-ROM List
	if !plan.CdromList.IsNull() && len(plan.CdromList.Elements()) > 0 {
		var cdroms []cdromModel
		plan.CdromList.ElementsAs(ctx, &cdroms, false)
		cdromStr := "cdroms"
		for _, cdrom := range cdroms {
			if !cdrom.IsoUuid.IsNull() && cdrom.IsoUuid.ValueString() != "" {
				cdromStr += fmt.Sprintf("::%s", cdrom.IsoUuid.ValueString())
			} else {
				cdromStr += "::Empty"
			}
		}
		// Fill remaining slots with None
		for i := len(cdroms); i < 3; i++ {
			cdromStr += "::None"
		}
		systemTags = append(systemTags, cdromStr)
	} else {
		systemTags = append(systemTags, "createWithoutCdRom::true")
	}

	// CPU Binding
	if !plan.CpuBindType.IsNull() && plan.CpuBindType.ValueString() == "structure" {
		if !plan.CpuBindListByVCpu.IsNull() && len(plan.CpuBindListByVCpu.Elements()) > 0 {
			var cpuBinds []cpuBindItemModel
			plan.CpuBindListByVCpu.ElementsAs(ctx, &cpuBinds, false)
			bindStr := ""
			for _, bind := range cpuBinds {
				if !bind.VCPU.IsNull() && !bind.PCPUList.IsNull() {
					var pcpus []string
					bind.PCPUList.ElementsAs(ctx, &pcpus, false)
					// Filter out NUMA node entries
					var filteredPcpus []string
					for _, pcpu := range pcpus {
						if !strings.Contains(pcpu, "NUMA node") {
							filteredPcpus = append(filteredPcpus, pcpu)
						}
					}
					if len(filteredPcpus) > 0 {
						bindStr += fmt.Sprintf("%s:%s;", bind.VCPU.ValueString(), strings.Join(filteredPcpus, ","))
					}
				}
			}
			if bindStr != "" {
				systemTags = append(systemTags, fmt.Sprintf("vmCpuPinning::%s", bindStr))
				systemTags = append(systemTags, "vmNumaEnable::true")
			}
		}
	}

	// Process vmNicConfig for additional system tags
	if !plan.VmNicConfig.IsNull() && len(plan.VmNicConfig.Elements()) > 0 {
		var nicConfigs []vmNicConfigModel
		plan.VmNicConfig.ElementsAs(ctx, &nicConfigs, false)
		for _, nic := range nicConfigs {
			l3uuid := nic.L3NetworkUuid.ValueString()

			// Custom MAC
			if !nic.CustomMac.IsNull() && nic.CustomMac.ValueString() != "" {
				systemTags = append(systemTags, fmt.Sprintf("customMac::%s::%s", l3uuid, nic.CustomMac.ValueString()))
			}

			// Static IP (IPv4)
			if !nic.StaticIp.IsNull() && nic.StaticIp.ValueString() != "" {
				ip := strings.ReplaceAll(nic.StaticIp.ValueString(), "::", "--")
				systemTags = append(systemTags, fmt.Sprintf("staticIp::%s::%s", l3uuid, ip))
			}

			// Static IP (IPv6)
			if !nic.StaticIpv6.IsNull() && nic.StaticIpv6.ValueString() != "" {
				ip := strings.ReplaceAll(nic.StaticIpv6.ValueString(), "::", "--")
				systemTags = append(systemTags, fmt.Sprintf("staticIp::%s::%s", l3uuid, ip))
			}

			// IPv4 Netmask
			if !nic.Ipv4Netmask.IsNull() && nic.Ipv4Netmask.ValueString() != "" {
				netmask := strings.ReplaceAll(nic.Ipv4Netmask.ValueString(), "::", "--")
				systemTags = append(systemTags, fmt.Sprintf("ipv4Netmask::%s::%s", l3uuid, netmask))
			}

			// IPv6 Prefix
			if !nic.Ipv6Prefix.IsNull() {
				systemTags = append(systemTags, fmt.Sprintf("ipv6Prefix::%s::%d", l3uuid, nic.Ipv6Prefix.ValueInt64()))
			}

			// IPv4 Gateway
			if !nic.Ipv4Gateway.IsNull() && nic.Ipv4Gateway.ValueString() != "" {
				gateway := strings.ReplaceAll(nic.Ipv4Gateway.ValueString(), "::", "--")
				systemTags = append(systemTags, fmt.Sprintf("ipv4Gateway::%s::%s", l3uuid, gateway))
			}

			// IPv6 Gateway
			if !nic.Ipv6Gateway.IsNull() && nic.Ipv6Gateway.ValueString() != "" {
				gateway := strings.ReplaceAll(nic.Ipv6Gateway.ValueString(), "::", "--")
				systemTags = append(systemTags, fmt.Sprintf("ipv6Gateway::%s::%s", l3uuid, gateway))
			}
		}
	}

	// Console Password
	if !plan.ConsolePassword.IsNull() && plan.ConsolePassword.ValueString() != "" {
		systemTags = append(systemTags, fmt.Sprintf("consolePassword::%s", plan.ConsolePassword.ValueString()))
	}

	// SSH Key
	if !plan.SshKey.IsNull() && plan.SshKey.ValueString() != "" {
		systemTags = append(systemTags, fmt.Sprintf("sshkey::%s", plan.SshKey.ValueString()))
	}

	// BIOS Time Sync
	if !plan.BiosTimeSync.IsNull() && plan.BiosTimeSync.ValueBool() {
		systemTags = append(systemTags, "resourceConfig::vm::vm.clock.track::guest")
	}

	// Clock Track
	if !plan.ClockTrack.IsNull() && plan.ClockTrack.ValueString() != "" {
		systemTags = append(systemTags, fmt.Sprintf("resourceConfig::vm::vm.clock.track::%s", plan.ClockTrack.ValueString()))
	}

	// Security Element
	if !plan.Se.IsNull() && plan.Se.ValueBool() {
		systemTags = append(systemTags, "securityElementEnable::true")
	}

	// Spice Streaming Mode
	if !plan.SpiceStreamingMode.IsNull() && plan.SpiceStreamingMode.ValueString() != "" {
		systemTags = append(systemTags, fmt.Sprintf("resourceConfig::vm::spiceStreamingMode::%s", plan.SpiceStreamingMode.ValueString()))
	}

	// Fault Strategy
	if !plan.FaultStrategy.IsNull() && plan.FaultStrategy.ValueString() != "" {
		systemTags = append(systemTags, fmt.Sprintf("resourceConfig::vm::crash.strategy::%s", plan.FaultStrategy.ValueString()))
	}

	// Emulate Hyper-V
	if !plan.EmulateHyperV.IsNull() && plan.EmulateHyperV.ValueBool() {
		systemTags = append(systemTags, "resourceConfig::vm::emulateHyperV::true")
	}

	// VM Port Off
	if !plan.VmPortOff.IsNull() && plan.VmPortOff.ValueBool() {
		systemTags = append(systemTags, "resourceConfig::vm::vmPortOff::true")
	}

	// Boot Menu Splash Timeout
	if !plan.BootMenuSplashTimeout.IsNull() && plan.BootMenuSplashTimeout.ValueString() != "" {
		systemTags = append(systemTags, fmt.Sprintf("resourceConfig::vm::bootMenuSplashTimeout::%s", plan.BootMenuSplashTimeout.ValueString()))
	}

	// CPU Hide KVM Mark
	if !plan.CpuHideKVMMark.IsNull() && plan.CpuHideKVMMark.ValueBool() {
		systemTags = append(systemTags, "resourceConfig::kvm::vm.cpu.hypervisor.feature::true")
	}

	// Migrate Auto Converge
	if !plan.MigrateAutoConverge.IsNull() && plan.MigrateAutoConverge.ValueBool() {
		systemTags = append(systemTags, "resourceConfig::kvm::migrate.autoConverge::true")
	}

	// Hot Plug Enabled
	if !plan.HotPlugEnabled.IsNull() && plan.HotPlugEnabled.ValueBool() {
		systemTags = append(systemTags, "resourceConfig::pciDevice::hotPlugEnabled::true")
	}

	// VM CPUID Vendor
	if !plan.VmCpuidVendor.IsNull() && plan.VmCpuidVendor.ValueString() != "" {
		systemTags = append(systemTags, fmt.Sprintf("resourceConfig::vm::vm.cpuid.vendor::%s", plan.VmCpuidVendor.ValueString()))
	}

	// Affinity Group
	if !plan.AffinityGroupUuid.IsNull() && plan.AffinityGroupUuid.ValueString() != "" {
		systemTags = append(systemTags, fmt.Sprintf("affinityGroupUuid::%s", plan.AffinityGroupUuid.ValueString()))
	}

	// HA Stick Strategy
	if !plan.HaStickStrategy.IsNull() && !plan.HaStickStrategy.ValueBool() && !plan.ClusterUuid.IsNull() {
		systemTags = append(systemTags, fmt.Sprintf("resourceBindings::Cluster:%s", plan.ClusterUuid.ValueString()))
	}

	return systemTags
}

// buildNetworkConfig builds network configuration from plan
func (r *vmResource) buildNetworkConfig(ctx context.Context, plan vmInstanceDataSourceModel, resp *resource.CreateResponse) ([]string, string, []NetworkInterfaceModel, error) {
	var inputNics []NetworkInterfaceModel
	if !plan.NetworkInterfaces.IsNull() && len(plan.NetworkInterfaces.Elements()) > 0 {
		diags := plan.NetworkInterfaces.ElementsAs(ctx, &inputNics, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return nil, "", nil, fmt.Errorf("failed to parse network_interfaces")
		}
	}

	var l3NetworkUuids []string
	var defaultL3Uuid string
	var createNics []NetworkInterfaceModel

	for _, nic := range inputNics {
		l3uuid := nic.L3NetworkUuid.ValueString()
		l3NetworkUuids = append(l3NetworkUuids, l3uuid)

		if nic.DefaultL3.ValueBool() {
			defaultL3Uuid = l3uuid
		}

		var staticIp types.String
		if nic.StaticIp.IsNull() || nic.StaticIp.ValueString() == "" {
			staticIp = types.StringNull()
		} else {
			staticIp = nic.StaticIp
		}

		createNics = append(createNics, NetworkInterfaceModel{
			L3NetworkUuid: nic.L3NetworkUuid,
			DefaultL3:     nic.DefaultL3,
			StaticIp:      staticIp,
		})
	}

	return l3NetworkUuids, defaultL3Uuid, createNics, nil
}

// buildNetworkConfigFromTemplate builds network configuration from template VM and user input
// Network interfaces are inherited from template, but user can override with same L3 UUID
// Additional network interfaces can be added by user
func (r *vmResource) buildNetworkConfigFromTemplate(ctx context.Context, plan vmInstanceDataSourceModel, templateVm *view.VmInstanceInventoryView, resp *resource.CreateResponse) ([]string, string, []NetworkInterfaceModel, error) {
	// Get user-specified nics
	var userNics []NetworkInterfaceModel
	if !plan.NetworkInterfaces.IsNull() && len(plan.NetworkInterfaces.Elements()) > 0 {
		diags := plan.NetworkInterfaces.ElementsAs(ctx, &userNics, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return nil, "", nil, fmt.Errorf("failed to parse network_interfaces")
		}
	}

	// Build user nics map for quick lookup
	userNicsMap := make(map[string]NetworkInterfaceModel)
	for _, nic := range userNics {
		userNicsMap[nic.L3NetworkUuid.ValueString()] = nic
	}

	var l3NetworkUuids []string
	var defaultL3Uuid string
	var createNics []NetworkInterfaceModel

	// Process template VM nics first
	for _, vmNic := range templateVm.VmNics {
		l3uuid := vmNic.L3NetworkUuid

		// Check if user specified this nic
		if userNic, ok := userNicsMap[l3uuid]; ok {
			// Use user-specified config (override)
			l3NetworkUuids = append(l3NetworkUuids, l3uuid)
			if userNic.DefaultL3.ValueBool() {
				defaultL3Uuid = l3uuid
			}
			createNics = append(createNics, NetworkInterfaceModel{
				L3NetworkUuid: userNic.L3NetworkUuid,
				DefaultL3:     userNic.DefaultL3,
				StaticIp:      userNic.StaticIp,
			})
		} else {
			// Use template nic config
			l3NetworkUuids = append(l3NetworkUuids, l3uuid)
			isDefault := templateVm.DefaultL3NetworkUuid != "" && l3uuid == templateVm.DefaultL3NetworkUuid
			createNics = append(createNics, NetworkInterfaceModel{
				L3NetworkUuid: types.StringValue(l3uuid),
				DefaultL3:     types.BoolValue(isDefault),
				StaticIp:      types.StringNull(),
			})
		}
	}

	// Add user-specified nics that are not in template
	for _, userNic := range userNics {
		l3uuid := userNic.L3NetworkUuid.ValueString()
		found := false
		for _, vmNic := range templateVm.VmNics {
			if vmNic.L3NetworkUuid == l3uuid {
				found = true
				break
			}
		}
		if !found {
			l3NetworkUuids = append(l3NetworkUuids, l3uuid)
			if userNic.DefaultL3.ValueBool() {
				defaultL3Uuid = l3uuid
			}
			createNics = append(createNics, NetworkInterfaceModel{
				L3NetworkUuid: userNic.L3NetworkUuid,
				DefaultL3:     userNic.DefaultL3,
				StaticIp:      userNic.StaticIp,
			})
		}
	}

	return l3NetworkUuids, defaultL3Uuid, createNics, nil
}

// validateImage validates the image before creation
func (r *vmResource) validateImage(ctx context.Context, imageUuid string, resp *resource.CreateResponse) error {
	image, err := r.client.GetImage(ctx, imageUuid)
	if err != nil {
		resp.Diagnostics.AddError("Params Error", fmt.Sprintf("failed to find image %s: %v", imageUuid, err))
		return err
	}

	if image.Status != "Ready" {
		resp.Diagnostics.AddError("Params Error", fmt.Sprintf("image %s status is %s, not Ready", imageUuid, image.Status))
		return fmt.Errorf("image not ready")
	}

	if image.State != "Enabled" {
		resp.Diagnostics.AddError("Params Error", fmt.Sprintf("image %s state is %s, not Enabled", imageUuid, image.State))
		return fmt.Errorf("image not enabled")
	}

	return nil
}

// handlePostCreation handles common post-creation processing
// fromImage: true if creating from image, false if creating from template
func (r *vmResource) handlePostCreation(ctx context.Context, plan vmInstanceDataSourceModel, instance *view.VmInstanceInventoryView, createNics []NetworkInterfaceModel, fromImage bool, resp *resource.CreateResponse) {
	plan.Uuid = types.StringValue(instance.UUID)
	plan.Name = types.StringValue(instance.Name)
	plan.Description = types.StringValue(instance.Description)
	plan.MemorySize = types.Int64Value(utils.BytesToMB(instance.MemorySize))
	plan.CPUNum = types.Int64Value(int64(instance.CpuNum))

	// Set fields that may be inherited from template
	if instance.Platform != "" {
		plan.Platform = types.StringValue(instance.Platform)
	}
	if instance.GuestOsType != "" {
		plan.GuestOsType = types.StringValue(instance.GuestOsType)
	}
	if instance.ImageUuid != "" {
		plan.ImageUuid = types.StringValue(instance.ImageUuid)
	}

	networkInterfaceAttrTypes := map[string]attr.Type{
		"port_group_uuid": types.StringType,
		"default_l3":      types.BoolType,
		"static_ip":       types.StringType,
	}

	if fromImage {
		// For image creation: preserve user-specified values from createNics
		// Map to track which L3 UUIDs we've processed
		processedNics := make(map[string]bool)
		var updatedNics []NetworkInterfaceModel

		// First, add all nics from createNics (user-specified values)
		for _, createNic := range createNics {
			l3uuid := createNic.L3NetworkUuid.ValueString()
			processedNics[l3uuid] = true
			updatedNics = append(updatedNics, createNic)
		}

		// Then, add any nics from API response that weren't in createNics
		for _, vmNic := range instance.VmNics {
			l3uuid := vmNic.L3NetworkUuid
			if !processedNics[l3uuid] {
				updatedNics = append(updatedNics, NetworkInterfaceModel{
					L3NetworkUuid: types.StringValue(l3uuid),
					DefaultL3:     types.BoolValue(instance.DefaultL3NetworkUuid != "" && l3uuid == instance.DefaultL3NetworkUuid),
					StaticIp:      types.StringNull(),
				})
			}
		}

		networkInterfacesList, diags := types.ListValueFrom(ctx,
			types.ObjectType{AttrTypes: networkInterfaceAttrTypes},
			updatedNics)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		plan.NetworkInterfaces = networkInterfacesList
	} else {
		// For template creation: build from API response (inherited from template)
		var updatedNics []NetworkInterfaceModel
		for _, vmNic := range instance.VmNics {
			updatedNics = append(updatedNics, NetworkInterfaceModel{
				L3NetworkUuid: types.StringValue(vmNic.L3NetworkUuid),
				DefaultL3:     types.BoolValue(instance.DefaultL3NetworkUuid != "" && vmNic.L3NetworkUuid == instance.DefaultL3NetworkUuid),
				StaticIp:      types.StringValue(vmNic.Ip),
			})
		}

		networkInterfacesList, diags := types.ListValueFrom(ctx,
			types.ObjectType{AttrTypes: networkInterfaceAttrTypes},
			updatedNics)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		plan.NetworkInterfaces = networkInterfacesList
	}

	var diskModelAttrTypes = map[string]attr.Type{
		"size":                 types.Int64Type,
		"primary_storage_uuid": types.StringType,
		"name":                 types.StringType,
		"boot":                 types.BoolType,
		"bus_type":             types.StringType,
	}

	if fromImage {
		// For image creation: preserve bus_type from plan since API response doesn't include it
		// Build root_disk
		var rootDiskPlan diskModel
		var planBusType types.String
		if !plan.RootDisk.IsNull() {
			plan.RootDisk.As(ctx, &rootDiskPlan, basetypes.ObjectAsOptions{})
			planBusType = rootDiskPlan.BusType
		}
		for _, disk := range instance.AllVolumes {
			if disk.Type == "Root" {
				rootDiskPlan.PrimaryStorageUuid = types.StringValue(disk.PrimaryStorageUuid)
				rootDiskPlan.Size = types.Int64Value(disk.Size)
				rootDiskPlan.Name = types.StringValue(disk.Name)
				rootDiskPlan.Boot = types.BoolValue(true)
				// Preserve bus_type from plan, use default "virtio" if not specified
				if !planBusType.IsNull() && planBusType.ValueString() != "" {
					rootDiskPlan.BusType = planBusType
				} else {
					rootDiskPlan.BusType = types.StringValue("virtio")
				}
				break
			}
		}

		rootDiskObj, diags := types.ObjectValueFrom(ctx, diskModelAttrTypes, rootDiskPlan)
		resp.Diagnostics.Append(diags...)
		if !resp.Diagnostics.HasError() {
			plan.RootDisk = rootDiskObj
		}

		// Build data_disks - auto-generate names and match bus_type by index
		var planDataDisks []diskModel
		if !plan.DataDisks.IsNull() {
			plan.DataDisks.ElementsAs(ctx, &planDataDisks, false)
		}

		// Collect all data disks from API response into a map by name
		// API returns disks with names: {vm-name}-1, {vm-name}-2, etc. (matching frontend format)
		apiDataDisksMap := make(map[string]view.VolumeInventoryView)
		for _, disk := range instance.AllVolumes {
			if disk.Type == "Data" {
				apiDataDisksMap[disk.Name] = disk
			}
		}

		// Build data disks in the same order as plan, using plan's expected names
		var dataDisksPlan []diskModel
		for i := range planDataDisks {
			// Generate the expected disk name based on index: {vm-name}-1, {vm-name}-2, etc.
			// This matches the frontend naming convention
			expectedName := fmt.Sprintf("%s-%d", plan.Name.ValueString(), i+1)

			// Get the disk from API response by name
			disk, _ := apiDataDisksMap[expectedName]

			// Match bus_type by index from plan
			busType := planDataDisks[i].BusType.ValueString()

			diskModel := diskModel{
				PrimaryStorageUuid: types.StringValue(disk.PrimaryStorageUuid),
				Size:               types.Int64Value(disk.Size),
				Name:               types.StringValue(disk.Name),
				Boot:               types.BoolValue(false),
				BusType:            types.StringValue(busType),
			}
			dataDisksPlan = append(dataDisksPlan, diskModel)
		}

		dataDisksList, diags := types.ListValueFrom(ctx, types.ObjectType{
			AttrTypes: diskModelAttrTypes,
		}, dataDisksPlan)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		plan.DataDisks = dataDisksList
	} else {
		// For template creation: build from API response (inherited from template)
		var rootDiskPlan diskModel
		for _, disk := range instance.AllVolumes {
			if disk.Type == "Root" {
				rootDiskPlan.PrimaryStorageUuid = types.StringValue(disk.PrimaryStorageUuid)
				rootDiskPlan.Size = types.Int64Value(disk.Size)
				rootDiskPlan.Name = types.StringValue(disk.Name)
				rootDiskPlan.Boot = types.BoolValue(true)
				break
			}
		}

		rootDiskObj, diags := types.ObjectValueFrom(ctx, diskModelAttrTypes, rootDiskPlan)
		resp.Diagnostics.Append(diags...)
		if !resp.Diagnostics.HasError() {
			plan.RootDisk = rootDiskObj
		}

		var dataDisksPlan []diskModel
		for _, disk := range instance.AllVolumes {
			if disk.Type == "Data" {
				dataDisksPlan = append(dataDisksPlan, diskModel{
					PrimaryStorageUuid: types.StringValue(disk.PrimaryStorageUuid),
					Size:               types.Int64Value(disk.Size),
					Name:               types.StringValue(disk.Name),
					Boot:               types.BoolValue(false),
				})
			}
		}

		dataDisksList, diags := types.ListValueFrom(ctx, types.ObjectType{
			AttrTypes: diskModelAttrTypes,
		}, dataDisksPlan)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		plan.DataDisks = dataDisksList
	}

	var vmNics []NicsModel
	for _, nic := range instance.VmNics {
		vmNics = append(vmNics, NicsModel{
			Uuid:    types.StringValue(nic.UUID),
			Ip:      types.StringValue(nic.Ip),
			Netmask: types.StringValue(nic.Netmask),
			Gateway: types.StringValue(nic.Gateway),
		})
	}

	plan.VMNics, _ = types.ListValueFrom(ctx, types.ObjectType{AttrTypes: networkModelAttrTypes}, vmNics)

	// Provider enhancement: Set Hostname after VM creation (requires VMtools)
	// Note: This is done asynchronously and may fail if VMtools is not ready
	if !plan.Hostname.IsNull() && plan.Hostname.ValueString() != "" {
		hostname := plan.Hostname.ValueString()
		tflog.Info(ctx, fmt.Sprintf("Setting hostname %s for VM %s", hostname, instance.UUID))

		setHostnameParam := param.SetVmHostnameParam{
			Params: param.SetVmHostnameParamDetail{
				Hostname: hostname,
			},
		}

		_, err := r.client.SetVmHostname(ctx, instance.UUID, setHostnameParam)
		if err != nil {
			// Don't fail the whole creation, just warn about hostname setting failure
			tflog.Warn(ctx, fmt.Sprintf("Failed to set hostname for VM %s: %v. This is expected if VMtools is not running yet.", instance.UUID, err))
		} else {
			tflog.Info(ctx, fmt.Sprintf("Successfully set hostname %s for VM %s", hostname, instance.UUID))
		}
	}

	// Provider enhancement: Set Static IP with Netmask and Gateway using API
	// This is more reliable than system tags for some configurations
	if len(createNics) > 0 && !plan.Netmask.IsNull() && plan.Netmask.ValueString() != "" {
		l3uuid := createNics[0].L3NetworkUuid.ValueString()
		vmNicUuid := ""
		for _, nic := range instance.VmNics {
			if nic.L3NetworkUuid == l3uuid {
				vmNicUuid = nic.UUID
				break
			}
		}

		if vmNicUuid != "" && !plan.Gateway.IsNull() && plan.Gateway.ValueString() != "" {
			staticIP := ""
			if !createNics[0].StaticIp.IsNull() && createNics[0].StaticIp.ValueString() != "" {
				staticIP = createNics[0].StaticIp.ValueString()
			}

			if staticIP != "" {
				netmask := plan.Netmask.ValueString()
				gateway := plan.Gateway.ValueString()

				setStaticIpParam := param.SetVmStaticIpParam{
					Params: param.SetVmStaticIpParamDetail{
						L3NetworkUuid: l3uuid,
						Ip:            &staticIP,
						Netmask:       &netmask,
						Gateway:       &gateway,
					},
				}

				tflog.Info(ctx, fmt.Sprintf("Setting static IP for VM %s: IP=%s, Netmask=%s, Gateway=%s", instance.UUID, staticIP, netmask, gateway))
				_, err := r.client.SetVmStaticIp(ctx, instance.UUID, setStaticIpParam)
				if err != nil {
					tflog.Warn(ctx, fmt.Sprintf("Failed to set static IP for VM %s: %v", instance.UUID, err))
				} else {
					// Apply network config
					updateNetworkParam := param.UpdateVmNetworkConfigParam{
						Params: param.UpdateVmNetworkConfigParamDetail{
							VmNicUuids: []string{vmNicUuid},
						},
					}
					_, err = r.client.UpdateVmNetworkConfig(ctx, instance.UUID, updateNetworkParam)
					if err != nil {
						tflog.Warn(ctx, fmt.Sprintf("Failed to apply network config for VM %s: %v", instance.UUID, err))
					} else {
						tflog.Info(ctx, fmt.Sprintf("Successfully configured network for VM %s", instance.UUID))
					}
				}
			}
		}
	}

	setDiags := resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(setDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

}

// executePostCreationTasks executes post-creation configuration tasks for image-based creation
func (r *vmResource) executePostCreationTasks(ctx context.Context, plan vmInstanceDataSourceModel, instance *view.VmInstanceInventoryView, resp *resource.CreateResponse) {
	// Use errgroup for concurrent execution of independent tasks
	g, ctx := errgroup.WithContext(ctx)

	// 1. Set VDI Monitor Number
	if !plan.VdiMonitorNumber.IsNull() && plan.VdiMonitorNumber.ValueInt64() > 0 {
		monitorNumber := int(plan.VdiMonitorNumber.ValueInt64())
		g.Go(func() error {
			tflog.Info(ctx, fmt.Sprintf("Setting VDI monitor number to %d for VM %s", monitorNumber, instance.UUID))
			_, err := r.client.SetVmMonitorNumber(ctx, instance.UUID, param.SetVmMonitorNumberParam{
				Params: param.SetVmMonitorNumberParamDetail{
					MonitorNumber: monitorNumber,
				},
			})
			if err != nil {
				tflog.Warn(ctx, fmt.Sprintf("Failed to set VDI monitor number for VM %s: %v", instance.UUID, err))
			}
			return nil
		})
	}

	// 2. Set Boot Order
	if !plan.BootOrders.IsNull() && len(plan.BootOrders.Elements()) > 0 {
		var bootOrders []string
		plan.BootOrders.ElementsAs(ctx, &bootOrders, false)
		var filteredBootOrders []string
		for _, order := range bootOrders {
			if order != "Empty" {
				filteredBootOrders = append(filteredBootOrders, order)
			}
		}
		if len(filteredBootOrders) > 0 {
			g.Go(func() error {
				tflog.Info(ctx, fmt.Sprintf("Setting boot order to %v for VM %s", filteredBootOrders, instance.UUID))
				_, err := r.client.SetVmBootOrder(ctx, instance.UUID, param.SetVmBootOrderParam{
					Params: param.SetVmBootOrderParamDetail{
						BootOrder: filteredBootOrders,
					},
				})
				if err != nil {
					tflog.Warn(ctx, fmt.Sprintf("Failed to set boot order for VM %s: %v", instance.UUID, err))
				}
				return nil
			})
		}
	}

	// 3. Set NUMA (only if explicitly enabled)
	if !plan.VnumaEnabled.IsNull() && plan.VnumaEnabled.ValueBool() {
		g.Go(func() error {
			tflog.Info(ctx, fmt.Sprintf("Enabling NUMA for VM %s", instance.UUID))
			_, err := r.client.SetVmNuma(ctx, instance.UUID, param.SetVmNumaParam{
				Params: param.SetVmNumaParamDetail{
					Enable: true,
				},
			})
			if err != nil {
				tflog.Warn(ctx, fmt.Sprintf("Failed to set NUMA for VM %s: %v", instance.UUID, err))
			}
			return nil
		})
	}

	// 4. Set Emulator Pinning
	if !plan.EmulatorPinning.IsNull() && plan.EmulatorPinning.ValueString() != "" {
		emulatorPinning := plan.EmulatorPinning.ValueString()
		g.Go(func() error {
			tflog.Info(ctx, fmt.Sprintf("Setting emulator pinning to %s for VM %s", emulatorPinning, instance.UUID))
			_, err := r.client.SetVmEmulatorPinning(ctx, instance.UUID, param.SetVmEmulatorPinningParam{
				Params: param.SetVmEmulatorPinningParamDetail{
					EmulatorPinning: emulatorPinning,
				},
			})
			if err != nil {
				tflog.Warn(ctx, fmt.Sprintf("Failed to set emulator pinning for VM %s: %v", instance.UUID, err))
			}
			return nil
		})
	}

	// Wait for all concurrent tasks to complete
	if err := g.Wait(); err != nil {
		tflog.Warn(ctx, fmt.Sprintf("Error during post-creation tasks for VM %s: %v", instance.UUID, err))
	}

	// 5. Attach USB Devices (sequential)
	if !plan.VmUSBConfig.IsNull() && len(plan.VmUSBConfig.Elements()) > 0 {
		var usbConfigs []usbConfigModel
		plan.VmUSBConfig.ElementsAs(ctx, &usbConfigs, false)
		for _, usb := range usbConfigs {
			if !usb.UsbDeviceUuid.IsNull() && usb.UsbDeviceUuid.ValueString() != "" {
				usbDeviceUuid := usb.UsbDeviceUuid.ValueString()
				tflog.Info(ctx, fmt.Sprintf("Attaching USB device %s to VM %s", usbDeviceUuid, instance.UUID))
				_, err := r.client.AttachUsbDeviceToVm(ctx, usbDeviceUuid, param.AttachUsbDeviceToVmParam{
					Params: param.AttachUsbDeviceToVmParamDetail{
						VmInstanceUuid: instance.UUID,
						AttachType:     usb.AttachType.ValueStringPointer(),
					},
				})
				if err != nil {
					tflog.Warn(ctx, fmt.Sprintf("Failed to attach USB device %s to VM %s: %v", usbDeviceUuid, instance.UUID, err))
				}
			}
		}
	}

	// 6. Set Clock Track (if specified)
	if !plan.ClockTrack.IsNull() && plan.ClockTrack.ValueString() != "" {
		track := plan.ClockTrack.ValueString()
		clockParam := param.SetVmClockTrackParam{
			Params: param.SetVmClockTrackParamDetail{
				Track: track,
			},
		}
		if !plan.ClockSyncAfterResume.IsNull() {
			syncAfterResume := plan.ClockSyncAfterResume.ValueBool()
			clockParam.Params.SyncAfterVMResume = &syncAfterResume
		}
		if !plan.ClockSyncInterval.IsNull() {
			interval := int(plan.ClockSyncInterval.ValueInt64())
			clockParam.Params.IntervalInSeconds = &interval
		}
		tflog.Info(ctx, fmt.Sprintf("Setting clock track to %s for VM %s", track, instance.UUID))
		_, err := r.client.SetVmClockTrack(ctx, instance.UUID, clockParam)
		if err != nil {
			tflog.Warn(ctx, fmt.Sprintf("Failed to set clock track for VM %s: %v", instance.UUID, err))
		}
	}
}

// executePostCreationTasksForTemplate executes post-creation configuration tasks for template-based creation
// This is different from image-based creation as per frontend implementation
func (r *vmResource) executePostCreationTasksForTemplate(ctx context.Context, plan vmInstanceDataSourceModel, instance *view.VmInstanceInventoryView, resp *resource.CreateResponse) {
	// Use errgroup for concurrent execution of independent tasks
	g, ctx := errgroup.WithContext(ctx)

	// 1. Set Emulator Pinning (only for template creation)
	if !plan.EmulatorPinning.IsNull() && plan.EmulatorPinning.ValueString() != "" {
		emulatorPinning := plan.EmulatorPinning.ValueString()
		g.Go(func() error {
			tflog.Info(ctx, fmt.Sprintf("Setting emulator pinning to %s for VM %s", emulatorPinning, instance.UUID))
			_, err := r.client.SetVmEmulatorPinning(ctx, instance.UUID, param.SetVmEmulatorPinningParam{
				Params: param.SetVmEmulatorPinningParamDetail{
					EmulatorPinning: emulatorPinning,
				},
			})
			if err != nil {
				tflog.Warn(ctx, fmt.Sprintf("Failed to set emulator pinning for VM %s: %v", instance.UUID, err))
			}
			return nil
		})
	}

	// Wait for all concurrent tasks to complete
	if err := g.Wait(); err != nil {
		tflog.Warn(ctx, fmt.Sprintf("Error during post-creation tasks for VM %s: %v", instance.UUID, err))
	}

	// 2. Attach USB Devices (sequential, for single VM creation only - matching frontend)
	if !plan.VmUSBConfig.IsNull() && len(plan.VmUSBConfig.Elements()) > 0 {
		var usbConfigs []usbConfigModel
		plan.VmUSBConfig.ElementsAs(ctx, &usbConfigs, false)
		for _, usb := range usbConfigs {
			if !usb.UsbDeviceUuid.IsNull() && usb.UsbDeviceUuid.ValueString() != "" {
				usbDeviceUuid := usb.UsbDeviceUuid.ValueString()
				tflog.Info(ctx, fmt.Sprintf("Attaching USB device %s to VM %s", usbDeviceUuid, instance.UUID))
				_, err := r.client.AttachUsbDeviceToVm(ctx, usbDeviceUuid, param.AttachUsbDeviceToVmParam{
					Params: param.AttachUsbDeviceToVmParamDetail{
						VmInstanceUuid: instance.UUID,
						AttachType:     usb.AttachType.ValueStringPointer(),
					},
				})
				if err != nil {
					tflog.Warn(ctx, fmt.Sprintf("Failed to attach USB device %s to VM %s: %v", usbDeviceUuid, instance.UUID, err))
				}
			}
		}
	}
}

// Read implements resource.Resource.
func (r *vmResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state vmInstanceDataSourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	vm, err := r.client.GetVmInstance(ctx, state.Uuid.ValueString())
	if err != nil {
		tflog.Warn(ctx, "cannot read vm, maybe it has been deleted, set uuid to 'empty'. vm was no longer managed by terraform. error: "+err.Error())
		state.Uuid = types.StringValue("")
		diags = resp.State.Set(ctx, &state)
		resp.Diagnostics.Append(diags...)
		return
	}

	state.Uuid = types.StringValue(vm.UUID)
	state.Name = types.StringValue(vm.Name)
	state.Description = types.StringValue(vm.Description)
	state.ImageUuid = types.StringValue(vm.ImageUuid)
	state.MemorySize = types.Int64Value(utils.BytesToMB(vm.MemorySize))
	state.CPUNum = types.Int64Value(int64(vm.CpuNum))

	var vmNics []NicsModel
	for _, nic := range vm.VmNics {
		vmNics = append(vmNics, NicsModel{
			Uuid:    types.StringValue(nic.UUID),
			Ip:      types.StringValue(nic.Ip),
			Netmask: types.StringValue(nic.Netmask),
			Gateway: types.StringValue(nic.Gateway),
		})
	}

	state.VMNics, _ = types.ListValueFrom(ctx, types.ObjectType{AttrTypes: networkModelAttrTypes}, vmNics)
	resp.Diagnostics.Append(diags...)

	var networkInterfaces []NetworkInterfaceModel
	originalNetworkInterfaces := make(map[string]NetworkInterfaceModel)
	if !state.NetworkInterfaces.IsNull() && len(state.NetworkInterfaces.Elements()) > 0 {
		var oldNics []NetworkInterfaceModel
		diags := state.NetworkInterfaces.ElementsAs(ctx, &oldNics, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		for _, oldNic := range oldNics {
			originalNetworkInterfaces[oldNic.L3NetworkUuid.ValueString()] = oldNic
		}
	}

	for _, nic := range vm.VmNics {
		staticIP := types.StringNull()
		defaultL3 := types.BoolValue(vm.DefaultL3NetworkUuid != "" && nic.L3NetworkUuid == vm.DefaultL3NetworkUuid)

		// Preserve original configuration if available
		if originalNic, ok := originalNetworkInterfaces[nic.L3NetworkUuid]; ok {
			// Always preserve static IP from original configuration
			// This ensures the value set by user is maintained
			staticIP = originalNic.StaticIp
			// Preserve original default_l3 value
			defaultL3 = originalNic.DefaultL3
		}

		networkInterfaces = append(networkInterfaces, NetworkInterfaceModel{
			L3NetworkUuid: types.StringValue(nic.L3NetworkUuid),
			DefaultL3:     defaultL3,
			StaticIp:      staticIP,
		})
	}
	state.NetworkInterfaces, _ = types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"port_group_uuid": types.StringType,
			"default_l3":      types.BoolType,
			"static_ip":       types.StringType,
		},
	}, networkInterfaces)

	// Read root_disk and data_disks from VM volumes
	var diskModelAttrTypes = map[string]attr.Type{
		"size":                 types.Int64Type,
		"primary_storage_uuid": types.StringType,
		"name":                 types.StringType,
		"boot":                 types.BoolType,
		"bus_type":             types.StringType,
	}

	// Get volumes for this VM
	queryParam := param.NewQueryParam()
	queryParam.AddQ(fmt.Sprintf("vmInstanceUuid=%s", vm.UUID))
	volumes, err := r.client.QueryVolume(ctx, &queryParam)
	if err == nil && len(volumes) > 0 {
		var rootDisk diskModel
		var dataDisks []diskModel

		for _, vol := range volumes {
			// Use size in bytes directly
			systemTagsQParm := param.NewQueryParam()
			systemTagsQParm.AddQ(fmt.Sprintf("resourceUuid=%s", vol.UUID))
			systemTagsQParm.AddQ("resourceType=VolumeVO")
			systemTags, err := r.client.QuerySystemTag(ctx, &systemTagsQParm)
			tflog.Info(ctx, fmt.Sprintf("Volume %s systemTags: %+v", vol.UUID, systemTags))
			if err != nil {
				tflog.Warn(ctx, fmt.Sprintf("Failed to query system tags for volume %s: %v", vol.UUID, err))
				continue
			}
			// Detect bus type from system tags
			busType := "virtio" // default bus type
			for _, systemTag := range systemTags {
				if systemTag.Tag == "capability::virtio-scsi" {
					busType = "virtio-scsi"
					break
				} else if systemTag.Tag == "capability::scsi" {
					busType = "scsi"
					break
				}
			}
			switch vol.Type {
			case "Root":
				rootDisk = diskModel{
					Size:               types.Int64Value(vol.Size),
					PrimaryStorageUuid: types.StringValue(vol.PrimaryStorageUuid),
					Name:               types.StringValue(vol.Name),
					Boot:               types.BoolValue(true),
					BusType:            types.StringValue(busType),
				}
			case "Data":
				dataDisk := diskModel{
					Size:               types.Int64Value(vol.Size),
					PrimaryStorageUuid: types.StringValue(vol.PrimaryStorageUuid),
					Name:               types.StringValue(vol.Name),
					Boot:               types.BoolValue(false),
					BusType:            types.StringValue(busType),
				}
				dataDisks = append(dataDisks, dataDisk)
			}
		}

		// Set root_disk
		if rootDisk.Size.ValueInt64() > 0 {
			rootDiskObj, diags := types.ObjectValueFrom(ctx, diskModelAttrTypes, rootDisk)
			resp.Diagnostics.Append(diags...)
			if !resp.Diagnostics.HasError() {
				state.RootDisk = rootDiskObj
			}
		}

		// Set data_disks - reorder according to existing state order
		if len(dataDisks) > 0 {
			// Get existing data disks from state to preserve order
			var existingDataDisks []diskModel
			if !state.DataDisks.IsNull() && len(state.DataDisks.Elements()) > 0 {
				state.DataDisks.ElementsAs(ctx, &existingDataDisks, false)
			}

			// Build a map of data disks by name
			dataDisksByName := make(map[string]diskModel)
			for _, disk := range dataDisks {
				dataDisksByName[disk.Name.ValueString()] = disk
			}

			// Reorder: first put disks in existing order, then append new ones
			var orderedDataDisks []diskModel
			usedNames := make(map[string]bool)

			// First, add disks in the same order as existing state
			for _, existingDisk := range existingDataDisks {
				name := existingDisk.Name.ValueString()
				if disk, ok := dataDisksByName[name]; ok {
					orderedDataDisks = append(orderedDataDisks, disk)
					usedNames[name] = true
				}
			}

			// Then append any new disks not in existing state
			for _, disk := range dataDisks {
				name := disk.Name.ValueString()
				if !usedNames[name] {
					orderedDataDisks = append(orderedDataDisks, disk)
				}
			}

			dataDisksList, diags := types.ListValueFrom(ctx, types.ObjectType{
				AttrTypes: diskModelAttrTypes,
			}, orderedDataDisks)
			resp.Diagnostics.Append(diags...)
			if !resp.Diagnostics.HasError() {
				state.DataDisks = dataDisksList
			}
		}
	}

	resp.Diagnostics.Append(diags...)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

func (r *vmResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state vmInstanceDataSourceModel

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

	// Preserve immutable fields from state
	plan.Uuid = state.Uuid
	plan.ImageUuid = state.ImageUuid

	// 1. Call Update API for basic fields (name, description, cpu, memory, platform, guestOsType)
	r.updateBasicFields(ctx, plan, state, resp)
	if resp.Diagnostics.HasError() {
		return
	}

	// 2. Update ResourceConfigs (cpuMode, gpuType, soundCard, cpuQuota, etc.)
	r.updateResourceConfigs(ctx, plan, state, resp)
	if resp.Diagnostics.HasError() {
		return
	}

	// 3. Update SystemTags (bootMode, cpuCores, vnuma, etc.)
	r.updateSystemTags(ctx, plan, state, resp)
	if resp.Diagnostics.HasError() {
		return
	}

	// 4. Update network configuration (static IP)
	r.updateNetworkConfig(ctx, plan, state, resp)
	if resp.Diagnostics.HasError() {
		return
	}

	// 5. Update post-creation configs (hostname, bootOrder, monitorNumber, clock, etc.)
	r.updatePostCreationConfigs(ctx, plan, state, resp)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read back the updated VM to sync all fields
	instance, err := r.client.GetVmInstance(ctx, plan.Uuid.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read VM instance", fmt.Sprintf("Error: %v", err))
		return
	}

	// Sync all fields from API response
	r.syncVMState(ctx, &plan, instance, state, resp)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// updateBasicFields updates basic VM fields using UpdateVmInstance API
func (r *vmResource) updateBasicFields(ctx context.Context, plan, state vmInstanceDataSourceModel, resp *resource.UpdateResponse) {
	updateParam := param.UpdateVmInstanceParam{
		Params: param.UpdateVmInstanceParamDetail{
			Name: plan.Name.ValueString(),
		},
	}

	// Add optional fields if they have changed
	if !plan.Description.IsNull() || !state.Description.IsNull() {
		description := plan.Description.ValueString()
		updateParam.Params.Description = &description
	}

	if !plan.Platform.IsNull() && !state.Platform.IsNull() && plan.Platform.ValueString() != state.Platform.ValueString() {
		platform := plan.Platform.ValueString()
		updateParam.Params.Platform = &platform
	}

	if !plan.GuestOsType.IsNull() && !state.GuestOsType.IsNull() && plan.GuestOsType.ValueString() != state.GuestOsType.ValueString() {
		guestOsType := plan.GuestOsType.ValueString()
		updateParam.Params.GuestOsType = &guestOsType
	}

	if !plan.CPUNum.IsNull() && !state.CPUNum.IsNull() && plan.CPUNum.ValueInt64() != state.CPUNum.ValueInt64() {
		cpuNum := int(plan.CPUNum.ValueInt64())
		updateParam.Params.CpuNum = &cpuNum
	}

	if !plan.MemorySize.IsNull() && !state.MemorySize.IsNull() && plan.MemorySize.ValueInt64() != state.MemorySize.ValueInt64() {
		memorySize := utils.MBToBytes(plan.MemorySize.ValueInt64())
		updateParam.Params.MemorySize = &memorySize
	}

	instance, err := r.client.UpdateVmInstance(ctx, plan.Uuid.ValueString(), updateParam)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update VM instance", fmt.Sprintf("Error: %v", err))
		return
	}

	// Update plan with response values
	plan.Name = types.StringValue(instance.Name)
	plan.Description = types.StringValue(instance.Description)
	plan.MemorySize = types.Int64Value(utils.BytesToMB(instance.MemorySize))
	plan.CPUNum = types.Int64Value(int64(instance.CpuNum))
	plan.Platform = types.StringValue(instance.Platform)
	plan.GuestOsType = types.StringValue(instance.GuestOsType)
}

// updateResourceConfigs updates resource configuration (cpuMode, gpuType, soundCard, etc.)
func (r *vmResource) updateResourceConfigs(ctx context.Context, plan, state vmInstanceDataSourceModel, resp *resource.UpdateResponse) {
	var resourceConfigs []param.UpdateResourceConfigs_ResourceConfigAOParam

	// CPU Mode
	if !plan.CpuMode.IsNull() && (!state.CpuMode.IsNull() && plan.CpuMode.ValueString() != state.CpuMode.ValueString()) {
		resourceConfigs = append(resourceConfigs, param.UpdateResourceConfigs_ResourceConfigAOParam{
			Category: utils.StringPointer("kvm"),
			Name:     "vm.cpuMode",
			Value:    utils.StringPointer(plan.CpuMode.ValueString()),
		})
	}

	// CPU Quota
	if !plan.CpuQuota.IsNull() && (!state.CpuQuota.IsNull() && plan.CpuQuota.ValueFloat64() != state.CpuQuota.ValueFloat64()) {
		quota := int(plan.CpuQuota.ValueFloat64() * 10000)
		resourceConfigs = append(resourceConfigs, param.UpdateResourceConfigs_ResourceConfigAOParam{
			Category: utils.StringPointer("kvm"),
			Name:     "vm.cpu.quota",
			Value:    utils.StringPointer(fmt.Sprintf("%d", quota)),
		})
	}

	// GPU Type
	if !plan.GpuType.IsNull() && (!state.GpuType.IsNull() && plan.GpuType.ValueString() != state.GpuType.ValueString()) {
		resourceConfigs = append(resourceConfigs, param.UpdateResourceConfigs_ResourceConfigAOParam{
			Category: utils.StringPointer("vm"),
			Name:     "videoType",
			Value:    utils.StringPointer(plan.GpuType.ValueString()),
		})
	}

	// Sound Card
	if !plan.SoundCard.IsNull() && (!state.SoundCard.IsNull() && plan.SoundCard.ValueString() != state.SoundCard.ValueString()) {
		resourceConfigs = append(resourceConfigs, param.UpdateResourceConfigs_ResourceConfigAOParam{
			Category: utils.StringPointer("vm"),
			Name:     "soundType",
			Value:    utils.StringPointer(plan.SoundCard.ValueString()),
		})
	}

	if len(resourceConfigs) > 0 {
		updateConfigParam := param.UpdateResourceConfigsParam{
			Params: param.UpdateResourceConfigsParamDetail{
				ResourceConfigs: resourceConfigs,
			},
		}

		_, err := r.client.UpdateResourceConfigs(ctx, plan.Uuid.ValueString(), updateConfigParam)
		if err != nil {
			resp.Diagnostics.AddError("Failed to update resource configs", fmt.Sprintf("Error: %v", err))
			return
		}
	}
}

// updateSystemTags updates system tags (bootMode, cpuCores, vnuma, etc.)
func (r *vmResource) updateSystemTags(ctx context.Context, plan, state vmInstanceDataSourceModel, resp *resource.UpdateResponse) {
	// Note: SystemTag updates require specific API calls
	// For now, this handles bootMode which is commonly updated

	// Boot Mode - handled via UpdateVmInstance or separate API in future
	if !plan.BootMode.IsNull() && (!state.BootMode.IsNull() && plan.BootMode.ValueString() != state.BootMode.ValueString()) {
		tflog.Info(ctx, fmt.Sprintf("Boot mode changed from %s to %s - requires VM reboot", state.BootMode.ValueString(), plan.BootMode.ValueString()))
	}
}

// updateNetworkConfig updates network configuration (static IP)
func (r *vmResource) updateNetworkConfig(ctx context.Context, plan, state vmInstanceDataSourceModel, resp *resource.UpdateResponse) {
	// Get current VM state
	vm, err := r.client.GetVmInstance(ctx, plan.Uuid.ValueString())
	if err != nil {
		tflog.Warn(ctx, fmt.Sprintf("Failed to get VM for network config update: %v", err))
		return
	}

	// Parse network interfaces from plan and state
	var planNics []NetworkInterfaceModel
	var stateNics []NetworkInterfaceModel

	if !plan.NetworkInterfaces.IsNull() && len(plan.NetworkInterfaces.Elements()) > 0 {
		plan.NetworkInterfaces.ElementsAs(ctx, &planNics, false)
	}
	if !state.NetworkInterfaces.IsNull() && len(state.NetworkInterfaces.Elements()) > 0 {
		state.NetworkInterfaces.ElementsAs(ctx, &stateNics, false)
	}

	// Build maps for comparison
	planNicsMap := make(map[string]NetworkInterfaceModel)
	stateNicsMap := make(map[string]NetworkInterfaceModel)

	for _, nic := range planNics {
		planNicsMap[nic.L3NetworkUuid.ValueString()] = nic
	}
	for _, nic := range stateNics {
		stateNicsMap[nic.L3NetworkUuid.ValueString()] = nic
	}

	// Update static IP for each NIC that changed
	for l3Uuid, planNic := range planNicsMap {
		stateNic, exists := stateNicsMap[l3Uuid]
		if !exists {
			continue
		}

		// Check if static IP changed
		planStaticIp := ""
		stateStaticIp := ""
		if !planNic.StaticIp.IsNull() {
			planStaticIp = planNic.StaticIp.ValueString()
		}
		if !stateNic.StaticIp.IsNull() {
			stateStaticIp = stateNic.StaticIp.ValueString()
		}

		if planStaticIp != stateStaticIp {
			// Find the VM NIC UUID
			var vmNicUuid string
			for _, vmNic := range vm.VmNics {
				if vmNic.L3NetworkUuid == l3Uuid {
					vmNicUuid = vmNic.UUID
					break
				}
			}

			if vmNicUuid != "" && planStaticIp != "" {
				staticIP := planStaticIp
				setStaticIpParam := param.SetVmStaticIpParam{
					Params: param.SetVmStaticIpParamDetail{
						L3NetworkUuid: l3Uuid,
						Ip:            &staticIP,
					},
				}

				tflog.Info(ctx, fmt.Sprintf("Setting static IP for VM %s NIC %s: IP=%s", plan.Uuid.ValueString(), vmNicUuid, planStaticIp))
				_, err := r.client.SetVmStaticIp(ctx, plan.Uuid.ValueString(), setStaticIpParam)
				if err != nil {
					tflog.Warn(ctx, fmt.Sprintf("Failed to set static IP for VM %s: %v", plan.Uuid.ValueString(), err))
				}
			}
		}
	}
}

// updatePostCreationConfigs updates post-creation configurations (hostname, bootOrder, etc.)
func (r *vmResource) updatePostCreationConfigs(ctx context.Context, plan, state vmInstanceDataSourceModel, resp *resource.UpdateResponse) {
	// Hostname
	if !plan.Hostname.IsNull() && (!state.Hostname.IsNull() && plan.Hostname.ValueString() != state.Hostname.ValueString()) {
		hostname := plan.Hostname.ValueString()
		setHostnameParam := param.SetVmHostnameParam{
			Params: param.SetVmHostnameParamDetail{
				Hostname: hostname,
			},
		}

		tflog.Info(ctx, fmt.Sprintf("Setting hostname for VM %s: %s", plan.Uuid.ValueString(), hostname))
		_, err := r.client.SetVmHostname(ctx, plan.Uuid.ValueString(), setHostnameParam)
		if err != nil {
			tflog.Warn(ctx, fmt.Sprintf("Failed to set hostname for VM %s: %v", plan.Uuid.ValueString(), err))
		}
	}

	// Boot Order
	if !plan.BootOrders.IsNull() && (!state.BootOrders.IsNull() || len(state.BootOrders.Elements()) == 0) {
		var planBootOrders []string
		plan.BootOrders.ElementsAs(ctx, &planBootOrders, false)

		if len(planBootOrders) > 0 {
			filteredBootOrders := []string{}
			for _, order := range planBootOrders {
				if order != "Empty" {
					filteredBootOrders = append(filteredBootOrders, order)
				}
			}

			if len(filteredBootOrders) > 0 {
				setBootOrderParam := param.SetVmBootOrderParam{
					Params: param.SetVmBootOrderParamDetail{
						BootOrder: filteredBootOrders,
					},
				}

				tflog.Info(ctx, fmt.Sprintf("Setting boot order for VM %s: %v", plan.Uuid.ValueString(), filteredBootOrders))
				_, err := r.client.SetVmBootOrder(ctx, plan.Uuid.ValueString(), setBootOrderParam)
				if err != nil {
					tflog.Warn(ctx, fmt.Sprintf("Failed to set boot order for VM %s: %v", plan.Uuid.ValueString(), err))
				}
			}
		}
	}

	// VDI Monitor Number
	if !plan.VdiMonitorNumber.IsNull() && (!state.VdiMonitorNumber.IsNull() && plan.VdiMonitorNumber.ValueInt64() != state.VdiMonitorNumber.ValueInt64()) {
		monitorNumber := int(plan.VdiMonitorNumber.ValueInt64())
		setMonitorParam := param.SetVmMonitorNumberParam{
			Params: param.SetVmMonitorNumberParamDetail{
				MonitorNumber: monitorNumber,
			},
		}

		tflog.Info(ctx, fmt.Sprintf("Setting VDI monitor number for VM %s: %d", plan.Uuid.ValueString(), monitorNumber))
		_, err := r.client.SetVmMonitorNumber(ctx, plan.Uuid.ValueString(), setMonitorParam)
		if err != nil {
			tflog.Warn(ctx, fmt.Sprintf("Failed to set VDI monitor number for VM %s: %v", plan.Uuid.ValueString(), err))
		}
	}
}

// syncVMState syncs all VM fields from API response
func (r *vmResource) syncVMState(ctx context.Context, plan *vmInstanceDataSourceModel, instance *view.VmInstanceInventoryView, state vmInstanceDataSourceModel, resp *resource.UpdateResponse) {
	// Update basic fields
	plan.Name = types.StringValue(instance.Name)
	plan.Description = types.StringValue(instance.Description)
	plan.MemorySize = types.Int64Value(utils.BytesToMB(instance.MemorySize))
	plan.CPUNum = types.Int64Value(int64(instance.CpuNum))
	if instance.Platform != "" {
		plan.Platform = types.StringValue(instance.Platform)
	}
	if instance.GuestOsType != "" {
		plan.GuestOsType = types.StringValue(instance.GuestOsType)
	}
	if instance.ImageUuid != "" {
		plan.ImageUuid = types.StringValue(instance.ImageUuid)
	}

	// Update network interfaces - preserve static_ip from plan
	var networkInterfaces []NetworkInterfaceModel
	planNicsMap := make(map[string]NetworkInterfaceModel)
	if !plan.NetworkInterfaces.IsNull() && len(plan.NetworkInterfaces.Elements()) > 0 {
		var nics []NetworkInterfaceModel
		diags := plan.NetworkInterfaces.ElementsAs(ctx, &nics, false)
		resp.Diagnostics.Append(diags...)
		for _, nic := range nics {
			planNicsMap[nic.L3NetworkUuid.ValueString()] = nic
		}
	}

	for _, nic := range instance.VmNics {
		staticIP := types.StringNull()
		defaultL3 := types.BoolValue(instance.DefaultL3NetworkUuid != "" && nic.L3NetworkUuid == instance.DefaultL3NetworkUuid)

		// Preserve static_ip from plan
		if planNic, ok := planNicsMap[nic.L3NetworkUuid]; ok {
			staticIP = planNic.StaticIp
			defaultL3 = planNic.DefaultL3
		}

		networkInterfaces = append(networkInterfaces, NetworkInterfaceModel{
			L3NetworkUuid: types.StringValue(nic.L3NetworkUuid),
			DefaultL3:     defaultL3,
			StaticIp:      staticIP,
		})
	}

	plan.NetworkInterfaces, _ = types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"port_group_uuid": types.StringType,
			"default_l3":      types.BoolType,
			"static_ip":       types.StringType,
		},
	}, networkInterfaces)

	// Update VM NICs
	var vmNics []NicsModel
	for _, nic := range instance.VmNics {
		vmNics = append(vmNics, NicsModel{
			Uuid:    types.StringValue(nic.UUID),
			Ip:      types.StringValue(nic.Ip),
			Netmask: types.StringValue(nic.Netmask),
			Gateway: types.StringValue(nic.Gateway),
		})
	}
	plan.VMNics, _ = types.ListValueFrom(ctx, types.ObjectType{AttrTypes: networkModelAttrTypes}, vmNics)
}

// Delete implements resource.Resource.
func (r *vmResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state vmInstanceDataSourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if state.Uuid == types.StringValue("") {
		tflog.Warn(ctx, "vm uuid is empty, so nothing to delete, skip it")
		return
	}
	if resp.Diagnostics.HasError() {
		return
	}

	vm, err := r.client.GetVmInstance(ctx, state.Uuid.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Could not read vm instance", "Error: "+err.Error(),
		)
		return
	}

	var volumeUuids []string
	for _, volume := range vm.AllVolumes {
		if volume.Type != "Data" {
			continue
		}
		volumeUuids = append(volumeUuids, volume.UUID)
	}

	tflog.Info(ctx, "Deleting vm instance "+state.Uuid.String())

	// Delete existing vm instance
	err = r.client.DestroyVmInstance(ctx, state.Uuid.ValueString(), param.DeleteModePermissive)
	if err != nil {
		resp.Diagnostics.AddError(
			"Could not destroy vm instance", "Error: "+err.Error(),
		)
		return
	}

	// Delete vm data volume
	for _, uuid := range volumeUuids {
		err = r.client.DeleteDataVolume(ctx, uuid, param.DeleteModePermissive)
		if err != nil {
			resp.Diagnostics.AddError(
				"Could not delete data volume", "Error: "+err.Error(),
			)
			return
		}
	}

	expunge := false
	if !state.Expunge.IsNull() && !state.Expunge.IsUnknown() {
		expunge = state.Expunge.ValueBool()
	}

	if expunge {
		tflog.Info(ctx, fmt.Sprintf("expunge instance %s", state.Uuid.ValueString()))
		// Expunge vm instance
		err = r.client.ExpungeVmInstance(ctx, state.Uuid.ValueString(), param.ExpungeVmInstanceParam{})
		if err != nil {
			resp.Diagnostics.AddError(
				"Could not expunge vm instance", "Error: "+err.Error(),
			)
			return
		}

		// Expunge vm data volume
		for _, uuid := range volumeUuids {
			err = r.client.ExpungeDataVolume(ctx, uuid, param.ExpungeDataVolumeParam{})
			if err != nil {
				resp.Diagnostics.AddError(
					"Could not expunge data volume", "Error: "+err.Error(),
				)
				return
			}
		}
	}

}

func isDiskParamValid(ctx context.Context, r *vmResource, model diskModel) error {
	if model.PrimaryStorageUuid.IsNull() || model.PrimaryStorageUuid.ValueString() == "" {
		return nil
	}

	dataDiskPrimaryStorageUuid := model.PrimaryStorageUuid.ValueString()

	qparam := param.NewQueryParam()
	qparam.AddQ("uuid=" + dataDiskPrimaryStorageUuid)
	qparam.AddQ("state=Enabled")
	qparam.Limit(1)
	primaryStorages, err := r.client.QueryPrimaryStorage(ctx, &qparam)
	if err != nil {
		return fmt.Errorf("failed to get primary storage %s, err: %v", dataDiskPrimaryStorageUuid, err)
	}

	if len(primaryStorages) == 0 {
		return fmt.Errorf("unable to find primary storage %s, err: %v", dataDiskPrimaryStorageUuid, err)
	}

	return nil
}
