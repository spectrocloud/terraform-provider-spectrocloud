
resource "spectrocloud_cluster_eks" "cluster" {
  name               = "eks-tf-dev"
  cluster_profile_id = spectrocloud_cluster_profile.profile.id
  cloud_account_id   = spectrocloud_cloudaccount_aws.account.id

  cloud_config {
    ssh_key_name = var.aws_ssh_key_name
    region       = var.aws_region
    vpc_id = "vpc-0e03ff84a894d40a2"
    endpoint_access = "public"

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

  machine_pool {
    control_plane           = true
    control_plane_as_worker = true
    name                    = "master-pool"
    count                   = 1
    instance_type           = "t3.large"
    disk_size_gb            = 62
    azs                     = ["us-west-2a", "us-west-2b"]
    subnets = {
      "us-west-2a" = "subnet-0d4978ddbff16c868",
      "us-west-2b" = "subnet-041a35c9c06eeb701"
    }
  }

  machine_pool {
    name          = "worker-basic"
    count         = 1
    instance_type = "t3.large"
    azs           = ["us-west-2a", "us-west-2b"]
    subnets = {
      "us-west-2a" = "subnet-0d4978ddbff16c868",
      "us-west-2b" = "subnet-041a35c9c06eeb701"
    }
  }
}
