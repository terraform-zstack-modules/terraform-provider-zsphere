# Copyright (c) ZStack.io, Inc.

run "create_datacenter_with_default_true" {
  command = apply

  variables {
    datacenter_is_default = true
  }

  assert {
    condition     = zsphere_datacenter.test.is_default == true
    error_message = "Datacenter is_default should be true"
  }
}

run "create_datacenter_with_empty_description" {
  command = apply

  variables {
    datacenter_name        = "test-datacenter-no-desc"
    datacenter_description = ""
    datacenter_is_default  = false
  }

  assert {
    condition     = zsphere_datacenter.test.name == "test-datacenter-no-desc"
    error_message = "Datacenter name should match"
  }

  assert {
    condition     = zsphere_datacenter.test.uuid != null
    error_message = "UUID should be assigned even with empty description"
  }
}

run "create_datacenter_special_chars" {
  command = apply

  variables {
    datacenter_name        = "test-datacenter-123"
    datacenter_description = "Test with numbers"
    datacenter_is_default  = false
  }

  assert {
    condition     = zsphere_datacenter.test.name == "test-datacenter-123"
    error_message = "Datacenter name with numbers should work"
  }
}
