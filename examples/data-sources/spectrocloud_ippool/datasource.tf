# Retrieve details of a specific IP Pool by name
data "spectrocloud_ippool" "example" {
  name                     = "my-ip-pool" # Specify the name of the IP pool
  private_cloud_gateway_id = "pcg-12345"  # Specify the ID of the associated Private Cloud Gateway
}

# Output the retrieved IP Pool name
output "ip_pool_name" {
  value = data.spectrocloud_ippool.example.name
}

# Output the associated Private Cloud Gateway ID
output "pcg_id" {
  value = data.spectrocloud_ippool.example.private_cloud_gateway_id
}
