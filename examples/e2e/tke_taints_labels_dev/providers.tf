terraform {
  required_providers {
    spectrocloud = {
      version = ">= 0.1"
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
}

variable "sc_project_name" {
  description = "Spectro Cloud Project (e.g: Default)"
  default     = "Default"
}

variable "sc_api_key" {
  description = "Spectro Cloud API KEY"
}

variable "sc_trace" {
  default = false
}

provider "spectrocloud" {
  host         = var.sc_host
  username     = var.sc_username
  password     = var.sc_password
  project_name = var.sc_project_name
  api_key      = var.sc_api_key
  trace        = var.sc_trace
}
