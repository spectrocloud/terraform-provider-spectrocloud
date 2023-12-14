


provider "spectrocloud" {
  project_name = var.SpectroCloudProject # Project name (e.g: Default)
  host         = var.SpectroCloudURI
  api_key      = var.SpectroCloudUsername
}


//  required_providers {
//    spectrocloud = {
//      version = ">= 0.1"
//      source  = "spectrocloud/spectrocloud"
//    }
//  }
//
