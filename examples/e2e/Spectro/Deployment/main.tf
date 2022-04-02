


provider "spectrocloud" {
  project_name = var.SpectroCloudProject # Project name (e.g: Default)
  host         = var.SpectroCloudURI
  username     = var.SpectroCloudUsername
  password     = var.SpectroCloudPassword
}


//  required_providers {
//    spectrocloud = {
//      version = ">= 0.1"
//      source  = "spectrocloud/spectrocloud"
//    }
//  }
//
