# Looks up a Helm registry by name.
data "spectrocloud_registry_helm" "my_helm_registry" {
  # Required lookup key.
  name = "my-helm-registry"
}

# Computed.
output "helm_registry_id" {
  value = data.spectrocloud_registry_helm.my_helm_registry.id
}
