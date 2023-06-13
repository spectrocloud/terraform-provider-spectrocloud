resource "spectrocloud_cluster_edge_vsphere" "cluster" {
  name = "nikwithcred-mar25"

  edge_host_uid = data.spectrocloud_appliance.virt_appliance.name

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
    ssh_key      = var.cluster_ssh_public_key
    # For Multiple ssh_keys
    # ssh_keys = ["ssh key1", "ssh key2"]
    static_ip    = false
    network_type = "VIP"
    vip          = "192.168.100.15"
    datacenter   = var.vsphere_datacenter
    folder       = var.vsphere_folder

  }

  machine_pool {
    control_plane           = true
    control_plane_as_worker = true
    name                    = "master-pool"
    count                   = 1

    placement {
      cluster       = var.vsphere_cluster
      resource_pool = var.vsphere_resource_pool
      datastore     = var.vsphere_datastore
      network       = var.vsphere_network
    }
    instance_type {
      disk_size_gb = 40
      memory_mb    = 4096
      cpu          = 2
    }
  }

  machine_pool {
    name = "worker-basic"

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

    placement {
      cluster       = var.vsphere_cluster
      resource_pool = var.vsphere_resource_pool
      datastore     = var.vsphere_datastore
      network       = var.vsphere_network
    }
    instance_type {
      disk_size_gb = 40
      memory_mb    = 8192
      cpu          = 4
    }
  }

}
