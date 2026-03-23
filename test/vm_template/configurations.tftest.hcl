# Copyright (c) ZStack.io, Inc.

# Configuration tests for VM Template
# Note: These tests require a running VM instance to convert to template

run "create_vm_template_with_description" {
  command = apply

  variables {
    vm_template_name        = "config-test-template"
    vm_template_description = "Configuration test VM template with description"
    vm_instance_uuid        = var.vm_instance_uuid
  }

  assert {
    condition     = zsphere_vm_template.test.name == "config-test-template"
    error_message = "VM template name should match"
  }

  assert {
    condition     = zsphere_vm_template.test.description == "Configuration test VM template with description"
    error_message = "VM template description should match"
  }

  assert {
    condition     = zsphere_vm_template.test.uuid != null
    error_message = "UUID should be computed"
  }

  assert {
    condition     = zsphere_vm_template.test.cluster_uuid != null
    error_message = "Cluster UUID should be computed"
  }
}

run "update_vm_template_description" {
  command = apply

  variables {
    vm_template_name        = "config-test-template"
    vm_template_description = "Updated configuration test VM template"
    vm_instance_uuid        = var.vm_instance_uuid
  }

  assert {
    condition     = zsphere_vm_template.test.description == "Updated configuration test VM template"
    error_message = "VM template description should be updated"
  }
}
