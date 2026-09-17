# Looks up an OCI, Helm, or Spectro (pack) registry by name.
#
# Lookup keys:
#   name - Required.
#   type - Optional, default "". Allowed: "", "oci", "helm", "spectro". If omitted, the
#          registry type is inferred by trying spectro/pack lookup first.
data "spectrocloud_registry" "my_registry" {
  name = "my-registry"
  # type = "helm"
}

# Computed outputs:
#   registry_id
#   registry_sync_status - Helm registries only - null/empty for oci and spectro registries.
output "registry_id" {
  value = data.spectrocloud_registry.my_registry.id
}

output "registry_sync_status" {
  value = data.spectrocloud_registry.my_registry.sync_status
}
