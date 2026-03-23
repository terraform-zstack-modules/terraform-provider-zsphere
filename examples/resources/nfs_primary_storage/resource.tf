resource "zsphere_nfs_primary_storage" "nfs_storage" {
  name            = "nfs-storage-from-terraform"
  description     = "Add an example NFS primary storage from terraform"
  datacenter_uuid = "datacenter-uuid-here"
  cluster_uuid    = "cluster-uuid-here"
  url             = "192.168.1.100:/share/nfs"
  cidr            = "192.168.1.0/24"
  mount_options   = "rw,sync"
}

output "zsphere_nfs_primary_storage" {
  value = zsphere_nfs_primary_storage.nfs_storage
}
