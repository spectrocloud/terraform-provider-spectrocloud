output "cluster_id" {
  value = spectrocloud_cluster_maas.cluster.id
}

# Computed, sensitive. Also written to disk as a file by kubeconfig.tf - see that file's outputs
# below for the file paths.
output "kubeconfig" {
  value     = spectrocloud_cluster_maas.cluster.kubeconfig
  sensitive = true
}

# Computed, sensitive. Cluster-admin kubeconfig.
output "admin_kube_config" {
  value     = spectrocloud_cluster_maas.cluster.admin_kube_config
  sensitive = true
}

# Computed. Deprecated in favor of cloud_config, but still populated.
output "cloud_config_id" {
  value = spectrocloud_cluster_maas.cluster.cloud_config_id
}

# Computed. Palette-derived geographic location of the cluster - not settable on this resource.
output "location_config" {
  value = spectrocloud_cluster_maas.cluster.location_config
}

# Local file paths written by kubeconfig.tf - export one of these and point kubectl at it:
#   export KUBECONFIG=$(terraform output -raw kubeconfig_path)
output "kubeconfig_path" {
  value = local_file.kubeconfig.filename
}

output "admin_kubeconfig_path" {
  value = local_file.admin_kubeconfig.filename
}
