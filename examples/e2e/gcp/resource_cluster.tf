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
    name                    = "master-pool"
    count                   = var.master_nodes.count
    instance_type           = var.master_nodes.instance_type
    disk_size_gb            = var.master_nodes.disk_size_gb
    azs                     = var.master_nodes.availability_zones
  }

  machine_pool {
    name          = "worker-basic"
    count         = var.worker_nodes.count
    instance_type = var.worker_nodes.instance_type
    disk_size_gb  = var.worker_nodes.disk_size_gb
    azs           = var.worker_nodes.availability_zones
  }

  # Custom timeouts for each CRUD operation
  #timeouts {
  #  create = "120m"
  #  update = "120m"
  #  delete = "120m"
  #}

}
