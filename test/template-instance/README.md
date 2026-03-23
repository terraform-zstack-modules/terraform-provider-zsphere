# Copyright (c) ZStack.io, Inc.
# SPDX-License-Identifier: MPL-2.0

# README for template-instance tests

# This test suite validates the creation of VM instances from a template.

## Prerequisites

1. A template VM must to exist in your ZStack environment
2. The template VM must to have the necessary configuration ( either:
   - Network configuration (port groups, L3 networks)
   - CPU and memory resources
   - Storage resources (for disk_aos)

## Test Files

- `basic.tftest.hcl` - Basic functionality tests
- `crud.tftest.hcl` - CRUD operation tests  
- `configurations.tftest.hcl` - Advanced configuration tests  
- `validation.tftest.hcl` - Input validation tests

## Running Tests

### Using Terraform CLI (1.6+)

```bash
cd test/template-instance
terraform init
terraform test -verbose
```

### Using the Test Script

```bash
cd scripts/test
./run_template_instance_tests.sh
```

### Using Environment Variables

```bash
export TF_VAR_zsphere_host="your-zsphere-host"
export TF_VAR_zsphere_access_key_id="your-access-key-id"
export TF_VAR_zsphere_access_key_secret="your-access-key-secret"
export TF_VAR_template_uuid="your-template-uuid"
export TF_VAR_port_group_name="your-port-group-name"
```

## Test Coverage

### Basic Tests (`basic.tftest.hcl`)
- ✅ Create instance from template
- ✅ Verify instance UUID is computed
- ✅ Verify CPU and memory configuration
- ✅ Verify outputs

### CRUD Tests (`crud.tftest.hcl`)
- ✅ Create instance from template
- ✅ Read instance
- ✅ Update instance description
- ✅ Update CPU and memory
- ✅ Verify VM NICs are computed

### Configuration Tests (`configurations.tftest.hcl`)
- ✅ Create instance with UEFI boot mode
- ✅ Create instance with hostname
- ✅ Create instance with advanced CPU configuration
- ✅ Create instance with GPU configuration

### Validation Tests (`validation.tftest.hcl`)
- ✅ Validate template_uuid is required
- ✅ Validate port_group_name is required
- ✅ Validate CPU and memory constraints
- ✅ Validate boot_mode values
- ✅ Validate strategy values

## Notes

- Template creation does NOT support `root_disk` and `data_disks` parameters
- Template creation does not support `platform`, `guest_os_type`, `root_disk`, `boot_orders` parameters
- Template creation supports `vm_nic_config` for advanced network configuration
- Template creation supports `disk_aos` for disk configuration
- Template creation supports `cdrom_list`, `vgpu_device`, `gpu_device_uuid_list`, and advanced configurations
