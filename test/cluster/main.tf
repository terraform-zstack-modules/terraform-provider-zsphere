# Copyright (c) ZStack.io, Inc.

resource "zsphere_cluster" "test" {
  name                        = var.cluster_name
  description                 = var.cluster_description
  hypervisor_type             = var.cluster_hypervisor_type
  type                        = var.cluster_type
  datacenter_uuid            = var.datacenter_uuid
  architecture                = var.cluster_architecture

  # Advanced features
  display_network_cidr        = var.cluster_display_network_cidr
  migrate_network_cidr       = var.cluster_migrate_network_cidr
  check_cpu_model            = var.cluster_check_cpu_model
  cpu_mode                   = var.cluster_cpu_mode
  network_hp                 = var.cluster_network_hp
  automation_level           = var.cluster_automation_level
  cpu_over_provisioning_ratio = var.cluster_cpu_over_provisioning_ratio
  vm_ha_level                = var.cluster_vm_ha_level

  # DRS settings
  drs_enabled                = var.cluster_drs_enabled
  drs_threshold_cpu          = var.cluster_drs_threshold_cpu
  drs_threshold_memory       = var.cluster_drs_threshold_memory
  drs_threshold_duration     = var.cluster_drs_threshold_duration
}

output "cluster_uuid" {
  value = zsphere_cluster.test.uuid
}

output "cluster_name" {
  value = zsphere_cluster.test.name
}

output "cluster_state" {
  value = zsphere_cluster.test.state
}

output "cluster_hypervisor_type" {
  value = zsphere_cluster.test.hypervisor_type
}

output "cluster_type" {
  value = zsphere_cluster.test.type
}

output "cluster_architecture" {
  value = zsphere_cluster.test.architecture
}

output "cluster_datacenter_uuid" {
  value = zsphere_cluster.test.datacenter_uuid
}
