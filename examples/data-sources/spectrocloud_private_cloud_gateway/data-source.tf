# Looks up a Private Cloud Gateway (PCG) by name and resolves its ID.
data "spectrocloud_private_cloud_gateway" "example_pcg" {
  # Lookup key, optional (conflicts with `id`), also Computed.
  name = "my-private-cloud-gateway"
}

# Computed.
output "pcg_id" {
  value = data.spectrocloud_private_cloud_gateway.example_pcg.id
}

output "pcg_name" {
  value = data.spectrocloud_private_cloud_gateway.example_pcg.name
}
