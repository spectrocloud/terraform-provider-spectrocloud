# Writes the cluster's kubeconfig files to local disk on every apply, so you don't have to run
# `terraform output -raw kubeconfig > ...` by hand. Uses the hashicorp/local provider.
#
# Day-2 note: both spectrocloud_cluster_edge_native.cluster.kubeconfig and .admin_kube_config are
# Computed - Palette regenerates them whenever the cluster changes, so these files are rewritten
# on every apply that touches the cluster.

# Non-admin (limited-scope) kubeconfig.
resource "local_file" "kubeconfig" {
  content         = spectrocloud_cluster_edge_native.cluster.kubeconfig
  filename        = "${var.kubeconfig_output_dir}/kubeconfig_${spectrocloud_cluster_edge_native.cluster.name}"
  file_permission = "0600"
}

# Cluster-admin kubeconfig - full cluster control, treat this file like any other secret.
resource "local_file" "admin_kubeconfig" {
  content         = spectrocloud_cluster_edge_native.cluster.admin_kube_config
  filename        = "${var.kubeconfig_output_dir}/admin_kubeconfig_${spectrocloud_cluster_edge_native.cluster.name}"
  file_permission = "0600"
}
