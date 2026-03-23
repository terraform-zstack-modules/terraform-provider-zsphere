resource "zsphere_datacenter" "datacenter" {
  name        = "datacenter_from_terraform"
  description = "Add an example datacenter from terraform"
  is_default  = false
}

output "zsphere_datacenter" {
  value = zsphere_datacenter.datacenter
}
