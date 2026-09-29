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

variable "cluster_cloud_account_name" {}
variable "cluster_cluster_profile_name" {}
variable "backup_storage_location_name" {}

variable "cluster_name" {}

variable "subnet_ids_eu_west_1c" {
  type    = list(string)
  default = ["subnet-04ab962a9fa3ca4b6", "subnet-039c3beb3da69172f"]
}
