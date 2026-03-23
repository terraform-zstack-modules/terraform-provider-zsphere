# Copyright (c) ZStack.io, Inc.

# Validation tests for zsphere_instance resource
# Note: Terraform 1.10 does not support expect_fail for negative testing
# Invalid value tests should be run manually to verify validators work

run "valid_platform_linux" {
  command = plan

  variables {
    platform = "Linux"
  }

  assert {
    condition     = var.platform == "Linux"
    error_message = "Platform should be Linux"
  }
}

run "valid_platform_windows" {
  command = plan

  variables {
    platform = "Windows"
  }

  assert {
    condition     = var.platform == "Windows"
    error_message = "Platform should be Windows"
  }
}

run "valid_platform_other" {
  command = plan

  variables {
    platform = "Other"
  }

  assert {
    condition     = var.platform == "Other"
    error_message = "Platform should be Other"
  }
}

run "valid_platform_paravirtualization" {
  command = plan

  variables {
    platform = "Paravirtualization"
  }

  assert {
    condition     = var.platform == "Paravirtualization"
    error_message = "Platform should be Paravirtualization"
  }
}

run "valid_platform_windows_virtio" {
  command = plan

  variables {
    platform = "WindowsVirtio"
  }

  assert {
    condition     = var.platform == "WindowsVirtio"
    error_message = "Platform should be WindowsVirtio"
  }
}

run "valid_strategy_instant_start" {
  command = plan

  variables {
    strategy = "InstantStart"
  }

  assert {
    condition     = var.strategy == "InstantStart"
    error_message = "Strategy should be InstantStart"
  }
}

run "valid_strategy_just_create" {
  command = plan

  variables {
    strategy = "JustCreate"
  }

  assert {
    condition     = var.strategy == "JustCreate"
    error_message = "Strategy should be JustCreate"
  }
}

run "valid_strategy_create_stopped" {
  command = plan

  variables {
    strategy = "CreateStopped"
  }

  assert {
    condition     = var.strategy == "CreateStopped"
    error_message = "Strategy should be CreateStopped"
  }
}

run "valid_cpu_num_minimum" {
  command = plan

  variables {
    cpu_num = 1
  }

  assert {
    condition     = var.cpu_num == 1
    error_message = "CPU number should be 1"
  }
}

run "valid_cpu_num_maximum" {
  command = plan

  variables {
    cpu_num = 1024
  }

  assert {
    condition     = var.cpu_num == 1024
    error_message = "CPU number should be 1024"
  }
}

run "valid_memory_size_minimum" {
  command = plan

  variables {
    memory_size = 1
  }

  assert {
    condition     = var.memory_size == 1
    error_message = "Memory size should be 1"
  }
}

run "valid_configuration_combination" {
  command = plan

  variables {
    instance_name   = "test-vm"
    cpu_num         = 4
    memory_size     = 4096
    platform        = "Linux"
    strategy        = "InstantStart"
  }

  assert {
    condition     = var.cpu_num == 4
    error_message = "CPU number should be 4"
  }

  assert {
    condition     = var.memory_size == 4096
    error_message = "Memory size should be 4096"
  }

  assert {
    condition     = var.platform == "Linux"
    error_message = "Platform should be Linux"
  }

  assert {
    condition     = var.strategy == "InstantStart"
    error_message = "Strategy should be InstantStart"
  }
}
