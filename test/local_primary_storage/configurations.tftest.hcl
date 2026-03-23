# Copyright (c) ZStack.io, Inc.

run "create_local_primary_storage_with_all_options" {
  command = apply

  variables {
    local_storage_name        = "config-test-local-storage"
    local_storage_description = "Configuration test local primary storage with all options"
    local_storage_url         = "/vms_ds"
  }

  assert {
    condition     = zsphere_local_primary_storage.test.name == "config-test-local-storage"
    error_message = "Local primary storage name should match"
  }

  assert {
    condition     = zsphere_local_primary_storage.test.description == "Configuration test local primary storage with all options"
    error_message = "Local primary storage description should match"
  }

  assert {
    condition     = zsphere_local_primary_storage.test.url == "/vms_ds"
    error_message = "URL should match"
  }
}

run "verify_local_primary_storage_computed_fields" {
  command = apply

  assert {
    condition     = zsphere_local_primary_storage.test.uuid != null
    error_message = "UUID should be computed"
  }

  assert {
    condition     = zsphere_local_primary_storage.test.type != null
    error_message = "Type should be computed"
  }

  assert {
    condition     = zsphere_local_primary_storage.test.state != null
    error_message = "State should be computed"
  }

  assert {
    condition     = zsphere_local_primary_storage.test.status != null
    error_message = "Status should be computed"
  }

  assert {
    condition     = zsphere_local_primary_storage.test.mount_path != null
    error_message = "Mount path should be computed"
  }
}

run "update_local_primary_storage_description" {
  command = apply

  variables {
    local_storage_description = "Updated description"
  }

  assert {
    condition     = zsphere_local_primary_storage.test.description == "Updated description"
    error_message = "Description should be updated"
  }
}
