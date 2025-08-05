# Example demonstrating workspace with GPU support, cluster-specific resource allocations, and cluster names

data "spectrocloud_cluster" "cluster1" {
  name = "api-aks-cazfl"
}

resource "spectrocloud_workspace" "workspace" {
  name        = "wsp-tf-123"
  description = "test123"
  workspace_quota {
    cpu    = 16
    memory = 32768
    gpu    = 4
  }
  clusters {
    uid = data.spectrocloud_cluster.cluster1.id
    # cluster_name is computed automatically by fetching cluster details from the API
  }

  cluster_rbac_binding {
    type = "ClusterRoleBinding"

    role = {
      kind = "ClusterRole"
      name = "testrole3"
    }
    subjects {
      type = "User"
      name = "testRoleUser4"
    }
    subjects {
      type = "Group"
      name = "testRoleGroup4"
    }
    subjects {
      type      = "ServiceAccount"
      name      = "testrolesubject3"
      namespace = "testrolenamespace"
    }
  }

  namespaces {
    name = "multi-cluster-ns"
    resource_allocation = {
      cpu_cores    = "8"
      memory_MiB   = "8192"
      gpu_limit    = "2"
      gpu_provider = "nvidia"
    }

    # Cluster-specific resource allocations
    cluster_resource_allocations {
      uid = data.spectrocloud_cluster.cluster1.id
      resource_allocation = {
        cpu_cores  = "4"
        memory_MiB = "4096"
        gpu_limit  = "1"
      }
    }

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
