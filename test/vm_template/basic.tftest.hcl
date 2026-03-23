# Copyright (c) ZStack.io, Inc.

run "create_vm_template" {
  command = apply

  variables {
    vm_template_name        = "test-vm-template"
    vm_template_description = "Basic test VM template"
    vm_instance_uuid        = var.vm_instance_uuid
  }

  assert {
    condition     = zsphere_vm_template.test.name == "test-vm-template"
    error_message = "VM template name should match"
  }

  assert {
    condition     = zsphere_vm_template.test.uuid != null
    error_message = "UUID should be computed"
  }

  assert {
    condition     = zsphere_vm_template.test.cluster_uuid != null
    error_message = "Cluster UUID should be set"
  }
}
