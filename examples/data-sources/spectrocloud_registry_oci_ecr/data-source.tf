# Data source to retrieve details of a specific SpectroCloud OCI registry by name
data "spectrocloud_registry_oci" "my_oci_registry" {
  name = "my-oci-registry" # Name of the OCI registry to look up
}

# Output the ID of the retrieved OCI registry
output "oci_registry_id" {
  value = data.spectrocloud_registry_oci.my_oci_registry.id
}