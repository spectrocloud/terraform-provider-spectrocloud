#############
# AWS Cluster
#############

resource "spectrocloud_cluster_aws" "dev_cluster" {
  count = var.deploy-aws ? 1 : 0

  name             = "tf-dev-cluster"
  cluster_timezone = "UTC"
  cloud_account_id = data.spectrocloud_cloudaccount_aws.account[0].id

  cloud_config {
    region       = var.aws-region
    ssh_key_name = var.aws-key-pair-name
  }

  cluster_template {
    id = spectrocloud_cluster_config_template.aws_template[0].id

    cluster_profile {
      id = spectrocloud_cluster_profile.aws_profile[0].id
      variables = {
        "app_replicas" = "1"
      }
    }
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
    create = "60m"
    delete = "60m"
  }
}

##################
# AWS Prod Cluster
##################

resource "spectrocloud_cluster_aws" "prod_cluster" {
  count = var.deploy-aws ? 1 : 0

  name             = "tf-prod-cluster"
  cluster_timezone = "UTC"
  cloud_account_id = data.spectrocloud_cloudaccount_aws.account[0].id

  cloud_config {
    region       = var.aws-region
    ssh_key_name = var.aws-key-pair-name
  }

  cluster_template {
    id = spectrocloud_cluster_config_template.aws_template[0].id

    cluster_profile {
      id = spectrocloud_cluster_profile.aws_profile[0].id
      variables = {
        "app_replicas" = "2"
      }
    }
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
    create = "60m"
    delete = "60m"
  }
}

###############
# Azure Cluster
###############

resource "spectrocloud_cluster_azure" "dev_cluster" {
  count = var.deploy-azure ? 1 : 0

  name             = "tf-dev-cluster-azure"
  cluster_timezone = "UTC"
  cloud_account_id = data.spectrocloud_cloudaccount_azure.account[0].id

  cloud_config {
    subscription_id = var.azure_subscription_id
    resource_group  = var.azure_resource_group
    region          = var.azure-region
    ssh_key         = tls_private_key.tutorial_ssh_key_azure[0].public_key_openssh
  }

  cluster_template {
    id = spectrocloud_cluster_config_template.azure_template[0].id

    cluster_profile {
      id = spectrocloud_cluster_profile.azure_profile[0].id
      variables = {
        "app_replicas" = "1"
      }
    }
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
    name                = "worker-pool"
    count               = var.azure_worker_nodes.count
    instance_type       = var.azure_worker_nodes.instance_type
    azs                 = var.azure-use-azs ? var.azure_worker_nodes.azs : [""]
    is_system_node_pool = var.azure_worker_nodes.is_system_node_pool
  }

  timeouts {
    create = "60m"
    delete = "60m"
  }
}

####################
# Azure Prod Cluster
####################

resource "spectrocloud_cluster_azure" "prod_cluster" {
  count = var.deploy-azure ? 1 : 0

  name             = "tf-prod-cluster-azure"
  cluster_timezone = "UTC"
  cloud_account_id = data.spectrocloud_cloudaccount_azure.account[0].id

  cloud_config {
    subscription_id = var.azure_subscription_id
    resource_group  = var.azure_resource_group
    region          = var.azure-region
    ssh_key         = tls_private_key.tutorial_ssh_key_azure[0].public_key_openssh
  }

  cluster_template {
    id = spectrocloud_cluster_config_template.azure_template[0].id

    cluster_profile {
      id = spectrocloud_cluster_profile.azure_profile[0].id
      variables = {
        "app_replicas" = "2"
      }
    }
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
    name                = "worker-pool"
    count               = var.azure_worker_nodes.count
    instance_type       = var.azure_worker_nodes.instance_type
    azs                 = var.azure-use-azs ? var.azure_worker_nodes.azs : [""]
    is_system_node_pool = var.azure_worker_nodes.is_system_node_pool
  }

  timeouts {
    create = "60m"
    delete = "60m"
  }
}
