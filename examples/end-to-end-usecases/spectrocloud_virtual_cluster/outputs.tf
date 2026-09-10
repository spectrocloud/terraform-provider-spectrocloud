output "cluster_id" {
  value = spectrocloud_virtual_cluster.cluster.id
}

# Computed, sensitive. Also written to disk as a file by kubeconfig.tf - see that file's outputs
# below for the file paths.
output "kubeconfig" {
  value     = spectrocloud_virtual_cluster.cluster.kubeconfig
  sensitive = true
}

# Computed, sensitive. Cluster-admin kubeconfig.
output "admin_kube_config" {
  value     = spectrocloud_virtual_cluster.cluster.admin_kube_config
  sensitive = true
}

# Computed. Deprecated - the provider notes this field must be of type `azure`, which does not
# apply to virtual clusters, so it is expected to remain empty here.
output "cloud_config_id" {
  value = spectrocloud_virtual_cluster.cluster.cloud_config_id
}

# Local file paths written by kubeconfig.tf - export one of these and point kubectl at it:
#   export KUBECONFIG=$(terraform output -raw kubeconfig_path)
output "kubeconfig_path" {
  value = local_file.kubeconfig.filename
}

output "admin_kubeconfig_path" {
  value = local_file.admin_kubeconfig.filename
}
