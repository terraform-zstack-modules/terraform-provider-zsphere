# Copyright (c) ZStack.io, Inc.

resource "zsphere_vm_template" "test" {
  name            = var.vm_template_name
  description     = var.vm_template_description
  vm_instance_uuid = var.vm_instance_uuid
}

output "vm_template_uuid" {
  value = zsphere_vm_template.test.uuid
}

output "vm_template_name" {
  value = zsphere_vm_template.test.name
}

output "vm_template_cluster_uuid" {
  value = zsphere_vm_template.test.cluster_uuid
}
