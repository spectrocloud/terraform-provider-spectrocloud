output "cluster_id" {
  value = spectrocloud_cluster_aws.cluster.id
}

output "cluster_kubeconfig" {
  value = local.cluster_kubeconfig
}

output "clusterprofile_id" {
  value = spectrocloud_cluster_profile.profile.id
}
