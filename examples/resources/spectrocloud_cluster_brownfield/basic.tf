# Basic Brownfield Cluster Registration (Day-1)
# This example shows the minimal required fields for registering an existing Kubernetes cluster

resource "spectrocloud_cluster_brownfield" "basic" {
  name        = "my-existing-cluster"
  cloud_type  = "generic" # Options: aws, eksa, azure, gcp, vsphere, openshift, generic
  context     = "project" # Optional, defaults to "project"
  import_mode = "full"

  description      = "My existing Kubernetes cluster"
  cluster_timezone = "Etc/UTC"
  tags             = ["environment:production", "team:platform", "managed-by:terraform"]
  # apply_setting = "DownloadAndInstall"
  cluster_profile {
    id = "696e05b775ded194bf2c14c1"
  }
  scan_policy {
    configuration_scan_schedule = "0 0 * * SUN"
    penetration_scan_schedule   = "0 0 * * SUN"
    conformance_scan_schedule   = "0 0 1 * *"
  }

  pause_agent_upgrades = "lock"
  machine_pool {
    name = "worker-pool"

    node {
      node_name = "cp-dev-worker"
      action    = "uncordon" # Options: "cordon" or "uncordon"
    }
  }

  #   machine_pool {
  #   name = "master-pool"

  #   node {
  #     node_name = "cp-dev-control-plane2"
  #     # node_id = "8f51f6d9-4cce-47fb-9124-2ac7bf760faa-03865"
  #     action  = "uncordon"  # Options: "cordon" or "uncordon"
  #   }
  # }
  # ClusterRoleBinding - Cluster-wide permissions
  cluster_rbac_binding {
    type = "ClusterRoleBinding"

    role = {
      kind = "ClusterRole"
      name = "cluster-admin"
    }

    # Subject type: User
    subjects {
      type = "User"
      name = "admin-user@example.com"
    }

    # Subject type: Group
    subjects {
      type = "Group"
      name = "platform-admins"
    }

    # Subject type: ServiceAccount (requires namespace)
    subjects {
      type      = "ServiceAccount"
      name      = "cluster-admin-sa"
      namespace = "kube-system"
    }
  }


  backup_policy {
    schedule                  = "0 0 * * SUN"
    backup_location_id        = "696f2b3d3154a9e6d65f6b54"
    prefix                    = "test-backup"
    expiry_in_hour            = 7200
    include_disks             = true
    include_cluster_resources = true
  }

  # RoleBinding - Namespace-specific permissions
  cluster_rbac_binding {
    type      = "RoleBinding"
    namespace = "production"
    role = {
      kind = "Role"
      name = "developer-role"
    }
    subjects {
      type = "User"
      name = "developer@example.com"
    }
    subjects {
      type = "Group"
      name = "developers"
    }
    subjects {
      type      = "ServiceAccount"
      name      = "app-service-account"
      namespace = "production"
    }
  }

}

# resource "kubectl_manifest" "cluster_import" {
#   depends_on = [
#     spectrocloud_cluster_brownfield.e2e,
#     data.http.manifest_content
#   ]

#   # Apply the manifest fetched from manifest_url
#   yaml_body = data.http.manifest_content.response_body

#   # Wait for the manifest to be applied successfully
#   wait = true

#   # Wait for rollouts to complete (for Deployments, StatefulSets, etc.)
#   wait_for_rollout = true
# }


# Output the manifest URL and kubectl command for easy access
output "manifest_url" {
  value     = spectrocloud_cluster_brownfield.basic.manifest_url
  sensitive = false
}

output "kubectl_command" {
  value     = spectrocloud_cluster_brownfield.basic.kubectl_command
  sensitive = false
}

output "cluster_status" {
  value = spectrocloud_cluster_brownfield.basic.status
}



# ============================================================================
# AUTOMATIC CLUSTER IMPORT MANIFEST APPLICATION
# ============================================================================
#
# Purpose:
#   This section automatically applies the cluster import manifest to your
#   Kubernetes cluster after the brownfield cluster registration is created.
#   The manifest contains the Spectro Cloud agent components needed to connect
#   your existing cluster to Spectro Cloud.
#
# How it works:
#   1. The spectrocloud_cluster_brownfield resource provides a manifest_url
#      (computed output) containing the import manifest URL
#   2. The http data source fetches the manifest YAML content from the URL
#   3. The kubectl_manifest resource applies the manifest to your cluster
#      using the configured kubectl provider
#
# Prerequisites:
#   - kubectl CLI must be installed and configured with access to your cluster
#   - kubectl provider must be configured in providers.tf (see below)
#   - http provider must be configured in providers.tf (see below)
#   - Your kubeconfig must be properly set up to access the target cluster
#

# Fetch the manifest content from the manifest_url
data "http" "manifest_content" {
  url    = spectrocloud_cluster_brownfield.basic.manifest_url
  method = "GET"

  depends_on = [
    spectrocloud_cluster_brownfield.basic
  ]
}

# Apply the manifest using kubectl provider
resource "kubectl_manifest" "cluster_import" {
  depends_on = [
    spectrocloud_cluster_brownfield.basic,
    data.http.manifest_content
  ]

  # Apply the manifest fetched from manifest_url
  yaml_body = data.http.manifest_content.response_body

  # Wait for the manifest to be applied successfully
  wait = true

  # Wait for rollouts to complete (for Deployments, StatefulSets, etc.)
  wait_for_rollout = true
}

# Alternative (Manual Application):
#   If you prefer to apply the manifest manually, you can:
#   1. Use the output kubectl_command: terraform output kubectl_command
#   2. Run the command directly on your cluster
#   3. Or use the manifest_url output to fetch and apply manually