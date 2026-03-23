# ============================================
# Complete Variable Definitions - Supports all fields in test.csv
# ============================================

# ZSphere API Configuration
variable "zsphere_host" {
  description = "ZSphere Cloud API endpoint address"
  type        = string
}

variable "access_key_id" {
  description = "ZSphere Cloud Access Key ID"
  type        = string
  sensitive   = true
}

variable "access_key_secret" {
  description = "ZSphere Cloud Access Key Secret"
  type        = string
  sensitive   = true
}

# ============================================
# CSV Field Related Variables
# ============================================

variable "csv_file_path" {
  description = "Path to CSV configuration file"
  type        = string
  default     = "test.csv"
}

# The following variables are used for name-based UUID lookup
variable "use_name_lookup" {
  description = "Whether to use name lookup for UUIDs (if CSV contains names instead of UUIDs)"
  type        = bool
  default     = false
}

variable "port_group_name" {
  description = "Port group name (used when use_name_lookup is true)"
  type        = string
  default     = ""
}

variable "image_name" {
  description = "Image name (used when use_name_lookup is true)"
  type        = string
  default     = ""
}

variable "datacenter_name" {
  description = "Datacenter name (used when use_name_lookup is true)"
  type        = string
  default     = ""
}

variable "cluster_name" {
  description = "Cluster name (used when use_name_lookup is true)"
  type        = string
  default     = ""
}

variable "host_name" {
  description = "Host name (used when use_name_lookup is true)"
  type        = string
  default     = ""
}

variable "root_storage_name" {
  description = "Root storage name (used when use_name_lookup is true)"
  type        = string
  default     = ""
}

variable "data_storage_name" {
  description = "Data storage name (used when use_name_lookup is true)"
  type        = string
  default     = ""
}

# ============================================
# Default Configuration Values
# ============================================

variable "default_root_disk_size_bytes" {
  description = "Default root disk size in bytes (5 GB)"
  type        = number
  default     = 5 * 1024 * 1024 * 1024
}

variable "default_bus_type" {
  description = "Default disk bus type"
  type        = string
  default     = "virtio"

  validation {
    condition     = contains(["virtio", "ide", "virtio-scsi", "scsi"], var.default_bus_type)
    error_message = "Bus type must be one of virtio, ide, virtio-scsi, or scsi."
  }
}

variable "default_strategy" {
  description = "Default startup strategy"
  type        = string
  default     = "InstantStart"

  validation {
    condition     = contains(["InstantStart", "JustCreate", "CreateStopped"], var.default_strategy)
    error_message = "Strategy must be one of InstantStart, JustCreate, or CreateStopped."
  }
}

variable "default_never_stop" {
  description = "Default setting for never stop"
  type        = bool
  default     = false
}

variable "default_cpu_mode" {
  description = "Default CPU mode"
  type        = string
  default     = "host-model"
}

variable "default_architecture" {
  description = "Default architecture"
  type        = string
  default     = "x86_64"
}
