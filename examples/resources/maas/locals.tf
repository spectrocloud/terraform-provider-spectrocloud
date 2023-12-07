locals {
  cluster_kubeconfig = spectrocloud_cluster_maas.cluster.kubeconfig
}
locals {
  cluster_admin_kubeconfig = spectrocloud_cluster_maas.cluster.admin_kube_config
}
