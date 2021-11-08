
resource "spectrocloud_cluster_maas" "cluster" {
  name               = "maas-picard-cluster"
  cluster_profile_id = spectrocloud_cluster_profile.profile.id
  cloud_account_id   = data.spectrocloud_cloudaccount_maas.account.id

  cloud_config {
    domain = var.maas_domain # "maas.sc"
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
      resource_pool = var.maas_resource_pool
    }
    instance_type {
      disk_size_gb = 61
      memory_mb    = 4096
      cpu          = 2
    }

    azs                     = [var.maas_region_az]
  }

  machine_pool {
    name          = "worker-basic"
    count         = 1
    placement {
      resource_pool = var.maas_resource_pool
    }
    instance_type {
      disk_size_gb = 61
      memory_mb    = 4096
      cpu          = 2
    }

    azs           = [var.maas_region_az]
  }

}
