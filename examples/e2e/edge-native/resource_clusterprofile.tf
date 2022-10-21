data "spectrocloud_cluster_profile" "profile" {
  name = "edge-native-infra"
}

output "same" {
  value = data.spectrocloud_cluster_profile.profile
}