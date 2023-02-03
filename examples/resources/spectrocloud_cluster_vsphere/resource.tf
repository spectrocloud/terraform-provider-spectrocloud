data "spectrocloud_cluster_profile" "vmware_profile"{
  name = "vmware-public-repo"
  version = "2.0.0"
  context = "tenant"
}
data "spectrocloud_cloudaccount_vsphere" "vmware_account"{
  name = "gm-pcg-wop-d"
}


resource "spectrocloud_cluster_vsphere" "cluster" {
  name               = "vsphere-picard-2"
  cluster_profile_id = data.spectrocloud_cluster_profile.vmware_profile.id
  cloud_account_id   = data.spectrocloud_cloudaccount_vsphere.vmware_account.id

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
    ssh_key = var.cluster_ssh_public_key

    datacenter = var.vsphere_datacenter
    folder     = var.vsphere_folder
    // For Dynamic DNS (network_type & network_search_domain value should set for DDNS)
    network_type          = "DDNS"
    network_search_domain = var.cluster_network_search
    // For Static (By Default static_ip is false, for static provisioning, it is set to be true. Not required to specify network_type & network_search_domain)
    # static_ip = true
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
    name  = "worker-basic"
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