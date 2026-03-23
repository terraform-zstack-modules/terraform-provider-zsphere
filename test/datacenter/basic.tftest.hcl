# Copyright (c) ZStack.io, Inc.

run "create_basic_datacenter" {
  command = apply

  assert {
    condition     = zsphere_datacenter.test.name == "test-datacenter"
    error_message = "Datacenter name should be 'test-datacenter'"
  }

  assert {
    condition     = zsphere_datacenter.test.state == "Enabled"
    error_message = "Datacenter state should be Enabled after creation"
  }

  assert {
    condition     = zsphere_datacenter.test.type == "zstack"
    error_message = "Datacenter type should be 'zstack'"
  }
}

run "verify_computed_attributes" {
  command = apply

  assert {
    condition     = zsphere_datacenter.test.uuid != null && zsphere_datacenter.test.uuid != ""
    error_message = "UUID should be computed and non-empty"
  }

  assert {
    condition     = zsphere_datacenter.test.state != null
    error_message = "State should be computed"
  }

  assert {
    condition     = zsphere_datacenter.test.type != null
    error_message = "Type should be computed"
  }
}

run "update_datacenter_name" {
  command = apply

  variables {
    datacenter_name = "test-datacenter-updated"
  }

  assert {
    condition     = zsphere_datacenter.test.name == "test-datacenter-updated"
    error_message = "Datacenter name should be updated"
  }

  assert {
    condition     = zsphere_datacenter.test.uuid != null
    error_message = "UUID should remain unchanged"
  }
}

run "update_datacenter_description" {
  command = apply

  variables {
    datacenter_description = "Updated description"
  }

  assert {
    condition     = zsphere_datacenter.test.description == "Updated description"
    error_message = "Datacenter description should be updated"
  }
}
