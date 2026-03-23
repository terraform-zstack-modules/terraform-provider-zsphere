# ============================================
# Provider Batch Deployment - Full Version
# Supports all fields in test.csv
# ============================================
# Supported CSV fields:
# vm_name, ip_address, netmask, gateway, port_uuid, image_uuid,
# datacenter_uuid, OS, BIOS, hostname, cpu, memorysize(G),
# datasize(G), datanum, cluster_uuid, host_uuid, ps_uuid, datavolume_ps_uuid
# ============================================

terraform {
  required_providers {
    zsphere = {
      source  = "terraform-zstack-modules/zsphere"
      version = "1.0.0"
    }
  }
}

provider "zsphere" {
  host              = var.zsphere_host
  access_key_id     = var.access_key_id
  access_key_secret = var.access_key_secret
}

# ============================================
# Read VM configurations from CSV
# ============================================

locals {
  # Read CSV file
  vm_csv_raw = csvdecode(file("${path.module}/${var.csv_file_path}"))

  # Transform data format to support all fields in test.csv
  vm_configs = [
    for vm in local.vm_csv_raw : {
      # Basic configuration
      name       = vm.vm_name
      ip_address = vm.ip_address
      netmask    = vm.netmask
      gateway    = vm.gateway
      hostname   = vm.hostname

      # Resource UUIDs (using UUIDs directly from CSV)
      port_uuid          = vm.port_uuid
      image_uuid         = vm.image_uuid
      datacenter_uuid    = vm.datacenter_uuid
      cluster_uuid       = vm.cluster_uuid != "" ? vm.cluster_uuid : null
      host_uuid          = vm.host_uuid != "" ? vm.host_uuid : null
      ps_uuid            = vm.ps_uuid != "" ? vm.ps_uuid : null
      datavolume_ps_uuid = vm.datavolume_ps_uuid != "" ? vm.datavolume_ps_uuid : null

      # OS and platform configuration
      os   = vm.OS
      bios = vm.BIOS

      # Compute resources
      cpu    = tonumber(vm.cpu)
      memory = tonumber(vm["memorysize(G)"]) # GB to MB conversion

      # Data disk configuration - convert GB to bytes
      data_disk_size = tonumber(vm["datasize(G)"]) * 1024 * 1024 * 1024 # Convert GB to bytes
      data_disk_num  = tonumber(vm.datanum)
    }
  ]

  # OS type mapping (supports all OS types from shell script)
  win_list = [
    "Windows", "Windows 10", "Windows 11", "Windows 7", "Windows 8",
    "Windows NT 4.0", "Windows xp", "WindowsServer 2003", "WindowsServer 2008",
    "WindowsServer 2012", "WindowsServer 2016", "WindowsServer 2019", "WindowsServer 2022"
  ]

  linux_list = [
    "Linux", "AlmaLinux 9", "Anolis OS 7", "Anolis OS 8", "CentOS 5", "CentOS 6",
    "CentOS 7", "CentOS 8", "CentOS 9", "Debian 10", "Debian 11", "Debian 12",
    "Debian 7", "Debian 8", "Debian 9", "Fedora", "Kylin 4", "Kylin V10",
    "Kylin V7", "NeoKylin V7", "NeoKylin V7 Update6", "openEuler 20.03",
    "openEuler 22.03", "openEuler 24.03", "openSUSE Leap 11", "openSUSE Leap 12",
    "openSUSE Leap 15", "SUSE Linux Enterprise Desktop 11", "SUSE Linux Enterprise Desktop 12",
    "SUSE Linux Enterprise Server 11", "SUSE Linux Enterprise Server 12",
    "SUSE Linux Enterprise Server 15", "Oracle Linux 7", "RHEL 5", "RHEL 6",
    "RHEL 7", "RHEL 8", "RHEL 9", "Rocky Linux 8", "Rocky Linux 9",
    "Ubuntu 14", "Ubuntu 16", "Ubuntu 18", "Ubuntu 20", "Ubuntu 22", "Ubuntu 24", "UOS 20"
  ]

  # OS list that requires host-model CPU mode
  change_cpu_mode_list = ["openEuler 24.03", "Rocky Linux 9"]

  # Windows versions that support hot plug
  hot_plug_list = ["WindowsServer 2008", "WindowsServer 2012", "WindowsServer 2016", "WindowsServer 2019", "WindowsServer 2022"]
}

# ============================================
# Batch create VMs
# ============================================

resource "zsphere_instance" "vms" {
  for_each = { for vm in local.vm_configs : vm.name => vm }

  name        = each.value.name
  description = "Created by Terraform - Provider Batch Deployment"
  image_uuid  = each.value.image_uuid
  expunge     = true

  # CPU and memory configuration
  cpu_num     = each.value.cpu
  memory_size = each.value.memory * 1024 # GB to MB conversion

  # Platform configuration (auto-detect based on OS)
  platform      = contains(local.win_list, each.value.os) ? "Windows" : "Linux"
  guest_os_type = each.value.os
  architecture  = var.default_architecture

  # Datacenter configuration
  datacenter_uuid = each.value.datacenter_uuid
  cluster_uuid    = each.value.cluster_uuid
  host_uuid       = each.value.host_uuid

  # Root disk configuration
  # Supported fields: size(bytes), primary_storage_uuid, bus_type
  root_disk = {
    size                 = var.default_root_disk_size_bytes
    primary_storage_uuid = each.value.ps_uuid
    bus_type             = "virtio" # var.default_bus_type
  }

  # Data disk configuration (created based on datanum and datasize)
  # Supported fields: size(bytes), primary_storage_uuid, bus_type
  data_disks = each.value.data_disk_num > 0 && each.value.data_disk_size > 0 ? [
    for i in range(each.value.data_disk_num) : {
      size                 = each.value.data_disk_size # Already in bytes
      primary_storage_uuid = each.value.datavolume_ps_uuid != null ? each.value.datavolume_ps_uuid : each.value.ps_uuid
      bus_type             = var.default_bus_type
    }
  ] : []

  # Network interface configuration
  network_interfaces = [
    {
      port_group_uuid = each.value.port_uuid
      default_l3      = true
      static_ip       = each.value.ip_address
    }
  ]

  # Startup strategy
  strategy   = var.default_strategy
  never_stop = var.default_never_stop

  # ============================================
  # Provider enhanced feature fields
  # ============================================

  # BIOS boot mode
  boot_mode = each.value.bios

  # VM machine type: auto-set to q35 in UEFI mode
  vm_machine_type = each.value.bios == "UEFI" ? "q35" : null

  # CPU mode: specific OS requires host-model
  cpu_mode = contains(local.change_cpu_mode_list, each.value.os) ? "host-model" : var.default_cpu_mode

  # Hostname configuration
  hostname = each.value.hostname

  # Network configuration (netmask and gateway)
  netmask = each.value.netmask
  gateway = each.value.gateway
}

# ============================================
# Outputs
# ============================================

output "created_vms" {
  description = "Information of created VMs"
  value = {
    for name, vm in zsphere_instance.vms : name => {
      uuid       = vm.uuid
      name       = vm.name
      ip         = vm.vm_nics[0].ip
      cpu        = vm.cpu_num
      memory_gb  = vm.memory_size / 1024
      boot_mode  = vm.boot_mode
      hostname   = vm.hostname
      platform   = vm.platform
      guest_os   = vm.guest_os_type
      data_disks = length(vm.data_disks)
    }
  }
}

output "vm_count" {
  description = "Total number of created VMs"
  value       = length(zsphere_instance.vms)
}

output "windows_vms" {
  description = "List of Windows VMs"
  value = {
    for name, vm in zsphere_instance.vms : name => vm.guest_os_type
    if vm.platform == "Windows"
  }
}

output "linux_vms" {
  description = "List of Linux VMs"
  value = {
    for name, vm in zsphere_instance.vms : name => vm.guest_os_type
    if vm.platform == "Linux"
  }
}

output "uefi_vms" {
  description = "List of VMs using UEFI boot"
  value = {
    for name, vm in zsphere_instance.vms : name => vm.boot_mode
    if vm.boot_mode == "UEFI"
  }
}
