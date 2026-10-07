terraform {
  required_providers {
    spectrocloud = {
      version = ">= 0.17.2"
      source  = "spectrocloud/spectrocloud"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "4.0.4"
    }

    local = {
      source  = "hashicorp/local"
      version = ">= 2.4.0"
    }
  }

  required_version = ">= 1.5"
}

variable "sc_host" {
  description = "The Palette API endpoint to connect to."
  default     = "api.spectrocloud.com"
}

variable "sc_api_key" {
  description = "Your Palette API key. Can also be set via the SPECTROCLOUD_APIKEY environment variable instead of this variable."
  default     = null
}

variable "sc_project_name" {
  description = "The Palette project to deploy resources into (e.g. Default)."
  default     = "Default"
}

provider "spectrocloud" {
  host         = var.sc_host
  api_key      = var.sc_api_key
  project_name = var.sc_project_name
}
