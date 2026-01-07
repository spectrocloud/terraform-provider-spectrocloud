# Apache CloudStack Cluster Example with Override Scaling Strategy
#
# This example demonstrates the use of the override_scaling feature for fine-grained
# control over rolling updates in Apache CloudStack clusters.
#
# The override_scaling feature allows you to specify custom max_surge and max_unavailable
# values to control how many nodes can be added or removed during a rolling update.

data "spectrocloud_cloudaccount_apache_cloudstack" "account" {
  name = var.cluster_cloud_account_name
}

data "spectrocloud_cluster_profile" "profile" {
  name = var.cluster_cluster_profile_name
}

resource "spectrocloud_cluster_apache_cloudstack" "cluster_with_override_scaling" {
  name             = "${var.cluster_name}-override-scaling"
  tags             = ["dev", "override-scaling", "cloudstack"]
  cloud_account_id = data.spectrocloud_cloudaccount_apache_cloudstack.account.id

  cloud_config {
    ssh_key_name = var.ssh_key_name

    zone {
      name = var.cloudstack_zone_name

      network {
        name = var.cloudstack_network_name
      }
    }
  }

  cluster_profile {
    id = data.spectrocloud_cluster_profile.profile.id
  }

  # Control Plane Pool with Standard Rolling Update
  machine_pool {
    control_plane           = true
    control_plane_as_worker = true
    name                    = "cp-pool"
    count                   = 1

    placement {
      zone         = var.cloudstack_zone_name
      compute      = var.cloudstack_compute_offering
      network_name = var.cloudstack_network_name
    }

    additional_labels = {
      "role" = "control-plane"
    }

    # Control planes typically use the default RollingUpdateScaleOut strategy
    update_strategy = "RollingUpdateScaleOut"
  }

  # Worker Pool with Override Scaling - Zero Downtime Updates
  # This configuration ensures no nodes are unavailable during updates
  # by creating new nodes before removing old ones.
  machine_pool {
    name  = "worker-pool-zero-downtime"
    count = 3
    min   = 2
    max   = 6

    placement {
      zone         = var.cloudstack_zone_name
      compute      = var.cloudstack_compute_offering_worker
      network_name = var.cloudstack_network_name
    }

    # IMPORTANT: When update_strategy is set to "OverrideScaling",
    # the override_scaling block MUST be specified.
    update_strategy = "OverrideScaling"

    # Zero-downtime configuration:
    # - max_surge = "1": Allow 1 extra node to be created during updates
    # - max_unavailable = "0": Never allow any nodes to be unavailable
    # This means during an update, a new node is created first, then the old one is removed.
    override_scaling {
      max_surge       = "1"
      max_unavailable = "0"
    }

    additional_labels = {
      "role"            = "worker"
      "update-strategy" = "zero-downtime"
      "environment"     = "production"
    }

    additional_annotations = {
      "scaling.company.com/strategy" = "override-zero-downtime"
      "update.company.com/method"    = "surge-first"
    }

    node_repave_interval = 90
  }

  # Worker Pool with Percentage-based Override Scaling
  # This configuration uses percentages for more flexible scaling
  # in larger clusters.
  machine_pool {
    name  = "worker-pool-percentage"
    count = 4
    min   = 2
    max   = 10

    placement {
      zone         = var.cloudstack_zone_name
      compute      = var.cloudstack_compute_offering_worker
      network_name = var.cloudstack_network_name
    }

    update_strategy = "OverrideScaling"

    # Percentage-based configuration:
    # - max_surge = "25%": Allow up to 25% more nodes during updates
    #   (for count=4, this means 1 extra node; for count=8, 2 extra nodes)
    # - max_unavailable = "25%": Allow up to 25% of nodes to be unavailable
    #   (for count=4, this means 1 node can be down; for count=8, 2 nodes)
    # This provides balanced updates with some downtime but faster completion.
    override_scaling {
      max_surge       = "25%"
      max_unavailable = "25%"
    }

    additional_labels = {
      "role"            = "worker"
      "update-strategy" = "percentage-balanced"
      "environment"     = "staging"
    }

    additional_annotations = {
      "scaling.company.com/strategy" = "override-percentage"
      "update.company.com/method"    = "balanced"
    }

    node_repave_interval = 60
  }

  # Worker Pool with Aggressive Scaling for Non-Production
  # This configuration prioritizes update speed over availability.
  machine_pool {
    name  = "worker-pool-aggressive"
    count = 3
    min   = 1
    max   = 5

    placement {
      zone         = var.cloudstack_zone_name
      compute      = var.cloudstack_compute_offering_worker
      network_name = var.cloudstack_network_name
    }

    update_strategy = "OverrideScaling"

    # Aggressive update configuration:
    # - max_surge = "2": Allow 2 extra nodes during updates
    # - max_unavailable = "1": Allow 1 node to be unavailable
    # This speeds up the update process but may cause brief service disruptions.
    override_scaling {
      max_surge       = "2"
      max_unavailable = "1"
    }

    additional_labels = {
      "role"            = "worker"
      "update-strategy" = "aggressive"
      "environment"     = "development"
    }

    additional_annotations = {
      "scaling.company.com/strategy" = "override-aggressive"
      "update.company.com/method"    = "fast-update"
    }

    node_repave_interval = 30
  }

  timeouts {
    create = "30m"
    update = "30m"
    delete = "30m"
  }
}

# Outputs
output "cluster_id_override_scaling" {
  value       = spectrocloud_cluster_apache_cloudstack.cluster_with_override_scaling.id
  description = "The unique ID of the Apache CloudStack cluster with override scaling"
}

output "cluster_kubeconfig_override_scaling" {
  value       = spectrocloud_cluster_apache_cloudstack.cluster_with_override_scaling.kubeconfig
  description = "Kubeconfig for the Apache CloudStack cluster with override scaling"
  sensitive   = true
}

# Usage Notes:
#
# 1. Update Strategy Types:
#    - RollingUpdateScaleOut (default): Adds new nodes before removing old ones
#    - RollingUpdateScaleIn: Removes old nodes before adding new ones
#    - OverrideScaling: Custom control with max_surge and max_unavailable
#
# 2. Override Scaling Values:
#    - Can be absolute numbers: "0", "1", "2", etc.
#    - Can be percentages: "10%", "25%", "50%", etc.
#    - max_surge: Maximum number of nodes that can be created above desired count
#    - max_unavailable: Maximum number of nodes that can be unavailable during update
#
# 3. Common Patterns:
#    - Zero Downtime: max_surge="1", max_unavailable="0"
#      → Always have full capacity, updates are slower
#    - Balanced: max_surge="25%", max_unavailable="25%"
#      → Good balance between speed and availability
#    - Fast Updates: max_surge="2", max_unavailable="1"
#      → Faster updates, may have brief capacity reduction
#
# 4. Important Validation:
#    - When update_strategy = "OverrideScaling", you MUST specify override_scaling block
#    - Both max_surge and max_unavailable must be provided in override_scaling
#    - Validation will fail if override_scaling is missing when using OverrideScaling strategy

