data "spectrocloud_cluster_profile" "profile" {
  name = "tfmod-tke-prof-infra-yx4n4"
  version = "2.2.2"
}

data "spectrocloud_cloudaccount_tencent" "tke_account"{
  name = "tf-tke-account-ukagc"
}

resource "spectrocloud_cluster_tke" "cluster" {
  name = "tke-demo-tf"

  cluster_profile {
    id = data.spectrocloud_cluster_profile.profile.id
  }

  cluster_rbac_binding {
    type = "ClusterRoleBinding"

    role = {
      kind = "ClusterRole"
      name = "testrole3"
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

  cloud_account_id = data.spectrocloud_cloudaccount_tencent.tke_account.id

  cloud_config {
    endpoint_access     = "public"
    public_access_cidrs = ["0.0.0.0/0"]
    ssh_key_name        = var.tke_ssh_key_name
    region              = var.tke_region
    vpc_id              = var.tke_vpc_id
    az_subnets          = var.master_tke_subnets_map
  }

  machine_pool {
    name          = "worker-basic"
    count         = 1
    min           = 1
    max           = 1
    instance_type = "S3.MEDIUM4"
    az_subnets    = var.worker_tke_subnets_map
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
