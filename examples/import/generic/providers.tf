terraform {
  required_version = ">= 0.14.0"

  required_providers {
    spectrocloud = {
      version = "~> 0.7.0"
      source  = "spectrocloud/spectrocloud"
    }
  }
}

variable "sc_host" {}
variable "sc_api_key" {
  sensitive   = true
}
variable "sc_project_name" {}

provider "spectrocloud" {
  host         = var.sc_host
  api_key      = var.sc_api_key
  project_name = var.sc_project_name
  trace = false
}
