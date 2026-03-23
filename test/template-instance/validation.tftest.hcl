# Copyright (c) ZStack.io, Inc.
# SPDX-License-Identifier: MPL-2.0

# Validation tests for creating instance from template
# These tests verify that the provider correctly validates input parameters

run "validate_template_uuid_or_name_required" {
  command = plan

  variables {
    template_name = ""
    template_uuid = ""
    create_basic_example = true
    create_advanced_example = false
  }

  expect_failures = [
    local.template_uuid,
  ]
}

run "validate_cpu_num_positive" {
  command = plan

  variables {
    cpu_num = 0
    create_basic_example = true
    create_advanced_example = false
  }

  expect_failures = [
    var.cpu_num,
  ]
}

run "validate_memory_size_positive" {
  command = plan

  variables {
    memory_size = 0
    create_basic_example = true
    create_advanced_example = false
  }

  expect_failures = [
    var.memory_size,
  ]
}

run "validate_boot_mode_values" {
  command = plan

  variables {
    boot_mode = "InvalidBootMode"
    create_basic_example = false
    create_advanced_example = true
  }

  expect_failures = [
    var.boot_mode,
  ]
}

run "validate_strategy_values" {
  command = plan

  variables {
    strategy = "InvalidStrategy"
    create_basic_example = true
    create_advanced_example = false
  }

  expect_failures = [
    var.strategy,
  ]
}

run "validate_platform_not_allowed_for_template" {
  command = plan

  variables {
    platform = "Linux"
    create_basic_example = true
    create_advanced_example = false
  }

  expect_failures = [
    zsphere_instance.basic_from_template,
  ]
}

run "validate_guest_os_type_not_allowed_for_template" {
  command = plan

  variables {
    guest_os_type = "Linux"
    create_basic_example = true
    create_advanced_example = false
  }

  expect_failures = [
    zsphere_instance.basic_from_template,
  ]
}

run "validate_root_disk_not_allowed_for_template" {
  command = plan

  variables {
    create_basic_example = true
    create_advanced_example = false
  }

  expect_failures = [
    zsphere_instance.basic_from_template,
  ]
}

run "validate_data_disks_not_allowed_for_template" {
  command = plan

  variables {
    create_basic_example = true
    create_advanced_example = false
  }

  expect_failures = [
    zsphere_instance.basic_from_template,
  ]
}
