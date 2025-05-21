data "spectrocloud_cluster_profile" "profile" {
  name = "tf-js13-azure-profile"
}

data "spectrocloud_cloudaccount_azure" "account" {
  name = "jayesh-azure-ca"
}

resource "spectrocloud_cluster_azure" "cluster" {
  name = "tf-azure-js-1"
  cluster_profile {
    id = data.spectrocloud_cluster_profile.profile.id
  }
  cloud_account_id = data.spectrocloud_cloudaccount_azure.account.id

  cloud_config {
    subscription_id = var.azure_subscription_id
    resource_group  = var.azure_resource_group
    region          = var.azure_region
    ssh_key         = var.cluster_ssh_public_key

    //Static placement config
    #    network_resource_group = "test-resource-group"
    #    virtual_network_name = "test-network-name"
    #    virtual_network_cidr_block = "10.0.0.9/10"
    #    control_plane_subnet {
    #      name="cp_subnet_name"
    #      cidr_block="10.0.0.9/16"
    #      security_group_name="cp_subnet_security_name"
    #    }
    #    worker_node_subnet {
    #      name="worker_subnet_name"
    #      cidr_block="10.0.0.9/16"
    #      security_group_name="worker_subnet_security_name"
    #    }
    #    private_api_server {
    #      resource_group = "test-resource-group"
    #      private_dns_zone = "test-private-dns-zone"
    #      static_ip = "10.11.12.51"
    #    }

  }

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
    count               = 2
    instance_type       = "Standard_D2_v3"
    azs                 = []
  }

}
