# Day-2 mutability: only `config.k8s_distribution` is ForceNew - changing the underlying
# distribution (e.g. "k3s" to "vcluster-generic") recreates the cluster group. Everything else -
# name, context, description, tags, the rest of config, and the clusters list - updates in place.
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
    # Optional, default "vcluster-generic", ForceNew. The Kubernetes distribution virtual
    # clusters in this group run on.
    k8s_distribution = "k3s"
  }

  # Optional. A host cluster profile can also be attached to the group itself (same block shape
  # as spectrocloud_cluster_profile's cluster_profile reference) - omitted here.
  # cluster_profile {
  #   id = data.spectrocloud_cluster_profile.host_profile.id
  # }

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