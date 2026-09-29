data "spectrocloud_cloudaccount_apache_cloudstack" "account" {
  # id = <uid>
  name = var.cluster_cloud_account_name
}

data "spectrocloud_cluster_profile" "profile" {
  # id = <uid>
  name = var.cluster_cluster_profile_name
}

# Day-2 mutability: only `name` and `cloud_account_id` are ForceNew on this resource - changing
# either recreates the cluster. Everything else below (cloud_config, cluster_profile,
# machine_pool, backup_policy, scan_policy, tags) updates in place, unlike some other cloud
# cluster resources where cloud_config itself is ForceNew.
resource "spectrocloud_cluster_apache_cloudstack" "cluster" {
  name             = var.cluster_name
  tags             = ["dev", "department:devops", "cloudstack"]
  cloud_account_id = data.spectrocloud_cloudaccount_apache_cloudstack.account.id

  # Optional: Update all worker pools in parallel for faster updates (default: false)
  # update_worker_pools_in_parallel = true

  # cloud_config:
  #   ssh_key_name - Optional. SSH key for cluster nodes.
  #   override_cluster_api_config - Optional. Raw Cluster API config overrides, merged into the
  #     generated cluster spec.
  #   project - Optional (V1CloudStackResource), alternative to leaving the project unset: places
  #     the cluster in a specific CloudStack project, by id or name.
  cloud_config {
    ssh_key_name = var.ssh_key_name

    override_cluster_api_config = <<-EOT
      spec:
        controlPlaneConfiguration:
          apiServer:
            extraArgs:
              authorization-mode: Node,RBAC
    EOT

    # project {
    #   id   = var.cloudstack_project_id    # CloudStack project ID
    #   name = var.cloudstack_project_name  # CloudStack project name
    # }

    # zone: Required. The CloudStack zone the cluster is provisioned into.
    zone {
      name = var.cloudstack_zone_name

      # network: Network configuration within the zone. Optional fields not shown: id, type
      #   ("shared" or "isolated"), gateway, netmask, offering, routing_mode.
      network {
        name = var.cloudstack_network_name
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

  # machine_pool (control plane):
  #   offering - Required. CloudStack compute offering (instance type/size) name for this pool.
  #   network - Optional. Selects the network within the cloud_config zone this pool's nodes
  #     attach to. If omitted, the zone's own network (cloud_config.zone.network) is used.
  #   instance_config - Computed (read-only). CloudStack returns the resolved disk/memory/CPU
  #     sizing for the chosen `offering` here; it cannot be set directly.
  #   template - Optional. Overrides the CloudStack template used for this pool's nodes.
  machine_pool {
    control_plane           = true
    control_plane_as_worker = true
    name                    = "cp-pool"
    count                   = 1

    offering = var.cloudstack_compute_offering

    network {
      network_name = var.cloudstack_network_name
    }

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

  # machine_pool (worker pool "worker-pool"):
  #   offering - Required. CloudStack compute offering (instance type/size) name for this pool.
  #   network - Optional. Selects the network within the cloud_config zone this pool's nodes
  #     attach to.
  #   override_kubeadm_configuration - Optional. Raw kubeadm config overrides (extra kubelet
  #     args, pre/post kubeadm commands) merged into the generated node config.
  #   override_cluster_api_config - Optional. Raw Cluster API config overrides, merged into the
  #     generated cluster spec.
  #   skip_k8s_upgrade - Optional, default "disabled". Set to "enabled" to skip the OS/Kubernetes
  #     version upgrade for this worker pool when the cluster profile is upgraded (worker pools
  #     only).
  #   update_strategy - Optional, default "RollingUpdateScaleOut". Allowed:
  #     "RollingUpdateScaleOut" (adds new nodes before removing old ones), "RollingUpdateScaleIn"
  #     (removes old nodes before adding new ones), "OverrideScaling" (custom control via
  #     max_surge/max_unavailable - when used, override_scaling MUST be specified).
  #   override_scaling - Optional, required when update_strategy = "OverrideScaling". max_surge/
  #     max_unavailable accept absolute counts or percentages; e.g. max_surge = "1",
  #     max_unavailable = "0" creates the replacement node before removing the old one
  #     (zero-downtime).
  #   node_repave_interval - Optional. Minutes to wait between repaving (recycling) nodes.
  #   override_health_check_configuration - Optional. Raw Machine Health Check overrides for this
  #     node pool.
  machine_pool {
    name  = "worker-pool"
    count = 2

    offering = var.cloudstack_compute_offering_worker

    network {
      network_name = var.cloudstack_network_name
    }

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

    override_cluster_api_config = <<-EOT
      spec:
        template:
          spec:
            nodeDrainTimeout: 5m
    EOT

    # skip_k8s_upgrade = "disabled"

    # Update Strategy Options:
    # - "RollingUpdateScaleOut" (default): Adds new nodes before removing old ones
    # - "RollingUpdateScaleIn": Removes old nodes before adding new ones
    # - "OverrideScaling": Custom control with max_surge and max_unavailable
    #   Note: When using "OverrideScaling", you MUST specify override_scaling block
    update_strategy = "RollingUpdateScaleOut"
    # IMPORTANT: When update_strategy is set to "OverrideScaling",
    # the override_scaling block MUST be specified.
    # update_strategy = "OverrideScaling"

    # Zero-downtime configuration:
    # - max_surge = "1": Allow 1 extra node to be created during updates
    # - max_unavailable = "0": Never allow any nodes to be unavailable
    # This means during an update, a new node is created first, then the old one is removed.
    # override_scaling {
    #   max_surge       = "1"
    #   max_unavailable = "0"
    # }
    node_repave_interval = 90

    # Optional: override Machine Health Check settings for this node pool
    override_health_check_configuration = <<-EOT
      maxUnhealthy: 40%
      nodeStartupTimeout: 10m
      unhealthyConditions:
        - type: Ready
          status: "False"
          timeout: 5m
        - type: Ready
          status: "Unknown"
          timeout: 5m
    EOT
  }

  # Optional: an autoscaled worker pool (min/max instead of a fixed count) using OverrideScaling -
  # see the "worker-pool" block above for what override_scaling's fields mean. max_surge/
  # max_unavailable also accept percentages (e.g. "25%") instead of absolute counts.
  # machine_pool {
  #   name            = "worker-pool-scalable"
  #   count           = 2
  #   min             = 1
  #   max             = 5
  #   offering        = var.cloudstack_compute_offering_worker
  #   update_strategy = "OverrideScaling"
  #   override_scaling {
  #     max_surge       = "1"
  #     max_unavailable = "0"
  #   }
  #   network {
  #     network_name = var.cloudstack_network_name
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
