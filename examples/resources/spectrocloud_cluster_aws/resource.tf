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

# Day-2 mutability: `name` and `cloud_account_id` are ForceNew - changing either recreates the
# cluster. `tags` (list form) is slated for deprecation in favor of `tags_map` (a map); the two
# are ConflictsWith each other - use only one. cluster_profile, backup_policy, scan_policy, and
# machine_pool update in place; see cloud_config's own header below for its mutability.
resource "spectrocloud_cluster_aws" "cluster" {
  name             = var.cluster_name
  tags             = ["dev", "department:devops", "owner:bob"]
  cloud_account_id = data.spectrocloud_cloudaccount_aws.account.id

  # cloud_config:
  #   ssh_key_name, region, vpc_id, control_plane_lb - ForceNew. Changing any of these recreates
  #     the cluster.
  #   override_cluster_api_config - Optional, NOT ForceNew. Raw Cluster API config overrides,
  #     merged into the generated cluster spec.
  cloud_config {
    ssh_key_name                = "spectro2022"
    region                      = "eu-west-1"
    vpc_id                      = "vpc-0a38a86f3bf3c6cf5"
    override_cluster_api_config = <<-EOT
      spec:
        controlPlaneConfiguration:
          apiServer:
            extraArgs:
              authorization-mode: Node,RBAC
    EOT
  }

  cluster_profile {
    id = data.spectrocloud_cluster_profile.profile.id

    variables = {
      "priority"    = "5",
      "default_cmd" = "pwd"
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

  # machine_pool (control plane):
  #   azs - Optional, alternative to az_subnets. Dynamic AZ provisioning by listing availability
  #     zone names (e.g. ["eu-west-1c","eu-west-1a"]); Palette selects subnets for you.
  #   az_subnets - Optional, alternative to azs. Static, per-AZ subnet provisioning: maps each
  #     availability zone name to its subnet ID(s).
  machine_pool {
    additional_labels = {
      "owner"   = "siva"
      "purpose" = "testing"
      "type"    = "cp"
    }
    control_plane           = true
    control_plane_as_worker = true
    name                    = "cp-pool"
    count                   = 1
    instance_type           = "m4.large"
    disk_size_gb            = 60
    # azs = ["eu-west-1c","eu-west-1a"]
    az_subnets = {
      "eu-west-1c" = join(",", var.subnet_ids_eu_west_1c)
      "eu-west-1a" = "subnet-08c7ad2affe1f1250,subnet-04dbeac9aba098d0e"
    }
  }

  # machine_pool (worker pool "worker-basic"):
  #   skip_k8s_upgrade - Optional, default "disabled". Set to "enabled" to skip the OS/Kubernetes
  #     version upgrade for this pool (N-3 skew).
  #   azs - Optional, alternative to az_subnets. Dynamic AZ provisioning by listing availability
  #     zone names.
  #   az_subnets - Optional, alternative to azs. Static, per-AZ subnet provisioning.
  #   override_health_check_configuration - Optional. Raw Machine Health Check overrides for this
  #     node pool.
  machine_pool {
    additional_labels = {
      "owner"   = "siva"
      "purpose" = "testing"
      "type"    = "worker"
    }
    name             = "worker-basic"
    count            = 1
    instance_type    = "m5.large"
    skip_k8s_upgrade = "disabled"
    # azs = ["eu-west-1c","eu-west-1a"]
    az_subnets = {
      "eu-west-1c" = "subnet-039c3beb3da69172f"
      "eu-west-1a" = "subnet-04dbeac9aba098d0e"
    }

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
