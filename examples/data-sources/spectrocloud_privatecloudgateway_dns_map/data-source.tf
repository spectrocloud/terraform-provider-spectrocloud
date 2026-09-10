# Looks up a private cloud gateway DNS mapping by search domain name (and, optionally, network
# when more than one mapping shares the same search domain name).
data "spectrocloud_privatecloudgateway_dns_map" "example" {
  # Required lookup key.
  search_domain_name = "example.com"
  # Optional lookup key, also Computed. Required only to disambiguate when more than one DNS
  # map shares the same search_domain_name.
  network = "VM-NETWORK2"
  # Required lookup key.
  private_cloud_gateway_id = "pcg-12345"
}

# Computed (also echoes the lookup key on success). Plain string - not an object.
output "dns_map_network" {
  value = data.spectrocloud_privatecloudgateway_dns_map.example.network
}

# Computed.
output "dns_map_data_center" {
  value = data.spectrocloud_privatecloudgateway_dns_map.example.data_center
}
