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

# Resource Variables - Common

variable "image_storage_name" {
  type        = string
  default     = "test-image-storage"
  description = "Name of the image storage"
}

variable "image_storage_description" {
  type        = string
  default     = "Test image storage created by Terraform"
  description = "Description of the image storage"
}

variable "image_storage_type" {
  type        = string
  default     = "ImageStore"
  description = "Type of the image storage: ImageStore or Ceph"
}

variable "image_storage_zone_uuid" {
  type        = string
  default     = ""
  description = "UUID of the zone to attach the image storage"
}

# ImageStore Type Specific Variables

variable "image_storage_hostname" {
  type        = string
  default     = "172.26.50.183" # "192.168.1.100"
  description = "Hostname for ImageStore type"
}

variable "image_storage_url" {
  type        = string
  default     = "/zstack/imagestore"
  description = "URL for ImageStore type"
}

variable "image_storage_ssh_port" {
  type        = number
  default     = 22
  description = "SSH port for ImageStore type"
}

variable "image_storage_username" {
  type        = string
  default     = "root"
  description = "Username for ImageStore type"
}

variable "image_storage_password" {
  type        = string
  description = "Password for ImageStore type"
}

variable "image_storage_import_images" {
  type        = bool
  default     = false
  description = "Import images for ImageStore type"
}

# Ceph Type Specific Variables

variable "image_storage_pool_name" {
  type        = string
  default     = "images"
  description = "Pool name for Ceph type"
}

variable "image_storage_mon_urls" {
  type        = list(string)
  default     = ["root:password@172.20.13.215:22"]
  description = "MON URLs for Ceph type"
}
