data "zsphere_instances" "test" {
}

output "zstack_secs" {
  value = data.zsphere_instances.test
}