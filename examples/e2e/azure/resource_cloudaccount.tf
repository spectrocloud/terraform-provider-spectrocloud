#
# If looking up a cloudaccount instead of creating one
# data "spectrocloud_cloudaccount_azure" "account" {
#   # id = <uid>
#   name = var.cluster_cloud_account_name
# }

resource "spectrocloud_cloudaccount_azure" "account" {
  name                = "az-picard-2"
  azure_tenant_id     = var.azure_tenant_id
  azure_client_id     = var.azure_client_id
  azure_client_secret = var.azure_client_secret
}
