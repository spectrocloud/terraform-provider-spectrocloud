variable "sc_host" {}
variable "sc_username" {}
variable "sc_password" {}
variable "sc_project_name" {}

variable "cluster_cloud_account_name" {}
variable "cluster_cluster_profile_name" {}
variable "backup_storage_location_name" {}

variable "cluster_name" {}

variable "subnet_ids_eu_west_1c" {
  type = list(string)
  default = ["subnet-04ab962a9fa3ca4b6","subnet-039c3beb3da69172f"]
}
