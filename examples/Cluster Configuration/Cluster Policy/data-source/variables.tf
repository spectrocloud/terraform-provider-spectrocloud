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

variable "policy_name" {
  description = "Name of the cluster config policy"
  type        = string
  default     = "tenant-policy"
}

variable "policy_context" {
  description = "Context of the cluster config policy (project or tenant)"
  type        = string
  default     = "tenant"
}

