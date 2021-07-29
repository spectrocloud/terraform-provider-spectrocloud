# If looking up a cluster profile instead of creating a new one
data "spectrocloud_cluster_profile" "profile" {
  name = "openstack-profile"
}