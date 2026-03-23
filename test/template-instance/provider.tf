# Copyright (c) ZStack.io, Inc.
# SPDX-License-Identifier: MPL-2.0

terraform {
  required_providers {
    zsphere = {
      source  = "local/zstack/zsphere"
      version = "1.0.0"
    }
  }
}

provider "zsphere" {
  host              = var.zsphere_host
  access_key_id     = var.zsphere_access_key_id
  access_key_secret = var.zsphere_access_key_secret
}
