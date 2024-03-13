data "spectrocloud_cloudaccount_gcp" "account" {
  # id = <uid>
  name = var.cluster_cloud_account_name
}

data "spectrocloud_cluster_profile" "profile" {
  # id = <uid>
  name = var.cluster_cluster_profile_name
}


resource "spectrocloud_cluster_gke" "cluster" {
  name             = var.cluster_name
  description = "Gke Cluster"
  tags             = ["dev", "department:pax"]
  cloud_account_id = data.spectrocloud_cloudaccount_gcp.account.id
  context = "project"

  cluster_profile {
    id = data.spectrocloud_cluster_profile.profile.id
  }

  cloud_config {
    project = "spectro-common-dev"
    region = "us-central1"
  }
  update_worker_pool_in_parallel = true
  machine_pool {
    name                 = "worker-basic"
    count                = 1
    instance_type        = "Standard_DS4"
    disk_size_gb         = 60
    is_system_node_pool  = true
    storage_account_type = "Standard_LRS"
  }
}
