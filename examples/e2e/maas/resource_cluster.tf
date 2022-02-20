
resource "spectrocloud_cluster_maas" "cluster" {
  name               = "maas-picard-cluster-1"

  cluster_profile {
    id = spectrocloud_cluster_profile.profile.id
  }
  cloud_account_id   = data.spectrocloud_cloudaccount_maas.account.id

  cloud_config {
    domain = var.maas_domain # "maas.sc"
  }

  machine_pool {
    control_plane           = true
    control_plane_as_worker = true
    name                    = "master-pool"
    count                   = 1
    placement {
      resource_pool = var.maas_resource_pool
    }
    instance_type {
      min_memory_mb    = 4096
      min_cpu          = 2
    }

    azs                     = var.maas_azs
  }

  machine_pool {
    name          = "worker-basic"
    count         = 1
    placement {
      resource_pool = var.maas_resource_pool # "Medium-Generic"
    }
    instance_type {
      min_memory_mb    = 4096
      min_cpu          = 2
    }

    azs           = ["az2"]
  }

}
