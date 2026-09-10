# Looks up an OCI registry by name.
data "spectrocloud_registry_oci" "my_oci_registry" {
  # Required lookup key.
  name = "my-oci-registry"
}

# Computed.
output "oci_registry_id" {
  value = data.spectrocloud_registry_oci.my_oci_registry.id
}
