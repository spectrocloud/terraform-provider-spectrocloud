#############
# AWS Cluster
#############
resource "spectrocloud_cluster_aws" "aws-cluster" {
  count = var.deploy-aws ? 1 : 0

  name             = "aws-cluster"
  tags             = concat(var.tags, ["env:aws"])
  cloud_account_id = data.spectrocloud_cloudaccount_aws.account[0].id

  cloud_config {
    region       = var.aws-region
    ssh_key_name = var.aws-key-pair-name
  }

  cluster_profile {
    id = var.deploy-aws && var.deploy-aws-kubecost ? resource.spectrocloud_cluster_profile.aws-profile-kubecost[0].id : resource.spectrocloud_cluster_profile.aws-profile[0].id
  }

  machine_pool {
    control_plane           = true
    control_plane_as_worker = true
    name                    = "control-plane-pool"
    count                   = var.aws_control_plane_nodes.count
    instance_type           = var.aws_control_plane_nodes.instance_type
    disk_size_gb            = var.aws_control_plane_nodes.disk_size_gb
    azs                     = var.aws_control_plane_nodes.availability_zones
  }

  machine_pool {
    name          = "worker-pool"
    count         = var.aws_worker_nodes.count
    instance_type = var.aws_worker_nodes.instance_type
    disk_size_gb  = var.aws_worker_nodes.disk_size_gb
    azs           = var.aws_worker_nodes.availability_zones
  }

  timeouts {
    create = "45m"
    delete = "45m"
  }
}
###############
# Azure Cluster
###############
resource "spectrocloud_cluster_azure" "azure-cluster" {
  count = var.deploy-azure ? 1 : 0

  name             = "azure-cluster"
  tags             = concat(var.tags, ["env:azure"])
  cloud_account_id = data.spectrocloud_cloudaccount_azure.account[0].id

  cloud_config {
    subscription_id = var.azure_subscription_id
    resource_group  = var.azure_resource_group
    region          = var.azure-region
    ssh_key         = tls_private_key.tutorial_ssh_key_azure[0].public_key_openssh
  }

  cluster_profile {
    id = var.deploy-azure && var.deploy-azure-kubecost ? resource.spectrocloud_cluster_profile.azure-profile-kubecost[0].id : resource.spectrocloud_cluster_profile.azure-profile[0].id
  }

  machine_pool {
    control_plane           = true
    control_plane_as_worker = true
    name                    = "control-plane-pool"
    count                   = var.azure_control_plane_nodes.count
    instance_type           = var.azure_control_plane_nodes.instance_type
    azs                     = var.azure-use-azs ? var.azure_control_plane_nodes.azs : [""]
    is_system_node_pool     = var.azure_control_plane_nodes.is_system_node_pool
    disk {
      size_gb = var.azure_control_plane_nodes.disk_size_gb
      type    = "Standard_LRS"
    }
  }

  machine_pool {
    name                = "worker-basic"
    count               = var.azure_worker_nodes.count
    instance_type       = var.azure_worker_nodes.instance_type
    azs                 = var.azure-use-azs ? var.azure_worker_nodes.azs : [""]
    is_system_node_pool = var.azure_worker_nodes.is_system_node_pool
  }

  timeouts {
    create = "45m"
    delete = "45m"
  }
}

#############
# GCP Cluster
#############
resource "spectrocloud_cluster_gcp" "gcp-cluster" {
  count = var.deploy-gcp ? 1 : 0

  name             = "gcp-cluster"
  tags             = concat(var.tags, ["env:gcp"])
  cloud_account_id = data.spectrocloud_cloudaccount_gcp.account[0].id

  cloud_config {
    project = var.gcp_project_name
    region  = var.gcp-region
  }

  cluster_profile {
    id = var.deploy-gcp && var.deploy-gcp-kubecost ? resource.spectrocloud_cluster_profile.gcp-profile-kubecost[0].id : resource.spectrocloud_cluster_profile.gcp-profile[0].id
  }

  machine_pool {
    control_plane           = true
    control_plane_as_worker = true
    name                    = "control-plane-pool"
    count                   = var.gcp_control_plane_nodes.count
    instance_type           = var.gcp_control_plane_nodes.instance_type
    disk_size_gb            = var.gcp_control_plane_nodes.disk_size_gb
    azs                     = var.gcp_control_plane_nodes.availability_zones
  }

  machine_pool {
    name          = "worker-pool"
    count         = var.gcp_worker_nodes.count
    instance_type = var.gcp_worker_nodes.instance_type
    disk_size_gb  = var.gcp_worker_nodes.disk_size_gb
    azs           = var.gcp_worker_nodes.availability_zones
  }

  timeouts {
    create = "45m"
    delete = "45m"
  }
}

################
# VMware Cluster
################

resource "spectrocloud_cluster_vsphere" "vmware-cluster" {
  count = var.deploy-vmware ? 1 : 0

  name             = "vmware-cluster"
  tags             = concat(var.tags, ["env:vmware"])
  cloud_account_id = data.spectrocloud_cloudaccount_vsphere.account[0].id

  cloud_config {
    ssh_keys              = [local.ssh_public_key]
    datacenter            = var.datacenter_name
    folder                = var.folder_name
    static_ip             = var.deploy-vmware-static # If true, the cluster will use static IP placement. If false, the cluster will use DDNS.
    network_search_domain = var.search_domain
  }

  cluster_profile {
    id = var.deploy-vmware && var.deploy-vmware-kubecost ? resource.spectrocloud_cluster_profile.vmware-profile-kubecost[0].id : resource.spectrocloud_cluster_profile.vmware-profile[0].id
  }

  scan_policy {
    configuration_scan_schedule = "0 0 * * SUN"
    penetration_scan_schedule   = "0 0 * * SUN"
    conformance_scan_schedule   = "0 0 1 * *"
  }

  machine_pool {
    name                    = "control-plane-pool"
    count                   = 1
    control_plane           = true
    control_plane_as_worker = true

    instance_type {
      cpu          = 4
      disk_size_gb = 60
      memory_mb    = 8000
    }

    placement {
      cluster       = var.vsphere_cluster
      datastore     = var.datastore_name
      network       = var.network_name
      resource_pool = var.resource_pool_name
      # Required for static IP placement.
      static_ip_pool_id = var.deploy-vmware-static ? resource.spectrocloud_privatecloudgateway_ippool.ippool[0].id : null
    }

  }

  machine_pool {
    name          = "worker-pool"
    count         = 1
    control_plane = false

    instance_type {
      cpu          = 4
      disk_size_gb = 60
      memory_mb    = 8000
    }

    placement {
      cluster       = var.vsphere_cluster
      datastore     = var.datastore_name
      network       = var.network_name
      resource_pool = var.resource_pool_name
      # Required for static IP placement.
      static_ip_pool_id = var.deploy-vmware-static ? resource.spectrocloud_privatecloudgateway_ippool.ippool[0].id : null
    }
  }

}
