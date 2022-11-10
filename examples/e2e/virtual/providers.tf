terraform {
  required_providers {
    spectrocloud = {
      version = ">= 0.10.0"
      source  = "spectrocloud/spectrocloud"
    }
  }
}

variable "sc_host" {
  description = "Spectro Cloud Endpoint"
  default     = "api.spectrocloud.com"
}

variable "sc_username" {
  description = "Spectro Cloud Username"
}

variable "sc_password" {
  description = "Spectro Cloud Password"
  sensitive   = true
  default     = ""
}

variable "sc_api_key" {
  description = "Spectro API key"
  sensitive   = true
  default     = ""
}

variable "sc_project_name" {
  description = "Spectro Cloud Project (e.g: Default)"
  default     = "Default"
}

provider "spectrocloud" {
  host         = var.sc_host
  username     = var.sc_username
  api_key      = var.sc_api_key
  project_name = var.sc_project_name
  ignore_insecure_tls_error = true
  trace = true
}
