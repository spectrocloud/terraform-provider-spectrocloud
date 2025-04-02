# Data source to retrieve details of a specific SpectroCloud Helm registry by name
data "spectrocloud_registry_helm" "my_helm_registry" {
  name = "my-helm-registry" # Name of the Helm registry to look up
}

# Output the ID of the retrieved Helm registry
output "helm_registry_id" {
  value = data.spectrocloud_registry_helm.my_helm_registry.id
}