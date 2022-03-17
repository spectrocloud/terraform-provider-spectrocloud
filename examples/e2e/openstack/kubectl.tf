resource "local_file" "kubeconfig" {
  content              = local.cluster_kubeconfig
  filename             = format("%s.kubeconfig", var.cluster_name)
  file_permission      = "0644"
  directory_permission = "0755"
}
