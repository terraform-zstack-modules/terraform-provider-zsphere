# Copyright (c) ZStack.io, Inc.

# Shared Provider Variables
# These variables are used by all resource tests.
# Set them via environment variables or terraform.tfvars.

variable "zsphere_host" {
  type        = string
  description = "ZStack management node host address"
}

variable "zsphere_access_key_id" {
  type        = string
  sensitive   = true
  description = "ZStack access key ID"
}

variable "zsphere_access_key_secret" {
  type        = string
  sensitive   = true
  description = "ZStack access key secret"
}
