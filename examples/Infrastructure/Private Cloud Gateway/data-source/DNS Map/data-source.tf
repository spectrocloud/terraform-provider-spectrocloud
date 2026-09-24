# Looks up a private cloud gateway DNS mapping by search domain name (and, optionally, network
# when more than one mapping shares the same search domain name).
#
# Lookup keys:
#   search_domain_name       - Required.
#   network                  - Optional, also Computed. Required only to disambiguate when more
#                               than one DNS map shares the same search_domain_name.
#   private_cloud_gateway_id - Required.
data "spectrocloud_privatecloudgateway_dns_map" "example" {
  search_domain_name       = "example.com"
  network                  = "VM-NETWORK2"
  private_cloud_gateway_id = "pcg-12345"
}

# Computed outputs:
#   dns_map_network    - Also echoes the lookup key on success. Plain string - not an object.
#   dns_map_data_center
output "dns_map_network" {
  value = data.spectrocloud_privatecloudgateway_dns_map.example.network
}

output "dns_map_data_center" {
  value = data.spectrocloud_privatecloudgateway_dns_map.example.data_center
}
