resource "zsphere_port_group" "example" {
  name         = "example-port-group"
  description  = "Example port group with full IPAM configuration"
  vswitch_uuid = "85f9149abb6f42bc9b8b700abc93be67" # L2Network UUID
  vlan         = 100
  vlan_mode    = "PVLAN" # ACCESS, NONE, PVLAN, TRUNK
  category     = "Private"
  dns_domain   = "example.com"
  ip_version   = 4 # 4 or 6
  enable_ipam  = true
  dhcp_service = true
  dhcp_ip      = "192.168.100.50"

  dns = ["8.8.8.8", "8.8.4.4"]

  ip_range {
    name                 = "primary-ip-range"
    start_ip             = "192.168.100.100"
    end_ip               = "192.168.100.200"
    netmask              = "255.255.255.0"
    gateway              = "192.168.100.1"
    ip_allocate_strategy = "RandomIpAllocator"
  }
}

output "port_group_example" {
  value = zsphere_port_group.example
}
