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

variable "maas_api_endpoint" {}
variable "maas_api_key" {}

variable "maas_pcg_name" {
  type        = string
  description = "Name of the Private Cloud Gateway used to reach this MAAS environment"
  default     = "System Private Gateway"
}
