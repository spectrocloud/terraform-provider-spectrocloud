terraform {
  required_providers {
    spectrocloud = {
      version = ">= 0.1"
      source  = "spectrocloud/spectrocloud"
    }
  }
}

provider "spectrocloud" {
  host         = var.sc_host
  username     = var.sc_username
  api_key     = var.sc_api_key
  project_name = var.sc_project_name
}
