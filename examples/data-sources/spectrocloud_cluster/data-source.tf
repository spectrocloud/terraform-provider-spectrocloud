# Looks up an existing cluster by name and retrieves its connection details.
data "spectrocloud_cluster" "example_cluster" {
  # Required lookup key.
  name = "my-cluster"
  # Optional lookup key, default "project". Allowed: "project", "tenant".
  context = "project"
  # Optional lookup key, default false. Set true to look up a virtual cluster instead of a
  # regular cluster.
  virtual = false
}

# Computed. State/health/timezone are populated automatically from the cluster.
output "cluster_state" {
  value = data.spectrocloud_cluster.example_cluster.state
}

output "cluster_health" {
  value = data.spectrocloud_cluster.example_cluster.health
}

output "cluster_timezone" {
  value = data.spectrocloud_cluster.example_cluster.cluster_timezone
}

resource "local_file" "kube_config" {
  # Computed, sensitive. Non-admin kubeconfig for this cluster.
  content              = data.spectrocloud_cluster.example_cluster.kube_config
  filename             = "client-101.kubeconfig"
  file_permission      = "0644"
  directory_permission = "0755"
}

resource "local_file" "admin_kube_config" {
  # Computed, sensitive. Cluster-admin kubeconfig for this cluster.
  content              = data.spectrocloud_cluster.example_cluster.admin_kube_config
  filename             = "admin-client-101.kubeconfig"
  file_permission      = "0644"
  directory_permission = "0755"
}
