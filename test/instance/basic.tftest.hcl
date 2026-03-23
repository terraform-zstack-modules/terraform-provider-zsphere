# Copyright (c) ZStack.io, Inc.

# Basic functionality tests for zsphere_instance resource

run "create_instance" {
  command = apply

  assert {
    condition     = zsphere_instance.test.name == var.instance_name
    error_message = "Instance name should match the configured value"
  }

  assert {
    condition     = zsphere_instance.test.uuid != null
    error_message = "Instance UUID should be computed"
  }
  
  assert {
    condition     = zsphere_instance.test.cpu_num == var.cpu_num
    error_message = "Instance CPU number should match the configured value"
  }

  assert {
    condition     = zsphere_instance.test.memory_size == var.memory_size
    error_message = "Instance memory size should match the configured value"
  }
}

run "verify_outputs" {
  command = plan

  assert {
    condition     = output.instance_uuid != null
    error_message = "Output instance_uuid should not be null"
  }

  assert {
    condition     = output.instance_name == var.instance_name
    error_message = "Output instance_name should match the configured value"
  }
}
