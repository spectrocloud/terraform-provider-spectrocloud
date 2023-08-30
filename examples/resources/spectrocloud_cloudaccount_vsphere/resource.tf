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
  type = string
}

variable "vsphere_vcenter" {
  type = string
}

variable "vsphere_username" {
  type = string
}

variable "vsphere_password" {
  type = string
}

output "same" {
  value = spectrocloud_cloudaccount_vsphere.account.id
}
