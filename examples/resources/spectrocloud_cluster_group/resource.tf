resource "spectrocloud_cluster_group" "cg" {
  name        = "ran-cp-cluster-group"
  context     = "tenant"
  description = "Cluster Group description updated"
  tags        = ["qa:dev"]

  config {
    host_endpoint_type       = "Ingress"
    cpu_millicore            = 12000
    memory_in_mb             = 16384
    storage_in_gb            = 12
    oversubscription_percent = 120
    values                   = ""
    k8s_distribution         = "k3s"
  }

  clusters {
    cluster_uid = "684fba868619ff5e691f0741"
    host_dns    = "*.dev.spectrocloud.com"
  }

}

# terraform import spectrocloud_cluster_group.cg "cluster_group_id:tenant"

# Or using the import block (Terraform 1.5+):
# import {
#   to = spectrocloud_cluster_group.cg
#   id = "cluster_group_id:context"
# }