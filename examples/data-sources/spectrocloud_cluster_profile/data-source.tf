data "spectrocloud_cluster_profile" "profile1" {
  name = "niktest_profile"
}

output "same" {
  value = data.spectrocloud_cluster_profile.profile1
}
