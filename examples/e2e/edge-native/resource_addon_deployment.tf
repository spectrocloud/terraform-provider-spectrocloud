resource "spectrocloud_addon_deployment" "depl" {
  cluster_uid = spectrocloud_cluster_edge_native.cluster.id
  context     = "project"

  cluster_profile {
    id = spectrocloud_cluster_profile.profile-addon.id
    pack {
      name   = data.spectrocloud_pack.dashboard.name
      tag    = data.spectrocloud_pack.dashboard.version
      uid    = data.spectrocloud_pack.dashboard.id
      values = data.spectrocloud_pack.dashboard.values
    }
  }
}
