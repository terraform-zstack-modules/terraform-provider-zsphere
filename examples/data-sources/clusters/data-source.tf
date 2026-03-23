data "zsphere_clusters" "test" {
}

output "zstack_secs" {
  value = data.zsphere_clusters.test
}

# example output:
# zstack_secs = {
#   "clusters" = tolist([
#     {
#       "hypervisor_type" = "KVM"
#       "name" = "Cluster-1"
#       "state" = "Enabled"
#       "type" = "zstack"
#       "uuid" = "5568b28dbaed465daf768b533b1fe72d"
#       "zone_uuid" = "86643f5442164890b8b71de5b582dc9f"
#     }
#   ])
#   "filter" = tolist([])
#   "name" = tostring(null)
#   "name_pattern" = tostring(null)
# }