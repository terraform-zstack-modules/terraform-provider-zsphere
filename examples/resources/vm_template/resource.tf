resource "zsphere_vm_template" "example" {
  name             = "my-updated-template"
  description      = "This template was managed by Terraform"
  vm_instance_uuid = var.vm_instance_uuid
}

output "vm_template" {
  value = zsphere_vm_template.example
}

# Note: Creating a VM Template is done by converting an existing VM instance.
# This resource is used to manage existing VM templates (read, update, delete).
# The creation of VM templates needs to be done through:
# 1. Converting an existing VM instance using ZStack UI
# 2. Using API: POST /vm-instances/{uuid}/create-templated-vmInstance
# 3. Or using zsphere_instance resource with convert_to_template = true
