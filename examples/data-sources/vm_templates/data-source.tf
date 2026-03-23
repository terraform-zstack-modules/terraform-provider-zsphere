data "zsphere_vm_templates" "test" {
}

output "zsphere_vm_templates" {
  value = data.zsphere_vm_templates.test
}

# Example output:
# zsphere_vm_templates = {
#   "filter" = tolist([])
#   "vm_templates" = tolist([
#     {
#       "account_uuid" = "36c27e8ff05c4780bf6d2fa65700f22e"
#       "name" = "my-vm-template"
#       "uuid" = "8c00337d47e94acab448e88afdc5c872"
#       "zone_uuid" = "b9608acd97a94af68611bcb058bc4f4d"
#     },
#   ])
#   "name" = tostring(null)
#   "name_pattern" = tostring(null)
# }

# Search by name
data "zsphere_vm_templates" "by_name" {
  name = "my-vm-template"
}

# Search by name pattern
data "zsphere_vm_templates" "by_pattern" {
  name_pattern = "my-template-*"
}

# Filter results
data "zsphere_vm_templates" "filtered" {
  filter {
    name   = "name"
    values = ["template-1", "template-2"]
  }
}
