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

variable "vm_template_name" {
  type        = string
  default     = "test-vm-template"
  description = "Name of the VM template"
}

variable "vm_template_description" {
  type        = string
  default     = "Test VM template created by Terraform"
  description = "Description of the VM template"
}

variable "vm_instance_uuid" {
  type        = string
  default     = ""
  description = "UUID of an existing VM instance to convert to a template"
}

variable "vm_template_uuid" {
  type        = string
  default     = ""
  description = "UUID of an existing VM template to manage"
}
