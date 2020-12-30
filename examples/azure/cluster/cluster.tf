
data "spectrocloud_cloudaccount_azure" "account" {
  # id = <uid>
  name = var.cluster_cloud_account_name
}

data "spectrocloud_cluster_profile" "profile" {
  # id = <uid>
  name = var.cluster_cluster_profile_name
}


# resource "spectrocloud_cloudaccount_azure" "azure-1" {
#   name                = "azure-1"
#   azure_tenant_id     = "<....>"
#   azure_client_id     = "<....>"
#   azure_client_secret = "<....>"
# }

resource "spectrocloud_cluster_azure" "cluster" {
  name               = var.cluster_name
  cluster_profile_id = data.spectrocloud_cluster_profile.profile.id
  cloud_account_id   = data.spectrocloud_cloudaccount_azure.account.id

  cloud_config {
    subscription_id = var.azure_subscription_id
    resource_group  = var.azure_resource_group
    location        = var.azure_location
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
    name                    = "master-pool"
    count                   = 1
    instance_type           = "Standard_D2_v3"
    disk {
      size_gb = 65
      type    = "Standard_LRS"
    }
  }

  machine_pool {
    name          = "worker-basic"
    count         = 1
    instance_type = "Standard_D2_v3"
  }

}


#module "psl" {
#source = "./coffee"
#
#coffee_name = "Packer Spiced Latte"
#}
#
#output "psl" {
#value = module.psl.coffee
#}

#data "hashicups_ingredients" "psl" {
#coffee_id = values(module.psl.coffee)[0].id
#}

# output "psl_i" {
#   value = data.hashicups_ingredients.psl
# }

//resource "spectrocloud_cloudaccounts" "new" {
//  items {
//    coffee {
//      id = 3
//    }
//    quantity = 2
//  }
//  items {
//    coffee {
//      id = 2
//    }
//    quantity = 2
//  }
//}
//

#output "new_order" {
#value = spectrocloud_order.new
#}


#data "hashicups_order" "first" {
#id = 1
#}
#
#output "first_order" {
#value = data.hashicups_order.first
#}
