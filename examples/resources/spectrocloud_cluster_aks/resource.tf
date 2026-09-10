data "spectrocloud_cloudaccount_azure" "account" {
  # id = <uid>
  name = var.cluster_cloud_account_name
}

data "spectrocloud_cluster_profile" "profile" {
  # id = <uid>
  name = var.cluster_cluster_profile_name
}

data "spectrocloud_backup_storage_location" "bsl" {
  name = var.backup_storage_location_name
}

# Day-2 mutability: `name` and `cloud_account_id` (below) are ForceNew - changing either
# recreates the cluster. So is every attribute inside `cloud_config` (subscription_id,
# resource_group, ssh_key, region, private_cluster, and all vnet/subnet/CIDR fields) - changing
# any of those also destroys and recreates the cluster, since they describe the underlying Azure
# infrastructure the cluster is provisioned onto. `override_cluster_api_config` is the one
# cloud_config field that is NOT ForceNew. Everything else on this resource - cluster_profile,
# backup_policy, scan_policy, machine_pool, tags, description - updates in place.
resource "spectrocloud_cluster_aks" "cluster" {
  name             = var.cluster_name
  tags             = ["dev", "department:devops", "owner:bob"]
  cloud_account_id = data.spectrocloud_cloudaccount_azure.account.id

  cloud_config {
    subscription_id = "subscription-id"
    resource_group  = "dev"
    ssh_key         = "ssh key value"
    region          = "centralus"
    # Optional. Not ForceNew - can be changed without recreating the cluster. Raw Cluster API
    # config overrides, merged into the generated cluster spec.
    override_cluster_api_config = <<-EOT
      spec:
        controlPlaneConfiguration:
          apiServer:
            extraArgs:
              authorization-mode: Node,RBAC
    EOT
    # Other optional, ForceNew cloud_config fields not shown here: private_cluster (bool),
    # vnet_name/vnet_resource_group/vnet_cidr_block (bring-your-own VNet), control_plane_cidr/
    # control_plane_subnet_name/control_plane_subnet_security_group_name, worker_cidr/
    # worker_subnet_name/worker_subnet_security_group_name (bring-your-own subnets).
  }

  cluster_profile {
    id = data.spectrocloud_cluster_profile.profile.id

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

  backup_policy {
    schedule                  = "0 0 * * SUN"
    backup_location_id        = data.spectrocloud_backup_storage_location.bsl.id
    prefix                    = "prod-backup"
    expiry_in_hour            = 7200
    include_disks             = true
    include_cluster_resources = true
  }

  scan_policy {
    configuration_scan_schedule = "0 0 * * SUN"
    penetration_scan_schedule   = "0 0 * * SUN"
    conformance_scan_schedule   = "0 0 1 * *"
  }

  machine_pool {
    name                 = "worker-basic"
    count                = 1
    instance_type        = "Standard_DS4"
    disk_size_gb         = 60
    is_system_node_pool  = true
    storage_account_type = "Standard_LRS"
    # os_sku = "Ubuntu" # Optional. Allowed: "Ubuntu", "AzureLinux", "Windows2022". The
    #                     description marks this immutable after creation, though it is not a
    #                     ForceNew schema field.
    # Other optional machine_pool fields not shown: additional_labels/additional_annotations
    # (key:value maps), taints, node (per-node overrides), min/max (autoscaling bounds),
    # update_strategy, os_type.
  }
}
