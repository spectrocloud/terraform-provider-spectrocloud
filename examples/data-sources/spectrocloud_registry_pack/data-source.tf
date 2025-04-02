
# Data source to retrieve details of a specific SpectroCloud registry pack by name
data "spectrocloud_registry_pack" "my_pack" {
  name = "my-pack" # Name of the registry pack to look up
}

# Output the ID of the retrieved registry pack
output "registry_pack_id" {
  value = data.spectrocloud_registry_pack.my_pack.id
}