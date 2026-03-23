data "zsphere_port_groups" "test" {
}

output "zstack_secs" {
  value = data.zsphere_port_groups.test
}

# example output:
# zstack_secs = {
#   "filter" = tolist([])
#   "name" = tostring(null)
#   "name_pattern" = tostring(null)
#   "port_groups" = tolist([
#     {
#       "category" = "Private"
#       "dns" = tolist([])
#       "free_ips" = tolist([
#         {
#           "gateway" = "192.168.1.1"
#           "ip" = "192.168.1.100"
#           "ip_range_uuid" = "9199704321e94de98f60eb06ba563b53"
#           "netmask" = "255.255.255.0"
#         },
#         {
#           "gateway" = "192.168.1.1"
#           "ip" = "192.168.1.101"
#           "ip_range_uuid" = "9199704321e94de98f60eb06ba563b53"
#           "netmask" = "255.255.255.0"
#         },
#         {
#           "gateway" = "192.168.1.1"
#           "ip" = "192.168.1.102"
#           "ip_range_uuid" = "9199704321e94de98f60eb06ba563b53"
#           "netmask" = "255.255.255.0"
#         },
#       ])
#       "ip_range" = tolist([
#         {
#           "cidr" = "192.168.1.0/24"
#           "end_ip" = "192.168.1.102"
#           "gateway" = "192.168.1.1"
#           "ip_range_name" = "192.168.1.100-192.168.1.102"
#           "netmask" = "255.255.255.0"
#           "start_ip" = "192.168.1.100"
#         },
#       ])
#       "name" = "Port-group-1"
#       "uuid" = "5cc7cd5963fd45149901336116a261a6"
#     },
#   ])
# }