terraform {
  required_providers {
    spectrocloud = {
      source  = "spectrocloud/spectrocloud"
      version = ">= 0.1"
    }
  }
}

provider "spectrocloud" {
  host         = var.sc_host
  api_key      = var.sc_api_key
  project_name = var.sc_project_name
}

variable "sc_project_name" {
  description = "Spectro Cloud project name"
  default     = "Default"
}
