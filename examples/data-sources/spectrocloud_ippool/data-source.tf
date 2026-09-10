# Looks up a private cloud gateway IP pool by name.
data "spectrocloud_ippool" "example" {
  # Required lookup key.
  name = "my-ip-pool"
  # Required lookup key. ID of the private cloud gateway that owns this IP pool.
  private_cloud_gateway_id = "pcg-12345"
}

# Computed.
output "ip_pool_id" {
  value = data.spectrocloud_ippool.example.id
}

output "ip_pool_name" {
  value = data.spectrocloud_ippool.example.name
}

output "pcg_id" {
  value = data.spectrocloud_ippool.example.private_cloud_gateway_id
}
