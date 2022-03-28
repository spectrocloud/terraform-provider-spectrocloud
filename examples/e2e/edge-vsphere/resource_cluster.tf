resource "spectrocloud_cluster_edge_vsphere" "cluster" {
  name = "nikwithcred-mar25"

  edge_host_uid = data.spectrocloud_appliance.virt_appliance.name

  cluster_profile {
    id = data.spectrocloud_cluster_profile.profile.id
  }

  cloud_config {
    ssh_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQCr3hE9IS5UUDPqNOiEWVJvVDS0v57QKjb1o9ubvvATQNg2T3x+inckfzfgX3et1H9X1oSp0FqY1+Mmy5nfTtTyIj5Get1cntcC4QqdZV8Op1tmpI01hYMj4lLn55WNaXgDt+35tJ47kWRr5RqTGV05MPNWN3klaVsePsqa+MgCjnLfCBiOz1tpBOgxqPNqtQPXh+/T/Ul6ZDUW/rySr9iNR9uGd04tYzD7wdTdvmZSRgWEre//IipNzMnnZC7El5KJCQn8ksF+DYY9eT9NtNFEMALTZC6hn8BnMc14zqxoJP/GNHftmig8TJC500Uofdr4OKTCRr1JwHS79Cx9LyZdAp/1D8mL6bIMyGOTPVQ8xUpmEYj77m1kdiCHCk22YtLyfUWuQ0SC+2p1soDoNfJUpmxcKboOTZsLq1HDCFrqSyLUWS1PrYZ/MzhsPrsDewB1iHLbYDt87r2odJOpxMO1vNWMOYontODdr5JPKBpCcd/noNyOy/m4Spntytfb/J3kM1oz3dpPfN0xXmC19uR1xHklmbtg1j784IMu7umI2ZCpUwLADAodkbxmbacdkp5I+1NFgrFamvnTjjQAvRexV31m4m9GielKFQ4tCCId2yagMBWRFn5taEhb3SKnRxBcAzaJLopUyErOtqxvSywGvb53v4MEShqBaQSUv4gHfw== spectro2022"
    static_ip = false
    network_type          = "VIP"
    vip = "192.168.100.15"
    datacenter = var.vsphere_datacenter
    folder     = var.vsphere_folder

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
