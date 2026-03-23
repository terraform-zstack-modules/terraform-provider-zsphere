# Copyright (c) ZStack.io, Inc.

resource "zsphere_distributed_switch" "test" {
  name            = var.distributed_switch_name
  description     = var.distributed_switch_description
  datacenter_uuid = var.datacenter_uuid
  vswitch_type    = var.distributed_switch_vswitch_type
  bonding_mode = "802.3ad"
  bonding_name = "Uplink1"
  xmit_hash_policy = "layer2"
  cluster_uuids       = var.distributed_switch_cluster_uuids

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

output "distributed_switch_uuid" {
  value = zsphere_distributed_switch.test.uuid
}

output "distributed_switch_name" {
  value = zsphere_distributed_switch.test.name
}

output "distributed_switch_description" {
  value = zsphere_distributed_switch.test.description
}

output "distributed_switch_datacenter_uuid" {
  value = zsphere_distributed_switch.test.datacenter_uuid
}

output "distributed_switch_vswitch_type" {
  value = zsphere_distributed_switch.test.vswitch_type
}

output "distributed_switch_type" {
  value = zsphere_distributed_switch.test.type
}

output "distributed_switch_bonding_name" {
  value = zsphere_distributed_switch.test.bonding_name
}

output "distributed_switch_attached_clusters" {
  value = zsphere_distributed_switch.test.attached_cluster_uuids
}
