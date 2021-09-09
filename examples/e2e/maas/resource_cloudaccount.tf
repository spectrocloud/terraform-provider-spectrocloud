#
# If looking up a cloudaccount instead of creating one
# data "spectrocloud_cloudaccount_maas" "account" {
#   # id = <uid>
#   name = var.cluster_cloud_account_name
# }

resource "spectrocloud_cloudaccount_maas" "account" {
  name = "maas-picard-3"
  maas_api_endpoint = var.maas_api_endpoint
  maas_api_key = var.maas_api_key

}
