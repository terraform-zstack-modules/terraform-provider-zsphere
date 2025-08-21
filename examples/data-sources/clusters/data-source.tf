data "zsphere_clusters" "test" {
}

output "zstack_secs" {
  value = data.zsphere_clusters.test
}