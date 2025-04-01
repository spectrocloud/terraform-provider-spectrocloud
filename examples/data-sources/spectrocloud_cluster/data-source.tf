# Retrieve cluster details by name
data "spectrocloud_cluster" "example_cluster" {
  name    = "my-cluster"   # Name of the cluster
  context = "project"      # Context can be "project" or "tenant"
  virtual = false          # Whether the cluster is virtual
}

resource "local_file" "kube_config" {
  content              = data.spectrocloud_cluster.cluster.kube_config
  filename             = "client-101.kubeconfig"
  file_permission      = "0644"
  directory_permission = "0755"
}

resource "local_file" "admin_kube_config" {
  content              = data.spectrocloud_cluster.cluster.admin_kube_config
  filename             = "admin-client-101.kubeconfig"
  file_permission      = "0644"
  directory_permission = "0755"
}
