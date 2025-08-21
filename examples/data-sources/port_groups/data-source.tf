data "zsphere_port_groups" "test" {

}

output "zstack_secs" {
  value = data.zsphere_port_groups.test
}