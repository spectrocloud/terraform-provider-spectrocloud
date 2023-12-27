terraform {
  required_providers {
    spectrocloud = {
      version = ">= 0.12.0"
      source  = "spectrocloud/spectrocloud"
    }
  }
}

provider "spectrocloud" {
  host         = var.sc_host
  api_key      = var.sc_api_key
  project_name = var.sc_project_name
  trace        = var.sc_trace
}