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

    additional_annotations = {
      "custom.io/annotation" = "value"
      "company.com/owner"    = "platform-team"
    }

    override_kubeadm_configuration = <<-EOT
      kubeletExtraArgs:
        node-labels: "env=production,tier=frontend"
        max-pods: "110"
      preKubeadmCommands:
        - echo 'Starting node setup'
        - sysctl -w net.ipv4.ip_forward=1
      postKubeadmCommands:
        - echo 'Node setup complete'
        - systemctl restart kubelet
    EOT

    # Update Strategy Options:
    # - "RollingUpdateScaleOut" (default): Adds new nodes before removing old ones
    # - "RollingUpdateScaleIn": Removes old nodes before adding new ones
    # - "OverrideScaling": Custom control with max_surge and max_unavailable
    #   Note: When using "OverrideScaling", you MUST specify override_scaling block
    update_strategy      = "RollingUpdateScaleOut"
    node_repave_interval = 90
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
  # }

  # Optional: Worker Pool with Override Scaling Strategy
  # This example demonstrates the use of override_scaling for fine-grained control
  # over rolling updates with custom surge and unavailability settings.
  # machine_pool {
  #   name  = "worker-pool-override-scaling"
  #   count = 3
  #   min   = 2
  #   max   = 5
  #
  #   placement {
  #     zone         = var.cloudstack_zone_name
  #     compute      = var.cloudstack_compute_offering_worker
  #     network_name = var.cloudstack_network_name
  #   }
  #
  #   # Use OverrideScaling strategy to control the rolling update behavior
  #   # Note: When using OverrideScaling, you MUST specify the override_scaling block
  #   update_strategy = "OverrideScaling"
  #
  #   # Override scaling configuration for rolling updates
  #   # max_surge: Maximum number of nodes that can be created above the desired count
  #   # max_unavailable: Maximum number of nodes that can be unavailable during update
  #   # Values can be absolute numbers (e.g., "1", "2") or percentages (e.g., "25%", "50%")
  #   override_scaling {
  #     max_surge       = "1"    # Allow 1 extra node during updates
  #     max_unavailable = "0"    # Ensure no nodes are unavailable (zero-downtime updates)
  #   }
  #
  #   additional_labels = {
  #     "role"            = "worker"
  #     "update-strategy" = "override-scaling"
  #   }
  #
  #   node_repave_interval = 90
  # }
  #
  # Alternative example with percentage-based scaling:
  # machine_pool {
  #   name  = "worker-pool-percentage-scaling"
  #   count = 4
  #   min   = 2
  #   max   = 10
  #
  #   placement {
  #     zone         = var.cloudstack_zone_name
  #     compute      = var.cloudstack_compute_offering_worker
  #     network_name = var.cloudstack_network_name
  #   }
  #
  #   update_strategy = "OverrideScaling"
  #
  #   # Using percentage values for scaling control
  #   override_scaling {
  #     max_surge       = "25%"  # Allow up to 25% more nodes during updates
  #     max_unavailable = "25%"  # Allow up to 25% of nodes to be unavailable
  #   }
  #
  #   additional_labels = {
  #     "role"            = "worker"
  #     "update-strategy" = "percentage-scaling"
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

