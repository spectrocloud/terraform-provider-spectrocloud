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
    ssh_key_name = "spectro22"
    region       = "us-east-2"
    vpc_id       = "vpc-0d2a3f1799e4d6065"
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
    instance_type           = "t3.large"
    disk_size_gb            = 62
    azs                     = ["us-east-2a"]
    az_subnet {
      id = "subnet-08864975df862eb58"
      az = "us-east-2a"
    }
    az_subnet {
      id = "subnet-031a7ff4ff5e7fb9a"
      az = "us-east-2a"
    }
  }

  machine_pool {
    name          = "worker-basic"
    count         = 1
    instance_type = "t3.large"
    azs           = ["us-east-2a"]
    az_subnet {
      id = "subnet-08864975df862eb58"
      az = "us-east-2a"
    }
  }

}
