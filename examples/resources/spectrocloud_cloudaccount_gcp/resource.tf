resource "spectrocloud_cloudaccount_gcp" "gcp-1" {
  name                 = "gcp-1"
  gcp_json_credentials = var.gcp_serviceaccount_json
}
