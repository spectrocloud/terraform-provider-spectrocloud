resource "spectrocloud_cluster_group" "cg" {
  name = "cluster-group-demo"
  context = "project"
  clusters {
    cluster_uid = data.spectrocloud_cluster.host_cluster0.id
    host_dns    = "*.test.com"
  }

  clusters {
    cluster_uid = data.spectrocloud_cluster.host_cluster1.id
    host_dns    = "*"
  }

  config {
    host_endpoint_type       = "Ingress"
    cpu_millicore            = 6000
    memory_in_mb             = 8192
    storage_in_gb            = 10
    oversubscription_percent = 120
  }
  cluster_profile{
    id = "6537d5b46ea0c476c9c80750"
  }
}
