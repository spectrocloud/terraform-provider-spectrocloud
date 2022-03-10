data "spectrocloud_registry_helm" "registry1" {
  name = "spectro-helm-repo"

}

output "test" {
  value = data.spectrocloud_registry_helm.registry1
}