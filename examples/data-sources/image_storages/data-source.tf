data "zsphere_image_storages" "test" {
}

output "zstack_secs" {
  value = data.zsphere_image_storages.test
}