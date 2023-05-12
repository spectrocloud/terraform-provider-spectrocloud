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

variable "credential_type" {}

variable "aws_access_key" {}
variable "aws_secret_key" {}

variable "arn" {}
variable "external_id" {}
