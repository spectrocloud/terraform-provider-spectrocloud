resource "spectrocloud_cluster_edge_vsphere" "cluster" {
  name = "nikwithcred-mar25"

  edge_host_uid = data.spectrocloud_appliance.virt_appliance.name

  cluster_profile {
    id = data.spectrocloud_cluster_profile.profile.id
  }

  cloud_config {
    ssh_key      = var.cluster_ssh_public_key
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
