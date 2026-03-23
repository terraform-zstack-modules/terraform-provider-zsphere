# Copyright (c) ZStack.io, Inc.

resource "zsphere_image" "test" {
  name                = var.image_name
  url                 = var.image_url
  format              = var.image_format
  media_type          = var.image_media_type
  architecture        = var.image_architecture
  boot_mode           = var.image_boot_mode
  platform            = var.image_platform
  guest_os_type       = var.image_guest_os_type
  description         = var.image_description
  virtio              = var.image_virtio
  expunge             = var.image_expunge
  image_storage_uuids = length(var.image_storage_uuids) > 0 ? var.image_storage_uuids : null
}

output "image_uuid" {
  value = zsphere_image.test.uuid
}

output "image_name" {
  value = zsphere_image.test.name
}

output "image_status" {
  value = zsphere_image.test.status
}

output "image_state" {
  value = zsphere_image.test.state
}

output "image_format" {
  value = zsphere_image.test.format
}

output "image_platform" {
  value = zsphere_image.test.platform
}

output "image_architecture" {
  value = zsphere_image.test.architecture
}

output "image_media_type" {
  value = zsphere_image.test.media_type
}

output "image_boot_mode" {
  value = zsphere_image.test.boot_mode
}
