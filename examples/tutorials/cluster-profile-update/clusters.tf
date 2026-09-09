#########################
# AWS Cluster Profile
########################
resource "spectrocloud_cluster_aws" "aws-cluster" {
  count = var.deploy-aws ? 1 : 0

  name             = "aws-cluster"
  tags             = concat(var.tags, ["env:aws", "service:hello-universe-frontend"])
  cloud_account_id = data.spectrocloud_cloudaccount_aws.account[0].id

  cloud_config {
    region       = var.aws-region
    ssh_key_name = var.aws-key-pair-name
  }

  cluster_profile {
    id = var.update_profile_version ? spectrocloud_cluster_profile.aws-profile-3tier[0].id : spectrocloud_cluster_profile.aws-profile[0].id
  }

  machine_pool {
    control_plane           = true
    control_plane_as_worker = true
    name                    = "control-pane-pool"
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
    create = "30m"
    delete = "15m"
  }
}

resource "spectrocloud_cluster_aws" "aws-cluster-api" {
  count = var.deploy-aws ? 1 : 0

  name             = "aws-cluster-api"
  tags             = concat(var.tags, ["env:aws", "service:hello-universe-backend"])
  cloud_account_id = data.spectrocloud_cloudaccount_aws.account[0].id

  cloud_config {
    region       = var.aws-region
    ssh_key_name = var.aws-key-pair-name
  }

  cluster_profile {
    id = spectrocloud_cluster_profile.aws-profile-api[0].id
  }

  machine_pool {
    control_plane           = true
    control_plane_as_worker = true
    name                    = "control-pane-pool"
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
    create = "30m"
    delete = "15m"
  }
}

resource "local_file" "aws-kubeconfig" {
  count = var.deploy-aws ? 1 : 0

  content              = data.spectrocloud_cluster.aws_cluster_api[0].kube_config
  filename             = "aws-cluster-api.kubeconfig"
  file_permission      = "0644"
  directory_permission = "0755"
}

#########################
# Azure Cluster Profile
#########################
resource "spectrocloud_cluster_azure" "azure-cluster" {
  count = var.deploy-azure ? 1 : 0

  name             = "azure-cluster"
  tags             = concat(var.tags, ["env:azure", "service:hello-universe-frontend"])
  cloud_account_id = data.spectrocloud_cloudaccount_azure.account[0].id

  cloud_config {
    subscription_id = var.azure_subscription_id
    resource_group  = var.azure_resource_group
    region          = var.azure-region
    ssh_key         = tls_private_key.tutorial_ssh_key[0].public_key_openssh
  }

  cluster_profile {
    id = var.update_profile_version ? spectrocloud_cluster_profile.azure-profile-3tier[0].id : spectrocloud_cluster_profile.azure-profile[0].id
  }

  machine_pool {
    control_plane           = true
    control_plane_as_worker = true
    name                    = "control-pane-pool"
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
    create = "30m"
    delete = "15m"
  }
}

resource "spectrocloud_cluster_azure" "azure-cluster-api" {
  count = var.deploy-azure ? 1 : 0

  name             = "azure-cluster-api"
  tags             = concat(var.tags, ["env:azure", "service:hello-universe-backend"])
  cloud_account_id = data.spectrocloud_cloudaccount_azure.account[0].id

  cloud_config {
    subscription_id = var.azure_subscription_id
    resource_group  = var.azure_resource_group
    region          = var.azure-region
    ssh_key         = tls_private_key.tutorial_ssh_key[0].public_key_openssh
  }

  cluster_profile {
    id = spectrocloud_cluster_profile.azure-profile-api[0].id
  }

  machine_pool {
    control_plane           = true
    control_plane_as_worker = true
    name                    = "control-pane-pool"
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
    create = "30m"
    delete = "15m"
  }
}

resource "local_file" "azure-kubeconfig" {
  count = var.deploy-azure ? 1 : 0

  content              = data.spectrocloud_cluster.azure_cluster_api[0].kube_config
  filename             = "azure-cluster-api.kubeconfig"
  file_permission      = "0644"
  directory_permission = "0755"
}

#########################
# GCP Cluster Profile
#########################
resource "spectrocloud_cluster_gcp" "gcp-cluster" {
  count = var.deploy-gcp ? 1 : 0

  name             = "gcp-cluster"
  tags             = concat(var.tags, ["env:gcp", "service:hello-universe-frontend"])
  cloud_account_id = data.spectrocloud_cloudaccount_gcp.account[0].id

  cloud_config {
    project = var.gcp_project_name
    region  = var.gcp-region
  }

  cluster_profile {
    id = var.update_profile_version ? spectrocloud_cluster_profile.gcp-profile-3tier[0].id : spectrocloud_cluster_profile.gcp-profile[0].id
  }

  machine_pool {
    control_plane           = true
    control_plane_as_worker = true
    name                    = "control-pane-pool"
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
    create = "30m"
    delete = "15m"
  }
}

resource "spectrocloud_cluster_gcp" "gcp-cluster-api" {
  count = var.deploy-gcp ? 1 : 0

  name             = "gcp-cluster-api"
  tags             = concat(var.tags, ["env:gcp", "service:hello-universe-backend"])
  cloud_account_id = data.spectrocloud_cloudaccount_gcp.account[0].id

  cloud_config {
    project = var.gcp_project_name
    region  = var.gcp-region
  }

  cluster_profile {
    id = spectrocloud_cluster_profile.gcp-profile-api[0].id
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
    create = "30m"
    delete = "15m"
  }
}

resource "local_file" "gcp-kubeconfig" {
  count = var.deploy-gcp ? 1 : 0

  content              = data.spectrocloud_cluster.gcp_cluster_api[0].kube_config
  filename             = "gcp-cluster-api.kubeconfig"
  file_permission      = "0644"
  directory_permission = "0755"
}
