# If looking up a cloudaccount instead of creating one
data "spectrocloud_cloudaccount_maas" "account" {
#   # id = <uid>
   name = "gateway-3"
}

/*
resource "spectrocloud_cloudaccount_maas" "account" {
  name = "maas-tf-account"
  private_cloud_gateway_id      = var.private_cloud_gateway_id
  maas_api_endpoint = var.maas_api_endpoint
  maas_api_key = var.maas_api_key
}*/
