data "zsphere_images" "test" {
}

output "zstack_secs" {
  value = data.zsphere_images.test
}