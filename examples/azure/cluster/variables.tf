variable "sc_host" {
  description = "Spectro Cloud Endpoint"
  default     = "console.spectrocloud.com"
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

variable "cluster_cloud_account_name" {
  description = "Cloud account name used for the cluster"
}

variable "cluster_cluster_profile_name" {
  description = "Cluster Profile name used for the the cluster"
}

variable "cluster_name" {
  description = "Name of the cluster"
  default     = "cluster1-azure"
}

variable "cluster_ssh_public_key" {
  description = "The public SSH key to inject into the nodes"
}

variable "azure_subscription_id" {
  description = "Azure subscription id (e.g: 871012....)"
}

variable "azure_resource_group" {
  description = "Azure resource group (e.g: rg1)"
}

variable "azure_location" {
  description = "Azure location (e.g: westus)"
}
