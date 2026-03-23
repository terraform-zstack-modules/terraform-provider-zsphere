data "zsphere_datacenter" "test" {
  name = "zone_example"
}

resource "zsphere_image_storage" "image_store" {
  name          = "image_store_from_terraform"
  description   = "Add An example image storage from terraform"
  type          = "ImageStore"
  hostname      = "192.168.1.100"
  url           = "/zstack/imagestore"
  ssh_port      = 22
  username      = "root"
  password      = "password"
  zone_uuid     = data.zsphere_datacenter.test.data_centers[0].uuid
  import_images = false
}

resource "zsphere_image_storage" "ceph" {
  name        = "ceph_backup_storage_from_terraform"
  description = "Add An example Ceph backup storage from terraform"
  type        = "Ceph"
  pool_name   = "images"
  mon_urls    = ["ceph-user:password@192.168.1.10:6789"]
  zone_uuid   = data.zsphere_zone.test.zones[0].uuid
}

output "zsphere_image_storage_image_store" {
  value = zsphere_image_storage.image_store
}

output "zsphere_image_storage_ceph" {
  value = zsphere_image_storage.ceph
}
