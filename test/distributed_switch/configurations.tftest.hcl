# Copyright (c) ZStack.io, Inc.

run "test_vswitch_type_linuxbridge" {
  command = apply

  variables {
    distributed_switch_name        = "linuxbridge-switch"
    distributed_switch_vswitch_type = "LinuxBridge"
  }

  assert {
    condition     = zsphere_distributed_switch.test.vswitch_type == "LinuxBridge"
    error_message = "vswitch_type should be LinuxBridge"
  }
}

run "test_empty_description" {
  command = apply

  variables {
    distributed_switch_name        = "no-desc-switch"
    distributed_switch_description = ""
  }

  assert {
    condition     = zsphere_distributed_switch.test.name == "no-desc-switch"
    error_message = "Name should be set"
  }
}

run "test_special_characters_in_name" {
  command = apply

  variables {
    distributed_switch_name        = "switch-test-123"
    distributed_switch_description = "Test @#$% special chars"
  }

  assert {
    condition     = zsphere_distributed_switch.test.name == "switch-test-123"
    error_message = "Name with numbers should work"
  }
}
