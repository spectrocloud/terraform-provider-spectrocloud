data "spectrocloud_cluster_profile" "profile"{
  name = "tf-js-azure-profile"
}

data "spectrocloud_cloudaccount_azure" "account"{
  name = "acc-azure-ca"
}

resource "spectrocloud_cluster_azure" "cluster" {
  name               = "tf-azure-js-1"
  cluster_profile {
    id = data.spectrocloud_cluster_profile.profile.id
  }
  cloud_account_id   = data.spectrocloud_cloudaccount_azure.account.id

  cloud_config {
    subscription_id = var.azure_subscription_id
    resource_group  = var.azure_resource_group
    region          = var.azure_region
    ssh_key         = var.cluster_ssh_public_key
  }

  machine_pool {
    control_plane           = true
    control_plane_as_worker = true
    name                    = "master-pool"
    count                   = 1
    instance_type           = "Standard_D2_v3"
    azs                     = []
    disk {
      size_gb = 65
      type    = "Standard_LRS"
    }
  }

  machine_pool {
    name          = "worker-basic"
    count         = 1
    instance_type = "Standard_D2_v3"
    azs           = []
  }

}
