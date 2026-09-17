# Looks up a Spectro pack registry by name.
#
# Lookup keys:
#   name - Required.
data "spectrocloud_registry_pack" "my_pack" {
  name = "my-pack"
}

# Computed outputs:
#   registry_pack_id
output "registry_pack_id" {
  value = data.spectrocloud_registry_pack.my_pack.id
}
