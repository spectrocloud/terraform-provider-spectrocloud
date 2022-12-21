resource "spectrocloud_cluster_group" "cg" {
  name = "cluster-group-demo"

  clusters {
    cluster_uid = data.spectrocloud_cluster.host_cluster0.id
    host        = "*.test.com"
  }

  clusters {
    cluster_uid = data.spectrocloud_cluster.host_cluster1.id
    host        = "*"
  }

  config {
    host_endpoint_type       = "LoadBalancer"
    cpu_millicore            = 6000
    memory_in_mb             = 8192
    storage_in_gb            = 10
    oversubscription_percent = 120
  }
}
