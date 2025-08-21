data "zsphere_images" "images" {
  name = "jiajian-test-from-terraform"
}

data "zsphere_port_groups" "networks" {
  name = "Pub-network-勿删"
}


resource "zsphere_instance" "vm" {
  name        = "vm_instance_from_terraform"
  description = "create a vm from terraform"
  image_uuid  = data.zsphere_images.images.images.0.uuid #"${data.zstack_images.images.images[0].uuid}" #"9b26312501614ec0b6dc731e6977dfb2"
  expunge     = true
  memory_size = 4096
  cpu_num     = 4

  data_disks = [
    {
      size = 10
    }
  ]
  network_interfaces = [
    {
      port_group_uuid = data.zsphere_port_groups.networks.port_groups.0.uuid
      default_l3      = true
      # static_ip       = "172.30.3.154"
    }
  ]

}

output "zsphere_instance" {
  value = zsphere_instance.vm
}