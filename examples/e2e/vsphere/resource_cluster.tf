## The following example provides a vSphere cluster with a cluster profile and two node pools: one master pool and one worker pool. Each node has 8 CPUs, 8Gb RAM, and 60Gb disk

resource "spectrocloud_cluster_vsphere" "cluster" {
  name = "vsphere-cluster-1"
  cluster_profile {
    id = spectrocloud_cluster_profile.profile.id
  }
  cloud_account_id = data.spectrocloud_cloudaccount_vsphere.account.id

  cloud_config {
    ssh_key = var.cluster_ssh_public_key

    datacenter = var.vsphere_datacenter
    folder     = var.vsphere_folder

    network_type          = "DDNS"
    network_search_domain = var.cluster_network_search
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
      disk_size_gb = 60
      memory_mb    = 8192
      cpu          = 8
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
      disk_size_gb = 60
      memory_mb    = 8192
      cpu          = 8
    }
  }
}
