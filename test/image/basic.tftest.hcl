# Copyright (c) ZStack.io, Inc.

run "create_basic_image" {
  command = apply

  assert {
    condition     = zsphere_image.test.name == "test-image"
    error_message = "Image name should be 'test-image'"
  }

  assert {
    condition     = zsphere_image.test.url == var.image_url
    error_message = "Image URL should match"
  }

  assert {
    condition     = zsphere_image.test.state == "Enabled"
    error_message = "Image state should be Enabled after creation"
  }
}

run "verify_default_values" {
  command = apply

  assert {
    condition     = zsphere_image.test.platform == "Linux"
    error_message = "Default platform should be Linux"
  }

  assert {
    condition     = zsphere_image.test.guest_os_type == "Linux"
    error_message = "Default guest_os_type should be Linux"
  }
}

run "verify_architecture_and_boot_mode" {
  command = apply

  assert {
    condition     = zsphere_image.test.architecture == "x86_64"
    error_message = "Architecture should be x86_64"
  }

  assert {
    condition     = zsphere_image.test.boot_mode == "Legacy"
    error_message = "Boot mode should be Legacy for x86_64"
  }
}

run "update_image_name" {
  command = apply

  variables {
    image_name = "test-image-updated"
  }

  assert {
    condition     = zsphere_image.test.name == "test-image-updated"
    error_message = "Image name should be updated"
  }

  assert {
    condition     = zsphere_image.test.url == var.image_url
    error_message = "Image URL should remain unchanged"
  }
}

run "verify_computed_attributes" {
  command = apply

  assert {
    condition     = zsphere_image.test.uuid != null && zsphere_image.test.uuid != ""
    error_message = "UUID should be computed and non-empty"
  }

  assert {
    condition     = zsphere_image.test.status != null
    error_message = "Status should be computed"
  }

  assert {
    condition     = zsphere_image.test.system != null
    error_message = "System should be computed"
  }
}
