data "spectrocloud_cluster_profile" "profile" {
   name = "withcredentials-full"
}

output "same" {
   value = data.spectrocloud_cluster_profile.profile
}