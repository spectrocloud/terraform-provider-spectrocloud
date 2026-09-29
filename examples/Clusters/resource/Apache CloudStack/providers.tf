terraform {
  required_providers {
    spectrocloud = {
      version = ">= 0.1"
      source  = "spectrocloud/spectrocloud"
    }
  }
}

provider "spectrocloud" {
  project_name = var.sc_project_name
  host         = var.sc_host
  api_key      = var.sc_api_key
}

