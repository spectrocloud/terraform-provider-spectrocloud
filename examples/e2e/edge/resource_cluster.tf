resource "spectrocloud_cluster_edge" "cluster" {
  name = "edge-dev"

  cluster_profile {
    id = data.spectrocloud_cluster_profile.profile.id
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

  cloud_config {
    ssh_key = "spectro2022"
  }

  machine_pool {
    control_plane           = true
    control_plane_as_worker = true
    name                    = "master-pool"
    count                   = 1

    placements {
      appliance_id = data.spectrocloud_appliance.virt_appliance.id
    }

  }

  machine_pool {
    name  = "worker-pool"

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

    count = 1

    placements {
      appliance_id = data.spectrocloud_appliance.virt_appliance.id
    }

  }

}
