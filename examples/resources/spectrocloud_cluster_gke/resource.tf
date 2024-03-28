data "spectrocloud_cloudaccount_gcp" "account" {
  name = var.gcp_cloud_account_name
}

data "spectrocloud_cluster_profile" "profile" {
  name = var.gke_cluster_profile_name
}


resource "spectrocloud_cluster_gke" "cluster" {
  name             = var.cluster_name
  description      = "Gke Cluster"
  tags             = ["dev", "department:pax"]
  cloud_account_id = data.spectrocloud_cloudaccount_gcp.account.id
  context          = "project"

  cluster_profile {
    id = data.spectrocloud_cluster_profile.profile.id
  }

  cloud_config {
    project = var.gcp_project
    region  = var.gcp_region
  }
  update_worker_pool_in_parallel = true
  machine_pool {
    name          = "worker-basic"
    count         = 3
    instance_type = "n2-standard-4"
  }
}
