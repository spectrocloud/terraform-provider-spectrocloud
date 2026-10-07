terraform {
  required_providers {
    spectrocloud = {
      version = ">= 0.16.1"
      source  = "spectrocloud/spectrocloud"
    }
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
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

# The aws provider is used only to look up available Availability Zones (see data.tf).
# It relies on the standard AWS credential chain (environment variables, ~/.aws/credentials, etc.)
# and needs no explicit configuration here.
