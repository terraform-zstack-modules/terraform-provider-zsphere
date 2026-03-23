# Copyright (c) ZStack.io, Inc.

run "create_iso_image" {
  command = apply

  variables {
    image_name         = "test-iso-image"
    image_url          = var.image_url_iso
    image_format       = "iso"
    image_media_type   = "ISO"
    image_architecture = "x86_64"
    image_platform     = "Linux"
  }

  assert {
    condition     = zsphere_image.test.media_type == "ISO"
    error_message = "Image media_type should be ISO"
  }

  assert {
    condition     = zsphere_image.test.format == "iso"
    error_message = "Image format should be iso"
  }

  assert {
    condition     = zsphere_image.test.expunge == false
    error_message = "Expunge should be false"
  }
}

run "create_aarch64_image" {
  command = apply

  variables {
    image_name         = "test-aarch64-image"
    image_url          = var.image_url_qcow2
    image_format       = "qcow2"
    image_architecture = "aarch64"
    image_boot_mode    = "UEFI"
    image_platform     = "Linux"
  }

  assert {
    condition     = zsphere_image.test.architecture == "aarch64"
    error_message = "Image architecture should be aarch64"
  }

  assert {
    condition     = zsphere_image.test.boot_mode == "UEFI"
    error_message = "Boot mode should be UEFI for aarch64"
  }

  assert {
    condition     = zsphere_image.test.expunge == false
    error_message = "Expunge should be false"
  }
}

run "create_windows_image" {
  command = apply

  variables {
    image_name          = "test-windows-image"
    image_url           = var.image_url_qcow2
    image_format        = "qcow2"
    image_architecture  = "x86_64"
    image_platform      = "Windows"
    image_guest_os_type = "Windows"
    image_virtio        = true
  }

  assert {
    condition     = zsphere_image.test.platform == "Windows"
    error_message = "Image platform should be Windows"
  }

  assert {
    condition     = zsphere_image.test.guest_os_type == "Windows"
    error_message = "Image guest_os_type should be Windows"
  }

  assert {
    condition     = zsphere_image.test.virtio == true
    error_message = "VirtIO should be enabled"
  }

  assert {
    condition     = zsphere_image.test.expunge == false
    error_message = "Expunge should be false"
  }
}

run "create_uefi_csm_image" {
  command = apply

  variables {
    image_name         = "test-uefi-csm-image"
    image_url          = var.image_url_qcow2
    image_format       = "qcow2"
    image_architecture = "x86_64"
    image_boot_mode    = "UEFI_WITH_CSM"
    image_platform     = "Linux"
  }

  assert {
    condition     = zsphere_image.test.boot_mode == "UEFI_WITH_CSM"
    error_message = "Boot mode should be UEFI_WITH_CSM"
  }

  assert {
    condition     = zsphere_image.test.expunge == false
    error_message = "Expunge should be false"
  }
}

run "create_data_volume_template" {
  command = apply

  variables {
    image_name         = "test-data-volume-template"
    image_url          = var.image_url_qcow2
    image_format       = "qcow2"
    image_media_type   = "DataVolumeTemplate"
    image_architecture = "x86_64"
    image_platform     = "Linux"
  }

  assert {
    condition     = zsphere_image.test.media_type == "DataVolumeTemplate"
    error_message = "Image media_type should be DataVolumeTemplate"
  }

  assert {
    condition     = zsphere_image.test.expunge == false
    error_message = "Expunge should be false"
  }
}

run "create_image_with_expunge" {
  command = apply

  variables {
    image_name         = "test-expunge-image"
    image_url          = var.image_url_qcow2
    image_format       = "qcow2"
    image_architecture = "x86_64"
    image_platform     = "Linux"
    image_expunge      = true
  }

  assert {
    condition     = zsphere_image.test.expunge == true
    error_message = "Expunge should be true"
  }
}
