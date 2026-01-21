---
page_title: "spectrocloud_cluster_brownfield Resource - terraform-provider-spectrocloud"
subcategory: ""
description: |-
  Register an existing Kubernetes cluster (brownfield) with Spectro Cloud. This resource creates a cluster registration and provides the import link and manifest URL needed to complete the cluster import process.
---

# spectrocloud_cluster_brownfield (Resource)

  Register an existing Kubernetes cluster (brownfield) with Spectro Cloud. This resource creates a cluster registration and provides the import link and manifest URL needed to complete the cluster import process.

## Example Usage
form
data "spectrocloud_cluster_profile" "profile" {
  # id = <uid>
  name = var.cluster_cluster_profile_name
}

data "spectrocloud_backup_storage_location" "bsl" {
  name = var.backup_storage_location_name
}

resource "spectrocloud_cluster_brownfield" "cluster" {
  name       = var.cluster_name
  cloud_type = "generic" # Options: aws, Eks-Anywhere, azure, gcp, vsphere, openshift, generic, apache-cloudstack, edge-native, maas, openstack
  context    = "project" # Optional, defaults to "project"
  import_mode = "full"   # Options: "read_only" or "full" (default: "full")
  
  description     = "My existing Kubernetes cluster"
  cluster_timezone = "America/New_York"
  tags            = ["environment:production", "team:platform", "managed-by:terraform"]

  cluster_profile {
    id = data.spectrocloud_cluster_profile.profile.id
  }

  scan_policy {
    configuration_scan_schedule = "0 0 * * SUN"
    penetration_scan_schedule   = "0 0 * * SUN"
    conformance_scan_schedule   = "0 0 1 * *"
  }

  backup_policy {
    schedule                  = "0 0 * * SUN"
    backup_location_id        = data.spectrocloud_backup_storage_location.bsl.id
    prefix                    = "cluster-backup"
    expiry_in_hour            = 7200
    include_disks             = true
    include_cluster_resources = true
  }

  cluster_rbac_binding {
    type = "ClusterRoleBinding"
    role = {
      kind = "ClusterRole"
      name = "cluster-admin"
    }
    subjects {
      type = "User"
      name = "admin-user@example.com"
    }
    subjects {
      type = "Group"
      name = "platform-admins"
    }
    subjects {
      type      = "ServiceAccount"
      name      = "cluster-admin-sa"
      namespace = "kube-system"
    }
  }

  namespaces {
    name = "production"
    resource_allocation = {
      cpu_cores  = "4"
      memory_MiB = "8192"
    }
  }

  machine_pool {
    name = "worker-pool"
    node {
      node_name = "worker-node-1"
      action    = "uncordon" # Options: "cordon" or "uncordon"
    }
  }
}

# Output the manifest URL and kubectl command for easy access
output "manifest_url" {
  value     = spectrocloud_cluster_brownfield.cluster.manifest_url
  sensitive = false
}

output "kubectl_command" {
  value     = spectrocloud_cluster_brownfield.cluster.kubectl_command
  sensitive = false
}
