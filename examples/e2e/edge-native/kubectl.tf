resource "local_file" "kubeconfig" {
  content              = local.cluster_kubeconfig
  filename             = "kubeconfig_ne-2"
  file_permission      = "0644"
  directory_permission = "0755"
}
