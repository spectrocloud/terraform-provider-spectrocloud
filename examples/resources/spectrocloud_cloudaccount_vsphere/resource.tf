resource "spectrocloud_cloudaccount_vsphere" "account" {
  name                          = "vs"
  context                       = "tenant"
  private_cloud_gateway_id      = var.private_cloud_gateway_id
  vsphere_vcenter               = var.vsphere_vcenter
  vsphere_username              = var.vsphere_username
  vsphere_password              = var.vsphere_password
  vsphere_ignore_insecure_error = true
}

variable "private_cloud_gateway_id" {
  type = string
}

output "same" {
  value = spectrocloud_cloudaccount_vsphere.account.id
}
