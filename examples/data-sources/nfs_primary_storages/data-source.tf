data "zsphere_nfs_primary_storages" "test" {
}

output "zsphere_nfs_storages" {
  value = data.zsphere_nfs_primary_storages.test
}

# Example output:
# zsphere_nfs_storages = {
#   "filter" = tolist([])
#   "nfs_storages" = tolist([
#     {
#       "available_capacity" = 50000000000
#       "available_physical_capacity" = 50000000000
#       "mount_path" = "/share/nfs"
#       "name" = "NFS-1"
#       "state" = "Enabled"
#       "status" = "Connected"
#       "system_used_capacity" = 0
#       "total_capacity" = 100000000000
#       "total_physical_capacity" = 100000000000
#       "type" = "NFS"
#       "url" = "192.168.1.100:/share/nfs"
#       "uuid" = "d1058cff5c864a84848e0ab896736578"
#       "datacenter_uuid" = "86643f5442164890b8b71de5b582dc9f"
#     },
#   ])
#   "name" = tostring(null)
#   "name_pattern" = tostring(null)
# }
