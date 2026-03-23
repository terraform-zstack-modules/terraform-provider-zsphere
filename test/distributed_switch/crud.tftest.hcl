# Copyright (c) ZStack.io, Inc.

run "create_distributed_switch_for_crud" {
  command = plan

  variables {
    distributed_switch_name        = "crud-test-switch"
    distributed_switch_description = "CRUD test distributed switch"
  }

  assert {
    condition     = zsphere_distributed_switch.test.name == "crud-test-switch"
    error_message = "Distributed switch name should be 'crud-test-switch'"
  }
}

run "apply_distributed_switch" {
  command = apply

  variables {
    distributed_switch_name        = "crud-test-switch"
    distributed_switch_description = "CRUD test distributed switch"
  }

  assert {
    condition     = zsphere_distributed_switch.test.uuid != null
    error_message = "UUID should be assigned after creation"
  }
}

run "read_distributed_switch" {
  command = apply

  assert {
    condition     = zsphere_distributed_switch.test.name == var.distributed_switch_name
    error_message = "Name should match"
  }

  assert {
    condition     = zsphere_distributed_switch.test.description == var.distributed_switch_description
    error_message = "Description should match"
  }
}

run "update_distributed_switch" {
  command = apply

  variables {
    distributed_switch_name        = "crud-test-switch-updated"
    distributed_switch_description = "Updated CRUD test description"
  }

  assert {
    condition     = zsphere_distributed_switch.test.name == "crud-test-switch-updated"
    error_message = "Name should be updated"
  }

  assert {
    condition     = zsphere_distributed_switch.test.description == "Updated CRUD test description"
    error_message = "Description should be updated"
  }

  assert {
    condition     = zsphere_distributed_switch.test.uuid != null
    error_message = "UUID should remain unchanged after update"
  }
}