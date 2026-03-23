# Copyright (c) ZStack.io, Inc.

# CRUD tests for VM Template
# Note: These tests require a running VM instance to convert to template

run "create_vm_template" {
  command = apply

  variables {
    vm_template_name        = "crud-test-template"
    vm_template_description = "CRUD test VM template"
    vm_instance_uuid        = var.vm_instance_uuid
  }

  assert {
    condition     = zsphere_vm_template.test.uuid != null
    error_message = "UUID should be computed after create"
  }

  assert {
    condition     = zsphere_vm_template.test.name == "crud-test-template"
    error_message = "VM template name should match"
  }
}

run "read_vm_template" {
  command = apply

  variables {
    vm_template_name        = "crud-test-template"
    vm_template_description = "CRUD test VM template"
    vm_instance_uuid        = var.vm_instance_uuid
  }

  assert {
    condition     = zsphere_vm_template.test.name == "crud-test-template"
    error_message = "VM template name should be readable"
  }

  assert {
    condition     = zsphere_vm_template.test.cluster_uuid != null
    error_message = "Cluster UUID should be readable"
  }
}

run "update_vm_template" {
  command = apply

  variables {
    vm_template_name        = "updated-crud-template"
    vm_template_description = "Updated CRUD test VM template"
    vm_instance_uuid        = var.vm_instance_uuid
  }

  assert {
    condition     = zsphere_vm_template.test.name == "updated-crud-template"
    error_message = "VM template name should be updated"
  }
}

run "verify_updated_vm_template" {
  command = apply

  variables {
    vm_template_name        = "updated-crud-template"
    vm_template_description = "Updated CRUD test VM template"
    vm_instance_uuid        = var.vm_instance_uuid
  }

  assert {
    condition     = zsphere_vm_template.test.name == "updated-crud-template"
    error_message = "Updated VM template name should persist"
  }
}
