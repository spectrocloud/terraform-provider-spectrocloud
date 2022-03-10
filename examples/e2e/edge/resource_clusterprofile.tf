data "spectrocloud_cluster_profile" "profile" {
   name = "profile-edge"
}

output "same" {
   value = data.spectrocloud_cluster_profile.profile
}