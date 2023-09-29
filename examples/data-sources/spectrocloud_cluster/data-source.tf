data "spectrocloud_cluster" "cluster" {
  name    = "client-101"
  context = "tenant"
}

resource "local_file" "kubeconfig" {
  content              = data.spectrocloud_cluster.cluster.kube_config
  filename             = "client-101.kubeconfig"
  file_permission      = "0644"
  directory_permission = "0755"
}

resource "local_file" "adminkubeconfig" {
  content              = data.spectrocloud_cluster.cluster.admin_kube_config
  filename             = "admin-client-101.kubeconfig"
  file_permission      = "0644"
  directory_permission = "0755"
}
