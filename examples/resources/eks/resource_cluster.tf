resource "spectrocloud_cluster_eks" "cluster" {
  name = "ran-tf-eks"
  context = "tenant"
  tags_map = {"QA" = "ranjithroy@_ .:/=+-@.@123"}

  cluster_profile {
    id = "68a6e0bc500766a5c9241784"
  }

  cloud_account_id = "68a6e0ec788fd02b1e0151a4"


  cloud_config {
    ssh_key_name = var.aws_ssh_key_name
    region       = var.aws_region
    vpc_id       = var.aws_vpc_id
    azs          = ["ap-south-1a", "ap-south-1b"]
  }

  machine_pool {
    name          = "worker-basic"
    count         = 1
    instance_type = "t3.xlarge"
    azs          = ["ap-south-1a", "ap-south-1b"]
    disk_size_gb  = 60
  }

  machine_pool {
    name          = "worker-basic-2"
    count         = 1
    instance_type = "t3.xlarge"
    azs          = ["ap-south-1a", "ap-south-1b"]
    disk_size_gb  = 60
  }

}
