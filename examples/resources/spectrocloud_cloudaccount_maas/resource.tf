resource "spectrocloud_cloudaccount_maas" "maas-1" {
  name           = "maas-1"
  maas_api_endpoint = var.maas_api_endpoint
  maas_api_key = var.maas_api_key
}
