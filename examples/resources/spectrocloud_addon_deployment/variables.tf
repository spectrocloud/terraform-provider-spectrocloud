variable "sc_host" {
  description = "Spectro Cloud API endpoint"
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

variable "cluster_uid" {
  description = "The unique identifier of the cluster to attach the addon deployment to"
  type        = string
}

