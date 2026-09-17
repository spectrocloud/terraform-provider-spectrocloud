# Looks up an OCI registry by name.
#
# Lookup keys:
#   name - Required.
data "spectrocloud_registry_oci" "my_oci_registry" {
  name = "my-oci-registry"
}

# Computed outputs:
#   oci_registry_id
output "oci_registry_id" {
  value = data.spectrocloud_registry_oci.my_oci_registry.id
}
