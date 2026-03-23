data "zsphere_images" "test" {
}

output "zstack_secs" {
  value = data.zsphere_images.test
}

# example output:
# zstack_secs = {
#   "filter" = tolist([])
#   "images" = tolist([
#     {
#       "architecture" = "x86_64"
#       "format" = "qcow2"
#       "name" = "Image-1"
#       "platform" = ""
#       "state" = "Enabled"
#       "status" = "Ready"
#       "uuid" = "8c00337d47e94acab448e88afdc5c872"
#     },
#   ])
#   "name" = tostring(null)
#   "name_pattern" = tostring(null)
# }