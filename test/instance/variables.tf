# Copyright (c) ZStack.io, Inc.

# Provider Variables (must be declared in each test directory)

variable "zsphere_host" {
  type        = string
  default     = "your-zsphere-host"
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

variable "instance_name" {
  type        = string
  default     = "test-vm-instance"
  description = "Name of the VM instance"
}

variable "instance_description" {
  type        = string
  default     = "Test VM instance created by Terraform"
  description = "Description of the VM instance"
}

variable "image_uuid" {
  type        = string
  default     = ""
  description = "UUID of the image to use for the VM"
}

variable "l3_network_uuid" {
  type        = string
  default     = ""
  description = "UUID of the L3 network to attach to the VM"
}

variable "host_uuid" {
  type        = string
  default     = ""
  description = "UUID of the host to deploy the VM"
}

variable "cpu_num" {
  type        = number
  default     = 2
  description = "Number of CPUs for the VM"
}

variable "memory_size" {
  type        = number
  default     = 2048
  description = "Memory size in MB for the VM"
}

variable "platform" {
  type        = string
  default     = "Linux"
  description = "Platform of the VM (Linux, Windows, Other, Paravirtualization, WindowsVirtio)"
}

variable "strategy" {
  type        = string
  default     = "CreateStopped"
  description = "Deployment strategy (InstantStart, JustCreate, CreateStopped)"
}
