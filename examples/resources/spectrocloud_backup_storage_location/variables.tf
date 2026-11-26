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

#variable "credential_type" {
#default = "secret"
#}
#
#variable "aws_access_key" {}
#variable "aws_secret_key" {}
#
#variable "arn" {}
#variable "external_id" {}

variable "aws_external_id" {
  description = "The tenant's STS external ID, shown in Palette when adding an AWS account with STS credentials"
  type        = string
  sensitive   = true
}
