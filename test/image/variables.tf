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

variable "image_name" {
  type        = string
  default     = "test-image"
  description = "Name of the image"
}

variable "image_url" {
  type        = string
  description = "URL of the image file (default for qcow2 format)"
}

variable "image_url_qcow2" {
  type        = string
  description = "URL of the qcow2 format image file"
}

variable "image_url_iso" {
  type        = string
  description = "URL of the iso format image file"
}

variable "image_url_raw" {
  type        = string
  description = "URL of the raw format image file"
}

variable "image_url_vmdk" {
  type        = string
  description = "URL of the vmdk format image file"
}

variable "image_format" {
  type        = string
  default     = "qcow2"
  description = "Format of the image (qcow2, iso, raw, vmdk)"
}

variable "image_media_type" {
  type        = string
  default     = "RootVolumeTemplate"
  description = "Media type of the image (ISO, RootVolumeTemplate, DataVolumeTemplate)"
}

variable "image_architecture" {
  type        = string
  default     = "x86_64"
  description = "Architecture of the image (x86_64, aarch64, mips64el, loongarch64)"
}

variable "image_boot_mode" {
  type        = string
  default     = "Legacy"
  description = "Boot mode of the image (Legacy, UEFI, UEFI_WITH_CSM)"
}

variable "image_platform" {
  type        = string
  default     = "Linux"
  description = "Platform of the image (Linux, Windows, Other, Paravirtualization, WindowsVirtio)"
}

variable "image_guest_os_type" {
  type        = string
  default     = "Linux"
  description = "Guest OS type"
}

variable "image_description" {
  type        = string
  default     = "Test image created by Terraform"
  description = "Description of the image"
}

variable "image_virtio" {
  type        = bool
  default     = false
  description = "Whether VirtIO is enabled"
}

variable "image_expunge" {
  type        = bool
  default     = false
  description = "Whether to expunge the image on deletion"
}

variable "image_storage_uuids" {
  type        = list(string)
  default     = []
  description = "List of backup storage UUIDs"
}
