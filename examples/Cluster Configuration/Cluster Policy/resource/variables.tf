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

