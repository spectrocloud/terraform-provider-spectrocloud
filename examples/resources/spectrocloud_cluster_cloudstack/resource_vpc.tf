terraform {
  required_providers {
    spectrocloud = {
      source  = "spectrocloud/spectrocloud"
      version = ">= 0.1"
    }
  }
}

# Example: Create a CloudStack cluster with VPC configuration
# This example shows how to use the advanced VPC networking features
resource "spectrocloud_cluster_cloudstack" "vpc_cluster" {
  name        = "cloudstack-vpc-cluster"
  description = "CloudStack cluster with VPC networking"
  tags        = ["env:production", "network:vpc"]

  cloud_account_id = var.cloudstack_cloud_account_id

  cluster_profile {
    id = var.cluster_profile_id
  }

  # CloudStack VPC configuration
  cloud_config {
    domain       = "domain1"
    project      = "vpc-project"
    ssh_key_name = "vpc-ssh-key"

    # Zone with VPC network configuration
    zone {
      name = "zone1"

      network {
        # Either 'id' or 'name' can be used to identify the network
        # id      = "network-uuid-vpc-123"
        name    = "vpc-network"
        type    = "Isolated"
        gateway = "10.0.1.1"
        netmask = "255.255.255.0"

        # Network offering for custom network capabilities
        offering = "DefaultNetworkOffering"

        # Routing mode configuration
        routing_mode = "Static"

        # VPC configuration
        vpc {
          name     = "production-vpc"
          cidr     = "10.0.0.0/16"
          offering = "Default VPC Offering"
        }
      }
    }

    # Multi-zone VPC deployment
    zone {
      name = "zone2"

      network {
        name    = "vpc-network-zone2"
        type    = "Isolated"
        gateway = "10.0.2.1"
        netmask = "255.255.255.0"

        offering     = "DefaultNetworkOffering"
        routing_mode = "Static"

        # Same VPC, different subnet
        vpc {
          name     = "production-vpc"
          cidr     = "10.0.0.0/16"
          offering = "Default VPC Offering"
        }
      }
    }
  }

  # Control plane in VPC
  machine_pool {
    name                    = "vpc-control-plane"
    count                   = 3
    control_plane           = true
    control_plane_as_worker = false

    offering = "Medium Instance"

    # VPC network configuration
    network {
      network_name = "vpc-network"
      ip_address   = "10.0.1.10" # Static IP in VPC subnet
    }
  }

  # Worker pool in VPC
  machine_pool {
    name          = "vpc-workers"
    count         = 3
    control_plane = false

    min = 3
    max = 10

    offering = "Large Instance"

    root_disk_size_gb = 200

    network {
      network_name = "vpc-network"
      # Dynamic IP assignment in VPC
    }

    additional_labels = {
      "vpc"      = "production-vpc"
      "workload" = "general"
    }
  }
}

# Example: CloudStack cluster with Shared VPC network
resource "spectrocloud_cluster_cloudstack" "shared_vpc_cluster" {
  name        = "cloudstack-shared-vpc-cluster"
  description = "CloudStack cluster using shared VPC network"

  cloud_account_id = var.cloudstack_cloud_account_id

  cluster_profile {
    id = var.cluster_profile_id
  }

  cloud_config {
    domain       = "domain1"
    ssh_key_name = "shared-ssh-key"

    zone {
      name = "zone1"

      network {
        name         = "shared-vpc-network"
        type         = "Shared"
        gateway      = "10.10.0.1"
        netmask      = "255.255.0.0"
        offering     = "SharedNetworkOffering"
        routing_mode = "Dynamic"

        vpc {
          name     = "shared-vpc"
          cidr     = "10.10.0.0/16"
          offering = "Shared VPC Offering"
        }
      }
    }
  }

  machine_pool {
    name                    = "control-plane"
    count                   = 1
    control_plane           = true
    control_plane_as_worker = true

    offering = "Small Instance"

    network {
      network_name = "shared-vpc-network"
    }
  }
}

# Variables
variable "cloudstack_cloud_account_id" {
  type        = string
  description = "CloudStack cloud account ID"
}

variable "cluster_profile_id" {
  type        = string
  description = "Cluster profile ID containing K8s version and add-ons"
}

# Outputs
output "vpc_cluster_id" {
  value       = spectrocloud_cluster_cloudstack.vpc_cluster.id
  description = "VPC cluster ID"
}

output "vpc_cluster_kubeconfig" {
  value       = spectrocloud_cluster_cloudstack.vpc_cluster.kubeconfig
  sensitive   = true
  description = "VPC cluster kubeconfig"
}

output "shared_vpc_cluster_id" {
  value       = spectrocloud_cluster_cloudstack.shared_vpc_cluster.id
  description = "Shared VPC cluster ID"
}

