resource "spectrocloud_cluster_edge_native" "cluster" {
  name            = "edge-native-example"
  skip_completion = true

  cluster_profile {
    id = data.spectrocloud_cluster_profile.profile.id
  }

  cloud_config {
    ssh_keys = ["spectro2022", "spectro2023"]
    vip      = "100.12.22.10"
    overlay_cidr_range = "100.12.22.11/12"
  }

  machine_pool {
    control_plane           = true
    control_plane_as_worker = true
    name                    = "master-pool"

    edge_host {
      host_uid  = spectrocloud_appliance.appliance0.uid
      static_ip = "126.10.10.23"
    }

  }

  machine_pool {
    name = "worker-pool"

    edge_host {
      host_uid  = spectrocloud_appliance.appliance1.uid
      static_ip = "136.10.10.24"
    }
  }

}
