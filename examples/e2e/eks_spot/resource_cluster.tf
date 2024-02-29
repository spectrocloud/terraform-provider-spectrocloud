resource "spectrocloud_cluster_eks" "cluster" {
  name = "eks-dev1"

  cluster_profile {
    id = spectrocloud_cluster_profile.profile.id
  }

  cluster_profile {
    id = spectrocloud_cluster_profile.profile-rbac.id
    pack {
      name   = "spectro-rbac"
      tag    = "1.0.0"
      values = file("rbac.yaml")
    }
  }

  cloud_account_id = spectrocloud_cloudaccount_aws.account.id

  cloud_config {
    ssh_key_name = var.aws_ssh_key_name
    region       = var.aws_region
    vpc_id       = var.aws_vpc_id
    azs          = var.azs != [] ? var.azs : null
    az_subnets   = var.master_azs_subnets_map != {} ? var.master_azs_subnets_map : null
  }

  machine_pool {
    name          = "worker-basic"
    count         = 1
    instance_type = "m3.large"
    capacity_type = "spot"
    max_price     = "0.015"
    azs           = var.azs != [] ? var.azs : null
    az_subnets    = var.master_azs_subnets_map != {} ? var.master_azs_subnets_map : null
    disk_size_gb  = 60
  }

  fargate_profile {
    name    = "fg-1"
    subnets = values(var.worker_azs_subnets_map)
    additional_tags = {
      hello = "yo"
    }
    selector {
      namespace = "fargate"
      labels = {
        abc = "cool"
      }

    }
  }

  # Custom timeouts for each CRUD operation
  #timeouts {
  #  create = "120m"
  #  update = "120m"
  #  delete = "120m"
  #}

}
