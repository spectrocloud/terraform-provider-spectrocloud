terraform {
  required_providers {
    spectrocloud = {
      source  = "spectrocloud/spectrocloud"
      version = ">= 0.1"
    }
  }
}

# Example: Create a CloudStack cluster
resource "spectrocloud_cluster_cloudstack" "cluster" {
  name        = "cloudstack-cluster-1"
  description = "CloudStack cluster managed by Terraform"
  tags        = ["env:production", "team:devops"]

  # Cloud account
  cloud_account_id = var.cloudstack_cloud_account_id

  # Cluster profile - contains the K8s version, CNI, CSI, etc.
  cluster_profile {
    id = var.cluster_profile_id
  }

  # CloudStack specific configuration
  cloud_config {
    domain       = "domain1"
    project      = "project1" # Optional
    ssh_key_name = "my-ssh-key"

    # Optional: Set control plane endpoint for static networks
    # control_plane_endpoint = "192.168.1.100"

    # Zone configuration - CloudStack zones for multi-AZ deployment
    zone {
      name = "zone1"

      # Optional: Network configuration for the zone
      network {
        # Either 'id' or 'name' can be used to identify the network
        # If both are specified, 'id' takes precedence
        # id      = "network-uuid-123"
        name    = "network1"
        type    = "Isolated"
        gateway = "192.168.1.1"
        netmask = "255.255.255.0"

        # Optional: Advanced network configuration
        # offering     = "DefaultNetworkOffering"
        # routing_mode = "Static"

        # Optional: VPC configuration (only for VPC-based deployments)
        # vpc {
        #   name     = "my-vpc"
        #   cidr     = "10.0.0.0/16"
        #   offering = "Default VPC Offering"
        # }
      }
    }

    # Additional zones for multi-AZ setup
    # zone {
    #   name = "zone2"
    #   network {
    #     name = "network2"
    #   }
    # }
  }

  # Control plane machine pool
  machine_pool {
    name                    = "control-plane-pool"
    count                   = 3
    control_plane           = true
    control_plane_as_worker = false

    offering = "Medium Instance" # CloudStack compute offering

    # Optional: Disk configuration
    # disk_offering = "Custom Disk"
    # root_disk_size_gb = 100

    # Optional: Affinity groups for VM placement
    # affinity_group_ids = ["affinity-group-1"]

    # Optional: Network configuration
    # network {
    #   network_name = "control-plane-network"
    #   ip_address   = "192.168.1.10"  # Static IP (optional)
    # }

    # Optional: Custom details
    # details = {
    #   "custom_key" = "custom_value"
    # }
  }

  # Worker machine pool
  machine_pool {
    name          = "worker-pool"
    count         = 3
    control_plane = false

    # Autoscaling configuration
    min = 3
    max = 10

    offering = "Large Instance"

    root_disk_size_gb = 200

    # Network configuration for workers
    network {
      network_name = "worker-network"
    }

    # Additional labels for the nodes
    additional_labels = {
      "workload" = "general"
      "tier"     = "backend"
    }

    # Taints (optional)
    # taints {
    #   key    = "dedicated"
    #   value  = "worker"
    #   effect = "NoSchedule"
    # }
  }

  # Optional: OS patching configuration
  os_patch_on_boot = false
  # os_patch_schedule = "0 2 * * SUN"  # Patch every Sunday at 2 AM

  # Optional: Backup policy
  # backup_policy {
  #   schedule                  = "0 1 * * *"
  #   backup_location_id        = var.backup_location_id
  #   prefix                    = "cloudstack-backup"
  #   expiry_in_hour           = 168  # 7 days
  #   include_disks            = true
  #   include_cluster_resources = true
  # }

  # Optional: Scan policy
  # scan_policy {
  #   configuration_scan_schedule = "0 3 * * *"
  #   penetration_scan_schedule   = "0 4 * * 6"
  #   conformance_scan_schedule   = "0 5 * * *"
  # }
}

# Output cluster ID and kubeconfig
output "cluster_id" {
  value = spectrocloud_cluster_cloudstack.cluster.id
}

output "kubeconfig" {
  value     = spectrocloud_cluster_cloudstack.cluster.kubeconfig
  sensitive = true
}

