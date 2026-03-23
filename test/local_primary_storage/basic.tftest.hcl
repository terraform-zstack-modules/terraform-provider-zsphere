# Copyright (c) ZStack.io, Inc.

run "create_local_primary_storage" {
  command = apply

  variables {
    local_storage_name        = "test-local-storage"
    local_storage_description = "Basic test local primary storage"
    local_storage_url         = "/vms_ds"
  }

  assert {
    condition     = zsphere_local_primary_storage.test.name == "test-local-storage"
    error_message = "Local primary storage name should match"
  }

  assert {
    condition     = zsphere_local_primary_storage.test.url == "/vms_ds"
    error_message = "Local primary storage URL should match"
  }

  assert {
    condition     = zsphere_local_primary_storage.test.uuid != null
    error_message = "UUID should be computed"
  }

  assert {
    condition     = zsphere_local_primary_storage.test.state == "Enabled"
    error_message = "Local primary storage state should be Enabled"
  }

  assert {
    condition     = zsphere_local_primary_storage.test.mount_path != null
    error_message = "Mount path should be computed"
  }
}
