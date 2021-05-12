resource "spectrocloud_cluster_eks" "cluster" {
  name               = "eks-dev1"
  cloud_account_id   = spectrocloud_cloudaccount_aws.account.id

  cloud_config {
    ssh_key_name = var.aws_ssh_key_name
    region       = var.aws_region
    vpc_id       = var.aws_vpc_id
    az_subnets   = var.master_azs_subnets_map
  }

  cluster_profile {
    cluster_profile_id = spectrocloud_cluster_profile.profile.id
  }

  machine_pool {
    name          = "worker-basic"
    count         = 3
    instance_type = "t3.large"
    az_subnets    = var.worker_azs_subnets_map
    disk_size_gb  = 60
  }
}
