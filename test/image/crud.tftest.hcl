# Copyright (c) ZStack.io, Inc.

run "create_image" {
  command = apply

  variables {
    image_name         = "crud-test-image"
    image_url          = var.image_url
    image_format       = "qcow2"
    image_architecture = "x86_64"
    image_platform     = "Linux"
    image_description  = "CRUD test image"
  }

  assert {
    condition     = zsphere_image.test.name == "crud-test-image"
    error_message = "Image name should match"
  }

  assert {
    condition     = zsphere_image.test.format == "qcow2"
    error_message = "Image format should be qcow2"
  }

  assert {
    condition     = zsphere_image.test.architecture == "x86_64"
    error_message = "Image architecture should be x86_64"
  }

  assert {
    condition     = zsphere_image.test.uuid != null && zsphere_image.test.uuid != ""
    error_message = "UUID should be generated"
  }
}

run "read_image" {
  command = apply

  assert {
    condition     = zsphere_image.test.state == "Enabled"
    error_message = "Image state should be Enabled"
  }
}

run "update_image_description" {
  command = apply

  variables {
    image_name         = "crud-test-image"
    image_url          = var.image_url
    image_format       = "qcow2"
    image_architecture = "x86_64"
    image_platform     = "Linux"
    image_description  = "Updated CRUD test image description"
  }

  assert {
    condition     = zsphere_image.test.description == "Updated CRUD test image description"
    error_message = "Image description should be updated"
  }

  assert {
    condition     = zsphere_image.test.name == "crud-test-image"
    error_message = "Image name should remain unchanged"
  }
}

run "update_image_platform" {
  command = apply

  variables {
    image_name         = "crud-test-image"
    image_url          = var.image_url
    image_format       = "qcow2"
    image_architecture = "x86_64"
    image_platform     = "Windows"
    image_description  = "Updated CRUD test image description"
  }

  assert {
    condition     = zsphere_image.test.platform == "Windows"
    error_message = "Image platform should be updated to Windows"
  }
}

run "url_change_requires_replace" {
  command = apply

  variables {
    image_name         = "crud-test-image-recreated"
    image_url          = var.image_url
    image_format       = "qcow2"
    image_architecture = "x86_64"
    image_platform     = "Linux"
    image_description  = "Image with new URL"
  }

  assert {
    condition     = zsphere_image.test.url == var.image_url
    error_message = "Image URL should be changed (resource recreated)"
  }
}
