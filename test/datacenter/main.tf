# Copyright (c) ZStack.io, Inc.

resource "zsphere_datacenter" "test" {
  name        = var.datacenter_name
  description = var.datacenter_description
  is_default  = var.datacenter_is_default
}

output "datacenter_uuid" {
  value = zsphere_datacenter.test.uuid
}

output "datacenter_name" {
  value = zsphere_datacenter.test.name
}

output "datacenter_state" {
  value = zsphere_datacenter.test.state
}

output "datacenter_type" {
  value = zsphere_datacenter.test.type
}

output "datacenter_is_default" {
  value = zsphere_datacenter.test.is_default
}
