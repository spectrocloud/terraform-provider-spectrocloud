output "cluster_id" {
  value = spectrocloud_cluster_azure.cluster.id
}

output "kubeconfig" {
  value     = spectrocloud_cluster_azure.cluster.kubeconfig
  sensitive = true
}

output "admin_kube_config" {
  value     = spectrocloud_cluster_azure.cluster.admin_kube_config
  sensitive = true
}

output "cloud_config_id" {
  value = spectrocloud_cluster_azure.cluster.cloud_config_id
}

output "kubeconfig_path" {
  value = local_file.kubeconfig.filename
}

output "admin_kubeconfig_path" {
  value = local_file.admin_kubeconfig.filename
}
