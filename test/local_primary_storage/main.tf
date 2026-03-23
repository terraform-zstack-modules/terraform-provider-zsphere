# Copyright (c) ZStack.io, Inc.

resource "zsphere_local_primary_storage" "test" {
  name         = var.local_storage_name
  description  = var.local_storage_description
  datacenter_uuid = var.datacenter_uuid
  cluster_uuid  = var.cluster_uuid
  url          = var.local_storage_url
}

output "local_storage_uuid" {
  value = zsphere_local_primary_storage.test.uuid
}

output "local_storage_name" {
  value = zsphere_local_primary_storage.test.name
}

output "local_storage_status" {
  value = zsphere_local_primary_storage.test.status
}
