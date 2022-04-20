data "spectrocloud_cluster_profile" "profile" {
  name = "bm-gpu-full"
}

output "same" {
  value = data.spectrocloud_cluster_profile.profile
}