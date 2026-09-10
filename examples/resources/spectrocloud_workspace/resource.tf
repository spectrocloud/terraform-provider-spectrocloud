# Example demonstrating workspace with GPU support, cluster-specific resource allocations, and cluster names

data "spectrocloud_cluster" "cluster1" {
  name = "api-aks-cazfl"
}

# Day-2 mutability: nothing on this resource is ForceNew - name, tags, description,
# workspace_quota, clusters, cluster_rbac_binding, namespaces, and backup_policy all update in
# place.
resource "spectrocloud_workspace" "workspace" {
  # Required.
  name = "wsp-tf-123"
  # Optional. Tags in `key:value` form.
  tags = ["dev", "department:devops", "owner:bob"]
  # Optional.
  description = "test123"

  # Optional, at most one block. Default resource limits for the whole workspace; 0 (default)
  # means no limit.
  workspace_quota {
    cpu    = 16    # vCPU
    memory = 32768 # MiB
    gpu    = 4
  }

  # Required, one or more. Clusters this workspace spans.
  clusters {
    uid = data.spectrocloud_cluster.cluster1.id
    # cluster_name is computed automatically by fetching cluster details from the API
  }

  # Optional, repeatable. Grants Kubernetes RBAC to users/groups/service accounts across the
  # workspace's clusters.
  cluster_rbac_binding {
    # Required. "RoleBinding" (namespace-scoped, requires `namespace` below) or
    # "ClusterRoleBinding" (cluster-scoped, used here).
    type = "ClusterRoleBinding"

    # Optional map with "kind" and "name" keys. Required if type = "RoleBinding".
    role = {
      kind = "ClusterRole"
      name = "testrole3"
    }
    subjects {
      # Required. "User", "Group", or "ServiceAccount".
      type = "User"
      name = "testRoleUser4"
    }
    subjects {
      type = "Group"
      name = "testRoleGroup4"
    }
    subjects {
      # namespace required when type = "ServiceAccount".
      type      = "ServiceAccount"
      name      = "testrolesubject3"
      namespace = "testrolenamespace"
    }
  }

  # Required, one or more. Kubernetes namespaces created/managed across this workspace's
  # clusters.
  namespaces {
    name = "multi-cluster-ns"
    # Required. Default per-cluster resource allocation for this namespace. Only cpu_cores,
    # memory_MiB, gpu, and gpu_provider are honored - any other key is silently ignored.
    resource_allocation = {
      cpu_cores    = "8"
      memory_MiB   = "8192"
      gpu          = "2"
      gpu_provider = "nvidia"
    }

    # Optional, at most one block. Overrides resource_allocation above for one specific cluster.
    # Note: gpu_provider is not supported here - only in the default resource_allocation.
    cluster_resource_allocations {
      uid = data.spectrocloud_cluster.cluster1.id
      resource_allocation = {
        cpu_cores  = "4"
        memory_MiB = "4096"
        gpu        = "1"
      }
    }

    # Optional. Container images disallowed in this namespace.
    images_blacklist = ["nginx:latest", "redis:latest"]
  }

  # backup_policy {
  #   schedule                  = "0 0 * * SUN"
  #   backup_location_id        = data.spectrocloud_backup_storage_location.bsl.id
  #   prefix                    = "prod-backup"
  #   expiry_in_hour            = 7200
  #   include_disks             = false
  #   include_cluster_resources = true

  #   namespaces           = ["test5ns", "multi-cluster-ns"]
  #   include_all_clusters = true
  #   cluster_uids         = [data.spectrocloud_cluster.cluster1.id]
  # }

}

# data "spectrocloud_backup_storage_location" "bsl" {
#   name = "test-aws-s3"
# }
