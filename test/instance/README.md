# Instance Resource Tests

This directory contains tests for the `zsphere_instance` resource.

## Test Files

- `basic.tftest.hcl` - Basic functionality tests (create, read)
- `crud.tftest.hcl` - CRUD operation tests (create, update, read)
- `validation.tftest.hcl` - Validation tests for field constraints

## Prerequisites

1. Build and install the provider:
   ```bash
   cd /root/terraform-provider-zsphere
   go build -o terraform-provider-zsphere
   mkdir -p ~/.terraform.d/plugins/local/zstack/zsphere/1.0.0/linux_amd64
   mv terraform-provider-zsphere ~/.terraform.d/plugins/local/zstack/zsphere/1.0.0/linux_amd64/
   ```

2. Set environment variables:
   ```bash
   export TF_VAR_zsphere_host="your-zstack-host"
   export TF_VAR_zsphere_access_key_id="your-access-key-id"
   export TF_VAR_zsphere_access_key_secret="your-access-key-secret"
   export TF_VAR_image_uuid="your-image-uuid"
   export TF_VAR_l3_network_uuid="your-l3-network-uuid"
   ```

## Running Tests

### Run all tests:
```bash
cd /root/terraform-provider-zsphere/test/instance
terraform init
terraform test -verbose
```

### Run specific test file:
```bash
terraform test -filter=basic.tftest.hcl -verbose
terraform test -filter=crud.tftest.hcl -verbose
terraform test -filter=validation.tftest.hcl -verbose
```

### Run specific test:
```bash
terraform test -run="create_instance" -verbose
```

## Test Coverage

### Basic Tests
- Create instance with required fields
- Verify UUID is computed
- Verify state is computed
- Verify CPU and memory settings
- **Verify data disk naming** (auto-generated as `{vm-name}-1`, `{vm-name}-2`, etc.)

### CRUD Tests
- Create instance
- Update instance name
- Update instance description
- Update CPU and memory
- Update platform
- **Verify disk ordering consistency** (disks matched by name, not index)

### Validation Tests
- Valid platform values (Linux, Windows, Other, Paravirtualization, WindowsVirtio)
- Valid strategy values (InstantStart, JustCreate, CreateStopped)
- Valid CPU numbers (1-1024)
- Valid memory size (>= 1)
- Configuration combination tests
- **Creation method validation** (image_uuid vs template_uuid mutual exclusion)

## Important Test Notes

### Data Disk Naming
Tests should verify that data disk names are auto-generated in the format `{vm-name}-{index}`:
- For a VM named "test-vm" with 2 data disks, the disks should be named:
  - `test-vm-1`
  - `test-vm-2`

### Creation Method Validation
Tests should verify that:
- Either `image_uuid` OR `template_uuid` must be specified (not both, not neither)
- Image-specific parameters are rejected when using template creation
- Template-specific parameters are rejected when using image creation

## Notes

- The `image_uuid` and `l3_network_uuid` are required and must be valid ZStack resources
- CRUD tests update the same instance to verify the Update method works correctly
- Validation tests only test valid values (Terraform 1.10 does not support `expect_fail` for negative testing)
- To test invalid values manually, run `terraform plan` with invalid values and verify that validators reject them
