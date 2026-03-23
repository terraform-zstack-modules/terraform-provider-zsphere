# Copyright (c) ZStack.io, Inc.

resource "zsphere_instance" "test" {
  name        = var.instance_name
  description = var.instance_description
  image_uuid  = var.image_uuid

  cpu_num     = var.cpu_num
  memory_size = var.memory_size
  platform    = var.platform
  strategy    = var.strategy
  guest_os_type = "Linux"
  host_uuid = var.host_uuid
  root_disk = {
    size = 5 * 1024 * 1024 * 1024
  }
}

output "instance_uuid" {
  value = zsphere_instance.test.uuid
}

output "instance_name" {
  value = zsphere_instance.test.name
}

output "instance_cpu_num" {
  value = zsphere_instance.test.cpu_num
}

output "instance_memory_size" {
  value = zsphere_instance.test.memory_size
}
