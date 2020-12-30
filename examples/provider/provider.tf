terraform {
  required_providers {
    spectrocloud = {
      version = ">= 0.1"
      source  = "spectrocloud/spectrocloud"
    }
  }
}

provider "spectrocloud" {
  host         = var.sc_host         # Spectro Cloud endpoint (defaults to api.spectrocloud.com)
  username     = var.sc_username     # Username of the user (or specify with SPECTROCLOUD_USERNAME env var)
  password     = var.sc_password     # Password (or specify with SPECTROCLOUD_PASSWORD env var)
  project_name = var.sc_project_name # Project name (e.g: Default)
}
