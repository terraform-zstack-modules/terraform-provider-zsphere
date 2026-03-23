data "zsphere_image_storages" "test" {
}

output "zstack_secs" {
  value = data.zsphere_image_storages.test
}

# example output:
# zstack_secs = {
#   "filter" = tolist([])
#   "image_storages" = tolist([
#     {
#       "available_capacity" = 263827959808
#       "name" = "Host-1"
#       "state" = "Enabled"
#       "status" = "Connected"
#       "total_capacity" = 312423686144
#       "uuid" = "e3f8b6038e274f0c9354962a78f46177"
#     },
#   ])
#   "name" = tostring(null)
#   "name_pattern" = tostring(null)
# }