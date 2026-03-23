# Copyright (c) ZStack.io, Inc.

run "create_host" {
  command = apply

  variables {
    host_name          = "crud-test-host"
    host_description   = "CRUD test host"
    host_management_ip = "192.168.1.101"
    host_username      = "root"
    host_password      = "password"
    host_ssh_port      = 22
    host_iommu         = false
    host_ept           = true
  }

  assert {
    condition     = zsphere_host.test.uuid != null
    error_message = "UUID should be computed after create"
  }
}

run "read_host" {
  command = apply

  variables {
    host_name          = "crud-test-host"
    host_management_ip = "192.168.1.101"
  }

  assert {
    condition     = zsphere_host.test.name == "crud-test-host"
    error_message = "Host name should be readable"
  }

  assert {
    condition     = zsphere_host.test.management_ip == "192.168.1.101"
    error_message = "Host management IP should be readable"
  }
}

run "update_host_name" {
  command = apply

  variables {
    host_name          = "updated-crud-host"
    host_management_ip = "192.168.1.101"
  }

  assert {
    condition     = zsphere_host.test.name == "updated-crud-host"
    error_message = "Host name should be updated"
  }
}

run "verify_updated_host" {
  command = apply

  variables {
    host_name          = "updated-crud-host"
    host_management_ip = "192.168.1.101"
  }

  assert {
    condition     = zsphere_host.test.name == "updated-crud-host"
    error_message = "Updated host name should persist"
  }
}