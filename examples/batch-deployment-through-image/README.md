# Batch Deployment Terraform - Full Version

This version fully supports all fields in `test.csv` and can directly replace the original Shell script batch deployment functionality.

## Supported CSV Fields

| CSV Field | Terraform Field | Description | Required |
|-----------|-----------------|-------------|----------|
| vm_name | `name` | VM name | ✅ |
| ip_address | `network_interfaces.static_ip` | Static IP address | ✅ |
| netmask | `netmask` | Netmask | ✅ |
| gateway | `gateway` | Gateway | ✅ |
| port_uuid | `network_interfaces.port_group_uuid` | Port group UUID | ✅ |
| image_uuid | `image_uuid` | Image UUID | ✅ |
| datacenter_uuid | `datacenter_uuid` | Datacenter UUID | ✅ |
| OS | `platform` + `guest_os_type` | Operating system type | ✅ |
| BIOS | `boot_mode` | Boot mode: Legacy/UEFI | ✅ |
| hostname | `hostname` | Hostname | ✅ |
| cpu | `cpu_num` | Number of CPUs | ✅ |
| memorysize(G) | `memory_size` | Memory size (GB) | ✅ |
| datasize(G) | `data_disks.size` | Data disk size (GB) | ✅ |
| datanum | `data_disks` list length | Number of data disks | ✅ |
| cluster_uuid | `cluster_uuid` | Cluster UUID | ❌ |
| host_uuid | `host_uuid` | Host UUID | ❌ |
| ps_uuid | `root_disk.primary_storage_uuid` | Root disk storage UUID | ❌ |
| datavolume_ps_uuid | `data_disks.primary_storage_uuid` | Data disk storage UUID | ❌ |

## File Structure

```
batch-deployment/
├── main-full.tf                    # Main configuration file (full version)
├── variables-full.tf               # Variable definitions (full version)
├── terraform-full.tfvars.example   # Configuration example
├── test.csv                        # CSV configuration file
├── cloud-init.tpl                  # cloud-init template
└── README-full.md                  # This file
```

## Usage

### 1. Configure CSV File

Edit the `test.csv` file and fill in the VM configuration information:

```csv
vm_name,ip_address,netmask,gateway,port_uuid,image_uuid,datacenter_uuid,OS,BIOS,hostname,cpu,memorysize(G),datasize(G),datanum,cluster_uuid,host_uuid,ps_uuid,datavolume_ps_uuid
zvm1,200.0.4.110,255.255.255.0,200.0.4.1,850ffa88e71c4ecc895a1f1fb71d4922,d4b9691eab08464e978df5e614545a0f,903f7684ba564b54b1b17f5e60f1a3ff,CentOS 7,Legacy,centos71,1,1,1,2,17394ee96f8a4153bc570e51685bbafd,,,
```

**Note**:
- When UUID fields in CSV are empty, the system will assign them randomly
- `cluster_uuid`, `host_uuid`, `ps_uuid`, `datavolume_ps_uuid` are optional fields

### 2. Configure Terraform Variables

```bash
cp terraform-full.tfvars.example terraform.tfvars
vim terraform.tfvars
```

Fill in the required configurations:

```hcl
zsphere_host      = "https://your-zsphere-api:8080"
access_key_id     = "your-access-key"
access_key_secret = "your-secret"
```

### 3. Initialize and Deploy

```bash
# Initialize Terraform
terraform init

# View execution plan
terraform plan

# Execute deployment
terraform apply
```

### 4. Batch Destroy

```bash
terraform destroy
```

## CSV Configuration Details

### Example 1: Basic Configuration (All Required Fields)

```csv
vm_name,ip_address,netmask,gateway,port_uuid,image_uuid,datacenter_uuid,OS,BIOS,hostname,cpu,memorysize(G),datasize(G),datanum,cluster_uuid,host_uuid,ps_uuid,datavolume_ps_uuid
vm1,192.168.1.10,255.255.255.0,192.168.1.1,port-uuid-xxx,image-uuid-xxx,dc-uuid-xxx,CentOS 7,Legacy,vm1-host,2,4,100,1,,,,
```

### Example 2: Specify Cluster and Storage

```csv
vm2,192.168.1.11,255.255.255.0,192.168.1.1,port-uuid-xxx,image-uuid-xxx,dc-uuid-xxx,CentOS 7,Legacy,vm2-host,4,8,200,2,cluster-uuid-xxx,,ps-uuid-xxx,
```

### Example 3: Specify Host

```csv
vm3,192.168.1.12,255.255.255.0,192.168.1.1,port-uuid-xxx,image-uuid-xxx,dc-uuid-xxx,WindowsServer 2019,Legacy,vm3-host,2,4,100,1,,host-uuid-xxx,,
```

### Example 4: UEFI Boot Mode

```csv
vm4,192.168.1.13,255.255.255.0,192.168.1.1,port-uuid-xxx,image-uuid-xxx,dc-uuid-xxx,Ubuntu 20,UEFI,vm4-host,2,4,100,1,,,,
```

### Example 5: Multiple Data Disks

```csv
vm5,192.168.1.14,255.255.255.0,192.168.1.1,port-uuid-xxx,image-uuid-xxx,dc-uuid-xxx,CentOS 7,Legacy,vm5-host,4,16,500,3,,,ps-uuid-xxx,data-ps-uuid-xxx
```

## Operating System Support

### Windows Series
- Windows, Windows 10, Windows 11, Windows 7, Windows 8
- Windows NT 4.0, Windows xp
- WindowsServer 2003, 2008, 2012, 2016, 2019, 2022

### Linux Series
- CentOS 5/6/7/8/9
- Debian 7/8/9/10/11/12
- Ubuntu 14/16/18/20/22/24
- RHEL 5/6/7/8/9
- Rocky Linux 8/9
- AlmaLinux 9
- openEuler 20.03/22.03/24.03
- And other supported Linux distributions

## Advanced Features

### 1. OS Auto-Detection

Terraform automatically determines the platform type based on the `OS` field in CSV:
- Contains "win" keyword (case-insensitive) → Windows
- Others → Linux

### 2. Special OS Handling

The following OS will automatically set CPU mode to `host-model`:
- openEuler 24.03
- Rocky Linux 9

### 3. UEFI Auto-Configuration

When `BIOS` field is set to `UEFI`:
- Automatically sets `vm_machine_type` to `q35`
- No manual configuration required

### 4. Data Disk Storage Inheritance

If `datavolume_ps_uuid` is empty but `ps_uuid` has a value:
- Data disks will automatically use the root disk's storage

## Output Information

After execution, Terraform will output the following information:

```
Outputs:

created_vms = {
  "zvm1" = {
    "boot_mode" = "Legacy"
    "cpu" = 1
    "data_disks" = 2
    "guest_os" = "CentOS 7"
    "hostname" = "centos71.6.1"
    "ip" = "200.0.4.110"
    "memory_gb" = 1
    "name" = "zvm1"
    "platform" = "Linux"
    "uuid" = "xxx"
  }
  ...
}

linux_vms = {
  "zvm1" = "CentOS 7"
  "zvm2" = "CentOS 7"
}

uefi_vms = {}

vm_count = 3

windows_vms = {
  "zvm3" = "WindowsServer 2019"
}
```

## Comparison with Shell Script

| Feature | Shell Script | Terraform |
|---------|--------------|-----------|
| CSV Configuration | ✅ | ✅ Fully compatible |
| BIOS Settings | ✅ | ✅ Supports Legacy/UEFI |
| Hostname Setting | ✅ | ✅ Supported |
| IP/Netmask/Gateway | ✅ | ✅ Full support |
| Data Disk Configuration | ✅ | ✅ Supports multiple disks |
| Storage Specification | ✅ | ✅ Independent root/data disk config |
| State Management | ❌ | ✅ Automatic management |
| Parallel Creation | ❌ Sequential | ✅ Parallel |
| Change Tracking | ❌ | ✅ Complete history |
| Rollback Capability | ❌ | ✅ One-click destroy |

## Notes

1. **Image Requirements**: Image must have VMtools installed, otherwise hostname setting may not take effect
2. **Network Requirements**: Static IP must be within the port group's network segment
3. **UUID Format**: UUIDs in CSV must be valid ZStack UUID format
4. **Empty Value Handling**: When optional fields in CSV are empty, the system will assign resources randomly
5. **Concurrency Limit**: When creating large numbers of VMs concurrently, you may need to adjust Terraform parallelism

## Troubleshooting

### Issue 1: Hostname Setting Failed

**Reason**: VMtools not running
**Solution**: Ensure the image has VMtools installed, or manually set the hostname

### Issue 2: Static IP Not Taking Effect

**Reason**: Port group configured with DHCP or IP conflict
**Solution**: Check port group configuration, ensure IP is in the correct subnet and not conflicting

### Issue 3: Data Disk Not Created

**Reason**: `datasize` or `datanum` set to 0
**Solution**: Check the values of these two fields in CSV
