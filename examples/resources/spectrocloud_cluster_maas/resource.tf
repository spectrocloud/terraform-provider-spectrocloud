data "spectrocloud_cloudaccount_maas" "account" {
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

resource "spectrocloud_cluster_maas" "cluster" {
  name             = var.cluster_name
  tags             = ["dev", "department:devops", "owner:bob"]
  cloud_account_id = data.spectrocloud_cloudaccount_maas.account.id

  cloud_config {
    domain        = "maas.mycompany.com"
    enable_lxd_vm = false
    ntp_servers   = ["0.pool.ntp.org", "1.pool.ntp.org", "time.google.com"]
    ssh_keys      = var.cluster_ssh_public_keys
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
    name                    = "cp-pool"
    count                   = 1

    placement {
      resource_pool = "Medium-Generic"
    }

    instance_type {
      min_memory_mb = 4096
      min_cpu       = 2
    }

    azs = ["az1"]
  }

  machine_pool {
    name  = "worker-basic"
    count = 1

    placement {
      resource_pool = "Medium-Generic"
    }

    instance_type {
      min_memory_mb = 4096
      min_cpu       = 2
    }

    azs = ["az2"]

    # Optional: override Machine Health Check settings for this node pool
    override_health_check_configuration = <<-EOT
      maxUnhealthy: 40%
      nodeStartupTimeout: 10m
      unhealthyConditions:
        - type: Ready
          status: "False"
          timeout: 5m
        - type: Ready
          status: "Unknown"
          timeout: 5m
    EOT
  }
}
