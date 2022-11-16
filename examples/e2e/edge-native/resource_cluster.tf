resource "spectrocloud_cluster_edge_native" "cluster" {
  name            = "edge-native-example"
  skip_completion = true

  cluster_profile {
    id = data.spectrocloud_cluster_profile.profile.id
  }

  cloud_config {
    ssh_key = "spectro2022"
    vip     = "192.168.100.15"
  }

  machine_pool {
    control_plane           = true
    control_plane_as_worker = true
    name                    = "master-pool"

    host_uids = [spectrocloud_appliance.appliance0.uid]

  }

  machine_pool {
    name      = "worker-pool"
    host_uids = [spectrocloud_appliance.appliance1.uid]

  }

}
