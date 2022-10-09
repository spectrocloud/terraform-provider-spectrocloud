
resource "spectrocloud_cluster_nested" "cluster" {
  name = "nested-cluster-demo"


  cluster_config {
    host_cluster_config {
      host_cluster {
        uid = var.host_cluster_uid
      }
      # cluster_group {
      #   uid = var.cluster_group_uid
      # }
    }
    resources {
      max_cpu = 2
      max_mem_in_mb = 6
      min_cpu = 0
      min_mem_in_mb = 0
    }
  }

  # uncomment the following 3 lines to deploy the tekton demo stack
  # cluster_profile {
  #   id = spectrocloud_cluster_profile.profile.id
  # }

  # optional nested cluster config
  # cloud_config {
  #   chart_name = var.chart_name
  #   chart_repo = var.chart_repo
  #   chart_version = var.chart_version
  #   chart_values = var.chart_values
  #   k8s_version = var.k8s_version
  # }

}
