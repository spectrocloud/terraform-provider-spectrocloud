output "cluster_id" {
  value = spectrocloud_cluster_aws.cluster.id
}

# Computed, sensitive. Also written to disk as a file by kubeconfig.tf.
output "kubeconfig" {
  value     = spectrocloud_cluster_aws.cluster.kubeconfig
  sensitive = true
}

output "admin_kube_config" {
  value     = spectrocloud_cluster_aws.cluster.admin_kube_config
  sensitive = true
}

# Computed. Deprecated in favor of cloud_config, but still populated.
output "cloud_config_id" {
  value = spectrocloud_cluster_aws.cluster.cloud_config_id
}

output "kubeconfig_path" {
  value = local_file.kubeconfig.filename
}

output "admin_kubeconfig_path" {
  value = local_file.admin_kubeconfig.filename
}
