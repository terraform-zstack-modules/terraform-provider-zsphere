resource "zsphere_host" "host" {
  name          = "host_from_terraform"
  description   = "Add an example host from terraform"
  management_ip = "192.168.1.100"
  cluster_uuid  = "uuid of the cluster"
  username      = "root"
  password      = "password"
  ssh_port      = 22
  iommu         = false
  ept           = true
}

output "zsphere_host" {
  value = zsphere_host.host
}

output "host_uuid" {
  value       = zsphere_host.host.uuid
  description = "The UUID of the host"
}

output "host_status" {
  value       = zsphere_host.host.status
  description = "The status of the host"
}
