data "spectrocloud_registry_oci" "registry1" {
  name = "test-nik"

}

output "test" {
  value = data.spectrocloud_registry_oci.registry1
}