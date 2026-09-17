# Looks up a private cloud gateway IP pool by name.
#
# Lookup keys (both required):
#   name                     - Name of the IP pool.
#   private_cloud_gateway_id - ID of the private cloud gateway that owns this IP pool.
data "spectrocloud_ippool" "example" {
  name                     = "my-ip-pool"
  private_cloud_gateway_id = "pcg-12345"
}

# Computed outputs:
#   ip_pool_id   - ID of the IP pool.
#   ip_pool_name - Name of the IP pool.
#   pcg_id       - ID of the owning private cloud gateway.
output "ip_pool_id" {
  value = data.spectrocloud_ippool.example.id
}

output "ip_pool_name" {
  value = data.spectrocloud_ippool.example.name
}

output "pcg_id" {
  value = data.spectrocloud_ippool.example.private_cloud_gateway_id
}
