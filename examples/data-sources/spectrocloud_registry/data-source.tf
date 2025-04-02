# Data source to retrieve details of a specific SpectroCloud registry by name
data "spectrocloud_registry" "my_registry" {
  name = "my-registry" # Name of the registry to look up
}

# Output the ID of the retrieved registry
output "registry_id" {
  value = data.spectrocloud_registry.my_registry.id
}
