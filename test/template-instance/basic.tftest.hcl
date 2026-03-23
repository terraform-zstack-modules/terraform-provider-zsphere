# Copyright (c) ZStack.io, Inc.
# SPDX-License-Identifier: MPL-2.0

# Basic functionality tests for creating instance from template

run "create_basic_instance_from_template" {
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

run "verify_basic_outputs" {
  command = plan

  assert {
    condition     = output.basic_instance_uuid != null
    error_message = "Output basic_instance_uuid should not be null"
  }

  assert {
    condition     = output.template_uuid_used != null
    error_message = "Output template_uuid_used should not be null"
  }
}
