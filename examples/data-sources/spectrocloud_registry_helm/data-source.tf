# Looks up a Helm registry by name.
#
# Lookup keys:
#   name - Required.
data "spectrocloud_registry_helm" "my_helm_registry" {
  name = "my-helm-registry"
}

# Computed outputs:
#   helm_registry_id
output "helm_registry_id" {
  value = data.spectrocloud_registry_helm.my_helm_registry.id
}
