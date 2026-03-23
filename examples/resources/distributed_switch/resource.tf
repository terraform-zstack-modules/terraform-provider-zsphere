resource "zsphere_distributed_switch" "distributed_switch" {
  name            = "distributed_switch_example"
  description     = "Distributed switch created from terraform"
  datacenter_uuid = "1d7fff50bd624a2f97ac1f8c834c6acc"
  vswitch_type    = "LinuxBridge"

  # use LACP (802.3ad) with hash policy
  bonding_name     = "Uplink1"
  bonding_mode     = "802.3ad"
  xmit_hash_policy = "layer2"

  # Option 3: Attach to clusters
  cluster_uuids = ["c54f0cef45854149970374da8b2d0750"]
  cluster_attachments = [
    {
      cluster_uuid = "c54f0cef45854149970374da8b2d0750"
      host_params = [
        {
          host_uuid          = "550a685224d349d0a9c29924aed0422d"
          physical_interface = "ens12"
        }
      ]
    }
  ]
}

output "zsphere_distributed_switch" {
  value = zsphere_distributed_switch.distributed_switch
}
