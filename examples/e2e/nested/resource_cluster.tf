
resource "spectrocloud_cluster_nested" "cluster" {
  name = "nested-cluster-tf-2"
  host_cluster_uid = data.spectrocloud_cluster.host_cluster.id
  #tags = ["skip_completion"]

  # Attach addon profile optionally.
  # cluster_profile {
  #   id = spectrocloud_cluster_profile.profile.id
  # }

  cloud_config {
    chart_name = ""
    chart_repo = ""
    chart_version = ""
    chart_values = ""
    k8s_version = "1.23.0"
  }

  machine_pool {
    control_plane           = true
    control_plane_as_worker = true
    name                    = "master-pool"
    count                   = 1
    resource_pool           = var.resource_pool
  }

}
