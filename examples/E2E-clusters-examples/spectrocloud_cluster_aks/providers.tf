terraform {
  required_providers {
    spectrocloud = {
      version = ">= 0.30.0"
      source  = "spectrocloud/spectrocloud"
    }
    # Used in kubeconfig.tf to write the cluster's kubeconfig files to local disk.
    local = {
      version = ">= 2.0"
      source  = "hashicorp/local"
    }
  }
}

variable "sc_host" {
  description = "Spectro Cloud Endpoint"
  default     = "api.spectrocloud.com"
}

variable "sc_api_key" {
  description = "Spectro Cloud API key"
  sensitive   = true
  default     = ""
}

variable "sc_project_name" {
  description = "Spectro Cloud Project (e.g: Default)"
  default     = "Default"
}

provider "spectrocloud" {
  host         = var.sc_host
  api_key      = var.sc_api_key
  project_name = var.sc_project_name
}
