# Copyright (c) ZStack.io, Inc.

resource "zsphere_nfs_primary_storage" "test" {
  name            = var.nfs_storage_name
  description     = var.nfs_storage_description
  datacenter_uuid = var.datacenter_uuid
  cluster_uuid    = var.cluster_uuid
  url             = var.nfs_storage_url
  cidr            = var.nfs_storage_cidr
  mount_options   = var.nfs_storage_mount_options
}

output "nfs_storage_uuid" {
  value = zsphere_nfs_primary_storage.test.uuid
}

output "nfs_storage_name" {
  value = zsphere_nfs_primary_storage.test.name
}

output "nfs_storage_status" {
  value = zsphere_nfs_primary_storage.test.status
}
