data "zsphere_primary_storages" "test" {

}

output "zstack_secs" {
  value = data.zsphere_primary_storages.test
}