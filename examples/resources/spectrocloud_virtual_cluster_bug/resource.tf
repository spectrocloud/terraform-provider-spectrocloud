resource "spectrocloud_virtual_cluster" "cluster" {
  name = "virtual-cluster-demo"

//  host_cluster_uid = var.host_cluster_uid
  cluster_group_uid = var.cluster_group_uid

  resources {
    max_cpu       = 6
    max_mem_in_mb = 6000
    min_cpu       = 0
    min_mem_in_mb = 0
  }
  # To pause luster set flag to true and for resume set to false, Default : False
  #  pause_cluster = false
}