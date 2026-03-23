data "zsphere_hosts" "test" {
}

output "zstack_secs" {
  value = data.zsphere_hosts.test
}

# example output:
# zstack_secs = {
#   "filter" = tolist([])
#   "hosts" = tolist([
#     {
#       "architecture" = "x86_64"
#       "cluster_uuid" = "5568b28dbaed465daf768b533b1fe72d"
#       "managementip" = "172.26.52.90"
#       "name" = "Host-1"
#       "state" = "Enabled"
#       "status" = "Connected"
#       "type" = "KVM"
#       "uuid" = "d1058cff5c864a84848e0ab896736578"
#       "zone_uuid" = "86643f5442164890b8b71de5b582dc9f"
#     },
#   ])
#   "name" = tostring(null)
#   "name_pattern" = tostring(null)
# }