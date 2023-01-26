resource "spectrocloud_cloudaccount_vsphere" "account" {
   name                 = "vs"
  context = "tenant"
   private_cloud_gateway_id      = var.private_cloud_gateway_id
   vsphere_vcenter               = "vcenter.spectrocloud.dev"
   vsphere_username              = "nikolay@vsphere.local"
   vsphere_password              = "VnC5z!KE"
   vsphere_ignore_insecure_error = true
}

variable "private_cloud_gateway_id" {
  type = string
}

output "same" {
  value = spectrocloud_cloudaccount_vsphere.account.id
}
