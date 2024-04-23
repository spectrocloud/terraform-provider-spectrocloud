data "spectrocloud_cloudaccount_gcp" "account" {
  name = var.gcp-cloud-account-name
}

resource "spectrocloud_cluster_gcp" "cluster" {
  name             = "gcp-cluster"
  tags             = ["gcp", "tutorial"]
  cloud_account_id = data.spectrocloud_cloudaccount_gcp.account.id

  cloud_config {
    project = var.gcp-cloud-account-name
    region  = var.region
  }

  cluster_profile {
    id = spectrocloud_cluster_profile.profile.id
  }

  machine_pool {
    control_plane           = true
    control_plane_as_worker = true
    name                    = "cp-pool"
    count                   = var.cp_nodes.count
    instance_type           = var.cp_nodes.instance_type
    disk_size_gb            = var.cp_nodes.disk_size_gb
    azs                     = var.cp_nodes.availability_zones
  }

  machine_pool {
    name          = "worker-basic"
    count         = var.worker_nodes.count
    instance_type = var.worker_nodes.instance_type
    disk_size_gb  = var.worker_nodes.disk_size_gb
    azs           = var.worker_nodes.availability_zones
  }
}
