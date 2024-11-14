terraform {
  required_providers {
    spectrocloud = {
      version = ">= 0.13.2"
      source  = "spectrocloud/spectrocloud"
    }
  }
}

# variable "sc_host" {}
# variable "sc_api_key" {}
# variable "sc_project_name" {}


provider "spectrocloud" {
  host         = "dev.spectrocloud.com"
  api_key      =  "MDNlOTQwYTkxYjU2ZDk3MGZlNjU1MThmYTc5Zjk4MzU="
  project_name = "srini-cluster"
}

#
# provider "spectrocloud" {
#   host         = var.sc_host
#   api_key      = var.sc_api_key
#   project_name = var.sc_project_name
# }