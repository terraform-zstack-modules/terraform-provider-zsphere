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

variable "local_storage_name" {
  type        = string
  default     = "test-local-storage"
  description = "Name of the local primary storage"
}

variable "local_storage_description" {
  type        = string
  default     = "Test local primary storage created by Terraform"
  description = "Description of the local primary storage"
}

variable "local_storage_url" {
  type        = string
  default     = "/vms_ds"
  description = "Mount path of the local primary storage"
}

variable "datacenter_uuid" {
  type        = string
  description = "UUID of the datacenter"
}

variable "cluster_uuid" {
  type        = string
  description = "UUID of the cluster to attach the local primary storage"
}
