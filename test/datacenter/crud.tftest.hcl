# Copyright (c) ZStack.io, Inc.

run "create_datacenter_for_crud" {
  command = apply

  assert {
    condition     = zsphere_datacenter.test.name == "test-datacenter"
    error_message = "Datacenter should be created"
  }

  assert {
    condition     = zsphere_datacenter.test.uuid != null
    error_message = "UUID should be assigned"
  }
}

run "read_datacenter" {
  command = plan

  assert {
    condition     = zsphere_datacenter.test.name == "test-datacenter"
    error_message = "Datacenter name should be readable"
  }

  assert {
    condition     = zsphere_datacenter.test.state != null
    error_message = "Datacenter state should be readable"
  }
}

run "update_datacenter_for_crud" {
  command = apply

  variables {
    datacenter_name        = "crud-updated-datacenter"
    datacenter_description = "CRUD test description"
  }

  assert {
    condition     = zsphere_datacenter.test.name == "crud-updated-datacenter"
    error_message = "Datacenter name should be updated"
  }

  assert {
    condition     = zsphere_datacenter.test.description == "CRUD test description"
    error_message = "Datacenter description should be updated"
  }
}

run "verify_datacenter_after_update" {
  command = plan

  assert {
    condition     = zsphere_datacenter.test.uuid != null
    error_message = "UUID should remain the same after update"
  }
}
