resource "spectrocloud_cluster_libvirt" "cluster" {
  name = "virt-nik"

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
    vip     = "192.168.100.15"
  }

  machine_pool {
    control_plane           = true
    control_plane_as_worker = true
    name                    = "cp-pool"
    count                   = 1

    placements {
      appliance_id        = data.spectrocloud_appliance.virt_appliance.id
      network_type        = "bridge"
      network_names       = "br0"
      image_storage_pool  = "ubuntu"
      target_storage_pool = "guest_images"
      data_storage_pool   = "tmp"
      network             = "br"
    }

    instance_type {
      disk_size_gb = 30
      memory_mb    = 8096
      cpu          = 4
      cpus_sets    = 1

      attached_disks {
        size_in_gb = "10"
      }
    }
  }

  machine_pool {
    name  = "worker-pool"
    count = 1

    placements {
      appliance_id        = data.spectrocloud_appliance.virt_appliance.id
      network_type        = "bridge"
      network_names       = "br0"
      image_storage_pool  = "ubuntu"
      target_storage_pool = "guest_images"
      data_storage_pool   = "tmp"
      network             = "br"
    }

    instance_type {
      disk_size_gb = 30
      memory_mb    = 8096
      cpu          = 2
      cpus_sets    = 1

      attached_disks {
        size_in_gb = "30"
        managed    = true
      }

      attached_disks {
        size_in_gb = "10"
        managed    = true
      }

    }
  }

}
