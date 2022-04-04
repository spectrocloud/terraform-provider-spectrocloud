resource "spectrocloud_cluster_libvirt" "cluster" {
  name = "virt-nik"

  cluster_profile {
    id = data.spectrocloud_cluster_profile.profile.id
  }

  cloud_config {
    ssh_key = "spectro2022"
    vip     = "192.168.100.15"
  }

  machine_pool {
    control_plane           = true
    control_plane_as_worker = true
    name                    = "master-pool"
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
      disk_size_gb           = 30
      memory_mb              = 8096
      cpu                    = 4
      cpus_sets              = 1

      attached_disks_size_gb {
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

      attached_disks_size_gb {
        size_in_gb = "30"
        managed = true
      }

      attached_disks_size_gb {
        size_in_gb = "10"
        managed = true
      }

    }
  }

}
