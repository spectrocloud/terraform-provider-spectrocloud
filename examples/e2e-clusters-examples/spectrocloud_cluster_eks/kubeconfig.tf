# Writes the cluster's kubeconfig files to local disk on every apply. See spectrocloud_cluster_vsphere's
# kubeconfig.tf for the general pattern.
resource "local_file" "kubeconfig" {
  content         = spectrocloud_cluster_eks.cluster.kubeconfig
  filename        = "${var.kubeconfig_output_dir}/kubeconfig_${spectrocloud_cluster_eks.cluster.name}"
  file_permission = "0600"
}

resource "local_file" "admin_kubeconfig" {
  content         = spectrocloud_cluster_eks.cluster.admin_kube_config
  filename        = "${var.kubeconfig_output_dir}/admin_kubeconfig_${spectrocloud_cluster_eks.cluster.name}"
  file_permission = "0600"
}
