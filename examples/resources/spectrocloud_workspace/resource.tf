data "spectrocloud_cluster" cluster1 {
  name = "vsphere-picard-2"
}

resource "spectrocloud_workspace" "workspace" {
  name = "wsp-tf"

  clusters {
    name = data.spectrocloud_cluster.cluster1.name
    uid = data.spectrocloud_cluster.cluster1.id
  }

  cluster_rbac_binding {
    type = "ClusterRoleBinding"

    role = {
      kind = "ClusterRole"
      name = "testrole3"
    }
    subjects {
      type = "User"
      name = "testRoleUser4"
    }
    subjects {
      type = "Group"
      name = "testRoleGroup4"
    }
    subjects {
      type      = "ServiceAccount"
      name      = "testrolesubject3"
      namespace = "testrolenamespace"
    }
  }

  cluster_rbac_binding {
    type      = "RoleBinding"
    namespace = "test5ns"
    role = {
      kind = "Role"
      name = "testrolefromns3"
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

  namespaces {
    name = "test5ns"
    resource_allocation = {
      cpu_cores  = "2"
      memory_MiB = "2048"
    }

    images_blacklist = ["1", "2", "3"]
  }

  backup_policy {
    schedule                  = "0 0 * * SUN"
    backup_location_id        = data.spectrocloud_backup_storage_location.bsl.id
    prefix                    = "prod-backup"
    expiry_in_hour            = 7200
    include_disks             = false
    include_cluster_resources = true

    //namespaces = ["test5ns"]
    include_all_clusters = true
    cluster_uids = [data.spectrocloud_cluster.cluster1.id]
  }

}

data "spectrocloud_backup_storage_location" bsl {
  name = "backups-nikolay"
}