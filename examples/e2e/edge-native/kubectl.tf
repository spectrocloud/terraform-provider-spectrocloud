resource "local_file" "kubeconfig" {
  content = spectrocloud_cluster_edge_native.cluster.kubeconfig
  filename             = "kubeconfig_ne-2"
  file_permission      = "0644"
  directory_permission = "0755"
#  sensitive_content = true
}
