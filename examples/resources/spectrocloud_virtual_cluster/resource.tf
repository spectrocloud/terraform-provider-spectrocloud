resource "spectrocloud_virtual_cluster" "cluster" {
  name = "virtual-cluster-demo"

  host_cluster_uid = var.host_cluster_uid

  resources {
    max_cpu       = 6
    max_mem_in_mb = 6000
    min_cpu       = 0
    min_mem_in_mb = 0
  }

}