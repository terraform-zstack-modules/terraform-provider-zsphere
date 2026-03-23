# Copyright (c) ZStack.io, Inc.

run "create_nfs_primary_storage_with_all_options" {
  command = apply

  variables {
    nfs_storage_name          = "config-test-nfs-storage"
    nfs_storage_description   = "Configuration test NFS primary storage with all options"
    nfs_storage_url           = "192.168.1.100:/share/nfs"
    nfs_storage_cidr          = "192.168.1.0/24"
    nfs_storage_mount_options  = "rw,sync"
  }

  assert {
    condition     = zsphere_nfs_primary_storage.test.name == "config-test-nfs-storage"
    error_message = "NFS primary storage name should match"
  }

  assert {
    condition     = zsphere_nfs_primary_storage.test.description == "Configuration test NFS primary storage with all options"
    error_message = "NFS primary storage description should match"
  }

  assert {
    condition     = zsphere_nfs_primary_storage.test.url == "192.168.1.100:/share/nfs"
    error_message = "URL should match"
  }
}

run "verify_nfs_primary_storage_computed_fields" {
  command = apply

  assert {
    condition     = zsphere_nfs_primary_storage.test.uuid != null
    error_message = "UUID should be computed"
  }

  assert {
    condition     = zsphere_nfs_primary_storage.test.type != null
    error_message = "Type should be computed"
  }

  assert {
    condition     = zsphere_nfs_primary_storage.test.state != null
    error_message = "State should be computed"
  }

  assert {
    condition     = zsphere_nfs_primary_storage.test.status != null
    error_message = "Status should be computed"
  }

  assert {
    condition     = zsphere_nfs_primary_storage.test.mount_path != null
    error_message = "Mount path should be computed"
  }
}

run "update_nfs_primary_storage_description" {
  command = apply

  variables {
    nfs_storage_description = "Updated description"
  }

  assert {
    condition     = zsphere_nfs_primary_storage.test.description == "Updated description"
    error_message = "Description should be updated"
  }
}
