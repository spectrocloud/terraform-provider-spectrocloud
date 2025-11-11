variable "sc_host" {
  description = "Spectro Cloud API host"
  type        = string
}

variable "sc_api_key" {
  description = "Spectro Cloud API key"
  type        = string
  sensitive   = true
}

variable "sc_project_name" {
  description = "Spectro Cloud project name"
  type        = string
  default     = "Default"
}

variable "cluster_profile_infra_id" {
  description = "UID of the infrastructure cluster profile"
  type        = string
}

variable "cluster_profile_addon_id" {
  description = "UID of the addon cluster profile"
  type        = string
}

variable "maintenance_policy_id" {
  description = "UID of the maintenance policy"
  type        = string
}

