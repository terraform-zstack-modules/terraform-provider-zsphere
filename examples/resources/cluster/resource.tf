resource "zsphere_cluster" "cluster" {
  name            = "cluster_from_terraform"
  description     = "Add an example cluster from terraform"
  hypervisor_type = "KVM"
  type            = "zstack"
  datacenter_uuid = "3b03dd8222294d86ac7f7ed6ca02baa6" # "uuid_of_parent_zone"
  architecture    = "x86_64"

  # Advanced features
  cpu_mode   = "hostPassthrough"
  network_hp = false

  # DRS settings (requires hypervisor_type = KVM)
  drs_enabled            = false
  drs_threshold_cpu      = 80
  drs_threshold_memory   = 80
  drs_threshold_duration = 300
}

output "zsphere_cluster" {
  value = zsphere_cluster.cluster
}
