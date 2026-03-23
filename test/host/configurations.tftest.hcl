# Copyright (c) ZStack.io, Inc.

run "create_host_with_all_options" {
  command = apply

  variables {
    host_name          = "config-test-host"
    host_description   = "Configuration test host with all options"
    host_management_ip = "192.168.1.102"
    host_username      = "admin"
    host_password      = "admin123"
    host_ssh_port      = 22
    host_iommu         = true
    host_ept           = false
  }

  assert {
    condition     = zsphere_host.test.name == "config-test-host"
    error_message = "Host name should match"
  }

  assert {
    condition     = zsphere_host.test.description == "Configuration test host with all options"
    error_message = "Host description should match"
  }

  assert {
    condition     = zsphere_host.test.management_ip == "192.168.1.102"
    error_message = "Management IP should match"
  }

  assert {
    condition     = zsphere_host.test.ssh_port == 22
    error_message = "SSH port should match"
  }
}

run "verify_host_computed_fields" {
  command = apply

  assert {
    condition     = zsphere_host.test.zone_uuid != null
    error_message = "Zone UUID should be computed"
  }

  assert {
    condition     = zsphere_host.test.hypervisor_type != null
    error_message = "Hypervisor type should be computed"
  }

  assert {
    condition     = zsphere_host.test.architecture != null
    error_message = "Architecture should be computed"
  }

  assert {
    condition     = zsphere_host.test.status != null
    error_message = "Status should be computed"
  }
}

run "update_host_description" {
  command = apply

  variables {
    host_description = "Updated description"
  }

  assert {
    condition     = zsphere_host.test.description == "Updated description"
    error_message = "Description should be updated"
  }
}
