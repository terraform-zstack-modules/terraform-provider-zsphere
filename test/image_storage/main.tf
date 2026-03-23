# Copyright (c) ZStack.io, Inc.

resource "zsphere_image_storage" "test" {
  name            = var.image_storage_name
  description     = var.image_storage_description
  type            = var.image_storage_type
  hostname        = var.image_storage_hostname
  url             = var.image_storage_url
  ssh_port        = var.image_storage_ssh_port
  username        = var.image_storage_username
  password        = var.image_storage_password
  zone_uuid       = var.image_storage_zone_uuid
  import_images   = var.image_storage_import_images
  pool_name       = var.image_storage_pool_name
  mon_urls        = var.image_storage_mon_urls
}

output "image_storage_uuid" {
  value = zsphere_image_storage.test.uuid
}

output "image_storage_name" {
  value = zsphere_image_storage.test.name
}

output "image_storage_type" {
  value = zsphere_image_storage.test.type
}

output "image_storage_state" {
  value = zsphere_image_storage.test.state
}

output "image_storage_status" {
  value = zsphere_image_storage.test.status
}

output "image_storage_total_capacity" {
  value = zsphere_image_storage.test.total_capacity
}

output "image_storage_available_capacity" {
  value = zsphere_image_storage.test.available_capacity
}
