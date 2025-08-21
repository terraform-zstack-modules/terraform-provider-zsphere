data "zsphere_image_storages" "test" {
  name = "BS-1-勿删"
}

resource "zsphere_image" "image" {
  name                = "test-from-terraform"
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