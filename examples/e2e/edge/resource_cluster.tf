resource "spectrocloud_cluster_edge" "cluster" {
  name = "edge-dev"

  cluster_profile {
    id = data.spectrocloud_cluster_profile.profile.id
  }

  cloud_config {
    ssh_key = "spectro2022"
  }

  machine_pool {
    control_plane           = true
    control_plane_as_worker = true
    name                    = "master-pool"
    count                   = 1

    placements {
      appliance_id = data.spectrocloud_appliance.virt_appliance.id
    }

  }

  machine_pool {
    name  = "worker-pool"
    count = 1

    placements {
      appliance_id = data.spectrocloud_appliance.virt_appliance.id
    }

  }

}
