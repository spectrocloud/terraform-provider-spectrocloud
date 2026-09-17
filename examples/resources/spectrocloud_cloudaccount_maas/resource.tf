# Nothing on this resource is ForceNew - every attribute below updates in place.
#
# Attributes:
#   name                     - Required.
#   context                  - Optional, default "project". Allowed: "project", "tenant".
#   private_cloud_gateway_id - Required. The Private Cloud Gateway used to reach this MAAS
#                              environment.
#   maas_api_endpoint        - Required. For example: http://maas:5240/MAAS.
#   maas_api_key             - Required, sensitive.

data "spectrocloud_private_cloud_gateway" "maas_pcg" {
  name = "System Private Gateway"
}

resource "spectrocloud_cloudaccount_maas" "maas-1" {
  name                     = "maas-1"
  context                  = "project"
  private_cloud_gateway_id = data.spectrocloud_private_cloud_gateway.maas_pcg.id
  maas_api_endpoint        = var.maas_api_endpoint
  maas_api_key             = var.maas_api_key
}
