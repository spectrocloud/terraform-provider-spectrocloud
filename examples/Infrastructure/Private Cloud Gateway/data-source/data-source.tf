# Looks up a Private Cloud Gateway (PCG) by name and resolves its ID.
#
# Lookup keys:
#   name - Optional (conflicts with `id`), also Computed.
data "spectrocloud_private_cloud_gateway" "example_pcg" {
  name = "my-private-cloud-gateway"
}

# Computed outputs:
#   pcg_id   - ID of the PCG.
#   pcg_name - Name of the PCG.
output "pcg_id" {
  value = data.spectrocloud_private_cloud_gateway.example_pcg.id
}

output "pcg_name" {
  value = data.spectrocloud_private_cloud_gateway.example_pcg.name
}
