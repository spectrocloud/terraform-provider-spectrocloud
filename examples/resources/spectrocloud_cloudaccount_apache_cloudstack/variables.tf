variable "sc_host" {
  description = "Spectro Cloud Endpoint"
  default     = "api.spectrocloud.com"
}

variable "sc_api_key" {
  description = "Spectro Cloud API key"
  type        = string
  sensitive   = true
}

variable "sc_project_name" {
  description = "Spectro Cloud Project (e.g: Default)"
  default     = "Default"
}

# Apache CloudStack Configuration
variable "cloudstack_api_url" {
  description = "CloudStack API endpoint URL (e.g., https://cloudstack.example.com:8080/client/api)"
  type        = string
}

variable "cloudstack_api_key" {
  description = "CloudStack API Key for authentication"
  type        = string
  sensitive   = true
}

variable "cloudstack_secret_key" {
  description = "CloudStack Secret Key for authentication"
  type        = string
  sensitive   = true
}

variable "cloudstack_domain" {
  description = "CloudStack domain for the cloud account (defaults to ROOT if not specified)"
  type        = string
  default     = "ROOT"
}

variable "private_cloud_gateway_id" {
  description = "Private Cloud Gateway ID for CloudStack connectivity"
  type        = string
}

