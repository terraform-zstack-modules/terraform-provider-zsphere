# Copyright (c) ZStack.io, Inc.
# SPDX-License-Identifier: MPL-2.0

# Advanced configuration tests for creating instance from template

run "create_advanced_instance_with_boot_config" {
  command = apply

  variables {
    create_basic_example = false
    create_advanced_example = true
    boot_mode    = "UEFI"
    hostname     = "test-vm-from-template"
    architecture = "x86_64"
  }

  assert {
    condition     = zsphere_instance.advanced_from_template[0].boot_mode == "UEFI"
    error_message = "Instance boot_mode should be UEFI"
  }

  assert {
    condition     = zsphere_instance.advanced_from_template[0].hostname == "test-vm-from-template"
    error_message = "Instance hostname should match the configured value"
  }

  assert {
    condition     = zsphere_instance.advanced_from_template[0].architecture == "x86_64"
    error_message = "Instance architecture should be x86_64"
  }
}

run "create_instance_with_cpu_config" {
  command = apply

  variables {
    create_basic_example = false
    create_advanced_example = true
    cpu_mode            = "host-model"
    cpu_quota           = 2.0
    socked_num          = 1
    cpu_resource_level  = "CpuHigh"
  }

  assert {
    condition     = zsphere_instance.advanced_from_template[0].cpu_mode == "host-model"
    error_message = "Instance cpu_mode should be host-model"
  }

  assert {
    condition     = zsphere_instance.advanced_from_template[0].cpu_resource_level == "CpuHigh"
    error_message = "Instance cpu_resource_level should be CpuHigh"
  }
}

run "create_instance_with_memory_config" {
  command = apply

  variables {
    create_basic_example = false
    create_advanced_example = true
    memory_resource_level = "High"
  }

  assert {
    condition     = zsphere_instance.advanced_from_template[0].memory_resource_level == "High"
    error_message = "Instance memory_resource_level should be High"
  }
}

run "create_instance_with_gpu_config" {
  command = apply

  variables {
    create_basic_example = false
    create_advanced_example = true
    gpu_type         = "qxl"
    total_gpu_memory = 128
    sound_card       = "ich6"
  }

  assert {
    condition     = zsphere_instance.advanced_from_template[0].gpu_type == "qxl"
    error_message = "Instance gpu_type should be qxl"
  }

  assert {
    condition     = zsphere_instance.advanced_from_template[0].total_gpu_memory == 128
    error_message = "Instance total_gpu_memory should be 128"
  }

  assert {
    condition     = zsphere_instance.advanced_from_template[0].sound_card == "ich6"
    error_message = "Instance sound_card should be ich6"
  }
}

run "create_instance_with_usb_config" {
  command = apply

  variables {
    create_basic_example = false
    create_advanced_example = true
    usb_redirect = true
  }

  assert {
    condition     = zsphere_instance.advanced_from_template[0].usb_redirect == true
    error_message = "Instance usb_redirect should be true"
  }
}

run "create_instance_with_vm_group" {
  command = apply

  variables {
    create_basic_example = false
    create_advanced_example = true
    never_stop = true
  }

  assert {
    condition     = zsphere_instance.advanced_from_template[0].never_stop == true
    error_message = "Instance never_stop should be true"
  }
}
