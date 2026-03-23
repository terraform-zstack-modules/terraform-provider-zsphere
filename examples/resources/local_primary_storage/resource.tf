resource "zsphere_local_primary_storage" "local_storage" {
  name            = "local-storage-from-terraform"
  description     = "Add an example local primary storage from terraform"
  datacenter_uuid = "datacenter-uuid-here"
  cluster_uuid    = "cluster-uuid-here"
  url             = "/vms_ds"
}

output "zsphere_local_primary_storage" {
  value = zsphere_local_primary_storage.local_storage
}
