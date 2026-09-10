terraform {
  required_providers {
    spectrocloud = {
      version = ">= 0.13.2"
      source  = "spectrocloud/spectrocloud"
    }
    local = {
      source  = "hashicorp/local"
      version = ">= 2.0"
    }
  }
}

variable "sc_host" {}
variable "sc_api_key" {}
variable "sc_project_name" {}

provider "spectrocloud" {
  host         = var.sc_host
  api_key      = var.sc_api_key
  project_name = var.sc_project_name
}