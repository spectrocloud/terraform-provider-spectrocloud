data "spectrocloud_cluster_profile" "profile" {
  name = "withcredentials-full"
}

data "spectrocloud_cluster_profile" "system" {
  name = "system-profile"
}

output "same" {
  value = data.spectrocloud_cluster_profile.profile
}