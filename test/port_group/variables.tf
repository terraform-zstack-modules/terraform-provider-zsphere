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

variable "port_group_name" {
  type        = string
  default     = "test-port-group"
  description = "Name of the port group"
}

variable "port_group_description" {
  type        = string
  default     = "Test port group created by Terraform"
  description = "Description of the port group"
}

variable "vswitch_uuid" {
  type        = string
  description = "UUID of the vSwitch (L2Network) to which the port group belongs"
}

variable "port_group_vlan" {
  type        = number
  default     = 100
  description = "VLAN ID for the port group (1-4094)"
}

variable "port_group_vlan_mode" {
  type        = string
  default     = "ACCESS"
  description = "VLAN mode: ACCESS, NONE, PVLAN, TRUNK"
}

variable "port_group_category" {
  type        = string
  default     = "Private"
  description = "Category of the port group: Private, Public"
}

variable "port_group_dns_domain" {
  type        = string
  default     = "test.local"
  description = "DNS domain for the port group"
}

variable "port_group_ip_version" {
  type        = number
  default     = 4
  description = "IP version: 4 or 6"
}

variable "port_group_enable_ipam" {
  type        = bool
  default     = false
  description = "Enable IPAM for the port group"
}

variable "port_group_dhcp_service" {
  type        = bool
  default     = false
  description = "Enable DHCP service for the port group"
}

variable "port_group_dhcp_ip" {
  type        = string
  default     = ""
  description = "DHCP service IP address"
}

variable "port_group_dns" {
  type        = list(string)
  default     = []
  description = "DNS servers for the port group"
}

variable "ip_range_name" {
  type        = string
  default     = "test-ip-range"
  description = "Name of the IP range"
}

variable "ip_range_start_ip" {
  type        = string
  default     = ""
  description = "Start IP address of the IP range"
}

variable "ip_range_end_ip" {
  type        = string
  default     = ""
  description = "End IP address of the IP range"
}

variable "ip_range_netmask" {
  type        = string
  default     = ""
  description = "Netmask for the IP range"
}

variable "ip_range_gateway" {
  type        = string
  default     = ""
  description = "Gateway for the IP range"
}

variable "ip_range_allocate_strategy" {
  type        = string
  default     = "RandomIpAllocator"
  description = "IP allocation strategy: RandomIpAllocator, FirstAvailableIpAllocator, AscDelayRecycleIpAllocator"
}
