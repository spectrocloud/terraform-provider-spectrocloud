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

  cloud_config {
    # CloudStack Network Configuration
    zone_name    = var.cloudstack_zone_name
    network_name = var.cloudstack_network_name

    # Optional: SSH key for cluster nodes
    ssh_key = var.ssh_key_name
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

