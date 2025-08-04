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

# Azure Public Cloud variables
variable "azure_tenant_id" {
  description = "Azure Tenant ID"
  type        = string
}

variable "azure_client_id" {
  description = "Azure Client ID"
  type        = string
}

variable "azure_client_secret" {
  description = "Azure Client Secret"
  type        = string
  sensitive   = true
}

# Azure US Government Cloud variables (optional)
variable "azure_gov_tenant_id" {
  description = "Azure US Government Tenant ID"
  type        = string
  default     = ""
}

variable "azure_gov_client_id" {
  description = "Azure US Government Client ID"
  type        = string
  default     = ""
}

variable "azure_gov_client_secret" {
  description = "Azure US Government Client Secret"
  type        = string
  default     = ""
  sensitive   = true
}

# Azure US Secret Cloud variables (optional)
variable "azure_secret_tenant_id" {
  description = "Azure US Secret Cloud Tenant ID"
  type        = string
  default     = ""
}

variable "azure_secret_client_id" {
  description = "Azure US Secret Cloud Client ID"
  type        = string
  default     = ""
}

variable "azure_secret_client_secret" {
  description = "Azure US Secret Cloud Client Secret"
  type        = string
  default     = ""
  sensitive   = true
}

variable "azure_secret_tls_cert" {
  description = "TLS certificate for Azure US Secret Cloud authentication"
  type        = string
  default     = ""
  sensitive   = true
}
