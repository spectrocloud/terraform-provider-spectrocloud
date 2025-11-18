resource "spectrocloud_cluster_aks" "aks1" {
  name             = "ran-tf-cluster-temp3"
  tags             = ["owner:ranjith"]
  cloud_account_id = "6914c73a23dc08bb241646f2"
  context          = "tenant"

  cloud_config {
    subscription_id = var.subscription_id
    resource_group  = var.resource_group
    ssh_key         = var.ssh_key
    region          = var.region
  }

  # cluster_template = spectrocloud_cluster_config_template.aws_template.id
  cluster_template {
    id = spectrocloud_cluster_config_template.aws_template.id

    cluster_profile {
      id = "691b556e50498bf5109ecf19"
      variables = {
        # image_tag = "6.7.0"
        pullPolicy = "IfNotPresent"
      }
    }
    cluster_profile {
      id = "691b556e50498bf514992e1f"
      variables = {
        pullPolicy      = "IfNotPresent"
        podantiaffinity = "soft"
      }
    }
  }



  machine_pool {
    name                 = "system"
    count                = 1
    instance_type        = "Standard_D4as_v4"
    disk_size_gb         = 50
    is_system_node_pool  = true
    storage_account_type = "Premium_LRS"
  }

}
