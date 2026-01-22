variable "sc_host" {
  description = "Spectro Cloud Endpoint"
  type        = string
  default     = "api.spectrocloud.com"
}

variable "sc_api_key" {
  description = "Spectro Cloud API key"
  type        = string
  sensitive   = true
}

variable "sc_project_name" {
  description = "Spectro Cloud Project (e.g: Default)"
  type        = string
  default     = "Default"
}


