# Looks up a Spectro pack registry by name.
data "spectrocloud_registry_pack" "my_pack" {
  # Required lookup key.
  name = "my-pack"
}

# Computed.
output "registry_pack_id" {
  value = data.spectrocloud_registry_pack.my_pack.id
}
