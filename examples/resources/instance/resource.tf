data "zsphere_images" "images" {
  name = "image_from_terraform"
}

data "zsphere_port_groups" "networks" {
  name = "port_group_from_terraform"
}

# Example 1: Basic VM creation from image
resource "zsphere_instance" "vm_basic" {
  name        = "vm_basic_from_terraform"
  description = "create a basic vm from terraform"
  image_uuid  = data.zsphere_images.images.images.0.uuid
  memory_size = 1024
  cpu_num     = 2

  # Platform configuration
  platform      = "Linux"
  guest_os_type = "CentOS 7"
  architecture  = "x86_64"

  root_disk = {
    size = 5 * 1024 * 1024 * 1024 # 5 GB in bytes
  }

  data_disks = [
    {
      size = 2 * 1024 * 1024 * 1024 # 2 GB in bytes
    }
  ]

  network_interfaces = [
    {
      port_group_uuid = data.zsphere_port_groups.networks.port_groups.0.uuid
      default_l3      = true
    }
  ]
}

# Example 2: Advanced VM with UEFI boot and custom disk bus types
resource "zsphere_instance" "vm_advanced" {
  name        = "vm_advanced_from_terraform"
  description = "create an advanced vm from terraform"
  image_uuid  = data.zsphere_images.images.images.0.uuid
  memory_size = 1024
  cpu_num     = 2

  # Platform configuration
  platform      = "Linux"
  guest_os_type = "CentOS 7"
  architecture  = "x86_64"

  # Boot configuration
  boot_mode       = "UEFI"
  vm_machine_type = "q35"
  cpu_mode        = "host-model"

  # Hostname configuration
  hostname = "my-server.example.com"

  # Root disk with virtio bus
  root_disk = {
    size     = 4 * 1024 * 1024 * 1024 # 4 GB in bytes
    name     = "root-disk"
    bus_type = "virtio"
  }

  # Data disks with different bus types
  data_disks = [
    {
      size     = 2 * 1024 * 1024 * 1024 # 2 GB in bytes
      bus_type = "virtio-scsi"
    },
    {
      size     = 1 * 1024 * 1024 * 1024 # 1 GB in bytes
      bus_type = "scsi"
    }
  ]

  network_interfaces = [
    {
      port_group_uuid = data.zsphere_port_groups.networks.port_groups.0.uuid
      default_l3      = true
      static_ip       = "192.168.1.100"
    }
  ]

  netmask = "255.255.255.0"
  gateway = "192.168.1.1"

  strategy   = "InstantStart"
  never_stop = false
  user_data  = "#!/bin/bash\necho 'Hello World'"
}

# Example 3: VM from template

resource "zsphere_instance" "vm_from_template" {
  name          = "vm_from_template"
  description   = "create a vm from template"
  template_uuid = "template uuid"
  expunge       = true
  memory_size   = 1024
  cpu_num       = 2

  network_interfaces = [
    {
      port_group_uuid = data.zsphere_port_groups.networks.port_groups.0.uuid
      default_l3      = true
    }
  ]
}

output "zsphere_instance_basic" {
  value     = zsphere_instance.vm_basic
  sensitive = true
}

output "zsphere_instance_advanced" {
  value     = zsphere_instance.vm_advanced
  sensitive = true
}

output "zsphere_instance_from_template" {
  value     = zsphere_instance.vm_from_template
  sensitive = true
}
