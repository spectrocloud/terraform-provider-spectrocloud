# Brownfield Cluster Registration
# This example registers an existing Kubernetes cluster and additionally exercises most of the
# resource's optional Day-2 attributes (cluster_profile, scan_policy, backup_policy,
# machine_pool node actions, cluster_rbac_binding) - the only truly required fields are `name`
# and `cloud_type`.
#
# Day-2 mutability: nothing on this resource is marked ForceNew in the schema (Terraform will
# always try an in-place update, never a destroy/recreate), but the provider's own docs state
# that `context` and `import_mode` "cannot be updated after creation" - changing them will not
# trigger a resource replacement, so re-applying with a different value may be a no-op or error
# at the API level rather than doing what you'd expect. Treat both as effectively set-once.

resource "spectrocloud_cluster_brownfield" "basic" {
  name = "my-existing-cluster"
  # Required. The schema's own description lists aws, eks-anywhere, azure, gcp, vsphere,
  # openshift, generic, maas - but the provider's registration logic (resource_cluster_brownfield.go,
  # resourceClusterBrownfieldImportCreate) only actually branches on aws, azure, gcp,
  # apache-cloudstack, and generic. Any other value (including several the schema itself lists,
  # like vsphere/maas/openshift/eks-anywhere) silently falls through to the generic import path
  # with no warning - validation for this field is disabled, so nothing catches the mismatch at
  # plan time. Stick to aws/azure/gcp/apache-cloudstack/generic to get the behavior you expect.
  cloud_type = "generic"
  # Optional, default "project". Allowed: "project", "tenant". Not updatable after creation.
  context = "project"
  # Optional, default "full" (documented; empty string on the wire). Allowed: "read_only", "full".
  # Not updatable after creation.
  import_mode = "full"

  description      = "My existing Kubernetes cluster"
  cluster_timezone = "Etc/UTC"
  tags             = ["environment:production", "team:platform", "managed-by:terraform"]
  apply_setting    = "DownloadAndInstall"
  cluster_profile {
    id = "CLUSTER_PROFILE_ID"
  }
  scan_policy {
    configuration_scan_schedule = "0 0 * * SUN"
    penetration_scan_schedule   = "0 0 * * SUN"
    conformance_scan_schedule   = "0 0 1 * *"
  }

  # Optional, default "unlock". Allowed: "lock" (pin the Palette agent version), "unlock".
  pause_agent_upgrades = "lock"
  machine_pool {
    name = "worker-pool"

    node {
      node_name = "cp-dev-worker"
      action    = "uncordon" # Options: "cordon" or "uncordon"
    }
  }

  machine_pool {
    name = "master-pool"

    node {
      node_name = "cp-dev-control-plane2"
      node_id   = "NODE_ID"
      action    = "uncordon" # Options: "cordon" or "uncordon"
    }
  }
  cluster_rbac_binding {
    type = "ClusterRoleBinding"
    role = {
      kind = "ClusterRole"
      name = "CLUSTER_ROLE_NAME"
    }
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
    backup_location_id        = "BACKUP_LOCATION_ID"
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