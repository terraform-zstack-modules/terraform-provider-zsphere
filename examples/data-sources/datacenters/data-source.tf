data "zsphere_datacenters" "test" {
}

output "zstack_secs" {
  value = data.zsphere_datacenters.test
}

# example out:
# zstack_secs = {
#   "data_centers" = tolist([
#     {
#       "name" = "Datacenter-1"
#       "state" = "Enabled"
#       "type" = "zstack"
#       "uuid" = "86643f5442164890b8b71de5b582dc9f"
#     },
#   ])
#   "filter" = tolist([])
#   "name" = tostring(null)
#   "name_pattern" = tostring(null)
# }