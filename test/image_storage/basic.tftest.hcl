# Copyright (c) ZStack.io, Inc.

run "create_basic_image_storage" {
  command = apply

  variables {
    image_storage_type = "ImageStore"
  }

  assert {
    condition     = zsphere_image_storage.test.name == "test-image-storage"
    error_message = "Image storage name should be 'test-image-storage'"
  }

  assert {
    condition     = zsphere_image_storage.test.type == "ImageStore"
    error_message = "Image storage type should be 'ImageStore'"
  }

  assert {
    condition     = zsphere_image_storage.test.hostname != null && zsphere_image_storage.test.hostname != ""
    error_message = "Image storage hostname should be set"
  }

  assert {
    condition     = zsphere_image_storage.test.url != null && zsphere_image_storage.test.url != ""
    error_message = "Image storage url should be set"
  }
}

run "verify_computed_attributes" {
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

  assert {
    condition     = zsphere_image_storage.test.total_capacity != null
    error_message = "Total capacity should be computed"
  }

  assert {
    condition     = zsphere_image_storage.test.available_capacity != null
    error_message = "Available capacity should be computed"
  }
}

run "update_image_storage_name" {
  command = apply

  variables {
    image_storage_name = "test-image-storage-updated"
  }

  assert {
    condition     = zsphere_image_storage.test.name == "test-image-storage-updated"
    error_message = "Image storage name should be updated"
  }

  assert {
    condition     = zsphere_image_storage.test.uuid != null
    error_message = "UUID should remain unchanged"
  }
}

run "update_image_storage_description" {
  command = apply

  variables {
    image_storage_description = "Updated description"
  }

  assert {
    condition     = zsphere_image_storage.test.description == "Updated description"
    error_message = "Image storage description should be updated"
  }
}
