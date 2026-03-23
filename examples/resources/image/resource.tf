data "zsphere_image_storages" "test" {
  name = "image_storage_from_terraform"
}

resource "zsphere_image" "image" {
  name                = "image_from_terraform"
  description         = "Add An example image from terraform"
  url                 = "http://minio.zstack.io:9001/packer/logserver-by-packer-image-compressed.qcow2"
  guest_os_type       = "Linux"
  platform            = "Linux"
  format              = "qcow2"
  architecture        = "x86_64"
  virtio              = true
  image_storage_uuids = [data.zsphere_image_storages.test.image_storages.0.uuid]
  boot_mode           = "Legacy"
  expunge             = true

}

output "zsphere_image" {
  value = zsphere_image.image
}