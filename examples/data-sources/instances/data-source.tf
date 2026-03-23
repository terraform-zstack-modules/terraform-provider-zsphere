data "zsphere_instances" "test" {
}

output "zstack_secs" {
  value = data.zsphere_instances.test
}

# example output
# zstack_secs = {
#   "filter" = tolist([])
#   "name" = tostring(null)
#   "name_pattern" = tostring(null)
#   "vminstances" = tolist([
#     {
#       "all_volumes" = tolist([
#         {
#           "volume_actual_size" = 0
#           "volume_description" = "Root volume for VM[uuid:0579c098c3294b3c8e587ceabb3c5870]"
#           "volume_format" = "qcow2"
#           "volume_size" = 40
#           "volume_state" = "Enabled"
#           "volume_status" = "Ready"
#           "volume_type" = "Root"
#           "volume_uuid" = "af6b08beb73f49489565d20ad00c4585"
#         },
#       ])
#       "architecture" = "x86_64"
#       "cluster_uuid" = "5568b28dbaed465daf768b533b1fe72d"
#       "cpu_num" = 2
#       "datacenter_uuid" = "86643f5442164890b8b71de5b582dc9f"
#       "host_uuid" = ""
#       "hypervisor_type" = "KVM"
#       "image_uuid" = "8c00337d47e94acab448e88afdc5c872"
#       "memory_size" = 2048
#       "name" = "VM-Instance-1"
#       "platform" = "Linux"
#       "state" = "Stopped"
#       "type" = "UserVm"
#       "uuid" = "0579c098c3294b3c8e587ceabb3c5870"
#       "vm_nics" = tolist(null) /* of object */
#     },
#     {
#       "all_volumes" = tolist([
#         {
#           "volume_actual_size" = 0
#           "volume_description" = "Root volume for VM[uuid:0c55e5b3f8204133afeaa98519a5e147]"
#           "volume_format" = "qcow2"
#           "volume_size" = 40
#           "volume_state" = "Enabled"
#           "volume_status" = "Deleted"
#           "volume_type" = "Root"
#           "volume_uuid" = "7b600b459fec44acbec2d6e913383583"
#         },
#       ])
#       "architecture" = "x86_64"
#       "cluster_uuid" = "5568b28dbaed465daf768b533b1fe72d"
#       "cpu_num" = 2
#       "datacenter_uuid" = "86643f5442164890b8b71de5b582dc9f"
#       "host_uuid" = ""
#       "hypervisor_type" = "KVM"
#       "image_uuid" = "8c00337d47e94acab448e88afdc5c872"
#       "memory_size" = 2048
#       "name" = "test-vm-instance"
#       "platform" = "Linux"
#       "state" = "Destroyed"
#       "type" = "UserVm"
#       "uuid" = "0c55e5b3f8204133afeaa98519a5e147"
#       "vm_nics" = tolist(null) /* of object */
#     },
#     {
#       "all_volumes" = tolist([
#         {
#           "volume_actual_size" = 0
#           "volume_description" = "Root volume for VM[uuid:6ada685eeb4c48ffa617b9e79bec5935]"
#           "volume_format" = "qcow2"
#           "volume_size" = 40
#           "volume_state" = "Enabled"
#           "volume_status" = "Deleted"
#           "volume_type" = "Root"
#           "volume_uuid" = "47e9f37bcc19402e9269b3c0b14d0826"
#         },
#       ])
#       "architecture" = "x86_64"
#       "cluster_uuid" = "5568b28dbaed465daf768b533b1fe72d"
#       "cpu_num" = 2
#       "datacenter_uuid" = "86643f5442164890b8b71de5b582dc9f"
#       "host_uuid" = ""
#       "hypervisor_type" = "KVM"
#       "image_uuid" = "8c00337d47e94acab448e88afdc5c872"
#       "memory_size" = 2048
#       "name" = "test-vm-instance"
#       "platform" = "Linux"
#       "state" = "Destroyed"
#       "type" = "UserVm"
#       "uuid" = "6ada685eeb4c48ffa617b9e79bec5935"
#       "vm_nics" = tolist(null) /* of object */
#     },
#   ])
# }