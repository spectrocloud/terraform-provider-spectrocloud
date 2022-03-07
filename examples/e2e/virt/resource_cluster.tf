resource "spectrocloud_cluster_libvirt" "cluster" {
  name = "virt-nik-new"

  cluster_profile {
    id = data.spectrocloud_cluster_profile.profile.id
  }

  cloud_config {
    ssh_key = "spectro2022"
    vip = "10.11.130.19"
  }

  machine_pool {
    control_plane           = true
    control_plane_as_worker = true
    name                    = "master-pool"
    count                   = 1

    placements {
      appliance_id = data.spectrocloud_appliance.virt_appliance.id
      network_type = "bridge"
      network_names = "br0"
      image_storage_pool = "ehl_images"
      target_storage_pool = "ehl_images"
      data_storage_pool = "ehl_data"
      network = "br"
    }

    instance_type {
      disk_size_gb    = 30
      memory_mb = 8096
      cpu          = 4
      cpus_sets = 1
      attached_disks_size_gb = "30, 10"
    }
  }

  machine_pool {
    name                    = "worker-pool"
    count                   = 1

    placements {
      appliance_id = data.spectrocloud_appliance.virt_appliance.id
      network_type = "bridge"
      network_names = "br0"
      image_storage_pool = "ehl_images"
      target_storage_pool = "ehl_images"
      data_storage_pool = "ehl_data"
      network = "br"
    }

    instance_type {
      disk_size_gb    = 30
      memory_mb = 8096
      cpu          = 2
      cpus_sets = 1
    }
  }

}
