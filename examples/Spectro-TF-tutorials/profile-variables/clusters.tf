##############
# AWS Cluster
##############
resource "spectrocloud_cluster_aws" "aws-cluster" {
  count = var.deploy-aws ? 1 : 0

  name             = "aws-profile-var-tf" #Enter your unique AWS Cluster name
  tags             = concat(var.tags, ["env:aws"])
  cloud_account_id = data.spectrocloud_cloudaccount_aws.account[0].id

  cloud_config {
    region       = var.aws-region
    ssh_key_name = var.aws-key-pair-name
  }

  cluster_profile {
    id = var.deploy-aws && var.deploy-aws-var ? resource.spectrocloud_cluster_profile.aws-profile-var[0].id : resource.spectrocloud_cluster_profile.aws-profile[0].id
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

  name             = "azure-cluster-var-tf"
  tags             = concat(var.tags, ["env:azure"])
  cloud_account_id = data.spectrocloud_cloudaccount_azure.account[0].id

  cloud_config {
    subscription_id = var.azure_subscription_id
    resource_group  = var.azure_resource_group
    region          = var.azure-region
    ssh_key         = tls_private_key.tutorial_ssh_key_azure[0].public_key_openssh
  }

  cluster_profile {
    id = var.deploy-azure && var.deploy-azure-var ? resource.spectrocloud_cluster_profile.azure-profile-var[0].id : resource.spectrocloud_cluster_profile.azure-profile[0].id
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

  name             = "gcp-cluster-var-tf"
  tags             = concat(var.tags, ["env:gcp"])
  cloud_account_id = data.spectrocloud_cloudaccount_gcp.account[0].id

  cloud_config {
    project = var.gcp_project_name
    region  = var.gcp-region
  }

  cluster_profile {
    id = var.deploy-gcp && var.deploy-gcp-var ? resource.spectrocloud_cluster_profile.gcp-profile-var[0].id : resource.spectrocloud_cluster_profile.gcp-profile[0].id
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
