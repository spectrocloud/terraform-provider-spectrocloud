data "spectrocloud_backup_storage_location" "bsl" {
  name = "shru-gcp-bck"
}

resource "spectrocloud_cluster_maas" "cluster" {
  name = "maas-picard-cluster-1"
  #tags = ["manikumar", "test"]
  context    = "project"

  cluster_meta_attribute = "{'clusterconfig':'true'}"
  cluster_profile {
    id = spectrocloud_cluster_profile.profile.id
  }
  cloud_account_id = data.spectrocloud_cloudaccount_maas.account.id

  cloud_config {
    domain = var.maas_domain # "maas.sc"
  }

  os_patch_on_boot    = false
  /*cluster_profile {
    id = "652e93ec369d9216acc04305"
  }*/

  /*cluster_rbac_binding {
    type = "ClusterRoleBinding"

    role = {
      kind = "ClusterRole"
      name = "testRole4"
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
    namespace = "test5ns"
    role = {
      kind = "Role"
      name = "testRoleFromNS6"
    }
    subjects {
      type = "User"
      name = "testUserRoleFromNS"
    }
    subjects {
      type = "Group"
      name = "testGroupFromNS"
    }
    subjects {
      type      = "ServiceAccount"
      name      = "testrolesubject"
      namespace = "testrolenamespace"
    }
  }*/

  /*backup_policy {
    schedule                  = ""
    backup_location_id        = data.spectrocloud_backup_storage_location.bsl.id
    prefix                    = "demo001-dmitry-backup"
    expiry_in_hour            = 7400
    include_disks             = true
    include_cluster_resources = true
  }*/

  /*scan_policy {
    configuration_scan_schedule = "0 0 * * MON"
    penetration_scan_schedule   = "0 0 * * SUN"
    conformance_scan_schedule   = "0 0 1 * *"
  }*/

  machine_pool {
    control_plane           = true
    control_plane_as_worker = false
    name                    = "master-pool"
    count                   = 1
    /*placement {
      resource_pool = var.maas_resource_pool
    }*/
    instance_type {
      min_memory_mb = 8192
      min_cpu       = 4
    }
    node_tags = ["sh-tag"]
    #azs = var.maas_azs
  }

  machine_pool {
    name  = "worker-basic"
    count = 1
   /* min   = 1
    max   = 3*/
    #additional_labels = {"app": "testpod"}
    /*placement {
      resource_pool = var.maas_worker_resource_pool # "Medium-Generic"
    }*/
    instance_type {
      min_memory_mb = 8192
      min_cpu       = 4
    }
    node_tags = ["sh-tag"]
    # azs = var.maas_worker_azs
    /*node {
      node_id = "653144352762558c95a11bfb-s66deg-3933011753580794236"
      action = "uncordon" #uncordon
    }*/
  }

}
