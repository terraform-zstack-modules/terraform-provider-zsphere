# Copyright (c) ZStack.io, Inc.

# Provider Variables

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

# Resource Variables

variable "datacenter_name" {
  type        = string
  default     = "test-datacenter"
  description = "Name of the datacenter"
}

variable "datacenter_description" {
  type        = string
  default     = "Test datacenter created by Terraform"
  description = "Description of the datacenter"
}

variable "datacenter_is_default" {
  type        = bool
  default     = false
  description = "Whether this is the default datacenter"
}
