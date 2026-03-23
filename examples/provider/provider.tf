terraform {
  required_providers {
    zsphere = {
      source  = "terraform-zstack-modules/zsphere"
      version = "1.0.0"
    }
  }
}

provider "zsphere" {
  host              = "ip address of zsphere cloud api endpoint"
  access_key_id     = "access_key_id of zsphere cloud"
  access_key_secret = "access_key_secret of zsphere cloud"
}