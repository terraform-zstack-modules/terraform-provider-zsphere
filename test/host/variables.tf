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

variable "host_name" {
  type        = string
  default     = "test-host"
  description = "Name of the host"
}

variable "host_description" {
  type        = string
  default     = "Test host created by Terraform"
  description = "Description of the host"
}

variable "host_management_ip" {
  type        = string
  description = "Management IP address of the host"
}

variable "host_username" {
  type        = string
  default     = "root"
  description = "Username for SSH authentication to the host"
}

variable "host_password" {
  type        = string
  sensitive   = true
  default     = "password"
  description = "Password for SSH authentication to the host"
}

variable "host_ssh_port" {
  type        = number
  default     = 22
  description = "SSH port for connecting to the host"
}

variable "host_iommu" {
  type        = bool
  default     = false
  description = "Enable scanning host IOMMU settings"
}

variable "host_ept" {
  type        = bool
  default     = true
  description = "Enable Intel EPT hardware-assisted virtualization"
}

variable "cluster_uuid" {
  type        = string
  description = "UUID of the cluster to which the host belongs"
}
