# Copyright (c) ZStack.io, Inc.

# Provider Variables

variable "zsphere_host" {
  type        = string
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

# Resource Variables

variable "cluster_name" {
  type        = string
  default     = "test-cluster"
  description = "Name of the cluster"
}

variable "cluster_description" {
  type        = string
  default     = "Test cluster created by Terraform"
  description = "Description of the cluster"
}

variable "cluster_hypervisor_type" {
  type        = string
  default     = "KVM"
  description = "Hypervisor type of the cluster"
}

variable "cluster_type" {
  type        = string
  default     = "zstack"
  description = "Type of the cluster"
}

variable "cluster_architecture" {
  type        = string
  default     = "x86_64"
  description = "CPU architecture of the cluster"
}

variable "datacenter_uuid" {
  type        = string
  description = "UUID of the datacenter (zone) to which the cluster belongs"
}

variable "cluster_display_network_cidr" {
  type        = string
  default     = ""
  description = "Display network CIDR"
}

variable "cluster_migrate_network_cidr" {
  type        = string
  default     = ""
  description = "Migration network CIDR"
}

variable "cluster_check_cpu_model" {
  type        = string
  default     = ""
  description = "CPU model validation: default, false, true"
}

variable "cluster_cpu_mode" {
  type        = string
  default     = ""
  description = "CPU mode: hostPassthrough, virtio, kvm"
}

variable "cluster_network_hp" {
  type        = bool
  default     = false
  description = "Enable network HP (ovsdpdk)"
}

variable "cluster_automation_level" {
  type        = string
  default     = ""
  description = "DRS automation level: Manual, Automatic, closed"
}

variable "cluster_cpu_over_provisioning_ratio" {
  type        = string
  default     = ""
  description = "CPU over-provisioning ratio, e.g., 4"
}

variable "cluster_vm_ha_level" {
  type        = string
  default     = ""
  description = "VM HA level: NeverStop, BestEffort"
}

variable "cluster_drs_enabled" {
  type        = bool
  default     = false
  description = "Enable Dynamic Resource Scheduling (DRS)"
}

variable "cluster_drs_threshold_cpu" {
  type        = number
  default     = 80
  description = "DRS CPU threshold percentage"
}

variable "cluster_drs_threshold_memory" {
  type        = number
  default     = 80
  description = "DRS memory threshold percentage"
}

variable "cluster_drs_threshold_duration" {
  type        = number
  default     = 300
  description = "DRS threshold duration in seconds"
}
