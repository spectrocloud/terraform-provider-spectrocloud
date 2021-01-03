data "spectrocloud_cloudaccount_vsphere" "account" {
  name = var.shared_vmware_cloud_account_name
}

# If creating a new cloud account, use this:
#
# resource "spectrocloud_cloudaccount_vsphere" "account" {
#   name                 = "vsphere-picard-2"
#   private_cloud_gateway_id      = var.private_cloud_gateway_id
#   vsphere_vcenter               = "<....>"
#   vsphere_username              = "<....>"
#   vsphere_password              = "<....>"
#   vsphere_ignore_insecure_error = true
# }
