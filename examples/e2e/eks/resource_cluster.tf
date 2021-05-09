
resource "spectrocloud_cluster_eks" "cluster" {
  name               = "eks-dev"
  cluster_profile_id = spectrocloud_cluster_profile.profile.id
  cloud_account_id   = spectrocloud_cloudaccount_aws.account.id

  cloud_config {
    ssh_key_name = var.aws_ssh_key_name
    region       = var.aws_region
    vpc_id       = var.aws_vpc_id
    az_subnets   = var.master_azs_subnets_map
  }

  # To override or specify values for a cluster:

  # pack {
  #   name   = "spectro-byo-manifest"
  #   tag    = "1.0.x"
  #   values = <<-EOT
  #     manifests:
  #       byo-manifest:
  #         contents: |
  #           # Add manifests here
  #           apiVersion: v1
  #           kind: Namespace
  #           metadata:
  #             labels:
  #               app: wordpress
  #               app2: wordpress2
  #             name: wordpress
  #   EOT
  # }

  # machine_pool {
  #   control_plane = true
  #   name          = "master-pool"
  #   count         = 1
  #   instance_type = "t3.large"
  #   disk_size_gb  = 62
  #   az_subnets    = var.master_azs_subnets_map
  # }

  machine_pool {
    name          = "worker-basic"
    count         = 1
    instance_type = "t3.large"
    az_subnets    = var.worker_azs_subnets_map
    disk_size_gb = 60
  }
}
