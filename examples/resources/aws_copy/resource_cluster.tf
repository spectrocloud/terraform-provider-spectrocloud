resource "spectrocloud_cluster_aws" "cluster" {
  name = "mani-aws-picard-1"
  context = "tenant"
  cluster_profile {
    id = spectrocloud_cluster_profile.profile.id
  }
  cloud_account_id = spectrocloud_cloudaccount_aws.account.id

  cloud_config {
    ssh_key_name = var.aws_ssh_key_name
    region       = var.aws_region
    #vpc_id       = "vpc-0989f5a78d60bdb54"
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
/*
  cluster_rbac_binding {
    type = "ClusterRoleBinding"

    role = {
      kind = "ClusterRole"
      name = "testRole3"
    }
    subjects {
      type = "User"
      name = "testRoleUser3"
    }
    subjects {
      type = "Group"
      name = "testRoleGroup3"
    }
    subjects {
      type      = "ServiceAccount"
      name      = "testrolesubject3"
      namespace = "testrolenamespace"
    }
  }

  namespaces {
    name = "test6ns"
    resource_allocation = {
      cpu_cores  = "2"
      memory_MiB = "2048"
    }
  }

  cluster_rbac_binding {
    type      = "RoleBinding"
    namespace = "test6ns"
    role = {
      kind = "Role"
      name = "testRoleFromNS3"
    }
    subjects {
      type = "User"
      name = "testUserRoleFromNS3"
    }
    subjects {
      type = "Group"
      name = "testGroupFromNS3"
    }
    subjects {
      type      = "ServiceAccount"
      name      = "testrolesubject3"
      namespace = "testrolenamespace"
    }
  }*/

  machine_pool {
    control_plane           = true
    control_plane_as_worker = true
    name                    = "master-pool"
    count                   = 1
    instance_type           = "t3.xlarge"
    disk_size_gb            = 62
    azs                     = [var.aws_region_az]
    #az_subnets    = var.master_azs_subnets_map != {} ? var.master_azs_subnets_map : null
    #additional_security_groups = ["sg-051e367608382537c","sg-0bb4b30ceab2091f3"]
  }

  machine_pool {
    name          = "worker-basic"
    count         = 1
    instance_type = "t3.medium"
    azs           = [var.aws_region_az]
    node_repave_interval  = 5
    #az_subnets    = var.worker_azs_subnets_map != {} ? var.worker_azs_subnets_map : null
    #min           = 2
    #max           = 2
  }

  /*machine_pool {
    name          = "worker-basic-1"
    count         = 1
    instance_type = "t3.medium"
    azs           = [var.aws_region_az]
    az_subnets    = var.worker_azs_subnets_map != {} ? var.worker_azs_subnets_map : null
    additional_security_groups = ["sg-051e367608382537c","sg-0bb4b30ceab2091f3"]
  }*/

}
