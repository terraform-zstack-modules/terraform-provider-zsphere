# Copyright (c) ZStack.io, Inc.

run "create_nfs_primary_storage" {
  command = apply

  variables {
    nfs_storage_name        = "crud-test-nfs-storage"
    nfs_storage_description = "CRUD test NFS primary storage"
    nfs_storage_url        = "192.168.1.100:/share/nfs"
  }

  assert {
    condition     = zsphere_nfs_primary_storage.test.uuid != null
    error_message = "UUID should be computed after create"
  }
}

run "read_nfs_primary_storage" {
  command = apply

  variables {
    nfs_storage_name = "crud-test-nfs-storage"
    nfs_storage_url = "192.168.1.100:/share/nfs"
  }

  assert {
    condition     = zsphere_nfs_primary_storage.test.name == "crud-test-nfs-storage"
    error_message = "NFS primary storage name should be readable"
  }

  assert {
    condition     = zsphere_nfs_primary_storage.test.url == "192.168.1.100:/share/nfs"
    error_message = "NFS primary storage URL should be readable"
  }
}

run "update_nfs_primary_storage_name" {
  command = apply

  variables {
    nfs_storage_name        = "updated-crud-nfs-storage"
    nfs_storage_description = "Updated CRUD test NFS primary storage"
    nfs_storage_url        = "192.168.1.100:/share/nfs"
  }

  assert {
    condition     = zsphere_nfs_primary_storage.test.name == "updated-crud-nfs-storage"
    error_message = "NFS primary storage name should be updated"
  }
}

run "verify_updated_nfs_primary_storage" {
  command = apply

  variables {
    nfs_storage_name        = "updated-crud-nfs-storage"
    nfs_storage_description = "Updated CRUD test NFS primary storage"
    nfs_storage_url        = "192.168.1.100:/share/nfs"
  }

  assert {
    condition     = zsphere_nfs_primary_storage.test.name == "updated-crud-nfs-storage"
    error_message = "Updated NFS primary storage name should persist"
  }
}
