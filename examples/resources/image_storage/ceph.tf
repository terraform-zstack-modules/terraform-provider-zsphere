data "zsphere_datacenter" "test" {
  name = "zone_example"
}

resource "zsphere_image_storage" "ceph" {
  name        = "ceph_backup_storage_from_terraform"
  description = "Add An example Ceph backup storage from terraform"
  type        = "Ceph"
  pool_name   = "images"
  mon_urls    = ["ceph-user:password@192.168.1.10:22"]
  zone_uuid   = data.zsphere_datacenter.test.data_centers[0].uuid
}

output "zsphere_image_storage_ceph" {
  value = zsphere_image_storage.ceph
}
