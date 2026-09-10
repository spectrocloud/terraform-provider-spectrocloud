data "spectrocloud_private_cloud_gateway" "gateway" {
  name = var.vsphere_pcg_name
}

# Nothing on this resource is ForceNew - every attribute below updates in place.
resource "spectrocloud_cloudaccount_vsphere" "account" {
  # Required.
  name = "vs"
  # Optional, default "project". Allowed: "project", "tenant".
  context = "tenant"
  # Required. Connects this account to the underlying vSphere environment through a PCG.
  private_cloud_gateway_id = data.spectrocloud_private_cloud_gateway.gateway.id
  # Required. The vCenter server address.
  vsphere_vcenter = var.vsphere_vcenter
  # Required.
  vsphere_username = var.vsphere_username
  # Required, sensitive.
  vsphere_password = var.vsphere_password
  # Optional, default false. Skips TLS verification against vCenter - only for
  # development/testing with self-signed certificates.
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
