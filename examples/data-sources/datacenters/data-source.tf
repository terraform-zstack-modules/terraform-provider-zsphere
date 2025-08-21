data "zsphere_datacenters" "test" {
}

output "zstack_secs" {
  value = data.zsphere_datacenters.test
}