# Comprehensive spectrocloud_cluster_group example - exercises the full attribute surface of the
# resource in one group, not just the minimal required fields. See
# examples/Cluster Group/resource for the minimal happy-path version.
#
# A cluster group ("hostCluster" type) logically groups one or more already-existing
# Palette-managed clusters so that Virtual Clusters (spectrocloud_virtual_cluster, via its
# cluster_group_uid attribute) can be scheduled across them collectively, sharing one resource
# quota and one add-on profile instead of configuring each host cluster individually.
#
# Day-2 mutability: only `config.k8s_distribution` is ForceNew - changing the underlying
# distribution recreates the cluster group. Everything else - name, context, tags, description,
# the rest of config, cluster_profile, and the clusters list - updates in place.
resource "spectrocloud_cluster_group" "cg" {
  # Required.
  name = "e2e-cluster-group"
  # Optional, default "tenant". Allowed: "project", "tenant".
  context = "tenant"
  # Optional. Tags in `key:value` form.
  tags = ["e2e-usecase", "team:platform", "owner:bob"]
  # Optional, default "". Free-text description.
  description = "Comprehensive end-to-end usecase cluster group exercising the full schema."

  # Required, MaxItems 1. Resource quota and endpoint settings shared by every host cluster in
  # the group.
  config {
    # Optional, default "Ingress". "LoadBalancer" is shown here as the alternative.
    host_endpoint_type = "LoadBalancer"
    # Optional. All 4 limit fields are independently optional - set only the ones you want to
    # bound.
    cpu_millicore            = 12000
    memory_in_mb             = 16384
    storage_in_gb            = 12
    oversubscription_percent = 120
    # Optional, default "". Free-form YAML values override applied to the group configuration.
    values = <<-EOT
      global:
        logLevel: info
    EOT
    # Optional, default "vcluster-generic", ForceNew. Allowed: "vcluster-generic", "cncf_k8s".
    # "k3s" also exists in the schema but is deprecated and rejected on create for new cluster
    # groups (existing k3s cluster groups remain supported/updatable, but should migrate).
    k8s_distribution = "cncf_k8s"
  }

  # Optional: an add-on profile applied across every host cluster in the group at once. Same
  # block shape as spectrocloud_cluster_profile's cluster_profile reference elsewhere in this
  # repo (id + variables + pack).
  cluster_profile {
    id = spectrocloud_cluster_profile.host_group_addon_profile.id
    # Supplies values for the profile_variables declared on the profile.
    variables = {
      app_version      = "1.2.0"
      environment_tier = "staging"
      notes            = "Deployed by the cluster-group end-to-end usecase example."
    }
    # Optional: override or add a pack's values just for this group, without touching the shared
    # profile definition.
    pack {
      name   = "byo-manifest-addon"
      type   = "manifest"
      values = local.byo_manifest_values
    }
  }

  # Optional, repeatable: the host clusters this group manages. Two are shown here (rather than
  # just one) to exercise the repeatable block - a single host cluster is just as valid.
  clusters {
    cluster_uid = var.host_cluster_uid_primary
    host_dns    = var.host_cluster_host_dns_primary
  }
  clusters {
    cluster_uid = var.host_cluster_uid_secondary
    host_dns    = var.host_cluster_host_dns_secondary
  }
}

# Import an existing cluster group. The ID must be "<group_uid_or_name>:<project|tenant>" - the
# context suffix is required, not optional; omitting it causes the import to fail.
#
# terraform import spectrocloud_cluster_group.cg "cluster_group_id:tenant"
#
# Or using the import block (Terraform 1.5+):
# import {
#   to = spectrocloud_cluster_group.cg
#   id = "cluster_group_id:tenant"
# }
