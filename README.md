# ZStack ZSphere Terraform Provider

The ZStack ZSphere provider enables Terraform to manage resources on ZStack ZSphere platform. This provider allows you to define and manage your infrastructure as code, including virtual machines, networks, storage, and more.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.10.4
- [Go](https://golang.org/doc/install) >= 1.25 (to build the provider plugin)

## Supported Versions

| Terraform Version | Minimum Provider Version | Maximum Provider Version |
|-------------------|--------------------------|--------------------------|
| >= 1.10.4         | 1.0.0                    | latest                   |

## Getting Started

### Provider Configuration

```hcl
terraform {
  required_providers {
    zsphere = {
      source  = "terraform-zstack-modules/zsphere"
      version = "~> 1.0"
    }
  }
}

provider "zsphere" {
  host              = "your-zstack-host"
  access_key_id     = "your-access-key-id"
  access_key_secret = "your-access-key-secret"
}
```

### Environment Variables

You can also configure the provider using environment variables:

```bash
export ZSPHERE_HOST="your-zstack-host"
export ZSPHERE_ACCESS_KEY_ID="your-access-key-id"
export ZSPHERE_ACCESS_KEY_SECRET="your-access-key-secret"
```

## Documentation

- [Provider Documentation](docs/index.md)
- [Resource Documentation](docs/resources/)
- [Data Source Documentation](docs/data-sources/)
- [Examples](examples/)

## Key Features

### Virtual Machine Instance Management

The `zsphere_instance` resource is the core component for managing VM instances. It supports two creation methods:

#### 1. Creating from Image

Create a VM instance from an existing image with full disk configuration:

```hcl
# Data sources to find resources
data "zsphere_images" "images" {
  name = "your-image-name"
}

data "zsphere_port_groups" "networks" {
  name = "your-port-group-name"
}

# Basic VM creation from image
resource "zsphere_instance" "vm_basic" {
  name        = "vm_basic_from_terraform"
  description = "create a basic vm from terraform"
  image_uuid  = data.zsphere_images.images.images.0.uuid
  expunge     = true
  memory_size = 1024
  cpu_num     = 2

  # Platform configuration
  platform      = "Linux"
  guest_os_type = "CentOS 7"
  architecture  = "x86_64"

  root_disk = {
    size = 5 * 1024 * 1024 * 1024  # 5 GB in bytes
  }

  data_disks = [
    {
      size = 2 * 1024 * 1024 * 1024  # 2 GB in bytes
    }
  ]

  network_interfaces = [
    {
      port_group_uuid = data.zsphere_port_groups.networks.port_groups.0.uuid
      default_l3      = true
    }
  ]
}
```

**Key Characteristics:**
- Full control over disk configuration (root disk and data disks)
- Supports disk bus types: `virtio`, `ide`, `virtio-scsi`, `scsi`
- Data disk names are auto-generated as `{vm-name}-1`, `{vm-name}-2`, etc.
- Supports all post-creation configurations (hostname, static IP, etc.)

**Advanced Example with UEFI and Custom Disks:**

```hcl
resource "zsphere_instance" "vm_advanced" {
  name        = "vm_advanced_from_terraform"
  description = "create an advanced vm from terraform"
  image_uuid  = data.zsphere_images.images.images.0.uuid
  expunge     = true
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
    size     = 4 * 1024 * 1024 * 1024  # 4 GB in bytes
    bus_type = "virtio"
  }

  # Data disks with different bus types
  # Names are auto-generated as: vm_advanced_from_terraform-1, vm_advanced_from_terraform-2
  data_disks = [
    {
      size     = 2 * 1024 * 1024 * 1024  # 2 GB in bytes
      bus_type = "virtio-scsi"
    },
    {
      size     = 1 * 1024 * 1024 * 1024  # 1 GB in bytes
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
```

#### 2. Creating from Template

Create a VM instance by cloning from an existing templated VM:

```hcl
# Data source to find the template VM
data "zsphere_instances" "template" {
  name = "your-template-name"
}

# Basic VM from template
resource "zsphere_instance" "vm_from_template" {
  name          = "vm_from_template"
  description   = "create a vm from template"
  template_uuid = data.zsphere_instances.template.instances.0.uuid
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
```

**Key Characteristics:**
- Disk configuration is inherited from the template
- Use `disk_aos` to add additional disks (only `new` type supported)
- Platform and guest OS type are inherited from template
- Some parameters cannot be overridden (see [limitations](#known-limitations))

**Advanced Example with Additional Disks:**

```hcl
resource "zsphere_instance" "advanced_from_template" {
  name          = "vm-advanced-from-template"
  description   = "Advanced VM from template"
  template_uuid = data.zsphere_instances.template.instances.0.uuid
  expunge       = true
  memory_size   = 4096
  cpu_num       = 4

  # Boot and system configuration
  boot_mode         = "UEFI"
  hostname          = "template-vm.example.com"
  architecture      = "x86_64"

  # CPU configuration
  cpu_mode            = "host-model"
  cpu_quota           = 100
  vnuma_enabled       = false
  cpu_resource_level  = "Normal"

  network_interfaces = [
    {
      port_group_uuid = data.zsphere_port_groups.networks.port_groups.0.uuid
      default_l3      = true
      static_ip       = "192.168.1.101"
    }
  ]

  # Additional disks beyond what the template provides
  disk_aos = [
    {
      size     = 50 * 1024 * 1024 * 1024  # 50 GB
      bus_type = "virtio"
    }
  ]

  strategy = "InstantStart"
}
```

### Important Notes for Instance Creation

1. **Mutual Exclusion**: You must specify either `image_uuid` OR `template_uuid`, not both.

2. **Data Disk Naming**: Data disk names are **auto-generated** as `{vm-name}-1`, `{vm-name}-2`, etc. to match ZStack UI behavior. The `name` field in `data_disks` is computed and cannot be set by the user.

3. **Parameter Validation**: The provider validates parameters based on creation method:
   - Image creation: Cannot use `vm_nic_params`, `disk_aos`, or `vm_nic_config`
   - Template creation: Cannot use `root_disk`, `data_disks`, `platform`, or `guest_os_type`

4. **Post-Creation Configuration**: Some configurations are applied after VM creation via separate API calls:
   - Hostname setting (requires VMtools)
   - Static IP configuration
   - Boot order configuration
   - NUMA settings

### Batch Deployment Example

For batch deployment from CSV, see the [batch-deployment-through-image example](examples/batch-deployment-through-image/):

```hcl
# Read VM configurations from CSV
locals {
  vm_csv_raw = csvdecode(file("${path.module}/vms.csv"))
  
  vm_configs = [
    for vm in local.vm_csv_raw : {
      name       = vm.vm_name
      ip_address = vm.ip_address
      cpu        = tonumber(vm.cpu)
      memory     = tonumber(vm.memory_gb)
      # ... other fields
    }
  ]
}

# Batch create VMs
resource "zsphere_instance" "vms" {
  for_each = { for vm in local.vm_configs : vm.name => vm }

  name        = each.value.name
  description = "Created by Terraform"
  image_uuid  = each.value.image_uuid
  cpu_num     = each.value.cpu
  memory_size = each.value.memory * 1024
  
  # ... other configuration
}
```

## Building the Provider

### Prerequisites

- Go 1.25 or higher
- Git

### Build Steps

```bash
# Clone the repository
git clone https://github.com/ZStack-Robot/terraform-provider-zstack.git
cd terraform-provider-zstack

# Build the provider
go build -o terraform-provider-zsphere

# Install the provider for local use
mkdir -p ~/.terraform.d/plugins/local/zstack/zsphere/1.0.0/linux_amd64
mv terraform-provider-zsphere ~/.terraform.d/plugins/local/zstack/zsphere/1.0.0/linux_amd64/
```

## Testing

### Test Prerequisites

Set the following environment variables before running tests:

```bash
export TF_VAR_zsphere_host="your-zstack-host"
export TF_VAR_zsphere_access_key_id="your-access-key-id"
export TF_VAR_zsphere_access_key_secret="your-access-key-secret"
export TF_VAR_image_uuid="your-image-uuid"
export TF_VAR_l3_network_uuid="your-l3-network-uuid"
```

For detailed testing documentation, see [test/instance/README.md](test/instance/README.md).

## Known Limitations

### Compared to ZStack UI (Frontend)

#### From Image Creation
- **Disk QoS**: IOPS and bandwidth limits not yet supported
- **Disk Provisioning Strategy**: Thin/thick provisioning not yet supported
- **Cache Mode**: Disk cache mode configuration not yet supported
- **AIO Mode**: Asynchronous I/O mode not yet supported

#### From Template Creation
- **Batch Creation**: Not supported (Terraform is declarative, not procedural)
- **Disk Create Types**: Only `new` type supported in `disk_aos` (frontend supports `new`/`image`/`created`/`rdm`)
- **Advanced Network Configuration**: MAC address, SR-IOV, bandwidth limits, security groups not yet supported

## Examples

See the [examples/](examples/) directory for complete working examples:

- [Basic VM Creation from Image](examples/resources/instance/resource.tf) - Tested in [test/instance](test/instance/)
- [Batch Deployment through Image](examples/batch-deployment-through-image/) - Production-ready batch deployment
- [VM Creation from Template](examples/resources/instance/resource.tf) - Tested in [test/template-instance](test/template-instance/)

## Contributing

We welcome contributions to the ZStack ZSphere Terraform Provider! Please see our [contributing guidelines](CONTRIBUTING.md) for details.

### Reporting Issues

If you encounter any issues or have feature requests, please [open an issue](https://github.com/ZStack-Robot/terraform-provider-zstack/issues/new) on GitHub.

## License

This project is licensed under the MPL-2.0 License - see the LICENSE file for details.

## Support

- **Documentation**: [Terraform Registry](https://registry.terraform.io/providers/terraform-zstack-modules/zsphere/latest/docs)
- **Issues**: [GitHub Issues](https://github.com/ZStack-Robot/terraform-provider-zstack/issues)
- **ZStack Documentation**: [www.zstack.io](https://www.zstack.io)
