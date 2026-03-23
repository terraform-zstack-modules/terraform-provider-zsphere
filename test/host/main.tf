# Copyright (c) ZStack.io, Inc.

resource "zsphere_host" "test" {
  name           = var.host_name
  description    = var.host_description
  management_ip  = var.host_management_ip
  cluster_uuid   = var.cluster_uuid
  username       = var.host_username
  password       = var.host_password
  ssh_port       = var.host_ssh_port
  iommu          = var.host_iommu
  ept            = var.host_ept
}

output "host_uuid" {
  value = zsphere_host.test.uuid
}

output "host_name" {
  value = zsphere_host.test.name
}

output "host_status" {
  value = zsphere_host.test.status
}
