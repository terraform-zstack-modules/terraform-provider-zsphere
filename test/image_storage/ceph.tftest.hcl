# Copyright (c) ZStack.io, Inc.

run "create_ceph_image_storage" {
  command = apply

  variables {
    image_storage_type = "Ceph"
  }

  assert {
    condition     = zsphere_image_storage.test.name == "test-image-storage"
    error_message = "Image storage name should be 'test-image-storage'"
  }

  assert {
    condition     = zsphere_image_storage.test.type == "Ceph"
    error_message = "Image storage type should be 'Ceph'"
  }

  assert {
    condition     = zsphere_image_storage.test.pool_name != null && zsphere_image_storage.test.pool_name != ""
    error_message = "Image storage pool_name should be set"
  }

  assert {
    condition     = zsphere_image_storage.test.mon_urls != null && length(zsphere_image_storage.test.mon_urls) > 0
    error_message = "Image storage mon_urls should be set"
  }
}

run "verify_ceph_computed_attributes" {
  command = apply

  assert {
    condition     = zsphere_image_storage.test.uuid != null && zsphere_image_storage.test.uuid != ""
    error_message = "UUID should be computed and non-empty"
  }

  assert {
    condition     = zsphere_image_storage.test.state != null
    error_message = "State should be computed"
  }

  assert {
    condition     = zsphere_image_storage.test.status != null
    error_message = "Status should be computed"
  }
}
