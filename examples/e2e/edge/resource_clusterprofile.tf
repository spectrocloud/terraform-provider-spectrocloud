data "spectrocloud_cluster_profile" "profile" {
   name = "libvirt"
}

output "same" {
   value = data.spectrocloud_cluster_profile.profile
}