terraform {
  required_providers {
    spectrocloud = {
      version = ">= 0.8.9"
      source  = "spectrocloud/spectrocloud"
    }
  }

  backend "http" {
      address = "https://gitlab.com/api/v4/projects/36258605/terraform/state/old-state-name"
      username = "nikolay@spectrocloud.com"
      password = "glpat-haTJ97tu6DyQUCikby3p"
  }
}

variable "sc_host" {
  description = "Spectro Cloud Endpoint"
  default     = "api.spectrocloud.com"
}

variable "sc_username" {
  description = "Spectro Cloud Username"
}

variable "sc_api_key" {
  description = "Spectro Cloud API key"
  //sensitive   = true
}

variable "sc_project_name" {
  description = "Spectro Cloud Project (e.g: Default)"
  default     = "Default"
}

provider "spectrocloud" {
  host         = var.sc_host
  username     = var.sc_username
  #password     = var.sc_password
  api_key = var.sc_api_key
  project_name = var.sc_project_name
  #trace = true
}

