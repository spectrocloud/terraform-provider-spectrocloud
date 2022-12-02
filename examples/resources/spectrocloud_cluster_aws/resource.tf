data "spectrocloud_cloudaccount_aws" "account" {
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

resource "spectrocloud_cluster_aws" "cluster" {
  name             = var.cluster_name
  tags             = ["dev", "department:devops", "owner:bob"]
  cloud_account_id = data.spectrocloud_cloudaccount_aws.account.id

  cloud_config {
    ssh_key_name = "spectro2022"
    region       = "eu-west-1"
    vpc_id       = "vpc-0a38a86f3bf3c6cf5"
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
    control_plane           = true
    control_plane_as_worker = true
    name                    = "master-pool"
    count                   = 1
    instance_type           = "m4.large"
    disk_size_gb            = 60
    #    Add azs for dynamic provisioning
    #    azs                     = ["us-east-2a"]
    #     Add az_subnet component for static provisioning
    az_subnets = {
      "eu-west-1c" = join(",", var.subnet_ids_eu_west_1c)
      "eu-west-1a" = "subnet-08c7ad2affe1f1250,subnet-04dbeac9aba098d0e"
    }
  }

  machine_pool {
    name          = "worker-basic"
    count         = 2
    instance_type = "m5.large"
    #    Add azs for dynamic provisioning
    #    azs           = ["us-east-2a"]
    #    Add az_subnet component for static provisioning
    az_subnets = {
      "eu-west-1c" = "subnet-039c3beb3da69172f"
      "eu-west-1a" = "subnet-04dbeac9aba098d0e"
    }
  }

}
