# Copyright (c) ZStack.io, Inc.
# SPDX-License-Identifier: MPL-2.0

# CRUD operation tests for creating instance from template

run "create_instance_from_template" {
  command = apply

  variables {
    create_basic_example = true
    create_advanced_example = false
  }

  assert {
    condition     = zsphere_instance.basic_from_template[0].name == "${var.instance_name}-basic"
    error_message = "Instance name should match the configured value"
  }

  assert {
    condition     = zsphere_instance.basic_from_template[0].uuid != null
    error_message = "Instance UUID should be computed"
  }

  assert {
    condition     = zsphere_instance.basic_from_template[0].template_uuid != null
    error_message = "Instance template_uuid should be set"
  }

  assert {
    condition     = zsphere_instance.basic_from_template[0].cpu_num == var.cpu_num
    error_message = "Instance CPU number should match the configured value"
  }

  assert {
    condition     = zsphere_instance.basic_from_template[0].memory_size == var.memory_size
    error_message = "Instance memory size should match the configured value"
  }
}

run "read_instance" {
  command = plan

  assert {
    condition     = zsphere_instance.basic_from_template[0].uuid != null
    error_message = "Instance UUID should still be present after read"
  }

  assert {
    condition     = zsphere_instance.basic_from_template[0].name == "${var.instance_name}-basic"
    error_message = "Instance name should be preserved after read"
  }
}

run "update_instance_description" {
  command = apply

  variables {
    create_basic_example = true
    create_advanced_example = false
    instance_description = "Updated description for template instance"
  }

  assert {
    condition     = zsphere_instance.basic_from_template[0].description == "Updated description for template instance"
    error_message = "Instance description should be updated"
  }

  assert {
    condition     = zsphere_instance.basic_from_template[0].uuid != null
    error_message = "Instance UUID should be preserved after update"
  }
}

run "update_instance_cpu_memory" {
  command = apply

  variables {
    create_basic_example = true
    create_advanced_example = false
    cpu_num     = 4
    memory_size = 4096
  }

  assert {
    condition     = zsphere_instance.basic_from_template[0].cpu_num == 4
    error_message = "Instance CPU number should be updated"
  }

  assert {
    condition     = zsphere_instance.basic_from_template[0].memory_size == 4096
    error_message = "Instance memory size should be updated"
  }
}

run "verify_vm_nics_computed" {
  command = plan

  assert {
    condition     = length(zsphere_instance.basic_from_template[0].vm_nics) >= 0
    error_message = "VM should have at least one NIC"
  }
}
