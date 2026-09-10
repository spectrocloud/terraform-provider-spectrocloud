# Looks up an OCI, Helm, or Spectro (pack) registry by name.
data "spectrocloud_registry" "my_registry" {
  # Required lookup key.
  name = "my-registry"
  # Optional lookup key, default "". Allowed: "", "oci", "helm", "spectro". If omitted, the
  # registry type is inferred by trying spectro/pack lookup first.
  # type = "helm"
}

# Computed.
output "registry_id" {
  value = data.spectrocloud_registry.my_registry.id
}

# Computed. Helm registries only - null/empty for oci and spectro registries.
output "registry_sync_status" {
  value = data.spectrocloud_registry.my_registry.sync_status
}
