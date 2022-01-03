output "cluster_id" {
  value = spectrocloud_cluster_virt.cluster.id
}

output "cluster_kubeconfig" {
  value = local.cluster_kubeconfig
}

output "clusterprofile_id" {
  value = spectrocloud_cluster_profile.profile.id
}
