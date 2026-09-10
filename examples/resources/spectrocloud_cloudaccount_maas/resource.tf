# Nothing on this resource is ForceNew - every attribute below updates in place.

data "spectrocloud_private_cloud_gateway" "maas_pcg" {
  name = "System Private Gateway"
}

resource "spectrocloud_cloudaccount_maas" "maas-1" {
  # Required.
  name = "maas-1"
  # Optional, default "project". Allowed: "project", "tenant".
  context = "project"
  # Required. The Private Cloud Gateway used to reach this MAAS environment.
  private_cloud_gateway_id = data.spectrocloud_private_cloud_gateway.maas_pcg.id
  # Required. For example: http://maas:5240/MAAS.
  maas_api_endpoint = var.maas_api_endpoint
  # Required, sensitive.
  maas_api_key = var.maas_api_key
}
