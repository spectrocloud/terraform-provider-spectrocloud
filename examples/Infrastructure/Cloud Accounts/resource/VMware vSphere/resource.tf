# Nothing on this resource is ForceNew - every attribute below updates in place.
#
# Attributes:
#   name                          - Required.
#   context                       - Optional, default "project". Allowed: "project", "tenant".
#   private_cloud_gateway_id      - Required. Connects this account to the underlying vSphere
#                                   environment through a PCG.
#   vsphere_vcenter               - Required. The vCenter server address.
#   vsphere_username              - Required.
#   vsphere_password              - Required, sensitive.
#   vsphere_ignore_insecure_error - Optional, default false. Skips TLS verification against
#                                   vCenter - only for development/testing with self-signed
#                                   certificates.

data "spectrocloud_private_cloud_gateway" "gateway" {
  name = var.vsphere_pcg_name
}

resource "spectrocloud_cloudaccount_vsphere" "account" {
  name                          = "vs"
  context                       = "tenant"
  private_cloud_gateway_id      = data.spectrocloud_private_cloud_gateway.gateway.id
  vsphere_vcenter               = var.vsphere_vcenter
  vsphere_username              = var.vsphere_username
  vsphere_password              = var.vsphere_password
  vsphere_ignore_insecure_error = true
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
