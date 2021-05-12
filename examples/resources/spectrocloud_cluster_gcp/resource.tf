data "spectrocloud_cloudaccount_gcp" "account" {
  # id = <uid>
  name = var.cluster_cloud_account_name
}

data "spectrocloud_cluster_profile" "profile" {
  # id = <uid>
  name = var.cluster_cluster_profile_name
}


resource "spectrocloud_cluster_gcp" "cluster" {
  name               = var.cluster_name
  cloud_account_id   = data.spectrocloud_cloudaccount_gcp.account.id

  cloud_config {
    network = var.gcp_network
    project = var.gcp_project
    region  = var.gcp_region
  }

  cluster_profile {
    cluster_profile_id = data.spectrocloud_cluster_profile.profile.id

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
  }

  machine_pool {
    control_plane           = true
    control_plane_as_worker = true
    name                    = "master-pool"
    count                   = 1
    instance_type           = "e2-standard-2"
    disk_size_gb            = 62
    azs                     = ["us-west3-a"]
  }

  machine_pool {
    name          = "worker-basic"
    count         = 1
    instance_type = "e2-standard-2"
    azs           = ["us-west3-a"]
  }

}
