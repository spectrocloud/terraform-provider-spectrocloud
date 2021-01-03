#
# If looking up a cloudaccount instead of creating one
# data "spectrocloud_cloudaccount_gcp" "account" {
#   # id = <uid>
#   name = var.cluster_cloud_account_name
# }

resource "spectrocloud_cloudaccount_gcp" "account" {
  name                 = "gcp-picard-2"
  gcp_json_credentials = var.gcp_serviceaccount_json
}
