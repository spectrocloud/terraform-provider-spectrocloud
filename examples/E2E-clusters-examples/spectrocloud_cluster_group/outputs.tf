# Unlike the cloud cluster resources in this repo, a cluster group has no kubeconfig of its own
# (virtual clusters scheduled onto it each have their own - see
# examples/E2E-clusters-examples/spectrocloud_virtual_cluster) - so its only computed attribute
# is its id.
output "cluster_group_id" {
  value = spectrocloud_cluster_group.cg.id
}

output "cluster_group_addon_profile_id" {
  value = spectrocloud_cluster_profile.host_group_addon_profile.id
}
