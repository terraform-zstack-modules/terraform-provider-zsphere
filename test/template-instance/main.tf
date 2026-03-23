# Copyright (c) ZStack.io, Inc.
# SPDX-License-Identifier: MPL-2.0

# Test configuration for creating VM instances from a template
# This example demonstrates all the new template-specific parameters

# Data source to find the template VM (only if template_uuid is not provided)
data "zsphere_instances" "template" {
  count = var.template_uuid == "" ? 1 : 0
  name  = var.template_name
}

# Local value to get template UUID (either from variable or data source)
locals {
  template_uuid = var.template_uuid != "" ? var.template_uuid : (
    length(data.zsphere_instances.template) > 0 ? data.zsphere_instances.template[0].vminstances[0].uuid : ""
  )
  template_name_val = var.template_uuid != "" ? var.instance_name : (
    length(data.zsphere_instances.template) > 0 ? data.zsphere_instances.template[0].vminstances[0].name : ""
  )
}

# Data source to find the port group (only if port_group_name is provided)
data "zsphere_port_groups" "network" {
  count = var.port_group_name != "" ? 1 : 0
  name  = var.port_group_name
}

# Local value to get port group UUID
locals {
  port_group_uuid = length(data.zsphere_port_groups.network) > 0 ? data.zsphere_port_groups.network[0].port_groups[0].uuid : ""
}

# Example 1: Basic VM from template (existing functionality)
resource "zsphere_instance" "basic_from_template" {
  count = var.create_basic_example ? 1 : 0

  name          = "${var.instance_name}-basic"
  description   = var.instance_description
  template_uuid = local.template_uuid
  expunge       = true

  cpu_num     = var.cpu_num
  memory_size = var.memory_size

  datacenter_uuid = var.datacenter_uuid != "" ? var.datacenter_uuid : null
  cluster_uuid    = var.cluster_uuid != "" ? var.cluster_uuid : null
  host_uuid       = var.host_uuid != "" ? var.host_uuid : null

  network_interfaces = local.port_group_uuid != "" ? [
    {
      port_group_uuid = local.port_group_uuid
      default_l3      = true
      static_ip       = var.static_ip != "" ? var.static_ip : null
    }
  ] : null

  strategy = var.strategy
}

# Example 2: Advanced VM from template with all new parameters
resource "zsphere_instance" "advanced_from_template" {
  count = var.create_advanced_example ? 1 : 0

  name          = "${var.instance_name}-advanced"
  description   = var.instance_description
  template_uuid = local.template_uuid
  expunge       = true

  cpu_num     = var.cpu_num
  memory_size = var.memory_size

  datacenter_uuid = var.datacenter_uuid != "" ? var.datacenter_uuid : null
  cluster_uuid    = var.cluster_uuid != "" ? var.cluster_uuid : null
  host_uuid       = var.host_uuid != "" ? var.host_uuid : null

  # Boot and system configuration
  boot_mode         = var.boot_mode
  hostname          = var.hostname
  architecture      = var.architecture
  motherboard_type  = var.motherboard_type

  # CPU configuration
  cpu_mode            = var.cpu_mode
  cpu_quota           = var.cpu_quota
  socked_num          = var.socked_num
  vnuma_enabled       = var.vnuma_enabled
  cpu_resource_level  = var.cpu_resource_level
  cpu_bind_type       = var.cpu_bind_type

  # Memory configuration
  memory_resource_level = var.memory_resource_level

  # GPU and display configuration
  gpu_type         = var.gpu_type
  total_gpu_memory = var.total_gpu_memory
  sound_card       = var.sound_card

  # USB configuration
  usb_redirect = var.usb_redirect

  # VM grouping and scheduling
  group         = var.group
  vm_group_uuid = var.vm_group_uuid
  never_stop    = var.never_stop

  # Network configuration using vm_nic_config
  vm_nic_config = local.port_group_uuid != "" ? [
    {
      l3_network_uuid    = local.port_group_uuid
      driver_type        = "virtio"
      multi_queue_num    = "1"
      state              = "enable"
      static_ip          = var.static_ip != "" ? var.static_ip : null
      ipv4_netmask       = var.netmask != "" ? var.netmask : null
      ipv4_gateway       = var.gateway != "" ? var.gateway : null
      inbound_bandwidth  = null
      outbound_bandwidth = null
      custom_mac         = null
    }
  ] : []

  # Alternative: Use vm_nic_params JSON string (advanced usage)
  # vm_nic_params = jsonencode([
  #   {
  #     l3NetworkUuid = local.port_group_uuid
  #     driverType    = "virtio"
  #     multiQueueNum = "1"
  #     state         = "enable"
  #   }
  # ])

  # Disk configuration (disk_aos)
  disk_aos = var.disk_aos

  # CD-ROM configuration
  cdrom_list = var.cdrom_list

  # USB device configuration
  vm_usb_config = var.vm_usb_config

  # GPU device configuration
  vgpu_device = var.vgpu_device

  # List of GPU device UUIDs
  gpu_device_uuid_list = var.gpu_device_uuid_list

  # CPU binding configuration
  cpu_bind_list_by_vcpu = var.cpu_bind_list_by_vcpu

  strategy = var.strategy
}

# Outputs
output "basic_instance_uuid" {
  value = var.create_basic_example ? zsphere_instance.basic_from_template[0].uuid : null
}

output "advanced_instance_uuid" {
  value = var.create_advanced_example ? zsphere_instance.advanced_from_template[0].uuid : null
}

output "template_uuid_used" {
  value = local.template_uuid
}

output "template_name_used" {
  value = local.template_name_val
}
