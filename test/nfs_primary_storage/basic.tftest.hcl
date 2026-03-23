# Copyright (c) ZStack.io, Inc.

run "create_nfs_primary_storage" {
  command = apply

  variables {
    nfs_storage_name        = "test-nfs-storage"
    nfs_storage_description = "Basic test NFS primary storage"
    nfs_storage_url         = "192.168.1.100:/share/nfs"
  }

  assert {
    condition     = zsphere_nfs_primary_storage.test.name == "test-nfs-storage"
    error_message = "NFS primary storage name should match"
  }

  assert {
    condition     = zsphere_nfs_primary_storage.test.url == "192.168.1.100:/share/nfs"
    error_message = "NFS primary storage URL should match"
  }

  assert {
    condition     = zsphere_nfs_primary_storage.test.uuid != null
    error_message = "UUID should be computed"
  }

  assert {
    condition     = zsphere_nfs_primary_storage.test.state == "Enabled"
    error_message = "NFS primary storage state should be Enabled"
  }

  assert {
    condition     = zsphere_nfs_primary_storage.test.mount_path != null
    error_message = "Mount path should be computed"
  }
}
