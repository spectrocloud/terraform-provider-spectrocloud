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

variable "vsphere_pcg_name" {
  type        = string
  description = "Name of the Private Cloud Gateway used to reach this vSphere environment"
}

variable "vsphere_vcenter" {
  type        = string
  description = "vCenter server address"
}

variable "vsphere_username" {
  type        = string
  description = "vSphere username"
}

variable "vsphere_password" {
  type        = string
  description = "vSphere password"
  sensitive   = true
}
