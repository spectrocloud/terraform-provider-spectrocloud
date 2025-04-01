# Retrieve details of a DNS map for a Private Cloud Gateway
data "spectrocloud_privatecloudgateway_dns_map" "example" {
  search_domain_name       = "example.com"  # Specify the domain name for DNS search
  # Option to filter with network, if more than one dns map in same search_domain_name.
  network                  = "VM-NETWORK2"
  private_cloud_gateway_id = "pcg-12345"    # Specify the associated Private Cloud Gateway ID
}

# Output the retrieved network
output "dns_map_network" {
  value = data.spectrocloud_privatecloudgateway_dns_map.example.network.id
}

# Output the associated data center
output "dns_map_data_center" {
  value = data.spectrocloud_privatecloudgateway_dns_map.example.data_center
}