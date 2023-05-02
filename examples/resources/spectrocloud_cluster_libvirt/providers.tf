terraform {
  required_providers {
    spectrocloud = {
      version = ">= 0.14.0"
      source  = "spectrocloud/spectrocloud"
    }
  }
}

provider "spectrocloud" {
  host         = var.sc_host
  api_key       = var.sc_api_key
  project_name = var.sc_project_name
  trace = true
}
