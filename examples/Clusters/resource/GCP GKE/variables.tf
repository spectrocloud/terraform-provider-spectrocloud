variable "sc_host" {
  description = "Spectro Cloud Endpoint"
  default     = "api.spectrocloud.com"
}

variable "sc_api_key" {
  description = "Spectro Cloud API key"
}

variable "sc_project_name" {
  description = "Spectro Cloud Project (e.g: Default)"
  default     = "Default"
}

variable "gcp_cloud_account_name" {}
variable "gke_cluster_profile_name" {}
variable "gcp_project" {}
variable "gcp_region" {}
variable "cluster_name" {}
