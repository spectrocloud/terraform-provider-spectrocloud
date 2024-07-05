resource "spectrocloud_cluster_vsphere" "cluster" {
  name             = "vsphere-static2"
  cloud_account_id = spectrocloud_cloudaccount_vsphere.account.id

  cluster_profile {
    id = data.spectrocloud_cluster_profile.profile.id
  }

  cloud_config {
    ssh_key               = var.cluster_ssh_public_key
    image_template_folder = var.vsphere_image_template_folder

    datacenter = var.vsphere_datacenter
    folder     = var.vsphere_folder

    static_ip = true
    #network_type          = "DDNS"
    network_search_domain = var.cluster_network_search
  }


  machine_pool {
    control_plane           = true
    control_plane_as_worker = true
    name                    = "cp-pool"
    count                   = 3

    placement {
      cluster           = var.vsphere_cluster
      resource_pool     = var.vsphere_resource_pool
      datastore         = var.vsphere_datastore
      network           = var.vsphere_network
      static_ip_pool_id = data.spectrocloud_ippool.ippool.id
    }
    instance_type {
      disk_size_gb = 40
      memory_mb    = 8192
      cpu          = 4
    }
  }

}
