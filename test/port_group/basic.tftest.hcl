# Copyright (c) ZStack.io, Inc.

run "create_basic_port_group" {
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
    condition     = zsphere_port_group.test.vlan_mode == "ACCESS"
    error_message = "Port group VLAN mode should be 'ACCESS'"
  }

  assert {
    condition     = zsphere_port_group.test.category == "Private"
    error_message = "Port group category should be 'Private'"
  }
}

run "verify_computed_attributes" {
  command = apply

  assert {
    condition     = zsphere_port_group.test.uuid != null && zsphere_port_group.test.uuid != ""
    error_message = "UUID should be computed and non-empty"
  }

  assert {
    condition     = zsphere_port_group.test.state != null
    error_message = "State should be computed"
  }

  assert {
    condition     = zsphere_port_group.test.datacenter_uuid != null && zsphere_port_group.test.datacenter_uuid != ""
    error_message = "Datacenter UUID should be set"
  }
}

run "update_port_group_name" {
  command = apply

  variables {
    port_group_name = "test-port-group-updated"
  }

  assert {
    condition     = zsphere_port_group.test.name == "test-port-group-updated"
    error_message = "Port group name should be updated"
  }

  assert {
    condition     = zsphere_port_group.test.uuid != null
    error_message = "UUID should remain unchanged"
  }
}

run "update_port_group_description" {
  command = apply

  variables {
    port_group_description = "Updated description"
  }

  assert {
    condition     = zsphere_port_group.test.description == "Updated description"
    error_message = "Port group description should be updated"
  }
}

run "update_port_group_category" {
  command = apply

  variables {
    port_group_category = "Public"
  }

  assert {
    condition     = zsphere_port_group.test.category == "Public"
    error_message = "Port group category should be updated"
  }
}
