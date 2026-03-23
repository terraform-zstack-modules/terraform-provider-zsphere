# Copyright (c) ZStack.io, Inc.

run "create_port_group" {
  command = apply

  assert {
    condition     = zsphere_port_group.test.name == "test-port-group"
    error_message = "Port group name should be 'test-port-group'"
  }

  assert {
    condition     = zsphere_port_group.test.vlan == 100
    error_message = "Port group VLAN should be 100"
  }

  assert {
    condition     = zsphere_port_group.test.uuid != null
    error_message = "UUID should be generated"
  }
}

run "read_port_group" {
  command = plan

  assert {
    condition     = zsphere_port_group.test.uuid == run.create_port_group.port_group_uuid
    error_message = "UUID should match"
  }
}

run "update_port_group" {
  command = apply

  variables {
    port_group_name        = "updated-port-group"
    port_group_description = "Updated description"
  }

  assert {
    condition     = zsphere_port_group.test.name == "updated-port-group"
    error_message = "Port group name should be updated"
  }

  assert {
    condition     = zsphere_port_group.test.description == "Updated description"
    error_message = "Port group description should be updated"
  }

  assert {
    condition     = zsphere_port_group.test.uuid == run.create_port_group.port_group_uuid
    error_message = "UUID should remain unchanged"
  }
}
