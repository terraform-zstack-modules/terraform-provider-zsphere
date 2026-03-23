# Copyright (c) ZStack.io, Inc.

resource "zsphere_port_group" "test" {
  name          = var.port_group_name
  description   = var.port_group_description
  vswitch_uuid  = var.vswitch_uuid
  vlan          = var.port_group_vlan
  vlan_mode     = var.port_group_vlan_mode
  category      = var.port_group_category
  dns_domain    = var.port_group_dns_domain
  ip_version    = var.port_group_ip_version
  enable_ipam   = var.port_group_enable_ipam
  dhcp_service  = var.port_group_dhcp_service
  dhcp_ip       = var.port_group_dhcp_ip != "" ? var.port_group_dhcp_ip : null

  # DNS servers list
  dns = length(var.port_group_dns) > 0 ? var.port_group_dns : null

  # IP Range (only when IPAM is enabled)
  dynamic "ip_range" {
    for_each = var.port_group_enable_ipam && var.ip_range_start_ip != "" ? [1] : []
    content {
      name               = var.ip_range_name
      start_ip           = var.ip_range_start_ip
      end_ip             = var.ip_range_end_ip
      netmask            = var.ip_range_netmask
      gateway            = var.ip_range_gateway != "" ? var.ip_range_gateway : null
    }
  }
}

output "port_group_uuid" {
  value = zsphere_port_group.test.uuid
}

output "port_group_name" {
  value = zsphere_port_group.test.name
}

output "port_group_state" {
  value = zsphere_port_group.test.state
}

output "port_group_vlan" {
  value = zsphere_port_group.test.vlan
}

output "port_group_vlan_mode" {
  value = zsphere_port_group.test.vlan_mode
}

output "port_group_category" {
  value = zsphere_port_group.test.category
}

output "port_group_datacenter_uuid" {
  value = zsphere_port_group.test.datacenter_uuid
}

output "port_group_ip_version" {
  value = zsphere_port_group.test.ip_version
}

output "port_group_enable_ipam" {
  value = zsphere_port_group.test.enable_ipam
}

output "port_group_dhcp_service" {
  value = zsphere_port_group.test.dhcp_service
}

output "port_group_dns_domain" {
  value = zsphere_port_group.test.dns_domain
}
