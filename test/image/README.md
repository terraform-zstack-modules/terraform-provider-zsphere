# Image Resource Test Cases

This directory contains Terraform test cases for the `zsphere_image` resource.

## Directory Structure

```
test/
├── provider.tf               # Shared provider configuration
├── variables.tf              # Shared provider variables
├── terraform.tfvars.example  # Shared provider credentials example
└── image/
    ├── provider.tf           # Provider config (references shared variables)
    ├── variables.tf          # Resource-specific variables
    ├── main.tf               # Resource definition and outputs
    ├── terraform.tfvars.example # Resource config example
    ├── basic.tftest.hcl      # Basic functionality tests
    ├── crud.tftest.hcl       # CRUD operation tests
    ├── configurations.tftest.hcl # Configuration combination tests
    └── README.md             # This file
```

## Configuration

### Provider Credentials (Shared)

Provider credentials are shared across all resource tests. Set them once:

**Option 1: Environment Variables (Recommended)**
```bash
export TF_VAR_zsphere_host="172.26.52.90"
export TF_VAR_zsphere_access_key_id="your-access-key-id"
export TF_VAR_zsphere_access_key_secret="your-access-key-secret"
```

**Option 2: terraform.tfvars in test/ directory**
```bash
cd test
cp terraform.tfvars.example terraform.tfvars
vim terraform.tfvars
```

### Resource-Specific Variables

Resource-specific variables (image_name, image_url, etc.) can be set in:
- `test/image/terraform.tfvars`
- Environment variables: `TF_VAR_image_name`, `TF_VAR_image_url`, etc.

## Prerequisites

1. Install Terraform >= 1.6.0
2. Build and install the zsphere provider:
   ```bash
   cd /root/terraform-provider-zsphere
   go build -o terraform-provider-zsphere
   mkdir -p ~/.terraform.d/plugins/local/zstack/zsphere/1.0.0/linux_amd64
   mv terraform-provider-zsphere ~/.terraform.d/plugins/local/zstack/zsphere/1.0.0/linux_amd64/
   ```

## Running Tests

### Using the Test Script

```bash
# Set provider credentials first
export TF_VAR_zsphere_host="172.26.52.90"
export TF_VAR_zsphere_access_key_id="your-key-id"
export TF_VAR_zsphere_access_key_secret="your-secret"

# Build and run all tests
./scripts/test/run_image_tests.sh all

# Run specific test suites
./scripts/test/run_image_tests.sh basic
./scripts/test/run_image_tests.sh crud
./scripts/test/run_image_tests.sh config

# Run a specific test by name
./scripts/test/run_image_tests.sh run create_basic_image
```

### Running Tests Manually

```bash
# Navigate to test directory
cd /root/terraform-provider-zsphere/test/image

# Run all tests
terraform test

# Run specific test file
terraform test -filter=basic.tftest.hcl

# Run specific test case
terraform test -run "create_basic_image"

# Verbose output
terraform test -verbose
```

## Test Cases

### basic.tftest.hcl - Basic Functionality Tests

| Test Case | Description |
|-----------|-------------|
| `create_basic_image` | Create basic image, verify name and URL |
| `verify_default_values` | Verify default values (platform, guest_os_type) |
| `verify_architecture_and_boot_mode` | Verify architecture and boot mode |
| `update_image_name` | Update image name |
| `verify_computed_attributes` | Verify computed attributes (uuid, status, system) |

### crud.tftest.hcl - CRUD Operation Tests

| Test Case | Description |
|-----------|-------------|
| `create_image` | Create image |
| `read_image` | Read image state |
| `update_image_description` | Update image description |
| `update_image_platform` | Update image platform |
| `url_change_requires_replace` | Verify URL change triggers resource recreation |

### configurations.tftest.hcl - Configuration Combination Tests

| Test Case | Description |
|-----------|-------------|
| `create_iso_image` | Create ISO type image |
| `create_aarch64_image` | Create aarch64 architecture image (UEFI boot) |
| `create_windows_image` | Create Windows platform image |
| `create_uefi_csm_image` | Create UEFI_WITH_CSM boot mode image |
| `create_data_volume_template` | Create DataVolumeTemplate type image |
| `create_image_with_expunge` | Create image with expunge option |

## Notes

1. **Provider credentials are shared** - Set them once in environment variables or test/terraform.tfvars
2. **Image URL**: Ensure the image URL is accessible, otherwise creation will fail
3. **Storage UUID**: If `image_storage_uuids` is not specified, the system will automatically select the first available backup storage
4. **Expunge**: Setting `expunge=true` will permanently delete the image, it cannot be recovered
5. **URL Change**: Changing `url` will trigger resource recreation (delete old image, create new one)
