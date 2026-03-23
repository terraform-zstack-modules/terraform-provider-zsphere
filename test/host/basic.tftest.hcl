# Copyright (c) ZStack.io, Inc.

run "create_host" {
  command = apply

  variables {
    host_name          = "test-host"
    host_description   = "Basic test host"
    host_management_ip = "192.168.1.100"
    host_username      = "root"
    host_password      = "password"
    host_ssh_port      = 22
    host_iommu         = false
    host_ept           = true
  }

  assert {
    condition     = zsphere_host.test.name == "test-host"
    error_message = "Host name should match"
  }

  assert {
    condition     = zsphere_host.test.management_ip == "192.168.1.100"
    error_message = "Host management IP should match"
  }

  assert {
    condition     = zsphere_host.test.uuid != null
    error_message = "UUID should be computed"
  }

  assert {
    condition     = zsphere_host.test.cluster_uuid != null
    error_message = "Cluster UUID should be set"
  }

  assert {
    condition     = zsphere_host.test.state == "Enabled"
    error_message = "Host state should be Enabled"
  }
}
