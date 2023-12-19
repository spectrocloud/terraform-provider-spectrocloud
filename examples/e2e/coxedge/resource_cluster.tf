resource "spectrocloud_cluster_coxedge" "cluster" {
  name             = "tf-coxedge-cluster01"
  cloud_account_id = data.spectrocloud_cloudaccount_coxedge.account.id

  cluster_profile {
    id = spectrocloud_cluster_profile.profile.id
  }

  cloud_config {
    environment     = "dev"
    organization_id = "abcd-efg-hij-klm"
    ssh_keys = [
      "ssh-rsa ",
    ]

    lb_config {
      pops = [
        "LAS",
      ]
    }

    worker_lb {
      pops = [
        "LAS",
      ]
    }
  }
  machine_pool {
    control_plane           = true
    control_plane_as_worker = true
    name                    = "master-pool"
    count                   = 1
    cox_config {
      spec = "SP-4"
    }
  }

  machine_pool {
    name  = "worker-basic"
    count = 1
    cox_config {
      spec = "SP-4"
    }
  }
}
