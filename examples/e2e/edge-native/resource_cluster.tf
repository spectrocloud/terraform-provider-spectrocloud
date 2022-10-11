resource "spectrocloud_cluster_edge_native" "cluster" {
  name = "edge-native-example"

  cluster_profile {
    id = data.spectrocloud_cluster_profile.profile.id
  }

  cloud_config {
    ssh_key = "spectro2022"
    host    = "192.168.100.15"
  }

  machine_pool {
    control_plane           = true
    control_plane_as_worker = true
    name                    = "master-pool"
    count                   = 1

    host_uids     = [spectrocloud_appliance.appliance1.uid]

    /*instance_type {
      name = "small"
      disk_size_gb = 30
      memory_mb    = 8096
      cpu          = 4
    }*/
  }

  machine_pool {
    name = "worker-pool"
    min           = 2
    max = 3
    count = 3

    host_uids     = [spectrocloud_appliance.appliance2.uid]

    /*instance_type {
      name = "large"
      disk_size_gb = 30
      memory_mb    = 8096
      cpu          = 2
    }*/
  }

}
