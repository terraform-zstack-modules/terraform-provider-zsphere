# Copyright (c) ZStack.io, Inc.

# CRUD operation tests for zsphere_instance resource

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
}

run "update_instance_name" {
  command = apply

  variables {
    instance_name = "updated-vm-instance"
  }

  assert {
    condition     = zsphere_instance.test.name == "updated-vm-instance"
    error_message = "Instance name should be updated"
  }
}

run "update_instance_description" {
  command = apply

  variables {
    instance_name        = "updated-vm-instance"
    instance_description = "Updated description"
  }

  assert {
    condition     = zsphere_instance.test.description == "Updated description"
    error_message = "Instance description should be updated"
  }
}

run "update_instance_resources" {
  command = apply

  variables {
    instance_name   = "updated-vm-instance"
    cpu_num         = 4
    memory_size     = 4096
  }

  assert {
    condition     = zsphere_instance.test.cpu_num == 4
    error_message = "Instance CPU number should be updated to 4"
  }

  assert {
    condition     = zsphere_instance.test.memory_size == 4096
    error_message = "Instance memory size should be updated to 4096 MB"
  }
}

run "update_instance_platform" {
  command = apply

  variables {
    instance_name = "updated-vm-instance"
    platform      = "Windows"
  }

  assert {
    condition     = zsphere_instance.test.platform == "Windows"
    error_message = "Instance platform should be updated to Windows"
  }
}
