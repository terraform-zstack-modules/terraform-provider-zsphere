# Copyright (c) ZStack.io, Inc.
# SPDX-License-Identifier: MPL-2.0

# Provider Variables (must be declared in each test directory)

variable "zsphere_host" {
  type        = string
  default     = "your-zsphere-host"
  description = "ZStack management node host address"
}

variable "zsphere_access_key_id" {
  type        = string
  sensitive   = true
  description = "ZStack access key ID"
}

variable "zsphere_access_key_secret" {
  type        = string
  sensitive   = true
  description = "ZStack access key secret"
}

# Example Control Variables

variable "create_basic_example" {
  type        = bool
  default     = true
  description = "Whether to create the basic example VM"
}

variable "create_advanced_example" {
  type        = bool
  default     = false
  description = "Whether to create the advanced example VM with all parameters"
}

# Resource Variables

variable "instance_name" {
  type        = string
  default     = "test-vm-from-template"
  description = "Name of the VM instance created from template"
}

variable "instance_description" {
  type        = string
  default     = "Test VM instance created from template by Terraform"
  description = "Description of the VM instance"
}

variable "template_name" {
  type        = string
  default     = ""
  description = "Name of the template VM to use for creating instances (used to lookup template)"
}

variable "template_uuid" {
  type        = string
  default     = ""
  description = "UUID of the template VM to use for creating instances (if provided, template_name is ignored)"
}

variable "port_group_name" {
  type        = string
  default     = ""
  description = "Name of the port group for network interfaces"
}

variable "datacenter_uuid" {
  type        = string
  default     = ""
  description = "UUID of the datacenter (zone) where the VM will be deployed"
}

variable "cluster_uuid" {
  type        = string
  default     = ""
  description = "UUID of the cluster where the VM will be deployed"
}

variable "host_uuid" {
  type        = string
  default     = ""
  description = "UUID of the host where the VM will be deployed (optional, if not specified, auto-select)"
}

variable "cpu_num" {
  type        = number
  default     = 2
  description = "Number of CPUs for the VM"
}

variable "memory_size" {
  type        = number
  default     = 2048
  description = "Memory size in MB for the VM"
}

variable "boot_mode" {
  type        = string
  default     = "Legacy"
  description = "Boot mode for the VM (Legacy, UEFI)"
}

variable "hostname" {
  type        = string
  default     = ""
  description = "Hostname for the VM"
}

variable "static_ip" {
  type        = string
  default     = ""
  description = "Static IP address for the VM"
}

variable "netmask" {
  type        = string
  default     = ""
  description = "Netmask for the VM network"
}

variable "gateway" {
  type        = string
  default     = ""
  description = "Gateway for the VM network"
}

variable "strategy" {
  type        = string
  default     = "CreateStopped"
  description = "Deployment strategy (InstantStart, JustCreate, CreateStopped)"
}

variable "never_stop" {
  type        = bool
  default     = false
  description = "Whether the VM should never stop automatically"
}

# New Template-Specific Variables

variable "architecture" {
  type        = string
  default     = "x86_64"
  description = "Architecture of the VM (x86_64, aarch64)"
}

variable "motherboard_type" {
  type        = string
  default     = ""
  description = "Motherboard type (q35, pc)"
}

variable "cpu_mode" {
  type        = string
  default     = "host-model"
  description = "CPU mode (host-model, host-passthrough, custom)"
}

variable "cpu_quota" {
  type        = number
  default     = null
  description = "CPU quota for the VM"
}

variable "socked_num" {
  type        = number
  default     = null
  description = "Number of CPU sockets"
}

variable "vnuma_enabled" {
  type        = bool
  default     = null
  description = "Whether vNUMA is enabled"
}

variable "cpu_resource_level" {
  type        = string
  default     = "Normal"
  description = "CPU resource level (Normal, CpuHigh)"
}

variable "cpu_bind_type" {
  type        = string
  default     = "none"
  description = "CPU binding type"
}

variable "memory_resource_level" {
  type        = string
  default     = "Normal"
  description = "Memory resource level (Normal, High)"
}

variable "gpu_type" {
  type        = string
  default     = ""
  description = "Type of GPU (qxl, vga)"
}

variable "total_gpu_memory" {
  type        = number
  default     = null
  description = "Total GPU memory in MB"
}

variable "sound_card" {
  type        = string
  default     = ""
  description = "Type of sound card (ich6, ac97)"
}

variable "usb_redirect" {
  type        = bool
  default     = false
  description = "Whether USB redirect is enabled"
}

variable "group" {
  type        = string
  default     = "default"
  description = "VM group identifier"
}

variable "vm_group_uuid" {
  type        = string
  default     = ""
  description = "UUID of the VM scheduling group"
}

# Disk AO Configuration
variable "disk_aos" {
  type = list(object({
    size                 = optional(number)
    primary_storage_uuid = optional(string)
    name                 = optional(string)
    boot                 = optional(bool)
    bus_type             = optional(string)
    source_type          = optional(string)
    source_uuid          = optional(string)
    system_tags          = optional(list(string))
  }))
  default     = []
  description = "List of disk AO configurations for template creation"
}

# CD-ROM Configuration
variable "cdrom_list" {
  type = list(object({
    cdrom   = string
    iso_uuid = optional(string)
  }))
  default     = []
  description = "List of CD-ROM configurations"
}

# USB Device Configuration
variable "vm_usb_config" {
  type = list(object({
    usb_device_uuid = string
    attach_type     = optional(string)
  }))
  default     = []
  description = "List of USB device configurations"
}

# vGPU Device Configuration
variable "vgpu_device" {
  type = object({
    uuid = string
    type = string
  })
  default     = null
  description = "vGPU device configuration"
}

# GPU Device UUID List
variable "gpu_device_uuid_list" {
  type        = list(string)
  default     = []
  description = "List of GPU device UUIDs"
}

# CPU Binding Configuration
variable "cpu_bind_list_by_vcpu" {
  type = list(object({
    vcpu      = string
    pcpu_list = list(string)
  }))
  default     = []
  description = "List of CPU binding configurations per vCPU"
}