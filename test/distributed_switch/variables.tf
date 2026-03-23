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

variable "distributed_switch_name" {
  type        = string
  default     = "test-distributed-switch"
  description = "Name of the distributed switch"
}

variable "distributed_switch_description" {
  type        = string
  default     = "Test distributed switch created by Terraform"
  description = "Description of the distributed switch"
}

variable "datacenter_uuid" {
  type        = string
  description = "UUID of the datacenter (zone) to which the distributed switch belongs"
}

variable "distributed_switch_vswitch_type" {
  type        = string
  default     = "LinuxBridge"
  description = "Type of the virtual switch: LinuxBridge, vSwitch, VLAN"
}

variable "distributed_switch_physical_interface" {
  type        = string
  default     = ""
  description = "Physical interface name (used when not using bonding)"
}

variable "distributed_switch_bonding_name" {
  type        = string
  default     = ""
  description = "Bonding aggregate port name (e.g., Uplink1)"
}

variable "distributed_switch_bonding_mode" {
  type        = string
  default     = ""
  description = "Bonding mode: balance-xor, 802.3ad, active-backup"
}

variable "distributed_switch_xmit_hash_policy" {
  type        = string
  default     = ""
  description = "Transmit hash policy for 802.3ad: layer2, layer2+3, layer3+4"
}

variable "distributed_switch_cluster_uuids" {
  type        = list(string)
  default     = []
  description = "List of cluster UUIDs to attach the distributed switch to"
}
