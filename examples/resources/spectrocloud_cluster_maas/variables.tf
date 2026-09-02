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

variable "cluster_ssh_public_keys" {
  description = "A list of SSH public keys to inject into MAAS nodes as authorized keys for the 'spectro' user."
  type        = list(string)
  default     = []
}
