resource "spectrocloud_cluster_gcp" "cluster" {
  name               = "gcp-picard-2"
  cluster_profile_id = spectrocloud_cluster_profile.profile.id
  cloud_account_id   = spectrocloud_cloudaccount_gcp.account.id
  os_patch_on_boot = true
  os_patch_schedule = "0 0 * * 0"
  #os_patch_after = "2021-02-03T14:59:37.000Z"

  cloud_config {
    network = var.gcp_network
    project = var.gcp_project
    region  = var.gcp_region
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
    instance_type           = "e2-standard-2"
    disk_size_gb            = 62
    azs                     = ["${var.gcp_region}-a"]
  }

  machine_pool {
    name          = "worker-basic"
    count         = 1
    instance_type = "e2-standard-2"
    azs           = ["${var.gcp_region}-a"]
  }

}
