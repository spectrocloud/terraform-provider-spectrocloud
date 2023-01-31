resource "spectrocloud_cloudaccount_vsphere" "account" {
  name                          = var.new_vmware_cloud_account_name
  context                       = "project"
  private_cloud_gateway_id      = data.spectrocloud_private_cloud_gateway.gateway.id
  vsphere_vcenter               = var.vsphere_vcenter
  vsphere_username              = var.vsphere_username
  vsphere_password              = var.vsphere_password
  vsphere_ignore_insecure_error = true
}