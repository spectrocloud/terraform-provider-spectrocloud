data "spectrocloud_cloudaccount_apache_cloudstack" "account" {
  # id = <uid>
  name = var.cluster_cloud_account_name
}

data "spectrocloud_cluster_profile" "profile" {
  # id = <uid>
  name = var.cluster_cluster_profile_name
}

resource "spectrocloud_cluster_apache_cloudstack" "cluster" {
  name             = var.cluster_name
  tags             = ["dev", "department:devops", "cloudstack"]
  cloud_account_id = data.spectrocloud_cloudaccount_apache_cloudstack.account.id

  # Optional: Update all worker pools in parallel for faster updates (default: false)
  # update_worker_pools_in_parallel = true

  cloud_config {
    # Optional: SSH key for cluster nodes
    ssh_key_name = var.ssh_key_name

    # Optional: CloudStack project (V1CloudStackResource)
    # project {
    #   id   = var.cloudstack_project_id    # CloudStack project ID
    #   name = var.cloudstack_project_name  # CloudStack project name
    # }

    # Zone configuration (required)
    zone {
      name = var.cloudstack_zone_name

      # Network configuration within the zone
      network {
        name = var.cloudstack_network_name
        # Optional fields:
        # id           = var.cloudstack_network_id
        # type         = "shared"  # or "isolated"
        # gateway      = "10.0.0.1"
        # netmask      = "255.255.255.0"
        # offering     = "DefaultNetworkOffering"
        # routing_mode = "static"
      }
    }
  }

  cluster_profile {
    id = data.spectrocloud_cluster_profile.profile.id

    # Optional: Override cluster profile variables
    # variables = {
    #   "priority"    = "5",
    #   "custom_var" = "value"
    # }

    # To override or specify values for a cluster profile pack:
    # pack {
    #   name   = "spectro-byo-manifest"
    #   tag    = "1.0.x"
    #   values = <<-EOT
    #     manifests:
    #       byo-manifest:
    #         contents: |
    #           # Add manifests here
    #           apiVersion: v1
    #           kind: Namespace
    #           metadata:
    #             name: custom-namespace
    #   EOT
    # }
  }

  # Alternative: Use cluster_template instead of cluster_profile
  # Note: cluster_template and cluster_profile are mutually exclusive
  # cluster_template {
  #   id = data.spectrocloud_cluster_config_template.template.id
  #
  #   # Optional: Override profile variables within the template
  #   cluster_profile {
  #     id = "profile-uid-1"
  #     variables = {
  #       "replicas" = "3"
  #     }
  #   }
  #   cluster_profile {
  #     id = "profile-uid-2"
  #     variables = {
  #       "namespace" = "production"
  #     }
  #   }
  # }

  # Optional: Backup Policy
  # backup_policy {
  #   schedule                  = "0 0 * * SUN"
  #   backup_location_id        = data.spectrocloud_backup_storage_location.bsl.id
  #   prefix                    = "prod-backup"
  #   expiry_in_hour            = 7200
  #   include_disks             = true
  #   include_cluster_resources = true
  # }

  # Optional: Scan Policy
  # scan_policy {
  #   configuration_scan_schedule = "0 0 * * SUN"
  #   penetration_scan_schedule   = "0 0 * * SUN"
  #   conformance_scan_schedule   = "0 0 1 * *"
  # }

  # Control Plane Pool
  machine_pool {
    control_plane           = true
    control_plane_as_worker = true
    name                    = "cp-pool"
    count                   = 1

    # Placement Configuration
    placement {
      zone         = var.cloudstack_zone_name
      compute      = var.cloudstack_compute_offering
      network_name = var.cloudstack_network_name

      # Optional: Static IP Pool
      # static_ip_pool_id = var.static_ip_pool_id
    }

    # Optional: Instance Configuration
    # instance_config {
    #   disk_gib   = 100
    #   memory_mib = 8192
    #   num_cpus   = 4
    # }

    # Optional: CloudStack Template Override
    # template {
    #   name = "ubuntu-22.04-template"
    # }

    additional_labels = {
      "role"    = "control-plane"
      "purpose" = "cluster-management"
    }

    # Optional: Node Taints (uncomment if needed)
    # taints {
    #   key    = "master"
    #   value  = "true"
    #   effect = "NoSchedule"
    # }
  }

  # Worker Pool
  machine_pool {
    name  = "worker-pool"
    count = 2

    # Placement Configuration
    placement {
      zone         = var.cloudstack_zone_name
      compute      = var.cloudstack_compute_offering_worker
      network_name = var.cloudstack_network_name

      # Optional: Static IP Pool
      # static_ip_pool_id = var.static_ip_pool_id
    }

    # Optional: Instance Configuration with custom resources
    # instance_config {
    #   disk_gib   = 200
    #   memory_mib = 16384
    #   num_cpus   = 8
    # }

    additional_labels = {
      "role"    = "worker"
      "purpose" = "workload-execution"
    }

    # Optional: Additional annotations for the worker pool nodes
    # additional_annotations = {
    #   "custom.io/annotation" = "value"
    #   "company.com/owner"    = "platform-team"
    # }

    # Optional: Override kubeadm configuration for worker nodes only
    # This YAML config can override kubeletExtraArgs, preKubeadmCommands, and postKubeadmCommands
    # override_kubeadm_configuration = <<-EOT
    #   kubeletExtraArgs:
    #     node-labels: "env=production,tier=frontend"
    #     max-pods: "110"
    #   preKubeadmCommands:
    #     - echo 'Starting node setup'
    #     - sysctl -w net.ipv4.ip_forward=1
    #   postKubeadmCommands:
    #     - echo 'Node setup complete'
    #     - systemctl restart kubelet
    # EOT

    # Optional: Rolling update strategy (default: RollingUpdateScaleOut)
    # rolling_update_strategy {
    #   type            = "OverrideScaling"        # Options: RollingUpdateScaleOut, RollingUpdateScaleIn, OverrideScaling
    #   max_surge       = "25%"                    # Max extra nodes during update (integer or percentage)
    #   max_unavailable = "1"                      # Max unavailable nodes during update (integer or percentage)
    # }

    # Deprecated: Use rolling_update_strategy instead
    # update_strategy = "RollingUpdateScaleOut"

    # Optional: Node repave interval in days (0 = disabled)
    # node_repave_interval = 90
  }

  # Optional: Additional Worker Pool with Minimum and Maximum Scaling
  # machine_pool {
  #   name    = "worker-pool-scalable"
  #   count   = 2
  #   min     = 1
  #   max     = 5
  #
  #   placement {
  #     zone         = var.cloudstack_zone_name
  #     compute      = var.cloudstack_compute_offering_worker
  #     network_name = var.cloudstack_network_name
  #   }
  #
  #   additional_labels = {
  #     "role"     = "worker"
  #     "scalable" = "true"
  #   }
  #
  #   additional_annotations = {
  #     "cluster-autoscaler.kubernetes.io/enabled" = "true"
  #   }
  #
  #   # Optional: Override kubeadm configuration for this pool
  #   override_kubeadm_configuration = <<-EOT
  #     kubeletExtraArgs:
  #       node-labels: "pool=scalable,priority=low"
  #   EOT
  #
  #   rolling_update_strategy {
  #     type            = "RollingUpdateScaleOut"
  #     max_surge       = "1"
  #     max_unavailable = "0"
  #   }
  # }

  timeouts {
    create = "30m"
    update = "30m"
    delete = "30m"
  }
}

# Output the cluster's kubeconfig
output "cluster_id" {
  value       = spectrocloud_cluster_apache_cloudstack.cluster.id
  description = "The unique ID of the Apache CloudStack cluster"
}

output "cluster_kubeconfig" {
  value       = spectrocloud_cluster_apache_cloudstack.cluster.kubeconfig
  description = "Kubeconfig for the Apache CloudStack cluster"
  sensitive   = true
}

