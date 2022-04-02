resource "spectrocloud_cluster_eks" "cluster" {
  name = "eks-dev-taints"

  cluster_profile {
    id = spectrocloud_cluster_profile.profile.id
  }

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
    name = "test5ns"
    resource_allocation = {
      cpu_cores  = "2"
      memory_MiB = "2048"
    }
  }

  cluster_rbac_binding {
    type      = "RoleBinding"
    namespace = "test5ns"
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
  }

  cloud_account_id = spectrocloud_cloudaccount_aws.account.id

  cloud_config {
    endpoint_access = "private"
    ssh_key_name    = var.aws_ssh_key_name
    region          = var.aws_region
    vpc_id          = var.aws_vpc_id
    az_subnets      = var.master_azs_subnets_map
  }

  machine_pool {
    name          = "worker-basic"
    count         = 1
    instance_type = "t3.large"
    az_subnets    = var.worker_azs_subnets_map
    disk_size_gb  = 30

    additional_labels = {
      addlabel = "addlabelval1"
    }

    taints {
      key    = "taintkey1"
      value  = "taintvalue1"
      effect = "PreferNoSchedule"
    }

    taints {
      key    = "taintkey2"
      value  = "taintvalue2"
      effect = "NoSchedule"
    }

  }
}
