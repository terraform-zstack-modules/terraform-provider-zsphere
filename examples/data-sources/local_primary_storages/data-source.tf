data "zsphere_local_primary_storages" "test" {
}

output "zsphere_local_storages" {
  value = data.zsphere_local_primary_storages.test
}

# Example output:
# zsphere_local_storages = {
#   "filter" = tolist([])
#   "local_storages" = tolist([
#     {
#       "available_capacity" = 50000000000
#       "available_physical_capacity" = 50000000000
#       "mount_path" = "/vms_ds"
#       "name" = "LocalStorage-1"
#       "state" = "Enabled"
#       "status" = "Connected"
#       "system_used_capacity" = 0
#       "total_capacity" = 100000000000
#       "total_physical_capacity" = 100000000000
#       "type" = "LocalStorage"
#       "url" = "/vms_ds"
#       "uuid" = "d1058cff5c864a84848e0ab896736578"
#       "zone_uuid" = "86643f5442164890b8b71de5b582dc9f"
#     },
#   ])
#   "name" = tostring(null)
#   "name_pattern" = tostring(null)
# }
