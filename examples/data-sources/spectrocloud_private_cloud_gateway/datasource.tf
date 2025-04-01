# Retrieve details of a specific Private Cloud Gateway (PCG) by name
data "spectrocloud_private_cloud_gateway" "example_pcg" {
  name = "my-private-cloud-gateway"  # Specify the name of the PCG
}

# Output the retrieved PCG ID
output "pcg_id" {
  value = data.spectrocloud_private_cloud_gateway.example_pcg.id
}

# Output the retrieved PCG name
output "pcg_name" {
  value = data.spectrocloud_private_cloud_gateway.example_pcg.name
}
