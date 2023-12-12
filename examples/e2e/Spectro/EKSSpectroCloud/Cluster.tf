

resource "spectrocloud_cluster_eks" "cluster" {
  name             = lower("${lookup(var.taggingstandard, "deployment")}-EKS")
  cloud_account_id = data.spectrocloud_cloudaccount_aws.account.id

  cloud_config {
    region       = var.aws_region
    ssh_key_name = ""
    az_subnets   = var.aws_subnets
    vpc_id       = var.aws_vpc_main_id
  }

  dynamic "cluster_profile" {
    for_each = data.spectrocloud_cluster_profile.profile
    content {
      id = cluster_profile.value["id"]
    }
  }

  dynamic "machine_pool" {
    for_each = var.SpectroConfig

    content {
      name          = lower(lookup(var.SpectroConfig[machine_pool.key], "WorkerPoolName"))
      count         = lookup(var.SpectroConfig[machine_pool.key], "WorkerPoolCount")
      instance_type = lookup(var.SpectroConfig[machine_pool.key], "WorkerPoolInstanceType")
      disk_size_gb  = lookup(var.SpectroConfig[machine_pool.key], "WorkerPoolDiskSize")
      az_subnets    = var.aws_subnets
    }
  }
}