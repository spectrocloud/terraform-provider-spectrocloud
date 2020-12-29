variable "sc_host" {
  description = "Spectro Cloud Endpoint"
  default     = "console.spectrocloud.com"
}

variable "sc_username" {
  description = "Spectro Cloud Username"
}

variable "sc_password" {
  description = "Spectro Cloud Password"
  sensitive = true
}

variable "sc_project_name" {
  description = "Spectro Cloud Project (e.g: Default)"
  default = "Default"
}

