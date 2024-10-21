data "spectrocloud_private_cloud_gateway" "gateway" {
  name = "sc-stagepcg"
}

data "spectrocloud_privatecloudgateway_dns_map" "gateway_dns_map" {
  search_domain_name = "spectrocloud2.dev"

  private_cloud_gateway_id = data.spectrocloud_private_cloud_gateway.gateway.id
}

output "dns_map" {
  value = data.spectrocloud_privatecloudgateway_dns_map.gateway_dns_map.id
}