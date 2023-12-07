resource "local_file" "kubeconfig" {
  content              = local.cluster_kubeconfig
  filename             = "kubeconfig_maas-1"
  file_permission      = "0644"
  directory_permission = "0755"
}
resource "local_file" "adminkubeconfig" {
  content              = local.cluster_admin_kubeconfig
  filename             = "admin-kubeconfig_maas-1"
  file_permission      = "0644"
  directory_permission = "0755"
}