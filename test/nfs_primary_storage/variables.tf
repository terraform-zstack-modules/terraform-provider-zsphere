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

variable "nfs_storage_name" {
  type        = string
  default     = "test-nfs-storage"
  description = "Name of the NFS primary storage"
}

variable "nfs_storage_description" {
  type        = string
  default     = "Test NFS primary storage created by Terraform"
  description = "Description of the NFS primary storage"
}

variable "nfs_storage_url" {
  type        = string
  default     = "192.168.1.100:/share/nfs"
  description = "Mount path of the NFS primary storage"
}

variable "nfs_storage_cidr" {
  type        = string
  default     = ""
  description = "CIDR for the NFS primary storage gateway"
}

variable "nfs_storage_mount_options" {
  type        = string
  default     = ""
  description = "Mount options for the NFS primary storage"
}

variable "datacenter_uuid" {
  type        = string
  description = "UUID of the datacenter"
}

variable "cluster_uuid" {
  type        = string
  description = "UUID of the cluster to attach the NFS primary storage"
}
