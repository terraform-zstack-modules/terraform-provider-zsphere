# Copyright (c) ZStack.io, Inc.

run "create_port_group" {
  command = apply

  assert {
    condition     = zsphere_port_group.test.name != ""
    error_message = "Port group name should not be empty"
  }

  assert {
    condition     = zsphere_port_group.test.vlan > 0
    error_message = "VLAN should be positive"
  }

  assert {
    condition     = can(regex("^[a-f0-9-]+$", zsphere_port_group.test.uuid))
    error_message = "UUID should be a valid UUID format"
  }
}

run "verify_configurations" {
  command = plan

  assert {
    condition     = zsphere_port_group.test.state == "Enabled"
    error_message = "Port group should be in Enabled state"
  }

  assert {
    condition     = zsphere_port_group.test.ip_version == 4
    error_message = "IP version should be 4"
  }

  assert {
    condition     = zsphere_port_group.test.enable_ipam == false
    error_message = "IPAM should be disabled by default"
  }
}
