data "spectrocloud_registry_pack" "registry1" {
  name = "Public Repo"

}

output "test" {
  value = data.spectrocloud_registry_pack.registry1
}