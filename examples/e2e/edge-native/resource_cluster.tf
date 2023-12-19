##  Here's an example on how to provision an edge cluster with:
##    - based on a profile "spectrocloud_cluster_profile"
##    - name: edge-native-tf-1
##    - single-node cluster with VIP and UID of edge device taken from variables

resource "spectrocloud_cluster_edge_native" "cluster" {
  name            = "edge-native-tf-1"
  skip_completion = false

  cluster_profile {
    id = spectrocloud_cluster_profile.profile.id
  }

  cloud_config {
    ssh_keys = ["spectro2023", "spectro2024"]
    vip      = var.vip
  }

  machine_pool {
    control_plane           = true
    control_plane_as_worker = true
    name                    = "master-pool"

    edge_host {
      host_uid = var.edge_id
      #static_ip = "123.45.67.89"
    }

  }
}
