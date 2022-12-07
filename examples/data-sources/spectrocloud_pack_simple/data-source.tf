data "spectrocloud_registry" "registry" {
  name = "Public Repo"
}

data "spectrocloud_pack_simple" "pack1" {
  type         = "operator-instance"
  name         = "mongodb-community-operator"
  version      = "0.7.6"
  registry_uid = data.spectrocloud_registry.registry.id

}

output "same" {
  value = data.spectrocloud_pack_simple.pack1
}
