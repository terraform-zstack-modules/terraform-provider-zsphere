# Copyright (c) ZStack.io, Inc.

run "create_local_primary_storage" {
  command = apply

  variables {
    local_storage_name        = "crud-test-local-storage"
    local_storage_description = "CRUD test local primary storage"
    local_storage_url         = "/vms_ds"
  }

  assert {
    condition     = zsphere_local_primary_storage.test.uuid != null
    error_message = "UUID should be computed after create"
  }
}

run "read_local_primary_storage" {
  command = apply

  variables {
    local_storage_name = "crud-test-local-storage"
    local_storage_url  = "/vms_ds"
  }

  assert {
    condition     = zsphere_local_primary_storage.test.name == "crud-test-local-storage"
    error_message = "Local primary storage name should be readable"
  }

  assert {
    condition     = zsphere_local_primary_storage.test.url == "/vms_ds"
    error_message = "Local primary storage URL should be readable"
  }
}

run "update_local_primary_storage_name" {
  command = apply

  variables {
    local_storage_name        = "updated-crud-local-storage"
    local_storage_description = "Updated CRUD test local primary storage"
    local_storage_url         = "/vms_ds"
  }

  assert {
    condition     = zsphere_local_primary_storage.test.name == "updated-crud-local-storage"
    error_message = "Local primary storage name should be updated"
  }
}

run "verify_updated_local_primary_storage" {
  command = apply

  variables {
    local_storage_name        = "updated-crud-local-storage"
    local_storage_description = "Updated CRUD test local primary storage"
    local_storage_url         = "/vms_ds"
  }

  assert {
    condition     = zsphere_local_primary_storage.test.name == "updated-crud-local-storage"
    error_message = "Updated local primary storage name should persist"
  }
}
