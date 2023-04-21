data "spectrocloud_cluster_kube_config" "cluster" {
  name = "client-101"
  context = "tenant"
}

resource "local_file" "kubeconfig" {
  content              = data.spectrocloud_cluster_kube_config.cluster.kube_config
  filename             = "client-101.kubeconfig"
  file_permission      = "0644"
  directory_permission = "0755"
}