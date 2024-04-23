resource "spectrocloud_cluster_eks" "cluster" {
  name = "eks-dev2"

  cluster_profile {
    id = spectrocloud_cluster_profile.profile.id
  }

  cluster_rbac_binding {
    type = "ClusterRoleBinding"

    role = {
      kind = "ClusterRole"
      name = "testRole2"
    }
    subjects {
      type = "User"
      name = "testRoleUser2"
    }
    subjects {
      type = "Group"
      name = "testRoleGroup2"
    }
    subjects {
      type      = "ServiceAccount"
      name      = "testrolesubject2"
      namespace = "testrolenamespace"
    }
  }

  namespaces {
    name = "test2ns"
    resource_allocation = {
      cpu_cores  = "2"
      memory_MiB = "2048"
    }
  }

  cluster_rbac_binding {
    type      = "RoleBinding"
    namespace = "test2ns"
    role = {
      kind = "Role"
      name = "testRoleFromNS2"
    }
    subjects {
      type = "User"
      name = "testUserRoleFromNS2"
    }
    subjects {
      type = "Group"
      name = "testGroupFromNS2"
    }
    subjects {
      type      = "ServiceAccount"
      name      = "testrolesubject2"
      namespace = "testrolenamespace"
    }
  }

  cloud_account_id = spectrocloud_cloudaccount_aws.account.id

  cloud_config {
    endpoint_access = "private"
    ssh_key_name    = var.aws_ssh_key_name
    region          = var.aws_region
    vpc_id          = var.aws_vpc_id
    azs             = var.azs != [] ? var.azs : null
    az_subnets      = var.cp_azs_subnets_map != {} ? var.cp_azs_subnets_map : null
  }

  machine_pool {
    name          = "worker-basic"
    count         = 1
    min           = 1
    max           = 3
    instance_type = "t3.large"
    azs           = var.azs != [] ? var.azs : null
    az_subnets    = var.cp_azs_subnets_map != {} ? var.cp_azs_subnets_map : null
    disk_size_gb  = 30
  }
}
