
data "zsphere_hosts" "test" {
}

output "zstack_secs" {
  value = data.zsphere_hosts.test
}