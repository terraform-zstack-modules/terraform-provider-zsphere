# Copyright (c) ZStack.io, Inc.

run "create_basic_distributed_switch" {
  command = apply

  assert {
    condition     = zsphere_distributed_switch.test.name == "test-distributed-switch"
    error_message = "Distributed switch name should be 'test-distributed-switch'"
  }

  assert {
    condition     = zsphere_distributed_switch.test.vswitch_type == "LinuxBridge"
    error_message = "Distributed switch vswitch type should be 'LinuxBridge'"
  }
}

run "verify_computed_attributes" {
  command = apply

  assert {
    condition     = zsphere_distributed_switch.test.uuid != null && zsphere_distributed_switch.test.uuid != ""
    error_message = "UUID should be computed and non-empty"
  }

  assert {
    condition     = zsphere_distributed_switch.test.type != null && zsphere_distributed_switch.test.type != ""
    error_message = "Type should be computed"
  }

  assert {
    condition     = zsphere_distributed_switch.test.datacenter_uuid != null && zsphere_distributed_switch.test.datacenter_uuid != ""
    error_message = "Datacenter UUID should be set"
  }
}

run "update_distributed_switch_name" {
  command = apply

  variables {
    distributed_switch_name = "test-distributed-switch-updated"
  }

  assert {
    condition     = zsphere_distributed_switch.test.name == "test-distributed-switch-updated"
    error_message = "Distributed switch name should be updated"
  }

  assert {
    condition     = zsphere_distributed_switch.test.uuid != null
    error_message = "UUID should remain unchanged"
  }
}

run "update_distributed_switch_description" {
  command = apply

  variables {
    distributed_switch_description = "Updated description"
  }

  assert {
    condition     = zsphere_distributed_switch.test.description == "Updated description"
    error_message = "Distributed switch description should be updated"
  }
}
