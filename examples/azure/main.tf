terraform {
  required_providers {
    spectrocloud = {
      version = ">= 0.1"
      source  = "spectrocloud.com/spectrocloud/spectrocloud"
    }
  }
}

provider "spectrocloud" {
  host        = "console.spectrocloud.com"
  username    = "<email>"
  password    = "<password"
  project_uid = "<project_uid>"
}

resource "spectrocloud_cloudaccount_azure" "azure-1" {
  name                = "azure-1"
  azure_tenant_id     = "<....>"
  azure_client_id     = "<....>"
  azure_client_secret = "<....>"
}

resource "spectrocloud_cluster_azure" "test5" {
  name               = "test5"
  cluster_profile_id = "5fd97f045e6ad657ac0935a2"
  cloud_account_id   = spectrocloud_cloudaccount_azure.azure-2.id

  cloud_config {
    subscription_id = "<....>"
    resource_group  = "saad-west1"
    location        = "westus"
    ssh_key         = "ssh-rsa <PUBKEY>"
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
