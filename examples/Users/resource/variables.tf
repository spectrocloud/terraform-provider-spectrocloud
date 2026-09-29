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

variable "ssh_key_value" {
  description = "ssh key value"
  default     = "ssh-rsa ...... == test@test.com"
}

variable "tenant_role_var" {
  type    = list(string)
  default = ["Tenant Admin", "Tenant User Admin"]
}

variable "app_role_var" {
  type    = list(string)
  default = ["App Deployment Admin", "App Deployment Editor"]
}

variable "workspace_role_var" {
  type    = list(string)
  default = ["Workspace Admin", "Workspace Operator"]
}

variable "resource_role_var" {
  type    = list(string)
  default = ["Resource Cluster Admin", "Resource Cluster Profile Admin"]
}

variable "resource_role_editor_var" {
  type    = list(string)
  default = ["Resource Cluster Editor", "Resource Cluster Profile Editor"]
}