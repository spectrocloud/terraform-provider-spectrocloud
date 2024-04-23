
resource "spectrocloud_cluster_azure" "cluster" {
  name = "az-picard-2"
  cluster_profile {
    id = spectrocloud_cluster_profile.profile.id
  }
  cloud_account_id = spectrocloud_cloudaccount_azure.account.id

  cloud_config {
    subscription_id = var.azure_subscription_id
    resource_group  = var.azure_resource_group
    region          = var.azure_region
    ssh_key         = var.cluster_ssh_public_key
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
    name                    = "cp-pool"
    count                   = 1
    instance_type           = "Standard_D2_v3"
    azs                     = []
    disk {
      size_gb = 65
      type    = "Standard_LRS"
    }
  }

  machine_pool {
    is_system_node_pool = true
    name                = "worker-basic"
    count               = 1
    instance_type       = "Standard_D2_v3"
    azs                 = []
  }

}
