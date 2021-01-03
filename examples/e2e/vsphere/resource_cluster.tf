resource "spectrocloud_cluster_vsphere" "cluster" {
  name               = "vsphere-picard-2"
  cluster_profile_id = spectrocloud_cluster_profile.profile.id
  cloud_account_id   = data.spectrocloud_cloudaccount_vsphere.account.id

  cloud_config {
    ssh_key = var.cluster_ssh_public_key

    datacenter = var.vsphere_datacenter
    folder     = var.vsphere_folder

    network_type          = "DDNS"
    network_search_domain = var.cluster_network_search
  }

  # To override or specify values for a cluster:

  # pack {
  #   name   = "spectro-byo-manifest"
  #   tag    = "1.0.x"
  #   values = <<-EOT
  #     manifests:
  #       byo-manifest:
  #         contents: |
  #           # Add manifests here
  #           apiVersion: v1
  #           kind: Namespace
  #           metadata:
  #             labels:
  #               app: wordpress
  #               app2: wordpress2
  #             name: wordpress
  #   EOT
  # }

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
      disk_size_gb = 61
      memory_mb    = 4096
      cpu          = 2
    }
  }

  machine_pool {
    name  = "worker-basic"
    count = 2

    placement {
      cluster       = var.vsphere_cluster
      resource_pool = var.vsphere_resource_pool
      datastore     = var.vsphere_datastore
      network       = var.vsphere_network
    }
    instance_type {
      disk_size_gb = 65
      memory_mb    = 8192
      cpu          = 4
    }
  }
}
