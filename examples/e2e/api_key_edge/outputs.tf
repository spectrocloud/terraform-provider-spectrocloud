output "cluster_id" {
  value = spectrocloud_cluster_edge.cluster.id
}

output "cluster_kubeconfig" {
  value = local.cluster_kubeconfig
}

